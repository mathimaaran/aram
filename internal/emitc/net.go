package emitc

import (
	_ "embed"
	"strings"

	"aram/internal/ast"
)

//go:embed net_runtime.inc
var netRuntimeC string

func (e *emitter) markNetNeeds(pkgNames []string) {
	for _, n := range pkgNames {
		if n == "வலை" {
			e.needNet = true
			e.needArena = true
			e.needSlice = true
			return
		}
	}
}

func (e *emitter) writeNetRuntime(b *strings.Builder) {
	if !e.needNet {
		return
	}
	e.needArena = true
	e.needSlice = true
	b.WriteString(netRuntimeC)
	b.WriteByte('\n')
}

func (e *emitter) writeNetIntrinsic(b *strings.Builder, fn *ast.FuncDecl) bool {
	if e.pkg != "வலை" || fn == nil || fn.Name == nil || fn.Recv != nil {
		return false
	}
	var call string
	switch fn.Name.Name {
	case "கேள்":
		call = "\treturn aram_net_listen(" + cIdent("முகவரி") + ");\n"
	case "ஏற்று":
		call = "\treturn aram_net_accept(" + cIdent("க") + ");\n"
	case "இணை":
		call = "\treturn aram_net_dial(" + cIdent("முகவரி") + ");\n"
	case "படி":
		call = "\treturn aram_net_read(" + cIdent("இ") + ", " + cIdent("இடம்") + ".data, " + cIdent("இடம்") + ".len);\n"
	case "எழுது":
		call = "\treturn aram_net_write(" + cIdent("இ") + ", " + cIdent("தரவு") + ".data, " + cIdent("தரவு") + ".len);\n"
	case "விடு":
		call = "\taram_net_close(" + cIdent("இ") + ");\n"
	case "நிறுத்து":
		call = "\taram_net_close(" + cIdent("க") + ");\n"
	case "முகவரி":
		call = "\treturn aram_net_addr(" + cIdent("க") + ");\n"
	default:
		return false
	}
	e.needNet = true
	e.needArena = true
	e.needSlice = true
	b.WriteString(e.cFuncSig(fn))
	b.WriteString(" {\n")
	b.WriteString(call)
	b.WriteString("}\n")
	return true
}
