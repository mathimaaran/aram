package emitc

import (
	_ "embed"
	"strings"
)

//go:embed gc_runtime.inc
var gcRuntimeC string

func (e *emitter) writeGCRuntime(b *strings.Builder) {
	if !e.needArena {
		b.WriteString("static void aram_gc_poll(void) {}\n")
		return
	}
	b.WriteString(gcRuntimeC)
	b.WriteByte('\n')
}

func (e *emitter) writeGCPoll(b *strings.Builder, level int) {
	indent(b, level)
	b.WriteString("aram_gc_poll();\n")
}
