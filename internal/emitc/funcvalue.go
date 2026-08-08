package emitc

import (
	"fmt"
	"strings"

	"aram/internal/ast"
	"aram/internal/check"
)

func (e *emitter) ensureFuncRuntime() {
	e.needFunc = true
	e.needArena = true
}

func (e *emitter) writeFuncTypedef(b *strings.Builder) {
	if !e.needFunc {
		return
	}
	b.WriteString("typedef struct { void *fn; void *env; } aram_fn;\n")
}

func (e *emitter) methodTrampKey(si *check.StructInfo, mi *check.MethodInfo) string {
	return cPkgIdent(si.Pkg, si.Name+"_"+mi.Name)
}

func (e *emitter) pkgFuncTrampKey(pkg, name string) string {
	return cPkgIdent(pkg, name)
}

func (e *emitter) ensureMethodFuncValue(si *check.StructInfo, mi *check.MethodInfo) {
	e.ensureFuncRuntime()
	key := e.methodTrampKey(si, mi)
	if e.funcTrampDone == nil {
		e.funcTrampDone = map[string]bool{}
	}
	if e.funcTrampDone[key] {
		return
	}
	e.funcTrampDone[key] = true
	tb := &e.funcTramps
	recvCT := cPkgIdent(si.Pkg, si.Name)
	tramp := "aram_tramp_" + key
	bindVal := "aram_bind_" + key + "_val"
	bindPtr := "aram_bind_" + key + "_ptr"
	results := mi.Results
	params := mi.Params

	// Trampoline
	e.writeFuncSigLine(tb, "static ", tramp, results, true, params)
	tb.WriteString(" {\n")
	if mi.RecvIsPtr {
		fmt.Fprintf(tb, "\t%s *recv = (%s *)env;\n", recvCT, recvCT)
	} else {
		fmt.Fprintf(tb, "\t%s recv = *(%s *)env;\n", recvCT, recvCT)
	}
	tb.WriteByte('\t')
	if len(results) > 0 {
		tb.WriteString("return ")
	}
	tb.WriteString(cPkgIdent(si.Pkg, si.Name+"_"+mi.Name))
	tb.WriteByte('(')
	if mi.RecvIsPtr {
		tb.WriteString("recv")
	} else {
		tb.WriteString("recv")
	}
	for i := range params {
		fmt.Fprintf(tb, ", a%d", i)
	}
	tb.WriteString(");\n}\n")

	// Bind helpers
	fmt.Fprintf(tb, "static aram_fn %s(%s recv) {\n", bindVal, recvCT)
	tb.WriteString("\taram_fn f;\n")
	fmt.Fprintf(tb, "\t%s *env = (%s *)aram_arena_alloc(sizeof(%s));\n", recvCT, recvCT, recvCT)
	tb.WriteString("\t*env = recv;\n")
	fmt.Fprintf(tb, "\tf.fn = (void *)%s;\n", tramp)
	tb.WriteString("\tf.env = env;\n")
	tb.WriteString("\treturn f;\n}\n")

	fmt.Fprintf(tb, "static aram_fn %s(%s *recv) {\n", bindPtr, recvCT)
	tb.WriteString("\taram_fn f;\n")
	fmt.Fprintf(tb, "\tf.fn = (void *)%s;\n", tramp)
	tb.WriteString("\tf.env = recv;\n")
	tb.WriteString("\treturn f;\n}\n")
}

