package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// footgunEntry mirrors the shape of shared/footguns.json.
type footgunEntry struct {
	ID            string        `json:"id"`
	Role          string        `json:"role"`
	Framework     string        `json:"framework"`
	Category      string        `json:"category"`
	Severity      string        `json:"severity"`
	Title         string        `json:"title"`
	Observation   string        `json:"observation"`
	AgentPrompt   string        `json:"agent_prompt"`
	FixConfidence string        `json:"fix_confidence"`
	DetectPattern detectPattern `json:"detect_pattern"`
}

type detectPattern struct {
	Type        string `json:"type"`    // "regex" | "text_search" | "ast_pattern" | "heuristic"
	Pattern     string `json:"pattern"` // set when type == "regex"
	Keywords    string `json:"keywords"` // set when type == "text_search"
	Description string `json:"description"`
	Language    string `json:"language"`
}

type footgunFile struct {
	Footguns []footgunEntry `json:"footguns"`
}

// compiled is a footgun with its regex pre-compiled (nil for non-regex types).
type compiledFootgun struct {
	entry footgunEntry
	re    *regexp.Regexp // nil when type != "regex"
}

var loadedFootguns []compiledFootgun

// LoadFootguns reads shared/footguns.json and compiles all regex patterns.
// Must be called once before CheckFootguns.
func LoadFootguns(sharedDir string) error {
	path := filepath.Join(sharedDir, "footguns.json")
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("footguns: open %s: %w", path, err)
	}
	defer f.Close()

	var ff footgunFile
	if err := json.NewDecoder(f).Decode(&ff); err != nil {
		return fmt.Errorf("footguns: parse: %w", err)
	}

	loadedFootguns = make([]compiledFootgun, 0, len(ff.Footguns))
	for _, fg := range ff.Footguns {
		cf := compiledFootgun{entry: fg}
		if fg.DetectPattern.Type == "regex" && fg.DetectPattern.Pattern != "" {
			re, err := regexp.Compile(fg.DetectPattern.Pattern)
			if err != nil {
				// Log bad patterns but don't abort loading
				continue
			}
			cf.re = re
		}
		loadedFootguns = append(loadedFootguns, cf)
	}
	return nil
}

// FootgunHit is a matched footgun with location info.
type FootgunHit struct {
	Entry      footgunEntry
	FilePath   string
	LineNumber int
	MatchText  string
}

// CheckFootguns scans file content for any matching footgun patterns.
// Returns up to 3 hits — the highest severity ones first.
func CheckFootguns(filePath, content string) []FootgunHit {
	if len(loadedFootguns) == 0 {
		return nil
	}

	base := strings.ToLower(filepath.Base(filePath))
	ext  := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))
	var hits []FootgunHit

	lines := splitLines(content)

	for _, cf := range loadedFootguns {
		if !languageMatches(cf.entry.DetectPattern.Language, ext, base) {
			continue
		}

		switch cf.entry.DetectPattern.Type {
		case "regex":
			if cf.re == nil {
				continue
			}
			for i, line := range lines {
				if cf.re.MatchString(line) {
					hits = append(hits, FootgunHit{
						Entry:      cf.entry,
						FilePath:   filePath,
						LineNumber: i + 1,
						MatchText:  strings.TrimSpace(line),
					})
					break // one hit per footgun per file
				}
			}

		case "text_search":
			needle := cf.entry.DetectPattern.Keywords
			if needle == "" {
				continue
			}
			for i, line := range lines {
				if strings.Contains(line, needle) {
					hits = append(hits, FootgunHit{
						Entry:      cf.entry,
						FilePath:   filePath,
						LineNumber: i + 1,
						MatchText:  strings.TrimSpace(line),
					})
					break
				}
			}

		// ast_pattern and heuristic require deeper analysis — skipped here,
		// handled by the AI observation path instead.
		}

		if len(hits) >= 3 {
			break
		}
	}

	return hits
}

// SubstitutePrompt replaces {file} and {line} placeholders in agent_prompt.
func SubstitutePrompt(template, filePath string, lineNum int) string {
	s := strings.ReplaceAll(template, "{file}", filePath)
	s = strings.ReplaceAll(s, "{line}", fmt.Sprintf("%d", lineNum))
	return s
}

// severityToConfidence converts a severity label to a base confidence score.
func severityToConfidence(severity string) float64 {
	switch severity {
	case "critical":
		return 0.95
	case "high":
		return 0.85
	case "medium":
		return 0.72
	default:
		return 0.60
	}
}

func splitLines(content string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// languageMatches returns true if the footgun's language applies to this file.
// ext is the lowercased extension without the dot; base is the lowercased filename.
func languageMatches(lang, ext, base string) bool {
	if lang == "" {
		return true // no language restriction
	}
	switch lang {
	case "javascript", "js":
		return ext == "js" || ext == "jsx" || ext == "ts" || ext == "tsx" || ext == "mjs" || ext == "cjs"
	case "typescript", "ts":
		return ext == "ts" || ext == "tsx"
	case "python", "py":
		return ext == "py"
	case "go":
		return ext == "go"
	case "rust":
		return ext == "rs"
	case "kotlin":
		return ext == "kt"
	case "swift":
		return ext == "swift"
	case "terraform", "hcl":
		return ext == "tf" || ext == "hcl"
	case "dockerfile":
		// Dockerfile has no extension — match by filename prefix
		return base == "dockerfile" || strings.HasPrefix(base, "dockerfile.")
	default:
		return ext == lang
	}
}
