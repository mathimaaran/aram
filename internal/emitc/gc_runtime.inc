/* Niraluli Tamil-0.60: stop-the-world conservative mark-sweep heap.
 * Linux + pthread. Safe points: uli_gc_poll and uli_gc_cond_wait.
 */
#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif
#include <pthread.h>
#include <setjmp.h>
#include <stdint.h>
#include <stdatomic.h>
#include <stdlib.h>
#include <string.h>
#ifdef __linux__
#include <unistd.h>
#endif

typedef struct uli_obj {
	size_t size;
	int mark;
	struct uli_obj *next;
} uli_obj;

#define ULI_PAYLOAD(o) ((unsigned char *)((o) + 1))

typedef struct uli_th {
	pthread_t id;
	void *stack_lo;
	void *stack_hi;
	int parked;
	jmp_buf regs;
	struct uli_th *next;
} uli_th;

static pthread_mutex_t uli_heap_mu = PTHREAD_MUTEX_INITIALIZER;
static pthread_mutex_t uli_gc_mu = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t uli_gc_cv = PTHREAD_COND_INITIALIZER;
static uli_obj *uli_objs;
static size_t uli_heap_used;
static size_t uli_gc_threshold = 256 * 1024;
static _Atomic int uli_gc_request;
static uli_obj **uli_gc_index;
static size_t uli_gc_index_len;
static size_t uli_gc_index_cap;
static int uli_thread_n;
static int uli_threads_parked;
static uli_th *uli_threads;
static _Thread_local uli_th *uli_self;

void uli_gc_poll(void);

static void uli_thread_stack_hi(void **hi) {
	char probe;
	*hi = &probe;
#ifdef __linux__
	pthread_attr_t attr;
	if (pthread_getattr_np(pthread_self(), &attr) == 0) {
		void *addr = NULL;
		size_t sz = 0;
		if (pthread_attr_getstack(&attr, &addr, &sz) == 0 && addr && sz)
			*hi = (char *)addr + sz;
		pthread_attr_destroy(&attr);
	}
#endif
}

