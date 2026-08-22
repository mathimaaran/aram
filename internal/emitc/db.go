package emitc

import (
	_ "embed"
	"strings"

	"niraluli/internal/ast"
)

//go:embed db_runtime.inc
var dbRuntimeC string

func (e *emitter) markDBNeeds(pkgNames []string) {
	for _, name := range pkgNames {
		if name == "தரவுத்தளம்" {
			e.needDB = true
			e.needArena = true
			e.needSlice = true
			return
		}
	}
}

func (e *emitter) writeDBRuntime(b *strings.Builder) {
	if !e.needDB {
		return
	}
	e.needArena = true
	e.needSlice = true
	b.WriteString(dbRuntimeC)
	b.WriteByte('\n')
}

func (e *emitter) dbTupleResult(fn *ast.FuncDecl, cResult, valueExpr string) string {
	errType := cPkgIdent("தரவுத்தளம்", "பிழை") + " *"
	ret := e.retCName(e.resultTypes(fn))
	return "\t" + cResult + " r = " + valueExpr + ";\n" +
		"\treturn (" + ret + "){ r.value, (" + errType + ")r.err };\n"
}

func (e *emitter) writeDBIntrinsic(b *strings.Builder, fn *ast.FuncDecl) bool {
	if e.pkg != "தரவுத்தளம்" || fn == nil || fn.Name == nil || fn.Recv != nil {
		return false
	}
	errType := cPkgIdent("தரவுத்தளம்", "பிழை") + " *"
	ret := e.retCName(e.resultTypes(fn))
	var call string
	switch fn.Name.Name {
	case "திற":
		call = e.dbTupleResult(fn, "uli_db_i64_result",
			"uli_db_open("+cIdent("இயக்கி")+", "+cIdent("முகவரி")+")")
	case "விடு":
		call = "\treturn (" + errType + ")uli_db_close(" + cIdent("இ") + ");\n"
	case "செயல்":
		resultType := cPkgIdent("தரவுத்தளம்", "செயல்முடிவு")
		call = "\tuli_db_exec_result r = uli_db_exec(" + cIdent("இ") + ", " + cIdent("வினா") + ");\n" +
			"\t" + resultType + " result;\n" +
			"\tresult." + cIdent("பாதித்தவை") + " = r.affected;\n" +
			"\tresult." + cIdent("சேர்க்கையெண்") + " = r.inserted;\n" +
			"\treturn (" + ret + "){ result, (" + errType + ")r.err };\n"
	case "தயாரி":
		call = e.dbTupleResult(fn, "uli_db_i64_result",
			"uli_db_prepare("+cIdent("இ")+", "+cIdent("வினா")+")")
	case "இன்மைபிணை":
		call = "\treturn (" + errType + ")uli_db_bind_null(" + cIdent("க") + ", " + cIdent("இடம்") + ");\n"
	case "முழுஎண்பிணை":
		call = "\treturn (" + errType + ")uli_db_bind_i64(" + cIdent("க") + ", " + cIdent("இடம்") + ", " + cIdent("மதிப்பு") + ");\n"
	case "மிதவைப்பிணை":
		call = "\treturn (" + errType + ")uli_db_bind_f64(" + cIdent("க") + ", " + cIdent("இடம்") + ", " + cIdent("மதிப்பு") + ");\n"
	case "சரம்பிணை":
		call = "\treturn (" + errType + ")uli_db_bind_text(" + cIdent("க") + ", " + cIdent("இடம்") + ", " + cIdent("மதிப்பு") + ");\n"
	case "இருமிப்பிணை":
		call = "\treturn (" + errType + ")uli_db_bind_blob(" + cIdent("க") + ", " + cIdent("இடம்") + ", " + cIdent("மதிப்பு") + ".data, " + cIdent("மதிப்பு") + ".len);\n"
	case "அடுத்து":
		call = e.dbTupleResult(fn, "uli_db_bool_result",
			"uli_db_step("+cIdent("க")+")")
	case "நெடுவரிசைகள்":
		call = e.dbTupleResult(fn, "uli_db_i64_result",
			"uli_db_column_count("+cIdent("க")+")")
	case "நெடுவரிசைப்பெயர்":
		call = e.dbTupleResult(fn, "uli_db_str_result",
			"uli_db_column_name("+cIdent("க")+", "+cIdent("இடம்")+")")
	case "இன்மையா":
		call = e.dbTupleResult(fn, "uli_db_bool_result",
			"uli_db_column_is_null("+cIdent("க")+", "+cIdent("இடம்")+")")
	case "முழுஎண்பெறு":
		call = e.dbTupleResult(fn, "uli_db_i64_result",
			"uli_db_column_i64("+cIdent("க")+", "+cIdent("இடம்")+")")
	case "மிதவைப்பெறு":
		call = e.dbTupleResult(fn, "uli_db_f64_result",
			"uli_db_column_f64("+cIdent("க")+", "+cIdent("இடம்")+")")
	case "சரம்பெறு":
		call = e.dbTupleResult(fn, "uli_db_str_result",
			"uli_db_column_text("+cIdent("க")+", "+cIdent("இடம்")+")")
	case "இருமிப்பெறு":
		call = "\tuli_db_blob_result r = uli_db_column_blob(" + cIdent("க") + ", " + cIdent("இடம்") + ");\n" +
			"\tuli_slice_u8 value = { r.data, r.len, r.len };\n" +
			"\treturn (" + ret + "){ value, (" + errType + ")r.err };\n"
	case "மீட்டமை":
		call = "\treturn (" + errType + ")uli_db_reset(" + cIdent("க") + ");\n"
	case "கூற்றுவிடு":
		call = "\treturn (" + errType + ")uli_db_finalize(" + cIdent("க") + ");\n"
	default:
		return false
	}
	e.needDB = true
	e.needArena = true
	e.needSlice = true
	b.WriteString(e.cFuncSig(fn))
	b.WriteString(" {\n")
	b.WriteString(call)
	b.WriteString("}\n")
	return true
}
