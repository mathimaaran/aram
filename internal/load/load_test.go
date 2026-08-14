package load_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"aram/internal/check"
	"aram/internal/emitc"
	"aram/internal/load"
)

func root(t *testing.T) string {
	t.Helper()
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Join(filepath.Dir(f), "..", "..")
}

func TestExpandDir(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "பல்கோப்பு")
	paths, err := load.Expand([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("got %d paths: %v", len(paths), paths)
	}
}

func TestMultiFileProgram(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "பல்கோப்பு")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	want := "7\n10\n"
	if string(got) != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestImportProgram(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "கொணர்_எடு")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	if len(prog.All) != 2 {
		t.Fatalf("packages: %d", len(prog.All))
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	want := "7\n5\nவணக்கம்\n8\n"
	if string(got) != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestUnexportedImportRejected(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "invalid", "தனிய_கொணர்")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	_, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) == 0 {
		t.Fatal("expected unexported name error")
	}
}

func TestUnusedImportRejected(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "invalid", "பயன்படா_கொணர்")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	_, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) == 0 {
		t.Fatal("expected unused import error")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "imported and not used") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("errs=%v", errs)
	}
}

func TestImportCycleRejected(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "invalid", "கொணர்_சுழற்சி")
	_, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) == 0 {
		t.Fatal("expected import cycle error")
	}
	found := false
	for _, e := range lerrs {
		if strings.Contains(e.Error(), "import cycle") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("errs=%v", lerrs)
	}
}

func TestImportAlias(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "கொணர்_மாற்று")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	want := "7\n"
	if string(got) != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestNetEchoProgram(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "வலை_எதிரொலி")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	if len(prog.All) != 2 {
		t.Fatalf("packages: %d", len(prog.All))
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cSrc, "aram_net_listen") {
		t.Fatalf("missing aram_net_listen\n%s", cSrc)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s\nC:\n%s", err, got, cSrc)
	}
	want := "அறம்\n"
	if string(got) != want {
		t.Fatalf("got %q want %q\nC:\n%s", got, want, cSrc)
	}
}

func TestHttpServeOne(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "வலைபரிமாற்றம்")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cSrc, "aram_http_mux_serve_one") {
		t.Fatalf("missing HTTP mux serve runtime\n%s", cSrc)
	}
	if !strings.Contains(cSrc, "aram_http_headers_set") {
		t.Fatalf("missing request-header map bridge\n%s", cSrc)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s\nC:\n%s", err, got, cSrc)
	}
	out := string(got)
	if !strings.Contains(out, "HTTP/1.0 200") {
		t.Fatalf("missing status line: %q\nC:\n%s", out, cSrc)
	}
	if !strings.Contains(out, "வணக்கம்") {
		t.Fatalf("missing body: %q\nC:\n%s", out, cSrc)
	}
	if !strings.Contains(out, "HTTP/1.0 404") || !strings.Contains(out, "காணவில்லை") {
		t.Fatalf("missing mux 404: %q\nC:\n%s", out, cSrc)
	}
	if strings.Contains(out, "பழையது") {
		t.Fatalf("duplicate route did not replace handler: %q\nC:\n%s", out, cSrc)
	}
}

func TestHttpHTML(t *testing.T) {
	dir := filepath.Join(root(t), "corpus", "tamil", "வலைபரிமாற்றம்_html")
	prog, lerrs := load.LoadProgram([]string{dir})
	if len(lerrs) != 0 {
		t.Fatal(lerrs)
	}
	merged, merrs := prog.MergedFiles()
	if len(merrs) != 0 {
		t.Fatal(merrs)
	}
	pi, errs := check.CheckProgram(merged, prog.Entry.Name)
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	cSrc, err := emitc.EmitProgram(pi)
	if err != nil {
		t.Fatal(err)
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		cc, err = exec.LookPath("cc")
		if err != nil {
			t.Skip("gcc/cc not available")
		}
	}
	tmp := t.TempDir()
	cFile := filepath.Join(tmp, "out.c")
	bin := filepath.Join(tmp, "out")
	if err := os.WriteFile(cFile, []byte(cSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=c11", "-O0", "-pthread", cFile, "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("cc: %v\n%s\n%s", err, out, cSrc)
	}
	got, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s\nC:\n%s", err, got, cSrc)
	}
	out := string(got)
	if !strings.Contains(out, "text/html") {
		t.Fatalf("missing content-type: %q", out)
	}
	if !strings.Contains(out, "<h1>அறம்</h1>") {
		t.Fatalf("missing html body: %q", out)
	}
}
