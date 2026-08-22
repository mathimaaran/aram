package parse_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"niraluli/internal/ast"
	"niraluli/internal/parse"
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
			"வணக்கம்.uli",
			"எண்கணிதம்.uli",
			"நிபந்தனை.uli",
			"கூட்டு.uli",
			"பதிப்பு.uli",
			"சுழல்.uli",
			"சரம்.uli",
			"பட்டியல்.uli",
			"பலபட்டியல்.uli",
			"ஒவ்வொரு.uli",
			"சேர்.uli",
			"சேர்_வளர்.uli",
			"ஆக்கு.uli",
			"அகராதி.uli",
			"அகராதி_எழுத்து.uli",
			"தள்ளிவை.uli",
			"தள்ளிவை_முறை.uli",
			"அகராதி_விசை_வளர்.uli",
			"இருமி.uli",
			"பொதுவகை.uli",
			"பகாஎண்.uli",
			"பலமதிப்பு.uli",
			"நிலையான.uli",
			"அமைப்பு.uli",
			"அமைப்பு_நிலை.uli",
			"சுட்டி.uli",
			"முறை.uli",
			"முறை_மதிப்பு.uli",
			"முறை_வெளிப்பாடு.uli",
			"மூடுகை.uli",
			"சுழல்_மூடுகை.uli",
			"இழை.uli",
			"தடத்தேர்வு.uli",
			"தடத்தேர்வு_காலி.uli",
			"தடத்தேர்வு_சந்திப்பு.uli",
			"அலறு.uli",
			"அலறு_இழை.uli",
			"அலறு_எண்.uli",
			"ஒவ்வொரு_தடம்.uli",
			"குவியல்.uli",
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
	f := mustParse(t, corpus(t, "கூட்டு.uli"))
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
	path := corpus(t, "invalid", "தொகுப்பு_இல்லை.uli")
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
	path := corpus(t, "invalid", "பழைய_else.uli")
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
	path := corpus(t, "invalid", "முழுமையற்ற_ஒதுக்கீடு.uli")
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
	f := mustParse(t, corpus(t, "தடத்தேர்வு.uli"))
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
		"தடத்தேர்வு_பிரிவு.uli",
		"தடத்தேர்வு_முழுமையற்ற.uli",
		"தடத்தேர்வு_வெற்று.uli",
		"திசைவி_பிரிவு.uli",
		"திசைவி_முழுமையற்ற.uli",
		"இழை_அழைப்பு_இல்லை.uli",
		"அமைப்பு_பிரிவு.uli",
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
			errs := parseWithTimeout(t, tc.name+".uli", tc.src)
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
