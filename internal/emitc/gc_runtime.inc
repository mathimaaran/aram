/* Aram Tamil-0.52: stop-the-world conservative mark-sweep heap.
 * Linux + pthread. Safe points: aram_gc_poll and aram_gc_cond_wait.
 */
#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif
#include <pthread.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#ifdef __linux__
#include <unistd.h>
#endif

typedef struct aram_obj {
	size_t size;
	int mark;
	struct aram_obj *next;
} aram_obj;

#define ARAM_PAYLOAD(o) ((unsigned char *)((o) + 1))

typedef struct aram_th {
	pthread_t id;
	void *stack_lo;
	void *stack_hi;
	int parked;
	struct aram_th *next;
} aram_th;

static pthread_mutex_t aram_heap_mu = PTHREAD_MUTEX_INITIALIZER;
static pthread_mutex_t aram_gc_mu = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t aram_gc_cv = PTHREAD_COND_INITIALIZER;
static aram_obj *aram_objs;
static size_t aram_heap_used;
static size_t aram_gc_threshold = 256 * 1024;
static int aram_gc_request;
static int aram_thread_n;
static int aram_threads_parked;
static aram_th *aram_threads;
static _Thread_local aram_th *aram_self;

static void aram_thread_stack_hi(void **hi) {
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

void aram_thread_register(void) {
	aram_th *t = (aram_th *)calloc(1, sizeof(*t));
	if (!t) abort();
	t->id = pthread_self();
	char probe;
	t->stack_lo = &probe;
	aram_thread_stack_hi(&t->stack_hi);
	pthread_mutex_lock(&aram_gc_mu);
	while (aram_gc_request) pthread_cond_wait(&aram_gc_cv, &aram_gc_mu);
	t->next = aram_threads;
	aram_threads = t;
	aram_thread_n++;
	aram_self = t;
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_thread_unregister(void) {
	pthread_mutex_lock(&aram_gc_mu);
	aram_th **pp = &aram_threads;
	while (*pp) {
		if (*pp == aram_self) {
			aram_th *x = *pp;
			*pp = x->next;
			aram_thread_n--;
			if (x->parked && aram_threads_parked > 0) aram_threads_parked--;
			free(x);
			break;
		}
		pp = &(*pp)->next;
	}
	aram_self = NULL;
	pthread_cond_broadcast(&aram_gc_cv);
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_gc_park_begin(void *sp) {
	pthread_mutex_lock(&aram_gc_mu);
	if (aram_self) {
		aram_self->stack_lo = sp;
		if (!aram_self->parked) {
			aram_self->parked = 1;
			aram_threads_parked++;
		}
	}
	pthread_cond_broadcast(&aram_gc_cv);
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_gc_park_end(void) {
	pthread_mutex_lock(&aram_gc_mu);
	while (aram_gc_request) pthread_cond_wait(&aram_gc_cv, &aram_gc_mu);
	if (aram_self && aram_self->parked) {
		aram_self->parked = 0;
		if (aram_threads_parked > 0) aram_threads_parked--;
	}
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_gc_poll(void) {
	if (!aram_gc_request) return;
	char sp;
	pthread_mutex_lock(&aram_gc_mu);
	if (!aram_gc_request) {
		pthread_mutex_unlock(&aram_gc_mu);
		return;
	}
	if (aram_self) {
		aram_self->stack_lo = &sp;
		if (!aram_self->parked) {
			aram_self->parked = 1;
			aram_threads_parked++;
		}
	}
	pthread_cond_broadcast(&aram_gc_cv);
	while (aram_gc_request) pthread_cond_wait(&aram_gc_cv, &aram_gc_mu);
	if (aram_self && aram_self->parked) {
		aram_self->parked = 0;
		if (aram_threads_parked > 0) aram_threads_parked--;
	}
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_gc_cond_wait(pthread_cond_t *cv, pthread_mutex_t *mu) {
	char sp;
	aram_gc_park_begin(&sp);
	pthread_cond_wait(cv, mu);
	aram_gc_park_end();
}

#define ARAM_GC_COND_WAIT(cv, mu) aram_gc_cond_wait((cv), (mu))

static aram_obj *aram_obj_of(void *p) {
	for (aram_obj *o = aram_objs; o; o = o->next) {
		void *b = ARAM_PAYLOAD(o);
		void *e = (char *)b + o->size;
		if (p >= b && p < e) return o;
	}
	return NULL;
}

static void aram_mark_ptr(void *p) {
	aram_obj *o = aram_obj_of(p);
	if (!o || o->mark) return;
	o->mark = 1;
	uintptr_t *w = (uintptr_t *)ARAM_PAYLOAD(o);
	size_t n = o->size / sizeof(uintptr_t);
	for (size_t i = 0; i < n; i++) aram_mark_ptr((void *)w[i]);
}

static void aram_mark_range(void *lo, void *hi) {
	if (!lo || !hi) return;
	if (lo > hi) {
		void *t = lo;
		lo = hi;
		hi = t;
	}
	uintptr_t *p = (uintptr_t *)((uintptr_t)lo & ~(sizeof(uintptr_t) - 1));
	uintptr_t *e = (uintptr_t *)hi;
	for (; p < e; p++) aram_mark_ptr((void *)*p);
}

static void aram_gc_mark_sweep(void) {
	for (aram_obj *o = aram_objs; o; o = o->next) o->mark = 0;
	for (aram_th *t = aram_threads; t; t = t->next) aram_mark_range(t->stack_lo, t->stack_hi);
	aram_obj **pp = &aram_objs;
	size_t live = 0;
	while (*pp) {
		aram_obj *o = *pp;
		if (o->mark) {
			live += o->size;
			pp = &o->next;
		} else {
			*pp = o->next;
			free(o);
		}
	}
	aram_heap_used = live;
	aram_gc_threshold = live * 2;
	if (aram_gc_threshold < 256 * 1024) aram_gc_threshold = 256 * 1024;
}

static void aram_gc_join_or_run(void) {
	char sp;
	pthread_mutex_lock(&aram_gc_mu);
	if (aram_gc_request) {
		if (aram_self) {
			aram_self->stack_lo = &sp;
			if (!aram_self->parked) {
				aram_self->parked = 1;
				aram_threads_parked++;
			}
		}
		pthread_cond_broadcast(&aram_gc_cv);
		while (aram_gc_request) pthread_cond_wait(&aram_gc_cv, &aram_gc_mu);
		if (aram_self && aram_self->parked) {
			aram_self->parked = 0;
			if (aram_threads_parked > 0) aram_threads_parked--;
		}
		pthread_mutex_unlock(&aram_gc_mu);
		return;
	}
	aram_gc_request = 1;
	if (aram_self) {
		aram_self->stack_lo = &sp;
		if (!aram_self->parked) {
			aram_self->parked = 1;
			aram_threads_parked++;
		}
	}
	pthread_cond_broadcast(&aram_gc_cv);
	while (aram_thread_n > 0 && aram_threads_parked < aram_thread_n)
		pthread_cond_wait(&aram_gc_cv, &aram_gc_mu);
	pthread_mutex_unlock(&aram_gc_mu);

	pthread_mutex_lock(&aram_heap_mu);
	aram_gc_mark_sweep();
	pthread_mutex_unlock(&aram_heap_mu);

	pthread_mutex_lock(&aram_gc_mu);
	aram_gc_request = 0;
	if (aram_self && aram_self->parked) {
		aram_self->parked = 0;
		if (aram_threads_parked > 0) aram_threads_parked--;
	}
	pthread_cond_broadcast(&aram_gc_cv);
	pthread_mutex_unlock(&aram_gc_mu);
}

void aram_gc(void) { aram_gc_join_or_run(); }

void *aram_alloc(size_t n) {
	if (n == 0) n = 1;
	for (;;) {
		aram_gc_poll();
		pthread_mutex_lock(&aram_heap_mu);
		pthread_mutex_lock(&aram_gc_mu);
		int req = aram_gc_request;
		pthread_mutex_unlock(&aram_gc_mu);
		if (req) {
			pthread_mutex_unlock(&aram_heap_mu);
			continue;
		}
		if (aram_heap_used + n > aram_gc_threshold) {
			pthread_mutex_unlock(&aram_heap_mu);
			aram_gc_join_or_run();
			continue;
		}
		aram_obj *o = (aram_obj *)calloc(1, sizeof(aram_obj) + n);
		if (!o) abort();
		o->size = n;
		o->next = aram_objs;
		aram_objs = o;
		aram_heap_used += n;
		pthread_mutex_unlock(&aram_heap_mu);
		return ARAM_PAYLOAD(o);
	}
}

size_t aram_heap_live(void) { return aram_heap_used; }

void aram_heap_init(void) { aram_thread_register(); }

void aram_heap_shutdown(void) {
	aram_obj *o = aram_objs;
	while (o) {
		aram_obj *n = o->next;
		free(o);
		o = n;
	}
	aram_objs = NULL;
	aram_heap_used = 0;
	aram_thread_unregister();
}

static void *aram_arena_alloc(size_t n) { return aram_alloc(n); }
static void aram_arena_free(void) { aram_heap_shutdown(); }
