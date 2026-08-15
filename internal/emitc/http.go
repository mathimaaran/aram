package emitc

import (
	_ "embed"
	"strings"

	"aram/internal/ast"
	"aram/internal/check"
)

//go:embed http_runtime.inc
var httpRuntimeC string

func (e *emitter) markHttpNeeds(pkgNames []string) {
	for _, n := range pkgNames {
		if n == "பரிமாற்றம்" {
			e.needHttp = true
			e.needNet = true
			e.needFunc = true
			e.needArena = true
			e.needSlice = true
			e.needMap = true
			return
		}
	}
}

func (e *emitter) writeHttpRuntime(b *strings.Builder) {
	if !e.needHttp {
		return
	}
	e.needNet = true
	e.needFunc = true
	e.needArena = true
	e.needSlice = true
	e.needMap = true
	b.WriteString("#include <ctype.h>\n")
	b.WriteString("#include <stdarg.h>\n")
	b.WriteString("#include <strings.h>\n")
	b.WriteString("static void *aram_http_headers_new(void);\n")
	b.WriteString("static void aram_http_headers_set(void *, const char *, const char *);\n")
	b.WriteString("typedef void (*aram_http_header_visit_fn)(const char *, const char *, void *);\n")
	b.WriteString("static void aram_http_headers_each(void *, aram_http_header_visit_fn, void *);\n")
	b.WriteString(httpRuntimeC)
	b.WriteByte('\n')
}

func (e *emitter) writeHttpMapBridge(b *strings.Builder) {
	if !e.needHttp || e.info == nil {
		return
	}
	for t, mi := range e.info.Maps {
		if mi.Key != check.TypeString || mi.Elem != check.TypeString {
			continue
		}
		mapType := e.mapCName(t)
		tabType := e.mapTabName(t)
		b.WriteString("static void *aram_http_headers_new(void) {\n")
		b.WriteString("\treturn (void *)" + e.mapMakeName(t) + "(0);\n")
		b.WriteString("}\n")
		b.WriteString("static void aram_http_headers_set(void *m, const char *k, const char *v) {\n")
		b.WriteString("\t" + e.mapSetName(t) + "((" + mapType + ")m, k, v);\n")
		b.WriteString("}\n")
		b.WriteString("static void aram_http_headers_each(void *m, aram_http_header_visit_fn fn, void *ctx) {\n")
		b.WriteString("\t" + mapType + " tab = (" + mapType + ")m;\n")
		b.WriteString("\tif (!tab || !fn) return;\n")
		b.WriteString("\tfor (int64_t i = 0; i < tab->cap; i++) {\n")
		b.WriteString("\t\tif (tab->entries[i].state != 1) continue;\n")
		b.WriteString("\t\tfn(tab->entries[i].key, tab->entries[i].val, ctx);\n")
		b.WriteString("\t}\n")
		b.WriteString("\t(void)sizeof(" + tabType + ");\n")
		b.WriteString("}\n")
		return
	}
}

func (e *emitter) httpClientRetCall(fn *ast.FuncDecl, callExpr string) string {
	errType := cPkgIdent("வலை", "பிழை") + " *"
	respType := cPkgIdent("பரிமாற்றம்", "பதில்")
	ret := e.retCName(e.resultTypes(fn))
	codeF := cIdent("குறியீடு")
	statusF := cIdent("விளக்கம்")
	hdrF := cIdent("தலைப்புகள்")
	bodyF := cIdent("உடல்")
	mapType := "void *"
	for t, mi := range e.info.Maps {
		if mi.Key == check.TypeString && mi.Elem == check.TypeString {
			mapType = e.mapCName(t)
			break
		}
	}
	var b strings.Builder
	b.WriteString("\taram_http_client_result r = " + callExpr + ";\n")
	b.WriteString("\t" + respType + " resp;\n")
	b.WriteString("\tmemset(&resp, 0, sizeof resp);\n")
	b.WriteString("\tresp." + codeF + " = r.code;\n")
	b.WriteString("\tresp." + statusF + " = r.status ? r.status : \"\";\n")
	b.WriteString("\tresp." + hdrF + " = (" + mapType + ")r.headers;\n")
	b.WriteString("\tresp." + bodyF + " = r.body;\n")
	b.WriteString("\treturn (" + ret + "){ resp, (" + errType + ")r.err };\n")
	return b.String()
}

func (e *emitter) writeHttpIntrinsic(b *strings.Builder, fn *ast.FuncDecl) bool {
	if e.pkg != "பரிமாற்றம்" || fn == nil || fn.Name == nil || fn.Recv != nil {
		return false
	}
	errType := cPkgIdent("வலை", "பிழை") + " *"
	var call string
	switch fn.Name.Name {
	case "எழுது":
		call = "\treturn (" + errType + ")aram_http_write(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", NULL, " + cIdent("உடல்") + ".data, " + cIdent("உடல்") + ".len);\n"
	case "எழுதுசரம்":
		call = "\treturn (" + errType + ")aram_http_write_str(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", NULL, " + cIdent("உடல்") + ");\n"
	case "எழுதுதலைப்புகள்":
		call = "\treturn (" + errType + ")aram_http_write(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("தலைப்புகள்") + ", " + cIdent("உடல்") + ".data, " + cIdent("உடல்") + ".len);\n"
	case "எழுதுசரந்தலைப்புகள்":
		call = "\treturn (" + errType + ")aram_http_write_str(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("தலைப்புகள்") + ", " + cIdent("உடல்") + ");\n"
	case "சேவைஒன்று":
		call = "\treturn (" + errType + ")aram_http_serve_one(" + cIdent("கேட்பி") + ", " + cIdent("க") + ");\n"
	case "கேட்டுசேவை":
		call = "\treturn (" + errType + ")aram_http_listen_and_serve(" + cIdent("முகவரி") + ", " + cIdent("க") + ");\n"
	case "பெறு":
		call = e.httpClientRetCall(fn, "aram_http_get("+cIdent("முகவரி")+")")
	case "பதிவிடு":
		call = e.httpClientRetCall(fn, "aram_http_post("+cIdent("முகவரி")+", "+cIdent("உள்ளடக்கம்")+", "+cIdent("உடல்")+".data, "+cIdent("உடல்")+".len)")
	case "கோரு":
		call = e.httpClientRetCall(fn, "aram_http_do("+cIdent("முறைமை")+", "+cIdent("முகவரி")+", "+cIdent("தலைப்புகள்")+", "+cIdent("உடல்")+".data, "+cIdent("உடல்")+".len)")
	default:
		return false
	}
	e.needHttp = true
	e.needNet = true
	e.needFunc = true
	e.needArena = true
	e.needSlice = true
	e.needMap = true
	b.WriteString(e.cFuncSig(fn))
	b.WriteString(" {\n")
	b.WriteString(call)
	b.WriteString("}\n")
	return true
}
