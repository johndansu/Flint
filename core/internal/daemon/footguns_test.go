package daemon

import (
	"regexp"
	"testing"
)

// injectFootguns replaces loadedFootguns for the duration of a test.
func injectFootguns(t *testing.T, fgs []compiledFootgun) {
	t.Helper()
	orig := loadedFootguns
	loadedFootguns = fgs
	t.Cleanup(func() { loadedFootguns = orig })
}

func makeRegexFootgun(id, lang, pattern, severity string) compiledFootgun {
	return compiledFootgun{
		entry: footgunEntry{
			ID:       id,
			Severity: severity,
			Title:    id,
			DetectPattern: detectPattern{
				Type:     "regex",
				Language: lang,
				Pattern:  pattern,
			},
		},
		re: regexp.MustCompile(pattern),
	}
}

func makeTextFootgun(id, lang, keyword, severity string) compiledFootgun {
	return compiledFootgun{
		entry: footgunEntry{
			ID:       id,
			Severity: severity,
			Title:    id,
			DetectPattern: detectPattern{
				Type:     "text_search",
				Language: lang,
				Keywords: keyword,
			},
		},
	}
}

// ---------------------------------------------------------------------------
// Regex detection
// ---------------------------------------------------------------------------

func TestCheckFootguns_RegexHit(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("log-password", "javascript", `console\.log.*password`, "high"),
	})

	hits := CheckFootguns("app.js", `console.log("password:", secret)`)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].Entry.ID != "log-password" {
		t.Errorf("unexpected hit ID: %s", hits[0].Entry.ID)
	}
	if hits[0].LineNumber != 1 {
		t.Errorf("expected line 1, got %d", hits[0].LineNumber)
	}
}

func TestCheckFootguns_RegexNoHit(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("log-password", "javascript", `console\.log.*password`, "high"),
	})

	hits := CheckFootguns("app.js", `console.log("user:", username)`)
	if len(hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}

func TestCheckFootguns_RegexMultiLine(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("todo", "go", `TODO`, "low"),
	})

	content := "func main() {\n\t// TODO fix this\n\tfoo()\n}"
	hits := CheckFootguns("main.go", content)
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	if hits[0].LineNumber != 2 {
		t.Errorf("expected match on line 2, got %d", hits[0].LineNumber)
	}
}

// ---------------------------------------------------------------------------
// Text search detection
// ---------------------------------------------------------------------------

func TestCheckFootguns_TextSearchHit(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeTextFootgun("hardcoded-secret", "python", "SECRET_KEY = \"", "critical"),
	})

	hits := CheckFootguns("settings.py", `SECRET_KEY = "abc123"`)
	if len(hits) != 1 {
		t.Fatalf("expected 1 text_search hit, got %d", len(hits))
	}
	if hits[0].Entry.Severity != "critical" {
		t.Errorf("expected critical severity, got %s", hits[0].Entry.Severity)
	}
}

// ---------------------------------------------------------------------------
// Language filtering
// ---------------------------------------------------------------------------

func TestCheckFootguns_LanguageFilter(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("go-only", "go", `fmt\.Println`, "low"),
	})

	// Should hit on .go file
	if hits := CheckFootguns("main.go", "fmt.Println(x)"); len(hits) != 1 {
		t.Errorf("expected hit on .go file, got %d", len(hits))
	}

	// Should not hit on .py file
	if hits := CheckFootguns("main.py", "fmt.Println(x)"); len(hits) != 0 {
		t.Errorf("expected 0 hits on .py file, got %d", len(hits))
	}
}

func TestCheckFootguns_NoLanguageRestriction(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("any-lang", "", `FIXME`, "low"),
	})

	for _, file := range []string{"main.go", "app.py", "index.ts", "Makefile"} {
		if hits := CheckFootguns(file, "FIXME: address this"); len(hits) != 1 {
			t.Errorf("expected hit on %s (no language restriction), got %d", file, len(hits))
		}
	}
}

