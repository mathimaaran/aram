package emitc

import (
	_ "embed"
	"strings"

	"aram/internal/ast"
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
	b.WriteString("#include <strings.h>\n")
	b.WriteString(httpRuntimeC)
	b.WriteByte('\n')
}

func (e *emitter) writeHttpIntrinsic(b *strings.Builder, fn *ast.FuncDecl) bool {
	if e.pkg != "பரிமாற்றம்" || fn == nil || fn.Name == nil || fn.Recv != nil {
		return false
	}
	var call string
	switch fn.Name.Name {
	case "எழுது":
		call = "\taram_http_write(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("உடல்") + ".data, " + cIdent("உடல்") + ".len);\n"
	case "எழுதுசரம்":
		call = "\taram_http_write_str(" + cIdent("ப") + ", " + cIdent("குறியீடு") + ", " + cIdent("உள்ளடக்கம்") + ", " + cIdent("உடல்") + ");\n"
	case "சேவைஒன்று":
		call = "\taram_http_serve_one(" + cIdent("கேட்பி") + ", " + cIdent("க") + ");\n"
	case "கேட்டுசேவை":
		call = "\taram_http_listen_and_serve(" + cIdent("முகவரி") + ", " + cIdent("க") + ");\n"
	default:
		return false
	}
	e.needHttp = true
	e.needNet = true
	e.needFunc = true
	e.needArena = true
	e.needSlice = true
	b.WriteString(e.cFuncSig(fn))
	b.WriteString(" {\n")
	b.WriteString(call)
	b.WriteString("}\n")
	return true
}