void uli_thread_register(void) {
	uli_th *t = (uli_th *)calloc(1, sizeof(*t));
	if (!t) abort();
	t->id = pthread_self();
	char probe;
	t->stack_lo = &probe;
	uli_thread_stack_hi(&t->stack_hi);
	pthread_mutex_lock(&uli_gc_mu);
	while (uli_gc_request) pthread_cond_wait(&uli_gc_cv, &uli_gc_mu);
	t->next = uli_threads;
	uli_threads = t;
	uli_thread_n++;
	uli_self = t;
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_thread_unregister(void) {
	uli_gc_poll();
	pthread_mutex_lock(&uli_gc_mu);
	uli_th **pp = &uli_threads;
	while (*pp) {
		if (*pp == uli_self) {
			uli_th *x = *pp;
			*pp = x->next;
			uli_thread_n--;
			if (x->parked && uli_threads_parked > 0) uli_threads_parked--;
			free(x);
			break;
		}
		pp = &(*pp)->next;
	}
	uli_self = NULL;
	pthread_cond_broadcast(&uli_gc_cv);
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_gc_park_begin(void *sp) {
	pthread_mutex_lock(&uli_gc_mu);
	if (uli_self) {
		(void)setjmp(uli_self->regs);
		uli_self->stack_lo = sp;
		if (!uli_self->parked) {
			uli_self->parked = 1;
			uli_threads_parked++;
		}
	}
	pthread_cond_broadcast(&uli_gc_cv);
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_gc_park_end(void) {
	pthread_mutex_lock(&uli_gc_mu);
	while (uli_gc_request) pthread_cond_wait(&uli_gc_cv, &uli_gc_mu);
	if (uli_self && uli_self->parked) {
		uli_self->parked = 0;
		if (uli_threads_parked > 0) uli_threads_parked--;
	}
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_gc_poll(void) {
	if (!uli_gc_request) return;
	char sp;
	pthread_mutex_lock(&uli_gc_mu);
	if (!uli_gc_request) {
		pthread_mutex_unlock(&uli_gc_mu);
		return;
	}
	if (uli_self) {
		(void)setjmp(uli_self->regs);
		uli_self->stack_lo = &sp;
		if (!uli_self->parked) {
			uli_self->parked = 1;
			uli_threads_parked++;
		}
	}
	pthread_cond_broadcast(&uli_gc_cv);
	while (uli_gc_request) pthread_cond_wait(&uli_gc_cv, &uli_gc_mu);
	if (uli_self && uli_self->parked) {
		uli_self->parked = 0;
		if (uli_threads_parked > 0) uli_threads_parked--;
	}
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_gc_cond_wait(pthread_cond_t *cv, pthread_mutex_t *mu) {
	char sp;
	uli_gc_park_begin(&sp);
	pthread_cond_wait(cv, mu);
	uli_gc_park_end();
}

#define ULI_GC_COND_WAIT(cv, mu) uli_gc_cond_wait((cv), (mu))

static int uli_obj_addr_cmp(const void *a, const void *b) {
	uintptr_t pa = (uintptr_t)ULI_PAYLOAD(*(uli_obj *const *)a);
	uintptr_t pb = (uintptr_t)ULI_PAYLOAD(*(uli_obj *const *)b);
	return pa < pb ? -1 : pa > pb;
}

static void uli_gc_build_index(void) {
	size_t n = 0;
	for (uli_obj *o = uli_objs; o; o = o->next) n++;
	if (n > uli_gc_index_cap) {
		size_t cap = uli_gc_index_cap ? uli_gc_index_cap : 256;
		while (cap < n) cap *= 2;
		uli_obj **grown = (uli_obj **)realloc(uli_gc_index, cap * sizeof(*grown));
		if (!grown) abort();
		uli_gc_index = grown;
		uli_gc_index_cap = cap;
	}
	uli_gc_index_len = n;
	size_t i = 0;
	for (uli_obj *o = uli_objs; o; o = o->next) uli_gc_index[i++] = o;
	if (n > 1) qsort(uli_gc_index, n, sizeof(*uli_gc_index), uli_obj_addr_cmp);
}

static uli_obj *uli_obj_of(void *p) {
	uintptr_t target = (uintptr_t)p;
	size_t lo = 0, hi = uli_gc_index_len;
	while (lo < hi) {
		size_t mid = lo + (hi - lo) / 2;
		uintptr_t base = (uintptr_t)ULI_PAYLOAD(uli_gc_index[mid]);
		if (base <= target) lo = mid + 1;
		else hi = mid;
	}
	if (lo == 0) return NULL;
	uli_obj *o = uli_gc_index[lo - 1];
	uintptr_t base = (uintptr_t)ULI_PAYLOAD(o);
	return target - base < o->size ? o : NULL;
}

static void uli_mark_ptr(void *p) {
	uli_obj *o = uli_obj_of(p);
	if (!o || o->mark) return;
	o->mark = 1;
	uintptr_t *w = (uintptr_t *)ULI_PAYLOAD(o);
	size_t n = o->size / sizeof(uintptr_t);
	for (size_t i = 0; i < n; i++) uli_mark_ptr((void *)w[i]);
}

static void uli_mark_range(void *lo, void *hi) {
	if (!lo || !hi) return;
	if (lo > hi) {
		void *t = lo;
		lo = hi;
		hi = t;
	}
	uintptr_t *p = (uintptr_t *)((uintptr_t)lo & ~(sizeof(uintptr_t) - 1));
	uintptr_t *e = (uintptr_t *)hi;
	for (; p < e; p++) uli_mark_ptr((void *)*p);
}

static void uli_gc_mark_sweep(void) {
	for (uli_obj *o = uli_objs; o; o = o->next) o->mark = 0;
	uli_gc_build_index();
	for (uli_th *t = uli_threads; t; t = t->next) {
		uli_mark_range(t->stack_lo, t->stack_hi);
		uli_mark_range(&t->regs, (char *)&t->regs + sizeof(t->regs));
	}
	uli_obj **pp = &uli_objs;
	size_t live = 0;
	while (*pp) {
		uli_obj *o = *pp;
		if (o->mark) {
			live += o->size;
			pp = &o->next;
		} else {
			*pp = o->next;
			free(o);
		}
	}
	uli_heap_used = live;
	uli_gc_threshold = live <= SIZE_MAX / 2 ? live * 2 : SIZE_MAX;
	if (uli_gc_threshold < 256 * 1024) uli_gc_threshold = 256 * 1024;
}

static void uli_gc_join_or_run(void) {
	char sp;
	pthread_mutex_lock(&uli_gc_mu);
	if (uli_gc_request) {
		if (uli_self) {
			(void)setjmp(uli_self->regs);
			uli_self->stack_lo = &sp;
			if (!uli_self->parked) {
				uli_self->parked = 1;
				uli_threads_parked++;
			}
		}
		pthread_cond_broadcast(&uli_gc_cv);
		while (uli_gc_request) pthread_cond_wait(&uli_gc_cv, &uli_gc_mu);
		if (uli_self && uli_self->parked) {
			uli_self->parked = 0;
			if (uli_threads_parked > 0) uli_threads_parked--;
		}
		pthread_mutex_unlock(&uli_gc_mu);
		return;
	}
	uli_gc_request = 1;
	if (uli_self) {
		(void)setjmp(uli_self->regs);
		uli_self->stack_lo = &sp;
		if (!uli_self->parked) {
			uli_self->parked = 1;
			uli_threads_parked++;
		}
	}
	pthread_cond_broadcast(&uli_gc_cv);
	while (uli_thread_n > 0 && uli_threads_parked < uli_thread_n)
		pthread_cond_wait(&uli_gc_cv, &uli_gc_mu);
	pthread_mutex_unlock(&uli_gc_mu);

	pthread_mutex_lock(&uli_heap_mu);
	uli_gc_mark_sweep();
	pthread_mutex_unlock(&uli_heap_mu);

	pthread_mutex_lock(&uli_gc_mu);
	uli_gc_request = 0;
	if (uli_self && uli_self->parked) {
		uli_self->parked = 0;
		if (uli_threads_parked > 0) uli_threads_parked--;
	}
	pthread_cond_broadcast(&uli_gc_cv);
	pthread_mutex_unlock(&uli_gc_mu);
}

void uli_gc(void) { uli_gc_join_or_run(); }

void *uli_alloc(size_t n) {
	if (n == 0) n = 1;
	if (n > SIZE_MAX - sizeof(uli_obj)) abort();
	int collected = 0;
	for (;;) {
		uli_gc_poll();
		pthread_mutex_lock(&uli_heap_mu);
		if (uli_gc_request) {
			pthread_mutex_unlock(&uli_heap_mu);
			continue;
		}
		if (n > SIZE_MAX - uli_heap_used) {
			pthread_mutex_unlock(&uli_heap_mu);
			abort();
		}
		if (uli_heap_used + n > uli_gc_threshold && !collected) {
			pthread_mutex_unlock(&uli_heap_mu);
			uli_gc_join_or_run();
			collected = 1;
			continue;
		}
		if (uli_heap_used + n > uli_gc_threshold) {
			size_t need = uli_heap_used + n;
			uli_gc_threshold = need <= SIZE_MAX / 2 ? need * 2 : SIZE_MAX;
		}
		uli_obj *o = (uli_obj *)calloc(1, sizeof(uli_obj) + n);
		if (!o) abort();
		o->size = n;
		o->next = uli_objs;
		uli_objs = o;
		uli_heap_used += n;
		pthread_mutex_unlock(&uli_heap_mu);
		return ULI_PAYLOAD(o);
	}
}

size_t uli_heap_live(void) { return uli_heap_used; }

void uli_heap_init(void) { uli_thread_register(); }

void uli_heap_shutdown(void) {
	uli_obj *o = uli_objs;
	while (o) {
		uli_obj *n = o->next;
		free(o);
		o = n;
	}
	uli_objs = NULL;
	uli_heap_used = 0;
	free(uli_gc_index);
	uli_gc_index = NULL;
	uli_gc_index_len = 0;
	uli_gc_index_cap = 0;
	uli_thread_unregister();
}

static void *uli_arena_alloc(size_t n) { return uli_alloc(n); }
static void uli_arena_free(void) { uli_heap_shutdown(); }