func TestCheckFootguns_DockerfileDetection(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("root-user", "dockerfile", `USER root`, "high"),
	})

	// Dockerfile has no extension — matched by filename
	if hits := CheckFootguns("Dockerfile", "USER root"); len(hits) != 1 {
		t.Errorf("expected hit on Dockerfile, got %d", len(hits))
	}
	if hits := CheckFootguns("Dockerfile.prod", "USER root"); len(hits) != 1 {
		t.Errorf("expected hit on Dockerfile.prod, got %d", len(hits))
	}
	// Should not match a .go file
	if hits := CheckFootguns("main.go", "USER root"); len(hits) != 0 {
		t.Errorf("expected 0 hits on .go file, got %d", len(hits))
	}
}

// ---------------------------------------------------------------------------
// Max-3-hits cap
// ---------------------------------------------------------------------------

func TestCheckFootguns_MaxThreeHits(t *testing.T) {
	fgs := make([]compiledFootgun, 10)
	for i := range fgs {
		fgs[i] = makeRegexFootgun("fg-"+string(rune('a'+i)), "", `MATCH`, "low")
	}
	injectFootguns(t, fgs)

	hits := CheckFootguns("file.txt", "MATCH")
	if len(hits) > 3 {
		t.Errorf("CheckFootguns should return at most 3 hits, got %d", len(hits))
	}
}

// ---------------------------------------------------------------------------
// One hit per footgun per file
// ---------------------------------------------------------------------------

func TestCheckFootguns_OneHitPerFootgun(t *testing.T) {
	injectFootguns(t, []compiledFootgun{
		makeRegexFootgun("dup", "", `MATCH`, "low"),
	})

	// Content has the pattern on multiple lines — should only produce 1 hit
	content := "MATCH line 1\nMATCH line 2\nMATCH line 3"
	hits := CheckFootguns("file.txt", content)
	if len(hits) != 1 {
		t.Errorf("expected 1 hit per footgun per file, got %d", len(hits))
	}
}

// ---------------------------------------------------------------------------
// languageMatches
// ---------------------------------------------------------------------------

func TestLanguageMatches(t *testing.T) {
	cases := []struct {
		lang, ext, base string
		want            bool
	}{
		{"", "go", "main.go", true},           // no restriction
		{"go", "go", "main.go", true},
		{"go", "py", "main.py", false},
		{"javascript", "js", "app.js", true},
		{"javascript", "ts", "app.ts", true},
		{"javascript", "tsx", "app.tsx", true},
		{"javascript", "go", "main.go", false},
		{"typescript", "ts", "app.ts", true},
		{"typescript", "js", "app.js", false},
		{"python", "py", "app.py", true},
		{"python", "go", "main.go", false},
		{"dockerfile", "", "dockerfile", true},
		{"dockerfile", "", "dockerfile.prod", true},
		{"dockerfile", "go", "main.go", false},
		{"terraform", "tf", "main.tf", true},
		{"terraform", "hcl", "vars.hcl", true},
		{"terraform", "go", "main.go", false},
		{"rust", "rs", "main.rs", true},
		{"rust", "go", "main.go", false},
	}

	for _, c := range cases {
		got := languageMatches(c.lang, c.ext, c.base)
		if got != c.want {
			t.Errorf("languageMatches(%q, %q, %q) = %v, want %v",
				c.lang, c.ext, c.base, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// severityToConfidence
// ---------------------------------------------------------------------------

func TestSeverityToConfidence(t *testing.T) {
	cases := []struct{ sev string; want float64 }{
		{"critical", 0.95},
		{"high", 0.85},
		{"medium", 0.72},
		{"low", 0.60},
		{"unknown", 0.60},
	}
	for _, c := range cases {
		got := severityToConfidence(c.sev)
		if got != c.want {
			t.Errorf("severityToConfidence(%q) = %.2f, want %.2f", c.sev, got, c.want)
		}
	}
}
