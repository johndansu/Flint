package daemon

import (
	"testing"
)

func TestIsPoorMessage(t *testing.T) {
	poor := []string{
		"fix",
		"wip",
		"update",
		"done",
		"ok",
		"tmp",
		"temp",
		"misc",
		"fixes",   // plural form
		"updated", // past tense
		"fixing",  // gerund
		"ab",      // under 8 chars
		"x",
	}
	for _, msg := range poor {
		if !isPoorMessage(msg) {
			t.Errorf("isPoorMessage(%q) = false, want true", msg)
		}
	}
}

func TestIsPoorMessage_GoodMessages(t *testing.T) {
	good := []string{
		"add user authentication via JWT",
		"fix: null pointer in payment handler",
		"refactor session classifier to use EMA",
		"bump deps to address CVE-2024-1234",
		"remove unused imports in api package",
	}
	for _, msg := range good {
		if isPoorMessage(msg) {
			t.Errorf("isPoorMessage(%q) = true, want false", msg)
		}
	}
}

func TestIsPoorMessage_EdgeCases(t *testing.T) {
	// Exactly 8 chars — border is < 8, so 8 is fine if not a known word
	if isPoorMessage("refactor") {
		t.Error("'refactor' is 8 chars and not a known poor word — should pass")
	}
	// 7 chars poor word
	if !isPoorMessage("cleanup") {
		t.Error("'cleanup' is a known poor word — should fail")
	}
}