func (e *emitter) ensurePkgFuncValue(pkg, name string, params, results []check.Type) {
	e.ensureFuncRuntime()
	key := e.pkgFuncTrampKey(pkg, name)
	if e.funcTrampDone == nil {
		e.funcTrampDone = map[string]bool{}
	}
	if e.funcTrampDone[key] {
		return
	}
	e.funcTrampDone[key] = true
	tb := &e.funcTramps
	tramp := "aram_tramp_" + key
	bind := "aram_fv_" + key

	e.writeFuncSigLine(tb, "static ", tramp, results, true, params)
	tb.WriteString(" {\n")
	tb.WriteString("\t(void)env;\n")
	tb.WriteByte('\t')
	if len(results) > 0 {
		tb.WriteString("return ")
	}
	tb.WriteString(cPkgIdent(pkg, name))
	tb.WriteByte('(')
	for i := range params {
		if i > 0 {
			tb.WriteString(", ")
		}
		fmt.Fprintf(tb, "a%d", i)
	}
	tb.WriteString(");\n}\n")

	fmt.Fprintf(tb, "static aram_fn %s(void) {\n", bind)
	tb.WriteString("\taram_fn f;\n")
	fmt.Fprintf(tb, "\tf.fn = (void *)%s;\n", tramp)
	tb.WriteString("\tf.env = NULL;\n")
	tb.WriteString("\treturn f;\n}\n")
}

// writeFuncCallHelper emits aram_fn_call_tN once.
func (e *emitter) writeFuncCallHelper(ft check.Type) string {
	e.ensureFuncRuntime()
	name := fmt.Sprintf("aram_fn_call_t%d", int(ft))
	if e.funcCallDone == nil {
		e.funcCallDone = map[string]bool{}
	}
	if e.funcCallDone[name] {
		return name
	}
	e.funcCallDone[name] = true
	fi := e.info.Funcs[ft]
	tb := &e.funcTramps

	ret := "void"
	if len(fi.Results) == 1 {
		ret = e.cTypeFrom(fi.Results[0])
	} else if len(fi.Results) > 1 {
		ret = e.retCName(fi.Results)
	}
	fmt.Fprintf(tb, "static %s %s(aram_fn f", ret, name)
	for i, p := range fi.Params {
		fmt.Fprintf(tb, ", %s a%d", e.cTypeFrom(p), i)
	}
	tb.WriteString(") {\n\t")
	if len(fi.Results) > 0 {
		tb.WriteString("return ")
	}
	tb.WriteString("((")
	tb.WriteString(ret)
	tb.WriteString("(*)(void *")
	for _, p := range fi.Params {
		tb.WriteString(", ")
		tb.WriteString(e.cTypeFrom(p))
	}
	tb.WriteString("))f.fn)(f.env")
	for i := range fi.Params {
		fmt.Fprintf(tb, ", a%d", i)
	}
	tb.WriteString(");\n}\n")
	return name
}

func (e *emitter) writeFuncSigLine(b *strings.Builder, prefix, name string, results []check.Type, withEnv bool, params []check.Type) {
	ret := "void"
	if len(results) == 1 {
		ret = e.cTypeFrom(results[0])
	} else if len(results) > 1 {
		ret = e.retCName(results)
	}
	b.WriteString(prefix)
	b.WriteString(ret)
	b.WriteByte(' ')
	b.WriteString(name)
	b.WriteByte('(')
	if withEnv {
		b.WriteString("void *env")
		for i, p := range params {
			fmt.Fprintf(b, ", %s a%d", e.cTypeFrom(p), i)
		}
	} else {
		for i, p := range params {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "%s a%d", e.cTypeFrom(p), i)
		}
	}
	b.WriteByte(')')
}

func (e *emitter) writeMethodValue(b *strings.Builder, expr *ast.SelectorExpr, mv *check.MethodValueInfo) {
	e.ensureMethodFuncValue(mv.Struct, mv.Method)
	key := e.methodTrampKey(mv.Struct, mv.Method)
	if mv.Method.RecvIsPtr {
		fmt.Fprintf(b, "aram_bind_%s_ptr(", key)
		if mv.TakeAddr {
			e.writeAddrOf(b, expr.X)
		} else {
			e.writeExpr(b, expr.X)
		}
		b.WriteByte(')')
		return
	}
	fmt.Fprintf(b, "aram_bind_%s_val(", key)
	if mv.RecvIsPtr {
		b.WriteByte('*')
		e.writeExpr(b, expr.X)
	} else {
		e.writeExpr(b, expr.X)
	}
	b.WriteByte(')')
}

