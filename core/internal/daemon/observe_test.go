package daemon

import (
	"testing"
)

// ---------------------------------------------------------------------------
// extractJSON
// ---------------------------------------------------------------------------

func TestExtractJSON_Clean(t *testing.T) {
	in := `{"worth_saying":true,"kind":"technical"}`
	got := extractJSON(in)
	if got != in {
		t.Errorf("clean JSON: want %q, got %q", in, got)
	}
}

func TestExtractJSON_ProseAround(t *testing.T) {
	in := `Here is my response: {"worth_saying":false} Hope that helps!`
	want := `{"worth_saying":false}`
	got := extractJSON(in)
	if got != want {
		t.Errorf("prose around: want %q, got %q", want, got)
	}
}

func TestExtractJSON_MarkdownFence(t *testing.T) {
	in := "```json\n{\"worth_saying\":true,\"kind\":\"win\"}\n```"
	want := `{"worth_saying":true,"kind":"win"}`
	got := extractJSON(in)
	if got != want {
		t.Errorf("markdown fence: want %q, got %q", want, got)
	}
}

func TestExtractJSON_Nested(t *testing.T) {
	in := `Some text {"outer":{"inner":42},"arr":[1,2,3]} trailing`
	want := `{"outer":{"inner":42},"arr":[1,2,3]}`
	got := extractJSON(in)
	if got != want {
		t.Errorf("nested: want %q, got %q", want, got)
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	in := "no json here at all"
	got := extractJSON(in)
	// Should return as-is so Unmarshal produces a clear error
	if got != in {
		t.Errorf("no JSON: want original string back, got %q", got)
	}
}

func TestExtractJSON_EmptyString(t *testing.T) {
	got := extractJSON("")
	if got != "" {
		t.Errorf("empty: want empty, got %q", got)
	}
}

func TestExtractJSON_OnlyOpenBrace(t *testing.T) {
	in := "prefix { no closing"
	// last } is -1, so returns as-is
	got := extractJSON(in)
	if got != in {
		t.Errorf("only open brace: want original, got %q", got)
	}
}

func TestExtractJSON_MultipleObjects(t *testing.T) {
	// Should return from first { to LAST } — outermost span
	in := `before {"a":1} middle {"b":2} after`
	want := `{"a":1} middle {"b":2}`
	got := extractJSON(in)
	if got != want {
		t.Errorf("multiple objects: want %q, got %q", want, got)
	}
}

// ---------------------------------------------------------------------------
// highestSeverity
// ---------------------------------------------------------------------------

func makeHit(severity string) FootgunHit {
	return FootgunHit{
		Entry: footgunEntry{Severity: severity, ID: severity},
	}
}

func TestHighestSeverity_Single(t *testing.T) {
	hits := []FootgunHit{makeHit("medium")}
	got := highestSeverity(hits)
	if got.Entry.Severity != "medium" {
		t.Errorf("single: want medium, got %s", got.Entry.Severity)
	}
}

func TestHighestSeverity_CriticalWins(t *testing.T) {
	hits := []FootgunHit{makeHit("low"), makeHit("critical"), makeHit("medium")}
	got := highestSeverity(hits)
	if got.Entry.Severity != "critical" {
		t.Errorf("critical wins: want critical, got %s", got.Entry.Severity)
	}
}

func TestHighestSeverity_HighOverMediumLow(t *testing.T) {
	hits := []FootgunHit{makeHit("medium"), makeHit("low"), makeHit("high")}
	got := highestSeverity(hits)
	if got.Entry.Severity != "high" {
		t.Errorf("high over medium/low: want high, got %s", got.Entry.Severity)
	}
}

func TestHighestSeverity_AllSame(t *testing.T) {
	hits := []FootgunHit{makeHit("low"), makeHit("low"), makeHit("low")}
	got := highestSeverity(hits)
	if got.Entry.Severity != "low" {
		t.Errorf("all same: want low, got %s", got.Entry.Severity)
	}
}

func TestHighestSeverity_FirstWhenTied(t *testing.T) {
	h1 := FootgunHit{Entry: footgunEntry{Severity: "high", ID: "first"}}
	h2 := FootgunHit{Entry: footgunEntry{Severity: "high", ID: "second"}}
	got := highestSeverity([]FootgunHit{h1, h2})
	if got.Entry.ID != "first" {
		t.Errorf("tied: want first hit, got %s", got.Entry.ID)
	}
}
