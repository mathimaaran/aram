package emitc_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"aram/internal/ast"
	"aram/internal/check"
	"aram/internal/emitc"
	"aram/internal/parse"
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
	if out, err := exec.Command(cc, "-std=c11", "-O0", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestEmitAndRunVanakkam(t *testing.T) {
	_, _, cSrc := compile(t, "வணக்கம்.aram")
	got := runC(t, cSrc)
	want := "வணக்கம், அறம்!\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestEmitAndRunAdd(t *testing.T) {
	_, _, cSrc := compile(t, "கூட்டு.aram")
	got := runC(t, cSrc)
	want := "9\n7\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunLoop(t *testing.T) {
	_, _, cSrc := compile(t, "சுழல்.aram")
	got := runC(t, cSrc)
	want := "0\n1\n2\n3\n4\n3\n2\n1\n0\n1\n2\n0\n1\n3\n4\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunString(t *testing.T) {
	_, _, cSrc := compile(t, "சரம்.aram")
	got := runC(t, cSrc)
	want := "வணக்கம், அறம்!\nமெய்\nபொய்\nஆம்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSlice(t *testing.T) {
	_, _, cSrc := compile(t, "பட்டியல்.aram")
	got := runC(t, cSrc)
	want := "3\n20\n99\n2\nஅறம்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRange(t *testing.T) {
	_, _, cSrc := compile(t, "ஒவ்வொரு.aram")
	got := runC(t, cSrc)
	want := "0\n10\n1\n20\n2\n30\n10\n20\n30\n60\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunAppendSliceString(t *testing.T) {
	_, _, cSrc := compile(t, "சேர்.aram")
	got := runC(t, cSrc)
	want := "3\n3\n2\n2\nell\n5\n0\n97\n1\n98\nவணக்கம், அறம்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunStruct(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு.aram")
	got := runC(t, cSrc)
	want := "3\n9\n0\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPointer(t *testing.T) {
	_, _, cSrc := compile(t, "சுட்டி.aram")
	got := runC(t, cSrc)
	want := "7\n9\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNil(t *testing.T) {
	_, _, cSrc := compile(t, "இன்மை.aram")
	got := runC(t, cSrc)
	want := "1\n5\n0\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRichStruct(t *testing.T) {
	_, _, cSrc := compile(t, "வளர்_அமைப்பு.aram")
	got := runC(t, cSrc)
	want := "1\n4\n20\n10\n3\n8\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSwitch(t *testing.T) {
	_, _, cSrc := compile(t, "திசைவி.aram")
	got := runC(t, cSrc)
	want := "20\n2\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunIfSwitchInit(t *testing.T) {
	_, _, cSrc := compile(t, "தொடக்கம்_நிபந்தனை.aram")
	got := runC(t, cSrc)
	want := "3\n20\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunStructEq(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு_சமம்.aram")
	got := runC(t, cSrc)
	want := "1\n2\n3\n4\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPrintStruct(t *testing.T) {
	_, _, cSrc := compile(t, "பதிப்பி_அமைப்பு.aram")
	got := runC(t, cSrc)
	want := "புள்ளி{x: 3, y: 4}\nசெவ்வகம்{அ: புள்ளி{x: 1, y: 2}, ஆ: புள்ளி{x: 5, y: 6}}\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMethod(t *testing.T) {
	_, _, cSrc := compile(t, "முறை.aram")
	got := runC(t, cSrc)
	want := "7\n10\n5\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSliceGrow(t *testing.T) {
	_, _, cSrc := compile(t, "சேர்_வளர்.aram")
	got := runC(t, cSrc)
	want := "6\n6\n2\n2\n3\n9\n5\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPositionalStruct(t *testing.T) {
	_, _, cSrc := compile(t, "அமைப்பு_நிலை.aram")
	got := runC(t, cSrc)
	want := "3\n4\n1\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNestedSlice(t *testing.T) {
	_, _, cSrc := compile(t, "பலபட்டியல்.aram")
	got := runC(t, cSrc)
	want := "2\n3\n2\n5\n3\n9\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMakeCopy(t *testing.T) {
	_, _, cSrc := compile(t, "ஆக்கு.aram")
	got := runC(t, cSrc)
	want := "3\n8\n0\n7\n3\n10\n30\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMultiReturn(t *testing.T) {
	_, _, cSrc := compile(t, "பலமதிப்பு.aram")
	got := runC(t, cSrc)
	want := "3\n2\n6\n2\n2\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunArray(t *testing.T) {
	_, _, cSrc := compile(t, "நிலையான.aram")
	got := runC(t, cSrc)
	want := "3\n3\n20\n99\n2\n99\n139\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunSliceStruct(t *testing.T) {
	_, _, cSrc := compile(t, "பட்டியல்_அமைப்பு.aram")
	got := runC(t, cSrc)
	want := "2\n1\n4\n3\n5\n2\n10\n10\n3\n1\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunNamedResults(t *testing.T) {
	_, _, cSrc := compile(t, "பெயர்_முடிவு.aram")
	got := runC(t, cSrc)
	want := "3\n2\n42\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunParallelAssign(t *testing.T) {
	_, _, cSrc := compile(t, "பரிமாற்றம்.aram")
	got := runC(t, cSrc)
	want := "2\n1\n30\n10\n20\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunTypeAlias(t *testing.T) {
	_, _, cSrc := compile(t, "வகை_மாற்று.aram")
	got := runC(t, cSrc)
	want := "7\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunFloat(t *testing.T) {
	_, _, cSrc := compile(t, "மிதவை.aram")
	got := runC(t, cSrc)
	want := "4\n3\n1\n3\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunTamilDigits(t *testing.T) {
	_, _, cSrc := compile(t, "தமிழ்_இலக்கம்.aram")
	got := runC(t, cSrc)
	want := "12\n3.5\n15\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunConversion(t *testing.T) {
	_, _, cSrc := compile(t, "மாற்றுதல்.aram")
	got := runC(t, cSrc)
	want := "7\n3\n7\n11\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunConversionExtra(t *testing.T) {
	_, _, cSrc := compile(t, "மாற்றுதல்_மேல்.aram")
	got := runC(t, cSrc)
	want := "மெய்\nA\n9\n5.5\n2.5\n1\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMap(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி.aram")
	got := runC(t, cSrc)
	want := "2\n7\n0\nபொய்\n7\nமெய்\n1\nஆ\n3\nஒன்று\n1\n0\nnil\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMapLiteral(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி_எழுத்து.aram")
	got := runC(t, cSrc)
	want := "2\n7\n3\n0\nபொய்\n0\nok\nஒன்று\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunDefer(t *testing.T) {
	_, _, cSrc := compile(t, "தள்ளிவை.aram")
	got := runC(t, cSrc)
	// start; 9; then LIFO: bare முடி, captured 1, அ, ஆ, முடிவு
	want := "தொடக்கம்\n9\nவெற்று\n1\nஅ\nஆ\nமுடிவு\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunDeferMethod(t *testing.T) {
	_, _, cSrc := compile(t, "தள்ளிவை_முறை.aram")
	got := runC(t, cSrc)
	// immediate 9, 7; LIFO: இ→8, ஆ→7, அ value copy→5
	want := "9\n7\n8\n7\n5\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunRichMapKeys(t *testing.T) {
	_, _, cSrc := compile(t, "அகராதி_விசை_வளர்.aram")
	got := runC(t, cSrc)
	want := "அரை\n2\n7\n0\nஆ\nசுட்டி\nமெய்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunByteRune(t *testing.T) {
	_, _, cSrc := compile(t, "இருமி.aram")
	got := runC(t, cSrc)
	want := "3\n65\nABC\n3\nABC\n2\n2949\nஅஆ\nA\nB\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunGenerics(t *testing.T) {
	_, _, cSrc := compile(t, "பொதுவகை.aram")
	got := runC(t, cSrc)
	want := "7\nவணக்கம்\n42\nஅ\n99\n9\n9\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunPrime(t *testing.T) {
	_, _, cSrc := compile(t, "பகாஎண்.aram")
	got := runC(t, cSrc)
	want := "மெய்\nபொய்\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestEmitAndRunMethodValue(t *testing.T) {
	_, _, cSrc := compile(t, "முறை_மதிப்பு.aram")
	got := runC(t, cSrc)
	want := "10\n14\n7\n"
	if got != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}