func (e *emitter) writePkgFuncValue(b *strings.Builder, fv *check.PkgFuncValueInfo) {
	e.ensurePkgFuncValue(fv.Pkg, fv.Name, fv.Params, fv.Results)
	fmt.Fprintf(b, "aram_fv_%s()", e.pkgFuncTrampKey(fv.Pkg, fv.Name))
}

func (e *emitter) writeFuncValueCall(b *strings.Builder, call *ast.CallExpr, ft check.Type) {
	helper := e.writeFuncCallHelper(ft)
	b.WriteString(helper)
	b.WriteByte('(')
	e.writeExpr(b, call.Fun)
	for _, a := range call.Args {
		b.WriteString(", ")
		e.writeExpr(b, a)
	}
	b.WriteByte(')')
}

func (e *emitter) capIdent(name string) string {
	return "_cap_" + cIdent(name)
}

func (e *emitter) isPromoted(name string) bool {
	_, ok := e.promoted[name]
	return ok
}

func (e *emitter) writeAddrOf(b *strings.Builder, expr ast.Expr) {
	if id, ok := expr.(*ast.Ident); ok && e.isPromoted(id.Name) {
		b.WriteString(e.capIdent(id.Name))
		return
	}
	b.WriteByte('&')
	e.writeExpr(b, expr)
}

func (e *emitter) writePromotedIdent(b *strings.Builder, name string) {
	b.WriteString("(*")
	b.WriteString(e.capIdent(name))
	b.WriteString(")")
}

func (e *emitter) closureKey(ci *check.ClosureInfo) string {
	return fmt.Sprintf("cl%d", ci.ID)
}

func (e *emitter) envStructName(ci *check.ClosureInfo) string {
	return "aram_env_" + e.closureKey(ci)
}

