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
	errType := cPkgIdent("வலை", "பிழை") + " *"
	ret := e.retCName(e.resultTypes(fn))
	var call string
	switch fn.Name.Name {
	case "கேள்":
		call = "\taram_net_i64_result r = aram_net_listen(" + cIdent("முகவரி") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "ஏற்று":
		call = "\taram_net_i64_result r = aram_net_accept(" + cIdent("க") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "இணை":
		call = "\taram_net_i64_result r = aram_net_dial(" + cIdent("முகவரி") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "படி":
		call = "\taram_net_i64_result r = aram_net_read(" + cIdent("இ") + ", " + cIdent("இடம்") + ".data, " + cIdent("இடம்") + ".len);\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "எழுது":
		call = "\taram_net_i64_result r = aram_net_write(" + cIdent("இ") + ", " + cIdent("தரவு") + ".data, " + cIdent("தரவு") + ".len);\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "விடு":
		call = "\treturn (" + errType + ")aram_net_close(" + cIdent("இ") + ");\n"
	case "நிறுத்து":
		call = "\treturn (" + errType + ")aram_net_close(" + cIdent("க") + ");\n"
	case "முகவரி":
		call = "\taram_net_str_result r = aram_net_addr(" + cIdent("க") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "தகவல்கேள்":
		call = "\taram_net_i64_result r = aram_net_udp_listen(" + cIdent("முகவரி") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "தகவல்முகவரி":
		call = "\taram_net_str_result r = aram_net_addr(" + cIdent("த") + ");\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "தகவலனுப்பு":
		call = "\taram_net_i64_result r = aram_net_udp_send(" + cIdent("த") + ", " + cIdent("முகவரி") + ", " + cIdent("தரவு") + ".data, " + cIdent("தரவு") + ".len);\n\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
	case "தகவல்பெறு":
		call = "\taram_net_udp_result r = aram_net_udp_recv(" + cIdent("த") + ", " + cIdent("இடம்") + ".data, " + cIdent("இடம்") + ".len);\n\treturn (" + ret + "){ r.n, r.addr, (" + errType + ")r.err };\n"
	case "தகவல்விடு":
		call = "\treturn (" + errType + ")aram_net_close(" + cIdent("த") + ");\n"
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
