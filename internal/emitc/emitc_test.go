package emitc_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"niraluli/internal/ast"
	"niraluli/internal/check"
	"niraluli/internal/emitc"
	"niraluli/internal/parse"
)

func root(t *testing.T) string {
	t.Helper()
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Join(filepath.Dir(f), "..", "..")
}

func compile(t *testing.T, name string) (*ast.File, *check.Info, string) {
	t.Helper()
	path := filepath.Join(root(t), "corpus", "tamil", name)
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatal(perrs)
	}
	info, errs := check.File(file)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.Emit(file, info)
	if err != nil {
		t.Fatal(err)
	}
	return file, info, cSrc
}

func runC(t *testing.T, cSrc string) string {
	return runCTimeout(t, cSrc, 10*time.Second)
}

func runCTimeout(t *testing.T, cSrc string, d time.Duration) string {
	t.Helper()
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	dir := t.TempDir()
	cFile := filepath.Join(dir, "out.c")
	bin := filepath.Join(dir, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O2", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin).CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("program timed out after %s\nC:\n%s", d, cSrc)
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func runCExpectFail(t *testing.T, cSrc string, d time.Duration) string {
	t.Helper()
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	dir := t.TempDir()
	cFile := filepath.Join(dir, "out.c")
	bin := filepath.Join(dir, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin).CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("program timed out after %s\nC:\n%s", d, cSrc)
	}
	if err == nil {
		t.Fatalf("expected non-zero exit, got %q\nC:\n%s", out, cSrc)
	}
	return string(out)
}

func TestEmitAndRunVanakkam(t *testing.T) {
	_, _, cSrc := compile(t, "வணக்கம்.uli")
	got := runC(t, cSrc)
	want := "வணக்கம், நிரலுளி!\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEmitAndRunAdd(t *testing.T) {
	_, _, cSrc := compile(t, "கூட்டு.uli")
	got := runC(t, cSrc)
	want := "9\n7\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunLoop(t *testing.T) {
	_, _, cSrc := compile(t, "சுழல்.uli")
	got := runC(t, cSrc)
	want := "0\n1\n2\n3\n4\n3\n2\n1\n0\n1\n2\n0\n1\n3\n4\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunString(t *testing.T) {
	_, _, cSrc := compile(t, "சரம்.uli")
	if !strings.Contains(cSrc, "uli_concat_many(3, __uli_parts)") {
		t.Fatalf("string chain was not flattened into one concat\n%s", cSrc)
	}
	got := runC(t, cSrc)
	want := "வணக்கம், நிரலுளி!\nமெய்\nபொய்\nஆம்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSlice(t *testing.T) {
	_, _, cSrc := compile(t, "பட்டியல்.uli")
	got := runC(t, cSrc)
	want := "3\n20\n99\n2\nநிரலுளி\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRange(t *testing.T) {
	_, _, cSrc := compile(t, "ஒவ்வொரு.uli")
	got := runC(t, cSrc)
	want := "0\n10\n1\n20\n2\n30\n10\n20\n30\n60\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRangeChan(t *testing.T) {
	_, _, cSrc := compile(t, "ஒவ்வொரு_தடம்.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "1\n2\n3\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunAppendSliceString(t *testing.T) {
	_, _, cSrc := compile(t, "சேர்.uli")
	got := runC(t, cSrc)
	want := "3\n3\n2\n2\nell\n5\n0\n97\n1\n98\nவணக்கம், நிரலுளி\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunStruct(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு.uli")
	got := runC(t, cSrc)
	want := "3\n9\n0\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPointer(t *testing.T) {
	_, _, cSrc := compile(t, "சுட்டி.uli")
	got := runC(t, cSrc)
	want := "7\n9\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNil(t *testing.T) {
	_, _, cSrc := compile(t, "இன்மை.uli")
	got := runC(t, cSrc)
	want := "1\n5\n0\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRichStruct(t *testing.T) {
	_, _, cSrc := compile(t, "வளர்_அமைப்பு.uli")
	got := runC(t, cSrc)
	want := "1\n4\n20\n10\n3\n8\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSwitch(t *testing.T) {
	_, _, cSrc := compile(t, "திசைவி.uli")
	got := runC(t, cSrc)
	want := "20\n2\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunIfSwitchInit(t *testing.T) {
	_, _, cSrc := compile(t, "தொடக்கம்_நிபந்தனை.uli")
	got := runC(t, cSrc)
	want := "3\n20\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunStructEq(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு_சமம்.uli")
	got := runC(t, cSrc)
	want := "1\n2\n3\n4\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPrintStruct(t *testing.T) {
	_, _, cSrc := compile(t, "பதிப்பி_அமைப்பு.uli")
	got := runC(t, cSrc)
	want := "புள்ளி{x: 3, y: 4}\nசெவ்வகம்{அ: புள்ளி{x: 1, y: 2}, ஆ: புள்ளி{x: 5, y: 6}}\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMethod(t *testing.T) {
	_, _, cSrc := compile(t, "முறை.uli")
	got := runC(t, cSrc)
	want := "7\n10\n5\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSliceGrow(t *testing.T) {
	_, _, cSrc := compile(t, "சேர்_வளர்.uli")
	got := runC(t, cSrc)
	want := "6\n6\n2\n2\n3\n9\n5\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPositionalStruct(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு_நிலை.uli")
	got := runC(t, cSrc)
	want := "3\n4\n1\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNestedSlice(t *testing.T) {
	_, _, cSrc := compile(t, "பலபட்டியல்.uli")
	got := runC(t, cSrc)
	want := "2\n3\n2\n5\n3\n9\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMakeCopy(t *testing.T) {
	_, _, cSrc := compile(t, "ஆக்கு.uli")
	got := runC(t, cSrc)
	want := "3\n8\n0\n7\n3\n10\n30\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMultiReturn(t *testing.T) {
	_, _, cSrc := compile(t, "பலமதிப்பு.uli")
	got := runC(t, cSrc)
	want := "3\n2\n6\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunArray(t *testing.T) {
	_, _, cSrc := compile(t, "நிலையான.uli")
	got := runC(t, cSrc)
	want := "3\n3\n20\n99\n2\n99\n139\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSliceStruct(t *testing.T) {
	_, _, cSrc := compile(t, "பட்டியல்_அமைப்பு.uli")
	got := runC(t, cSrc)
	want := "2\n1\n4\n3\n5\n2\n10\n10\n3\n1\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNamedResults(t *testing.T) {
	_, _, cSrc := compile(t, "பெயர்_முடிவு.uli")
	got := runC(t, cSrc)
	want := "3\n2\n42\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunParallelAssign(t *testing.T) {
	_, _, cSrc := compile(t, "பரிமாற்றம்.uli")
	got := runC(t, cSrc)
	want := "2\n1\n30\n10\n20\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunTypeAlias(t *testing.T) {
	_, _, cSrc := compile(t, "வகை_மாற்று.uli")
	got := runC(t, cSrc)
	want := "7\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunFloat(t *testing.T) {
	_, _, cSrc := compile(t, "மிதவை.uli")
	got := runC(t, cSrc)
	want := "4\n3\n1\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunTamilDigits(t *testing.T) {
	_, _, cSrc := compile(t, "தமிழ்_இலக்கம்.uli")
	got := runC(t, cSrc)
	want := "12\n3.5\n15\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunConversion(t *testing.T) {
	_, _, cSrc := compile(t, "மாற்றுதல்.uli")
	got := runC(t, cSrc)
	want := "7\n3\n7\n11\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunConversionExtra(t *testing.T) {
	_, _, cSrc := compile(t, "மாற்றுதல்_மேல்.uli")
	got := runC(t, cSrc)
	want := "மெய்\nA\n9\n5.5\n2.5\n1\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMap(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி.uli")
	got := runC(t, cSrc)
	want := "2\n7\n0\nபொய்\n7\nமெய்\n1\nஆ\n3\nஒன்று\n1\n0\nnil\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMapLiteral(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி_எழுத்து.uli")
	got := runC(t, cSrc)
	want := "2\n7\n3\n0\nபொய்\n0\nok\nஒன்று\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunDefer(t *testing.T) {
	_, _, cSrc := compile(t, "தள்ளிவை.uli")
	got := runC(t, cSrc)
	// start; 9; then LIFO: bare முடி, captured 1, அ, ஆ, முடிவு
	want := "தொடக்கம்\n9\nவெற்று\n1\nஅ\nஆ\nமுடிவு\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunDeferMethod(t *testing.T) {
	_, _, cSrc := compile(t, "தள்ளிவை_முறை.uli")
	got := runC(t, cSrc)
	// immediate 9, 7; LIFO: இ→8, ஆ→7, அ value copy→5
	want := "9\n7\n8\n7\n5\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRichMapKeys(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி_விசை_வளர்.uli")
	got := runC(t, cSrc)
	want := "அரை\n2\n7\n0\nஆ\nசுட்டி\nமெய்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunByteRune(t *testing.T) {
	_, _, cSrc := compile(t, "இருமி.uli")
	got := runC(t, cSrc)
	want := "3\n65\nABC\n3\nABC\n2\n2949\nஅஆ\nA\nB\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunGenerics(t *testing.T) {
	_, _, cSrc := compile(t, "பொதுவகை.uli")
	got := runC(t, cSrc)
	want := "7\nவணக்கம்\n42\nஅ\n99\n9\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPrime(t *testing.T) {
	_, _, cSrc := compile(t, "பகாஎண்.uli")
	got := runC(t, cSrc)
	want := "மெய்\nபொய்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMethodValue(t *testing.T) {
	_, _, cSrc := compile(t, "முறை_மதிப்பு.uli")
	got := runC(t, cSrc)
	want := "10\n14\n7\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMethodExpr(t *testing.T) {
	_, _, cSrc := compile(t, "முறை_வெளிப்பாடு.uli")
	got := runC(t, cSrc)
	want := "7\n10\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunClosure(t *testing.T) {
	_, _, cSrc := compile(t, "மூடுகை.uli")
	got := runC(t, cSrc)
	want := "2\n12\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunLoopClosure(t *testing.T) {
	_, _, cSrc := compile(t, "சுழல்_மூடுகை.uli")
	got := runC(t, cSrc)
	want := "0\n1\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

