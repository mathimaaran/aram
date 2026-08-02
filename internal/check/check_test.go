package check_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"aram/internal/check"
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

func TestCheckValidCorpus(t *testing.T) {
	names := []string{"வணக்கம்.aram", "எண்கணிதம்.aram", "நிபந்தனை.aram", "கூட்டு.aram", "பதிப்பு.aram", "சுழல்.aram", "சரம்.aram", "பட்டியல்.aram", "பலபட்டியல்.aram", "பட்டியல்_அமைப்பு.aram", "ஒவ்வொரு.aram", "சேர்.aram", "சேர்_வளர்.aram", "ஆக்கு.aram", "அகராதி.aram", "அகராதி_எழுத்து.aram", "அகராதி_விசை_வளர்.aram", "தள்ளிவை.aram", "தள்ளிவை_முறை.aram", "இருமி.aram", "பொதுவகை.aram", "பகாஎண்.aram", "பலமதிப்பு.aram", "பெயர்_முடிவு.aram", "பரிமாற்றம்.aram", "வகை_மாற்று.aram", "மிதவை.aram", "தமிழ்_இலக்கம்.aram", "மாற்றுதல்.aram", "மாற்றுதல்_மேல்.aram", "நிலையான.aram", "அமைப்பு.aram", "அமைப்பு_நிலை.aram", "சுட்டி.aram", "முறை.aram", "இன்மை.aram", "வளர்_அமைப்பு.aram", "திசைவி.aram", "தொடக்கம்_நிபந்தனை.aram", "அமைப்பு_சமம்.aram", "பதிப்பி_அமைப்பு.aram"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root(t), "corpus", "tamil", name)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			file, perrs := parse.ParseFile(path, string(src))
			if len(perrs) != 0 {
				t.Fatalf("parse: %v", perrs)
			}
			if _, errs := check.File(file); len(errs) != 0 {
				t.Fatalf("check: %v", errs)
			}
		})
	}
}

func TestCheckUndeclaredAssign(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அறிவிக்கப்படாத_ஒதுக்கீடு.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected type errors")
	}
}

func TestCheckNilShortDecl(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "இன்மை_குறுக்கு.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected type errors for p := இன்மை")
	}
}

func TestCheckValueRecursiveStruct(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "மதிப்பு_சுழல்.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected recursive value type error")
	}
}

func TestCheckSliceStructNotComparable(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "துண்டு_சமம்.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected non-comparable struct error")
	}
}

func TestCheckThreeIndexString(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "மூன்று_துண்டு_சரம்.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected three-index on சரம் error")
	}
}

func TestCheckMixedStructLit(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அமைப்பு_கலப்பு.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected mixed keyed/unkeyed struct literal error")
	}
}

func TestCheckBadMapKey(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அகராதி_விசை.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected non-comparable map key error")
	}
}

func TestCheckMapLitMissingKey(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அகராதி_எழுத்து_விசை.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	file, perrs := parse.ParseFile(path, string(src))
	if len(perrs) != 0 {
		t.Fatalf("parse: %v", perrs)
	}
	_, errs := check.File(file)
	if len(errs) == 0 {
		t.Fatal("expected missing key in map literal error")
	}
}
