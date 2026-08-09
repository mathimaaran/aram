package parse_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"aram/internal/ast"
	"aram/internal/parse"
)

func corpus(t *testing.T, elems ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(append([]string{root, "corpus", "tamil"}, elems...)...)
}

func mustParse(t *testing.T, path string) *ast.File {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f, errs := parse.ParseFile(path, string(src))
	if len(errs) != 0 {
		t.Fatalf("%s: %v", path, errs)
	}
	return f
}

func TestParseValidCorpus(t *testing.T) {
		files := []string{
			"வணக்கம்.aram",
			"எண்கணிதம்.aram",
			"நிபந்தனை.aram",
			"கூட்டு.aram",
			"பதிப்பு.aram",
			"சுழல்.aram",
			"சரம்.aram",
			"பட்டியல்.aram",
			"பலபட்டியல்.aram",
			"ஒவ்வொரு.aram",
			"சேர்.aram",
			"சேர்_வளர்.aram",
			"ஆக்கு.aram",
			"அகராதி.aram",
			"அகராதி_எழுத்து.aram",
			"தள்ளிவை.aram",
			"தள்ளிவை_முறை.aram",
			"அகராதி_விசை_வளர்.aram",
			"இருமி.aram",
			"பொதுவகை.aram",
			"பகாஎண்.aram",
			"பலமதிப்பு.aram",
			"நிலையான.aram",
			"அமைப்பு.aram",
			"அமைப்பு_நிலை.aram",
			"சுட்டி.aram",
			"முறை.aram",
			"முறை_மதிப்பு.aram",
			"முறை_வெளிப்பாடு.aram",
			"மூடுகை.aram",
			"சுழல்_மூடுகை.aram",
			"இழை.aram",
			"தடத்தேர்வு.aram",
		}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			f := mustParse(t, corpus(t, name))
			if f.Package == nil || f.Package.Name.Name != "தொடக்கம்" {
				t.Fatalf("package name: %#v", f.Package)
			}
			if len(f.Decls) == 0 {
				t.Fatal("expected declarations")
			}
		})
	}
}

func TestParseAddFunc(t *testing.T) {
	f := mustParse(t, corpus(t, "கூட்டு.aram"))
	var fn *ast.FuncDecl
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "கூட்டு" {
			fn = fd
			break
		}
	}
	if fn == nil {
		t.Fatal("missing கூட்டு")
	}
	if len(fn.Params) != 2 || len(fn.Results) != 1 {
		t.Fatalf("signature: params=%d results=%v", len(fn.Params), fn.Results)
	}
	if fn.Results[0].Name != nil {
		t.Fatalf("signature: unexpected named result %v", fn.Results[0].Name)
	}
	tn, ok := fn.Results[0].Type.(*ast.TypeName)
	if !ok || tn.Name != "முழுஎண்" {
		t.Fatalf("signature: result=%v", fn.Results[0].Type)
	}
}

func TestParseInvalidMissingPackage(t *testing.T) {
	path := corpus(t, "invalid", "தொகுப்பு_இல்லை.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, errs := parse.ParseFile(path, string(src))
	if len(errs) == 0 {
		t.Fatal("expected parse errors")
	}
}

func TestParseInvalidOldElseStillParses(t *testing.T) {
	// இல்லையெனில் is not the ELSE keyword — it lexes as IDENT.
	// The file is therefore syntactically a bare ident + block after if;
	// rejecting it is a later semantic/style concern, not the parser's job.
	path := corpus(t, "invalid", "பழைய_else.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, errs := parse.ParseFile(path, string(src))
	if len(errs) != 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}
}

func TestParseInvalidIncompleteAssign(t *testing.T) {
	path := corpus(t, "invalid", "முழுமையற்ற_ஒதுக்கீடு.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, errs := parse.ParseFile(path, string(src))
	if len(errs) == 0 {
		t.Fatal("expected parse errors for incomplete assignment")
	}
}

func TestParseSelectCorpus(t *testing.T) {
	f := mustParse(t, corpus(t, "தடத்தேர்வு.aram"))
	if len(f.Decls) == 0 {
		t.Fatal("expected decls")
	}
}

func parseWithTimeout(t *testing.T, path, src string) []error {
	t.Helper()
	done := make(chan []error, 1)
	go func() {
		_, errs := parse.ParseFile(path, src)
		done <- errs
	}()
	select {
	case errs := <-done:
		return errs
	case <-time.After(2 * time.Second):
		t.Fatalf("parse hung: %s", path)
		return nil
	}
}

// Invalid select/switch/struct snippets must finish quickly (error ⇒ progress).
func TestParseMustTerminateInvalid(t *testing.T) {
	names := []string{
		"தடத்தேர்வு_பிரிவு.aram",
		"தடத்தேர்வு_முழுமையற்ற.aram",
		"தடத்தேர்வு_வெற்று.aram",
		"திசைவி_பிரிவு.aram",
		"திசைவி_முழுமையற்ற.aram",
		"இழை_அழைப்பு_இல்லை.aram",
		"அமைப்பு_பிரிவு.aram",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := corpus(t, "invalid", name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			errs := parseWithTimeout(t, path, string(src))
			if len(errs) == 0 {
				t.Fatal("expected parse errors")
			}
		})
	}
}

func TestParseMustTerminateInline(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "select_garbage",
			src: `தொகுப்பு தொடக்கம்
செயல்பாடு தொடக்கம்() {
    தடத்தேர்வு { x }
}
`,
		},
		{
			name: "switch_garbage",
			src: `தொகுப்பு தொடக்கம்
செயல்பாடு தொடக்கம்() {
    திசைவி 1 { x }
}
`,
		},
		{
			name: "select_only_default_ok_shape",
			src: `தொகுப்பு தொடக்கம்
செயல்பாடு தொடக்கம்() {
    தடத்தேர்வு {
        மற்றபடி { பதிப்பி(1) }
    }
}
`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := parseWithTimeout(t, tc.name+".aram", tc.src)
			if tc.name == "select_only_default_ok_shape" {
				if len(errs) != 0 {
					t.Fatalf("unexpected parse errors: %v", errs)
				}
				return
			}
			if len(errs) == 0 {
				t.Fatal("expected parse errors")
			}
		})
	}
}