// Emit-only: structural checks without executing (kept for fast CI).
func TestEmitSelectAndGo(t *testing.T) {
	for _, name := range []string{"இழை.uli", "தடத்தேர்வு.uli"} {
		t.Run(name, func(t *testing.T) {
			_, _, cSrc := compile(t, name)
			if !strings.Contains(cSrc, "uli_chan_make") {
				t.Fatalf("%s: missing uli_chan_make", name)
			}
			if name == "இழை.uli" && !strings.Contains(cSrc, "uli_go(") {
				t.Fatalf("இழை: missing uli_go")
			}
			if name == "தடத்தேர்வு.uli" {
				if !strings.Contains(cSrc, "uli_select(") {
					t.Fatalf("தடத்தேர்வு: missing uli_select")
				}
				if !strings.Contains(cSrc, "uli_sel_kick") {
					t.Fatalf("தடத்தேர்வு: missing uli_sel_kick (cond wake)")
				}
				// has_def=1 → default sentinel is nComm (not mixed clause index)
				if !strings.Contains(cSrc, "uli_select(__sc_0, 1, 1)") &&
					!strings.Contains(cSrc, "uli_select(__sc_1, 1, 1)") {
					// ids may be 0,1 depending on goID; accept any __sc_N with has_def 1
					if !strings.Contains(cSrc, ", 1, 1)") {
						t.Fatalf("தடத்தேர்வு: expected uli_select(..., nComm=1, has_def=1)\n%s", cSrc)
					}
				}
			}
		})
	}
}

