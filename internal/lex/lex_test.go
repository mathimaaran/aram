package lex_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"aram/internal/lex"
	"aram/internal/token"
)

func corpusFile(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, "corpus", "tamil", name)
}

func TestLexVanakkam(t *testing.T) {
	path := corpusFile(t, "வணக்கம்.aram")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	l := lex.New(path, string(src))
	toks := l.All()
	if len(l.Errors()) != 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}

	wantKinds := []token.Kind{
		token.PACKAGE, token.IDENT, token.SEMICOLON, // தொகுப்பு தொடக்கம்
		token.FUNC, token.IDENT, token.LPAREN, token.RPAREN, token.LBRACE,
		token.PRINT, token.LPAREN, token.STRING, token.RPAREN, token.SEMICOLON,
		token.RBRACE, token.SEMICOLON,
		token.EOF,
	}
	if len(toks) != len(wantKinds) {
		t.Fatalf("got %d tokens, want %d\n%v", len(toks), len(wantKinds), summarize(toks))
	}
	for i, k := range wantKinds {
		if toks[i].Kind != k {
			t.Fatalf("token %d: got %s %q, want %s", i, toks[i].Kind, toks[i].Lit, k)
		}
	}
	if toks[1].Lit != "தொடக்கம்" {
		t.Fatalf("package name: got %q", toks[1].Lit)
	}
	if toks[4].Lit != "தொடக்கம்" {
		t.Fatalf("func name: got %q", toks[4].Lit)
	}
	if toks[10].Kind != token.STRING || toks[10].Lit != "வணக்கம், அறம்!" {
		t.Fatalf("string lit: got %s %q", toks[10].Kind, toks[10].Lit)
	}
}

func TestLexKeywordsAndDefine(t *testing.T) {
	src := "மாறி x முழுஎண் = 1\ny := 2\nஎனில் மெய் { } இல்லையேல் { }\n"
	l := lex.New("t.aram", src)
	toks := l.All()
	if len(l.Errors()) != 0 {
		t.Fatalf("lex errors: %v", l.Errors())
	}
	kinds := map[token.Kind]bool{}
	for _, tok := range toks {
		kinds[tok.Kind] = true
	}
	for _, k := range []token.Kind{token.VAR, token.TYPE_INT, token.ASSIGN, token.DEFINE, token.IF, token.TRUE, token.ELSE} {
		if !kinds[k] {
			t.Fatalf("missing kind %s in %v", k, summarize(toks))
		}
	}
}

func TestLexIllegalElseKeywordStillIdentifierOrKeyword(t *testing.T) {
	// Rejected keyword இல்லையெனில் is not reserved — lexes as IDENT.
	src := "இல்லையெனில்"
	l := lex.New("t.aram", src)
	tok := l.Next()
	if tok.Kind != token.IDENT || tok.Lit != "இல்லையெனில்" {
		t.Fatalf("got %s %q", tok.Kind, tok.Lit)
	}
}

func summarize(toks []token.Token) string {
	s := ""
	for i, tok := range toks {
		if i > 0 {
			s += " "
		}
		s += tok.Kind.String()
		if tok.Lit != "" && tok.Kind != token.EOF {
			s += "(" + tok.Lit + ")"
		}
	}
	return s
}
