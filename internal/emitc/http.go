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
	b.WriteString("#include <strings.h>\n")
	b.WriteString("static void *aram_http_headers_new(void);\n")
	b.WriteString("static void aram_http_headers_set(void *, const char *, const char *);\n")
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
		b.WriteString("static void *aram_http_headers_new(void) {\n")
		b.WriteString("\treturn (void *)" + e.mapMakeName(t) + "(0);\n")
		b.WriteString("}\n")
		b.WriteString("static void aram_http_headers_set(void *m, const char *k, const char *v) {\n")
		b.WriteString("\t" + e.mapSetName(t) + "((" + mapType + ")m, k, v);\n")
		b.WriteString("}\n")
		return
	}
}

func (e *emitter) writeHttpIntrinsic(b *strings.Builder, fn *ast.FuncDecl) bool {
	if e.pkg != "பரிமாற்றம்" || fn == nil || fn.Name == nil || fn.Recv != nil {
		return false
	}
	errType := cPkgIdent("வலை", "பிழை") + " *"
	var call string
	switch fn.Name.Name {
	case "புதியவழிப்படுத்தி":
		call = "\treturn aram_http_mux_new();\n"
	case "பதிவு":
		call = "\taram_http_mux_handle(" + cIdent("வ") + ", " + cIdent("பாதை") + ", " + cIdent("க") + ");\n"
	case "வழிசேவைஒன்று":
		call = "\treturn (" + errType + ")aram_http_mux_serve_one(" + cIdent("கேட்பி") + ", " + cIdent("வ") + ");\n"
	case "வழிசேவை":
		call = "\treturn (" + errType + ")aram_http_mux_listen_and_serve(" + cIdent("முகவரி") + ", " + cIdent("வ") + ");\n"
	case "எழுது":
		call = "\treturn (" + errType + ")aram_http_write(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("உடல்") + ".data, " + cIdent("உடல்") + ".len);\n"
	case "எழுதுசரம்":
		call = "\treturn (" + errType + ")aram_http_write_str(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("உடல்") + ");\n"
	case "சேவைஒன்று":
		call = "\treturn (" + errType + ")aram_http_serve_one(" + cIdent("கேட்பி") + ", " + cIdent("க") + ");\n"
	case "கேட்டுசேவை":
		call = "\treturn (" + errType + ")aram_http_listen_and_serve(" + cIdent("முகவரி") + ", " + cIdent("க") + ");\n"
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