func TestEmitAndRunFunctionAggregateFields(t *testing.T) {
	_, _, cSrc := compile(t, "செயல்பாடு_புலம்.uli")
	got := runC(t, cSrc)
	want := "5\n10\n15\nபெட்டிகள்{கணக்குகள்: [<செயல்பாடு>, <செயல்பாடு>]}\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunGo(t *testing.T) {
	_, _, cSrc := compile(t, "இழை.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "வணக்கம்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSelect(t *testing.T) {
	_, _, cSrc := compile(t, "தடத்தேர்வு.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "7\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPanic(t *testing.T) {
	_, _, cSrc := compile(t, "அலறு.uli")
	if !strings.Contains(cSrc, "uli_panic(") {
		t.Fatalf("அலறு: missing uli_panic\n%s", cSrc)
	}
	if !strings.Contains(cSrc, "uli_recover") {
		t.Fatalf("அலறு: missing uli_recover\n%s", cSrc)
	}
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "பிழை\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPanicInt(t *testing.T) {
	_, _, cSrc := compile(t, "அலறு_எண்.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "42\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPanicGo(t *testing.T) {
	_, _, cSrc := compile(t, "அலறு_இழை.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "இழை\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitUncaughtPanicTrace(t *testing.T) {
	src := `தொகுப்பு தொடக்கம்
செயல்பாடு உள்() {
    அலறு("x")
}
செயல்பாடு தொடக்கம்() {
    உள்()
}
`
	file, perrs := parse.ParseFile("அலறு_தடம்.uli", src)
	if len(perrs) != 0 {
		t.Fatal(perrs)
	}
	info, errs := check.File(file)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.Emit(file, info)
	if err != nil {
		t.Fatal(err)
	}
	out := runCExpectFail(t, cSrc, 5*time.Second)
	if !strings.Contains(out, "அலறு: x") {
		t.Fatalf("missing panic message: %q\nC:\n%s", out, cSrc)
	}
	if !strings.Contains(out, "உள்") || !strings.Contains(out, "தொடக்கம்") {
		t.Fatalf("missing stack frames: %q\nC:\n%s", out, cSrc)
	}
}

func TestEmitAndRunSelectUnbufferedDefault(t *testing.T) {
	_, _, cSrc := compile(t, "தடத்தேர்வு_காலி.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "1\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSelectUnbufferedRendezvous(t *testing.T) {
	_, _, cSrc := compile(t, "தடத்தேர்வு_சந்திப்பு.uli")
	got := runCTimeout(t, cSrc, 5*time.Second)
	want := "7\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunHeap(t *testing.T) {
	_, _, cSrc := compile(t, "குவியல்.uli")
	if !strings.Contains(cSrc, "uli_alloc") {
		t.Fatalf("குவியல்: missing uli_alloc\n%s", cSrc)
	}
	if !strings.Contains(cSrc, "uli_gc_poll") {
		t.Fatalf("குவியல்: missing uli_gc_poll\n%s", cSrc)
	}
	if !strings.Contains(cSrc, "uli_heap_init") {
		t.Fatalf("குவியல்: missing uli_heap_init\n%s", cSrc)
	}
	got := runCTimeout(t, cSrc, 15*time.Second)
	want := "2400\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestGCRuntimeReclaim(t *testing.T) {
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	dir := t.TempDir()
	srcPath := filepath.Join(root(t), "internal", "emitc", "gc_runtime.inc")
	runtime, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	driver := string(runtime) + `
int main(void) {
	uli_heap_init();
	volatile unsigned char *root = (unsigned char *)uli_alloc(64);
	root[0] = 77;
	size_t base = uli_heap_live();
	for (int i = 0; i < 100; i++) (void)uli_alloc(64);
	size_t mid = uli_heap_live();
	uli_gc();
	size_t after = uli_heap_live();
	if (mid <= base) return 2;
	if (after >= mid) return 3;
	if (after > base + 256) return 4;
	if (root[0] != 77) return 5;
	unsigned char *large = (unsigned char *)uli_alloc(1024 * 1024);
	large[0] = 1;
	large[1024 * 1024 - 1] = 2;
	if (large[0] != 1 || large[1024 * 1024 - 1] != 2) return 6;
	uli_heap_shutdown();
	return 0;
}
`
	cFile := filepath.Join(dir, "gc_test.c")
	bin := filepath.Join(dir, "gc_test")
	if err := os.WriteFile(cFile, []byte(driver), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O2", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s", err, out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, bin).CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("gc runtime test timed out")
	}
	if err != nil {
		t.Fatalf("gc reclaim failed: %v\n%s", err, out)
	}
}
