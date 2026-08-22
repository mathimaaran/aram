package check_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"niraluli/internal/check"
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

func TestCheckValidCorpus(t *testing.T) {
	names := []string{"வணக்கம்.uli", "எண்கணிதம்.uli", "நிபந்தனை.uli", "கூட்டு.uli", "பதிப்பு.uli", "சுழல்.uli", "சரம்.uli", "பட்டியல்.uli", "பலபட்டியல்.uli", "பட்டியல்_அமைப்பு.uli", "ஒவ்வொரு.uli", "சேர்.uli", "சேர்_வளர்.uli", "ஆக்கு.uli", "அகராதி.uli", "அகராதி_எழுத்து.uli", "அகராதி_விசை_வளர்.uli", "தள்ளிவை.uli", "தள்ளிவை_முறை.uli", "இருமி.uli", "பொதுவகை.uli", "பகாஎண்.uli", "பலமதிப்பு.uli", "பெயர்_முடிவு.uli", "பரிமாற்றம்.uli", "வகை_மாற்று.uli", "மிதவை.uli", "தமிழ்_இலக்கம்.uli", "மாற்றுதல்.uli", "மாற்றுதல்_மேல்.uli", "நிலையான.uli", "அமைப்பு.uli", "அமைப்பு_நிலை.uli", "சுட்டி.uli", "முறை.uli", "முறை_மதிப்பு.uli", "முறை_வெளிப்பாடு.uli", "மூடுகை.uli", "சுழல்_மூடுகை.uli", "இன்மை.uli", "வளர்_அமைப்பு.uli", "திசைவி.uli", "தொடக்கம்_நிபந்தனை.uli", "அமைப்பு_சமம்.uli", "பதிப்பி_அமைப்பு.uli", "செயல்பாடு_புலம்.uli", "இழை.uli", "தடத்தேர்வு.uli", "தடத்தேர்வு_காலி.uli", "தடத்தேர்வு_சந்திப்பு.uli", "அலறு.uli", "அலறு_இழை.uli", "அலறு_எண்.uli", "ஒவ்வொரு_தடம்.uli", "குவியல்.uli"}
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அறிவிக்கப்படாத_ஒதுக்கீடு.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "இன்மை_குறுக்கு.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "மதிப்பு_சுழல்.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "துண்டு_சமம்.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "மூன்று_துண்டு_சரம்.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அமைப்பு_கலப்பு.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அகராதி_விசை.uli")
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
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "அகராதி_எழுத்து_விசை.uli")
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

func TestCheckMethodValueNonAddressable(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "முறை_மதிப்பு_சுட்டி.uli")
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
		t.Fatal("expected non-addressable pointer method value error")
	}
}

func TestCheckMethodExprPointerOnValue(t *testing.T) {
	path := filepath.Join(root(t), "corpus", "tamil", "invalid", "முறை_வெளிப்பாடு_சுட்டி.uli")
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
		t.Fatal("expected pointer method on T method-expression error")
	}
}

func TestCheckInvalidConcurrency(t *testing.T) {
	names := []string{
		"தடம்_அனுப்பு.uli",
		"தடத்தேர்வு_பெறு_அல்ல.uli",
		"ஒவ்வொரு_தடம்_இரு.uli",
		"ஒவ்வொரு_தடம்_அனுப்பு.uli",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root(t), "corpus", "tamil", "invalid", name)
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
				t.Fatal("expected typecheck errors")
			}
		})
	}
}

func TestCheckInvalidPanic(t *testing.T) {
	names := []string{
		"அலறு_தருமதிப்பு.uli",
		"அலறு_பட்டியல்.uli",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root(t), "corpus", "tamil", "invalid", name)
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
				t.Fatal("expected typecheck errors")
			}
		})
	}
}
