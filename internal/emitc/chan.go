package emitc

import (
	"fmt"
	"strings"

	"niraluli/internal/ast"
	"niraluli/internal/check"
	"niraluli/internal/token"
)

func (e *emitter) writeChanRuntime(b *strings.Builder) {
	if !e.needChan && !e.needGo {
		return
	}
	b.WriteString("#include <pthread.h>\n\n")
	if e.needGo {
		b.WriteString("static pthread_mutex_t uli_go_mu = PTHREAD_MUTEX_INITIALIZER;\n")
		b.WriteString("static pthread_t *uli_go_threads;\n")
		b.WriteString("static size_t uli_go_n, uli_go_cap;\n")
		b.WriteString("static void uli_go(void *(*fn)(void *), void *arg) {\n")
		b.WriteString("\tpthread_t th;\n")
		b.WriteString("\tif (pthread_create(&th, NULL, fn, arg) != 0) abort();\n")
		b.WriteString("\tpthread_mutex_lock(&uli_go_mu);\n")
		b.WriteString("\tif (uli_go_n + 1 > uli_go_cap) {\n")
		b.WriteString("\t\tsize_t nc = uli_go_cap ? uli_go_cap * 2 : 8;\n")
		b.WriteString("\t\tuli_go_threads = (pthread_t *)realloc(uli_go_threads, nc * sizeof(pthread_t));\n")
		b.WriteString("\t\tif (!uli_go_threads) abort();\n")
		b.WriteString("\t\tuli_go_cap = nc;\n")
		b.WriteString("\t}\n")
		b.WriteString("\tuli_go_threads[uli_go_n++] = th;\n")
		b.WriteString("\tpthread_mutex_unlock(&uli_go_mu);\n")
		b.WriteString("}\n")
		b.WriteString("static void uli_go_wait_all(void) {\n")
		b.WriteString("\tpthread_mutex_lock(&uli_go_mu);\n")
		b.WriteString("\tsize_t n = uli_go_n;\n")
		b.WriteString("\tpthread_t *ts = uli_go_threads;\n")
		b.WriteString("\tuli_go_n = 0;\n")
		b.WriteString("\tpthread_mutex_unlock(&uli_go_mu);\n")
		b.WriteString("\tfor (size_t i = 0; i < n; i++) pthread_join(ts[i], NULL);\n")
		b.WriteString("\tfree(ts);\n")
		b.WriteString("\tuli_go_threads = NULL; uli_go_cap = 0;\n")
		b.WriteString("}\n\n")
	}
	if !e.needChan {
		return
	}
	// Wakes blocked தடத்தேர்வு (no CPU spin / nanosleep).
	b.WriteString("static pthread_mutex_t uli_sel_mu = PTHREAD_MUTEX_INITIALIZER;\n")
	b.WriteString("static pthread_cond_t uli_sel_cv = PTHREAD_COND_INITIALIZER;\n")
	b.WriteString("static unsigned uli_sel_gen;\n")
	b.WriteString("static void uli_sel_kick(void) {\n")
	b.WriteString("\tpthread_mutex_lock(&uli_sel_mu);\n")
	b.WriteString("\tuli_sel_gen++;\n")
	b.WriteString("\tpthread_cond_broadcast(&uli_sel_cv);\n")
	b.WriteString("\tpthread_mutex_unlock(&uli_sel_mu);\n")
	b.WriteString("}\n\n")
	b.WriteString("typedef struct uli_chan {\n")
	b.WriteString("\tpthread_mutex_t mu;\n")
	b.WriteString("\tpthread_cond_t send_cv;\n")
	b.WriteString("\tpthread_cond_t recv_cv;\n")
	b.WriteString("\tint elem_size;\n")
	b.WriteString("\tint cap;\n")
	b.WriteString("\tint len;\n")
	b.WriteString("\tint closed;\n")
	b.WriteString("\tint recv_waiters;\n")
	b.WriteString("\tint head;\n")
	b.WriteString("\tint tail;\n")
	b.WriteString("\tchar *buf;\n")
	b.WriteString("\tint has_slot;\n")
	b.WriteString("\tchar *slot;\n")
	b.WriteString("} uli_chan;\n\n")
	b.WriteString("static uli_chan *uli_chan_make(int elem_size, int64_t cap) {\n")
	b.WriteString("\tif (elem_size <= 0 || cap < 0) abort();\n")
	b.WriteString("\tuli_chan *c = (uli_chan *)uli_alloc(sizeof(uli_chan));\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_init(&c->mu, NULL);\n")
	b.WriteString("\tpthread_cond_init(&c->send_cv, NULL);\n")
	b.WriteString("\tpthread_cond_init(&c->recv_cv, NULL);\n")
	b.WriteString("\tc->elem_size = elem_size;\n")
	b.WriteString("\tc->cap = (int)cap;\n")
	b.WriteString("\tif (c->cap > 0) {\n")
	b.WriteString("\t\tc->buf = (char *)uli_alloc((size_t)c->cap * (size_t)elem_size);\n")
	b.WriteString("\t\tif (!c->buf) abort();\n")
	b.WriteString("\t} else {\n")
	b.WriteString("\t\tc->slot = (char *)uli_alloc((size_t)elem_size);\n")
	b.WriteString("\t\tif (!c->slot) abort();\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn c;\n")
	b.WriteString("}\n\n")
	b.WriteString("static void uli_chan_close(uli_chan *c) {\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\tif (c->closed) { pthread_mutex_unlock(&c->mu); abort(); }\n")
	b.WriteString("\tc->closed = 1;\n")
	b.WriteString("\tpthread_cond_broadcast(&c->send_cv);\n")
	b.WriteString("\tpthread_cond_broadcast(&c->recv_cv);\n")
	b.WriteString("\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\tuli_sel_kick();\n")
	b.WriteString("}\n\n")
	b.WriteString("static int uli_chan_try_send(uli_chan *c, const void *data) {\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\tif (c->closed) { pthread_mutex_unlock(&c->mu); abort(); }\n")
	b.WriteString("\tif (c->cap == 0) {\n")
	b.WriteString("\t\t/* Unbuffered try-send only if a receiver is waiting (Go select). */\n")
	b.WriteString("\t\tif (c->has_slot || c->recv_waiters <= 0) { pthread_mutex_unlock(&c->mu); return 0; }\n")
	b.WriteString("\t\tmemcpy(c->slot, data, (size_t)c->elem_size);\n")
	b.WriteString("\t\tc->has_slot = 1;\n")
	b.WriteString("\t\tpthread_cond_signal(&c->recv_cv);\n")
	b.WriteString("\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\tuli_sel_kick();\n")
	b.WriteString("\t\treturn 1;\n")
	b.WriteString("\t}\n")
	b.WriteString("\tif (c->len == c->cap) { pthread_mutex_unlock(&c->mu); return 0; }\n")
	b.WriteString("\tmemcpy(c->buf + c->tail * c->elem_size, data, (size_t)c->elem_size);\n")
	b.WriteString("\tc->tail = (c->tail + 1) % c->cap;\n")
	b.WriteString("\tc->len++;\n")
	b.WriteString("\tpthread_cond_signal(&c->recv_cv);\n")
	b.WriteString("\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\tuli_sel_kick();\n")
	b.WriteString("\treturn 1;\n")
	b.WriteString("}\n\n")
	b.WriteString("static void uli_chan_send(uli_chan *c, const void *data) {\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\tfor (;;) {\n")
	b.WriteString("\t\tif (c->closed) { pthread_mutex_unlock(&c->mu); abort(); }\n")
	b.WriteString("\t\tif (c->cap == 0) {\n")
	b.WriteString("\t\t\tif (!c->has_slot) {\n")
	b.WriteString("\t\t\t\tmemcpy(c->slot, data, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\t\tc->has_slot = 1;\n")
	b.WriteString("\t\t\t\tpthread_cond_signal(&c->recv_cv);\n")
	b.WriteString("\t\t\t\tuli_sel_kick();\n")
	b.WriteString("\t\t\t\twhile (c->has_slot && !c->closed) ULI_GC_COND_WAIT(&c->send_cv, &c->mu);\n")
	b.WriteString("\t\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\t\treturn;\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t} else if (c->len < c->cap) {\n")
	b.WriteString("\t\t\tmemcpy(c->buf + c->tail * c->elem_size, data, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\tc->tail = (c->tail + 1) % c->cap;\n")
	b.WriteString("\t\t\tc->len++;\n")
	b.WriteString("\t\t\tpthread_cond_signal(&c->recv_cv);\n")
	b.WriteString("\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\tuli_sel_kick();\n")
	b.WriteString("\t\t\treturn;\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tULI_GC_COND_WAIT(&c->send_cv, &c->mu);\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString("static int uli_chan_try_recv(uli_chan *c, void *out, int *ok) {\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\tif (c->cap == 0) {\n")
	b.WriteString("\t\tif (c->has_slot) {\n")
	b.WriteString("\t\t\tmemcpy(out, c->slot, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\tc->has_slot = 0;\n")
	b.WriteString("\t\t\tpthread_cond_signal(&c->send_cv);\n")
	b.WriteString("\t\t\tif (ok) *ok = 1;\n")
	b.WriteString("\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\tuli_sel_kick();\n")
	b.WriteString("\t\t\treturn 1;\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tif (c->closed) {\n")
	b.WriteString("\t\t\tmemset(out, 0, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\tif (ok) *ok = 0;\n")
	b.WriteString("\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\treturn 1;\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\treturn 0;\n")
	b.WriteString("\t}\n")
	b.WriteString("\tif (c->len > 0) {\n")
	b.WriteString("\t\tmemcpy(out, c->buf + c->head * c->elem_size, (size_t)c->elem_size);\n")
	b.WriteString("\t\tc->head = (c->head + 1) % c->cap;\n")
	b.WriteString("\t\tc->len--;\n")
	b.WriteString("\t\tpthread_cond_signal(&c->send_cv);\n")
	b.WriteString("\t\tif (ok) *ok = 1;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\tuli_sel_kick();\n")
	b.WriteString("\t\treturn 1;\n")
	b.WriteString("\t}\n")
	b.WriteString("\tif (c->closed) {\n")
	b.WriteString("\t\tmemset(out, 0, (size_t)c->elem_size);\n")
	b.WriteString("\t\tif (ok) *ok = 0;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\treturn 1;\n")
	b.WriteString("\t}\n")
	b.WriteString("\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\treturn 0;\n")
	b.WriteString("}\n\n")
	b.WriteString("static int uli_chan_recv(uli_chan *c, void *out) {\n")
	b.WriteString("\tint ok = 0;\n")
	b.WriteString("\tif (!c) abort();\n")
	b.WriteString("\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\tfor (;;) {\n")
	b.WriteString("\t\tif (c->cap == 0) {\n")
	b.WriteString("\t\t\tif (c->has_slot) {\n")
	b.WriteString("\t\t\t\tmemcpy(out, c->slot, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\t\tc->has_slot = 0;\n")
	b.WriteString("\t\t\t\tpthread_cond_signal(&c->send_cv);\n")
	b.WriteString("\t\t\t\tok = 1;\n")
	b.WriteString("\t\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\t\tuli_sel_kick();\n")
	b.WriteString("\t\t\t\treturn ok;\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t\tif (c->closed) {\n")
	b.WriteString("\t\t\t\tmemset(out, 0, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\t\treturn 0;\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t} else if (c->len > 0) {\n")
	b.WriteString("\t\t\tmemcpy(out, c->buf + c->head * c->elem_size, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\tc->head = (c->head + 1) % c->cap;\n")
	b.WriteString("\t\t\tc->len--;\n")
	b.WriteString("\t\t\tpthread_cond_signal(&c->send_cv);\n")
	b.WriteString("\t\t\tok = 1;\n")
	b.WriteString("\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\tuli_sel_kick();\n")
	b.WriteString("\t\t\treturn ok;\n")
	b.WriteString("\t\t} else if (c->closed) {\n")
	b.WriteString("\t\t\tmemset(out, 0, (size_t)c->elem_size);\n")
	b.WriteString("\t\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\t\treturn 0;\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc->recv_waiters++;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&c->mu);\n")
	b.WriteString("\t\tuli_sel_kick();\n")
	b.WriteString("\t\tpthread_mutex_lock(&c->mu);\n")
	b.WriteString("\t\tif (c->cap == 0) {\n")
	b.WriteString("\t\t\tif (!c->has_slot && !c->closed) ULI_GC_COND_WAIT(&c->recv_cv, &c->mu);\n")
	b.WriteString("\t\t} else {\n")
	b.WriteString("\t\t\tif (c->len == 0 && !c->closed) ULI_GC_COND_WAIT(&c->recv_cv, &c->mu);\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc->recv_waiters--;\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
	b.WriteString("typedef struct { uli_chan *ch; int is_send; void *data; int *ok; } uli_sel_case;\n")
	b.WriteString("static int uli_select_try(uli_sel_case *cs, int n) {\n")
	b.WriteString("\tfor (int i = 0; i < n; i++) {\n")
	b.WriteString("\t\tif (cs[i].is_send) {\n")
	b.WriteString("\t\t\tif (uli_chan_try_send(cs[i].ch, cs[i].data)) return i;\n")
	b.WriteString("\t\t} else {\n")
	b.WriteString("\t\t\tif (uli_chan_try_recv(cs[i].ch, cs[i].data, cs[i].ok)) return i;\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn -1;\n")
	b.WriteString("}\n")
	b.WriteString("static void uli_select_arm(uli_sel_case *cs, int n) {\n")
	b.WriteString("\tfor (int i = 0; i < n; i++) {\n")
	b.WriteString("\t\tif (cs[i].is_send || !cs[i].ch) continue;\n")
	b.WriteString("\t\tpthread_mutex_lock(&cs[i].ch->mu);\n")
	b.WriteString("\t\tcs[i].ch->recv_waiters++;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&cs[i].ch->mu);\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	b.WriteString("static void uli_select_disarm(uli_sel_case *cs, int n) {\n")
	b.WriteString("\tfor (int i = 0; i < n; i++) {\n")
	b.WriteString("\t\tif (cs[i].is_send || !cs[i].ch) continue;\n")
	b.WriteString("\t\tpthread_mutex_lock(&cs[i].ch->mu);\n")
	b.WriteString("\t\tif (cs[i].ch->recv_waiters > 0) cs[i].ch->recv_waiters--;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&cs[i].ch->mu);\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	// has_def: if set and no case ready, return n (default sentinel).
	// Else arm recv waiters, re-try, then park on uli_sel_cv (gen avoids lost wakeup).
	b.WriteString("static int uli_select(uli_sel_case *cs, int n, int has_def) {\n")
	b.WriteString("\tfor (;;) {\n")
	b.WriteString("\t\tint i = uli_select_try(cs, n);\n")
	b.WriteString("\t\tif (i >= 0) return i;\n")
	b.WriteString("\t\tif (has_def) return n;\n")
	b.WriteString("\t\tuli_select_arm(cs, n);\n")
	b.WriteString("\t\tuli_sel_kick();\n")
	b.WriteString("\t\tpthread_mutex_lock(&uli_sel_mu);\n")
	b.WriteString("\t\tunsigned gen = uli_sel_gen;\n")
	b.WriteString("\t\tpthread_mutex_unlock(&uli_sel_mu);\n")
	b.WriteString("\t\ti = uli_select_try(cs, n);\n")
	b.WriteString("\t\tif (i >= 0) { uli_select_disarm(cs, n); return i; }\n")
	b.WriteString("\t\tpthread_mutex_lock(&uli_sel_mu);\n")
	b.WriteString("\t\twhile (uli_sel_gen == gen) ULI_GC_COND_WAIT(&uli_sel_cv, &uli_sel_mu);\n")
	b.WriteString("\t\tpthread_mutex_unlock(&uli_sel_mu);\n")
	b.WriteString("\t\tuli_select_disarm(cs, n);\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")
}

func (e *emitter) markChanNeeds() {
	if e.info != nil && len(e.info.Chans) > 0 {
		e.needChan = true
		e.needArena = true
	}
}

func (e *emitter) writeSend(b *strings.Builder, s *ast.SendStmt, level int) {
	e.needChan = true
	ct := e.typeOf(s.Chan)
	ci := e.info.Chans[ct]
	elemCT := e.cTypeFrom(ci.Elem)
	tmp := fmt.Sprintf("__send_%d", e.goID)
	e.goID++
	indent(b, level)
	fmt.Fprintf(b, "%s %s = ", elemCT, tmp)
	e.writeExpr(b, s.Value)
	b.WriteString(";\n")
	indent(b, level)
	b.WriteString("uli_chan_send(")
	e.writeExpr(b, s.Chan)
	fmt.Fprintf(b, ", &%s);\n", tmp)
}

func (e *emitter) writeRecvExpr(b *strings.Builder, u *ast.UnaryExpr) {
	e.needChan = true
	ct := e.typeOf(u.X)
	ci := e.info.Chans[ct]
	elemCT := e.cTypeFrom(ci.Elem)
	// GNU statement expression
	b.WriteString("({\n\t")
	fmt.Fprintf(b, "%s __rv;\n\t", elemCT)
	b.WriteString("uli_chan_recv(")
	e.writeExpr(b, u.X)
	b.WriteString(", &__rv);\n\t__rv;\n})")
}

func (e *emitter) writeRecvCommaOk(b *strings.Builder, u *ast.UnaryExpr, valName, okName string) {
	e.needChan = true
	ct := e.typeOf(u.X)
	ci := e.info.Chans[ct]
	elemCT := e.cTypeFrom(ci.Elem)
	fmt.Fprintf(b, "%s %s;\n", elemCT, valName)
	fmt.Fprintf(b, "int %s = uli_chan_recv(", okName)
	e.writeExpr(b, u.X)
	fmt.Fprintf(b, ", &%s);\n", valName)
}

func (e *emitter) writeGoStmt(b *strings.Builder, s *ast.GoStmt, level int) {
	e.needGo = true
	e.needArena = true
	id := e.goID
	e.goID++
	call := s.Call
	if call == nil {
		indent(b, level)
		b.WriteString("/* bad இழை */;\n")
		return
	}

	envName := fmt.Sprintf("uli_go_env_%d", id)
	trampName := fmt.Sprintf("uli_go_tramp_%d", id)

	ft := e.typeOf(call.Fun)
	isFnVal := check.IsFunc(ft)
	var helper string
	if isFnVal {
		e.needFunc = true
		helper = e.writeFuncCallHelper(ft)
	}

	tb := &e.funcTramps
	fmt.Fprintf(tb, "typedef struct { ")
	if isFnVal {
		tb.WriteString("uli_fn f; ")
	}
	for i, a := range call.Args {
		fmt.Fprintf(tb, "%s a%d; ", e.cTypeFrom(e.typeOf(a)), i)
	}
	fmt.Fprintf(tb, "} %s;\n", envName)
	fmt.Fprintf(tb, "static void *%s(void *p) {\n", trampName)
	tb.WriteString("\tuli_thread_register();\n")
	fmt.Fprintf(tb, "\t%s *e = (%s *)p;\n", envName, envName)
	if e.needPanic {
		tb.WriteString("\tuli_panic_push(\"இழை\");\n")
	}
	tb.WriteByte('\t')
	if isFnVal {
		tb.WriteString(helper)
		tb.WriteString("(e->f")
		for i := range call.Args {
			fmt.Fprintf(tb, ", e->a%d", i)
		}
		tb.WriteString(");\n")
	} else if idn, ok := call.Fun.(*ast.Ident); ok {
		tb.WriteString(cPkgIdent(e.pkg, idn.Name))
		tb.WriteByte('(')
		for i := range call.Args {
			if i > 0 {
				tb.WriteString(", ")
			}
			fmt.Fprintf(tb, "e->a%d", i)
		}
		tb.WriteString(");\n")
	} else if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if e.typeOf(sel.X) == check.TypeInvalid {
			if pid, ok := sel.X.(*ast.Ident); ok {
				tb.WriteString(cPkgIdent(e.realPkg(pid.Name), sel.Sel.Name))
				tb.WriteByte('(')
				for i := range call.Args {
					if i > 0 {
						tb.WriteString(", ")
					}
					fmt.Fprintf(tb, "e->a%d", i)
				}
				tb.WriteString(");\n")
			}
		} else {
			tb.WriteString("/* go method call unsupported */(void)e;\n")
		}
	} else {
		tb.WriteString("/* go call */(void)e;\n")
	}
	if e.needPanic {
		tb.WriteString("\tuli_panic_pop();\n")
	}
	tb.WriteString("\tuli_thread_unregister();\n")
	tb.WriteString("\treturn NULL;\n}\n")

	indent(b, level)
	fmt.Fprintf(b, "%s *__goe_%d = (%s *)uli_arena_alloc(sizeof(%s));\n", envName, id, envName, envName)
	if isFnVal {
		indent(b, level)
		fmt.Fprintf(b, "__goe_%d->f = ", id)
		e.writeExpr(b, call.Fun)
		b.WriteString(";\n")
	}
	for i, a := range call.Args {
		indent(b, level)
		fmt.Fprintf(b, "__goe_%d->a%d = ", id, i)
		e.writeExpr(b, a)
		b.WriteString(";\n")
	}
	indent(b, level)
	fmt.Fprintf(b, "uli_go(%s, __goe_%d);\n", trampName, id)
}

func (e *emitter) writeSelect(b *strings.Builder, s *ast.SelectStmt, level int) {
	e.needChan = true
	id := e.goID
	e.goID++

	type caseInfo struct {
		isDefault bool
		isSend    bool
		clause    *ast.CommClause
		chanExpr  ast.Expr
		valExpr   ast.Expr
		recvUnary *ast.UnaryExpr
		assign    ast.Stmt
	}
	var cases []caseInfo
	hasDef := false
	for _, cl := range s.Body {
		if cl.Default {
			hasDef = true
			cases = append(cases, caseInfo{isDefault: true, clause: cl})
			continue
		}
		ci := caseInfo{clause: cl}
		switch comm := cl.Comm.(type) {
		case *ast.SendStmt:
			ci.isSend = true
			ci.chanExpr = comm.Chan
			ci.valExpr = comm.Value
		case *ast.ExprStmt:
			if u, ok := comm.X.(*ast.UnaryExpr); ok && u.Op == token.ARROW {
				ci.recvUnary = u
				ci.chanExpr = u.X
			}
		case *ast.AssignStmt, *ast.ShortVarDecl:
			ci.assign = comm
			var recv ast.Expr
			switch a := comm.(type) {
			case *ast.AssignStmt:
				if len(a.Values) == 1 {
					recv = a.Values[0]
				}
			case *ast.ShortVarDecl:
				if len(a.Values) == 1 {
					recv = a.Values[0]
				}
			}
			if u, ok := recv.(*ast.UnaryExpr); ok && u.Op == token.ARROW {
				ci.recvUnary = u
				ci.chanExpr = u.X
			}
		}
		cases = append(cases, ci)
	}

	nComm := 0
	for _, c := range cases {
		if !c.isDefault {
			nComm++
		}
	}
	hasDefI := 0
	if hasDef {
		hasDefI = 1
	}
	indent(b, level)
	fmt.Fprintf(b, "uli_sel_case __sc_%d[%d];\n", id, max(nComm, 1))
	commI := 0
	for i, c := range cases {
		if c.isDefault {
			continue
		}
		if c.chanExpr == nil || e.info == nil {
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ch = NULL;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].is_send = 0;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].data = NULL;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ok = NULL;\n", id, commI)
			commI++
			continue
		}
		ct := e.typeOf(c.chanExpr)
		ci, ok := e.info.Chans[ct]
		if !ok || !check.IsChan(ct) {
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ch = NULL;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].is_send = 0;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].data = NULL;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ok = NULL;\n", id, commI)
			commI++
			continue
		}
		elem := ci.Elem
		if c.isSend {
			tmp := fmt.Sprintf("__scd_%d_%d", id, i)
			indent(b, level)
			fmt.Fprintf(b, "%s %s = ", e.cTypeFrom(elem), tmp)
			e.writeExpr(b, c.valExpr)
			b.WriteString(";\n")
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ch = ", id, commI)
			e.writeExpr(b, c.chanExpr)
			b.WriteString(";\n")
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].is_send = 1;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].data = &%s;\n", id, commI, tmp)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ok = NULL;\n", id, commI)
		} else {
			tmp := fmt.Sprintf("__scd_%d_%d", id, i)
			okt := fmt.Sprintf("__sco_%d_%d", id, i)
			indent(b, level)
			fmt.Fprintf(b, "%s %s;\n", e.cTypeFrom(elem), tmp)
			indent(b, level)
			fmt.Fprintf(b, "int %s = 0;\n", okt)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ch = ", id, commI)
			e.writeExpr(b, c.chanExpr)
			b.WriteString(";\n")
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].is_send = 0;\n", id, commI)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].data = &%s;\n", id, commI, tmp)
			indent(b, level)
			fmt.Fprintf(b, "__sc_%d[%d].ok = &%s;\n", id, commI, okt)
		}
		commI++
	}
	indent(b, level)
	// has_def → runtime returns nComm as default sentinel
	fmt.Fprintf(b, "int __si_%d = uli_select(__sc_%d, %d, %d);\n", id, id, nComm, hasDefI)
	indent(b, level)
	fmt.Fprintf(b, "switch (__si_%d) {\n", id)

	commI = 0
	for i, c := range cases {
		if c.isDefault {
			indent(b, level)
			fmt.Fprintf(b, "case %d: {\n", nComm)
			e.writeBlock(b, c.clause.Body, level+1)
			indent(b, level+1)
			b.WriteString("break;\n")
			indent(b, level)
			b.WriteString("}\n")
			continue
		}
		indent(b, level)
		fmt.Fprintf(b, "case %d: {\n", commI)
		if !c.isSend && c.assign != nil && c.chanExpr != nil {
			tmp := fmt.Sprintf("__scd_%d_%d", id, i)
			okt := fmt.Sprintf("__sco_%d_%d", id, i)
			switch a := c.assign.(type) {
			case *ast.AssignStmt:
				if len(a.LHS) == 1 {
					indent(b, level+1)
					e.writeExpr(b, a.LHS[0])
					fmt.Fprintf(b, " = %s;\n", tmp)
				} else if len(a.LHS) == 2 {
					indent(b, level+1)
					e.writeExpr(b, a.LHS[0])
					fmt.Fprintf(b, " = %s;\n", tmp)
					indent(b, level+1)
					e.writeExpr(b, a.LHS[1])
					fmt.Fprintf(b, " = %s;\n", okt)
				}
			case *ast.ShortVarDecl:
				if len(a.Names) >= 1 {
					ct := e.typeOf(c.chanExpr)
					elem := e.info.Chans[ct].Elem
					indent(b, level+1)
					fmt.Fprintf(b, "%s %s = %s;\n", e.cTypeFrom(elem), cIdent(a.Names[0].Name), tmp)
				}
				if len(a.Names) >= 2 {
					indent(b, level+1)
					fmt.Fprintf(b, "int %s = %s;\n", cIdent(a.Names[1].Name), okt)
				}
			}
		}
		e.writeBlock(b, c.clause.Body, level+1)
		indent(b, level+1)
		b.WriteString("break;\n")
		indent(b, level)
		b.WriteString("}\n")
		commI++
	}
	indent(b, level)
	b.WriteString("}\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