func (e *emitter) ensureClosure(ci *check.ClosureInfo) {
	if ci == nil || ci.Lit == nil {
		return
	}
	e.ensureFuncRuntime()
	key := e.closureKey(ci)
	if e.funcTrampDone == nil {
		e.funcTrampDone = map[string]bool{}
	}
	if e.funcTrampDone[key] {
		return
	}
	e.funcTrampDone[key] = true

	var tb strings.Builder
	envName := e.envStructName(ci)
	tramp := "aram_tramp_" + key
	bind := "aram_bind_" + key

	// Env struct
	if len(ci.Captures) > 0 {
		fmt.Fprintf(&tb, "typedef struct { ")
		for i, cap := range ci.Captures {
			fmt.Fprintf(&tb, "%s *c%d; ", e.cTypeFrom(cap.Type), i)
		}
		fmt.Fprintf(&tb, "} %s;\n", envName)
	}

	// Trampoline
	e.writeFuncSigLine(&tb, "static ", tramp, ci.Results, true, ci.Params)
	tb.WriteString(" {\n")
	prevPromo := e.promoted
	e.promoted = map[string]check.Type{}
	if len(ci.Captures) == 0 {
		tb.WriteString("\t(void)env;\n")
	} else {
		fmt.Fprintf(&tb, "\t%s *_e = (%s *)env;\n", envName, envName)
		for i, cap := range ci.Captures {
			fmt.Fprintf(&tb, "\t%s *%s = _e->c%d;\n", e.cTypeFrom(cap.Type), e.capIdent(cap.Name), i)
			e.promoted[cap.Name] = cap.Type
		}
	}
	if e.info != nil {
		for name, t := range e.info.PromoteInLit[ci.Lit] {
			e.promoted[name] = t
		}
	}
	for i, p := range ci.Lit.Params {
		if p.Name == nil {
			continue
		}
		name := p.Name.Name
		ct := e.cTypeFrom(ci.Params[i])
		if e.isPromoted(name) && !captureHas(ci, name) {
			fmt.Fprintf(&tb, "\t%s *%s = (%s *)aram_arena_alloc(sizeof(%s));\n", ct, e.capIdent(name), ct, ct)
			fmt.Fprintf(&tb, "\t*%s = a%d;\n", e.capIdent(name), i)
		} else if !captureHas(ci, name) {
			fmt.Fprintf(&tb, "\t%s %s = a%d;\n", ct, cIdent(name), i)
		}
	}
	prevFn := e.curFn
	e.curFn = &ast.FuncDecl{Params: ci.Lit.Params, Results: ci.Lit.Results, Body: ci.Lit.Body}
	prevDefer := e.fnHasDefer
	e.fnHasDefer = hasDeferStmt(ci.Lit.Body)
	if e.fnHasDefer {
		e.needDefer = true
		tb.WriteString("\taram_defer_frame *_defers = NULL;\n")
	}
	e.writeBlock(&tb, ci.Lit.Body, 1)
	if e.fnHasDefer {
		e.writeDeferEpilogue(&tb, e.curFn)
	}
	e.curFn = prevFn
	e.fnHasDefer = prevDefer
	e.promoted = prevPromo
	tb.WriteString("}\n")

	// Bind helper
	fmt.Fprintf(&tb, "static aram_fn %s(", bind)
	if len(ci.Captures) == 0 {
		tb.WriteString("void")
	} else {
		for i, cap := range ci.Captures {
			if i > 0 {
				tb.WriteString(", ")
			}
			fmt.Fprintf(&tb, "%s *c%d", e.cTypeFrom(cap.Type), i)
		}
	}
	tb.WriteString(") {\n\taram_fn f;\n")
	fmt.Fprintf(&tb, "\tf.fn = (void *)%s;\n", tramp)
	if len(ci.Captures) == 0 {
		tb.WriteString("\tf.env = NULL;\n")
	} else {
		fmt.Fprintf(&tb, "\t%s *e = (%s *)aram_arena_alloc(sizeof(%s));\n", envName, envName, envName)
		for i := range ci.Captures {
			fmt.Fprintf(&tb, "\te->c%d = c%d;\n", i, i)
		}
		tb.WriteString("\tf.env = e;\n")
	}
	tb.WriteString("\treturn f;\n}\n")

	e.funcTramps.WriteString(tb.String())
}

func captureHas(ci *check.ClosureInfo, name string) bool {
	for _, c := range ci.Captures {
		if c.Name == name {
			return true
		}
	}
	return false
}

func (e *emitter) writeFuncLit(b *strings.Builder, lit *ast.FuncLit) {
	if e.info == nil {
		b.WriteString("/*bad func lit*/(aram_fn){0}")
		return
	}
	ci := e.info.Closures[lit]
	if ci == nil {
		b.WriteString("/*bad func lit*/(aram_fn){0}")
		return
	}
	e.ensureClosure(ci)
	fmt.Fprintf(b, "aram_bind_%s(", e.closureKey(ci))
	for i, cap := range ci.Captures {
		if i > 0 {
			b.WriteString(", ")
		}
		if e.isPromoted(cap.Name) {
			b.WriteString(e.capIdent(cap.Name))
		} else {
			// Should have been promoted; take address of stack local as fallback.
			b.WriteByte('&')
			b.WriteString(cIdent(cap.Name))
		}
	}
	b.WriteByte(')')
}

func (e *emitter) writePromoteAlloc(b *strings.Builder, name string, t check.Type, init string, level int) {
	e.ensureFuncRuntime()
	ct := e.cTypeFrom(t)
	indent(b, level)
	fmt.Fprintf(b, "%s *%s = (%s *)aram_arena_alloc(sizeof(%s));\n", ct, e.capIdent(name), ct, ct)
	indent(b, level)
	fmt.Fprintf(b, "*%s = %s;\n", e.capIdent(name), init)
}
