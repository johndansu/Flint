package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"flint/core/internal/errorlog"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"
)

const (
	osvBatchURL     = "https://api.osv.dev/v1/querybatch"
	depCheckInterval = 7 * 24 * time.Hour // re-check weekly
	depCheckKey      = "deps.last_check"
	depReportedKey   = "deps.reported_cves" // comma-separated CVE IDs
)

// dep is a single parsed dependency entry ready for OSV lookup.
type dep struct {
	Name      string
	Version   string
	Ecosystem string // "npm" | "Go" | "PyPI" | "crates.io"
	DepFile   string // source file for error log location
}

// DepMonitor checks project dependencies against the OSV vulnerability database.
// CVE findings bypass all daemon time gates and fire immediately.
type DepMonitor struct {
	server        *ipc.Server
	db            *store.DB
	el            *errorlog.Logger
	workspaceRoot string
	httpClient    *http.Client

	mu           sync.Mutex
	reportedCVEs map[string]bool // CVE IDs fired this session
}

// NewDepMonitor creates a DepMonitor. Call Run to start the background check loop.
func NewDepMonitor(server *ipc.Server, db *store.DB, el *errorlog.Logger, workspaceRoot string) *DepMonitor {
	// Seed reported set from persisted baselines so we don't re-fire on restart
	reported := make(map[string]bool)
	if v, _ := db.GetBaseline(depReportedKey); v != "" {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				reported[id] = true
			}
		}
	}

	return &DepMonitor{
		server:        server,
		db:            db,
		el:            el,
		workspaceRoot: workspaceRoot,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		reportedCVEs:  reported,
	}
}

// Run checks deps on startup then re-checks every 7 days. Blocks until ctx is done.
func (d *DepMonitor) Run(ctx context.Context) {
	go func() {
		// Check immediately if it's been more than 7 days since the last check
		if d.shouldCheck() {
			if err := d.check(ctx); err != nil {
				log.Printf("deps: check failed: %v", err)
			}
		}

		ticker := time.NewTicker(depCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := d.check(ctx); err != nil {
					log.Printf("deps: check failed: %v", err)
				}
			}
		}
	}()
}

func (d *DepMonitor) shouldCheck() bool {
	v, _ := d.db.GetBaseline(depCheckKey)
	if v == "" {
		return true
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return true
	}
	return time.Since(t) >= depCheckInterval
}

// ---------------------------------------------------------------------------
// Main check
// ---------------------------------------------------------------------------

func (d *DepMonitor) check(ctx context.Context) error {
	deps := d.parseDeps(d.workspaceRoot)
	if len(deps) == 0 {
		log.Printf("deps: no dependency files found in %s", d.workspaceRoot)
		return nil
	}

	log.Printf("deps: checking %d dependencies against OSV...", len(deps))
	found, err := d.queryOSV(ctx, deps)
	if err != nil {
		return fmt.Errorf("osv query: %w", err)
	}

	// Record timestamp regardless of findings
	_ = d.db.SetBaseline(depCheckKey, time.Now().UTC().Format(time.RFC3339))

	log.Printf("deps: %d vulnerable packages found", len(found))
	for _, v := range found {
		d.fireVuln(ctx, v)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Dependency file parsing
// ---------------------------------------------------------------------------

func (d *DepMonitor) parseDeps(root string) []dep {
	var deps []dep
	deps = append(deps, parsePackageJSON(filepath.Join(root, "package.json"))...)
	deps = append(deps, parseGoMod(filepath.Join(root, "go.mod"))...)
	deps = append(deps, parseRequirementsTxt(filepath.Join(root, "requirements.txt"))...)
	deps = append(deps, parseCargoToml(filepath.Join(root, "Cargo.toml"))...)
	return deps
}

func parsePackageJSON(path string) []dep {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(b, &pkg); err != nil {
		return nil
	}
	var deps []dep
	for name, ver := range pkg.Dependencies {
		deps = append(deps, dep{Name: name, Version: cleanVersion(ver), Ecosystem: "npm", DepFile: path})
	}
	for name, ver := range pkg.DevDependencies {
		deps = append(deps, dep{Name: name, Version: cleanVersion(ver), Ecosystem: "npm", DepFile: path})
	}
	return deps
}

func parseGoMod(path string) []dep {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// Match: <module-path> v<semver>
	re := regexp.MustCompile(`^\s+(\S+)\s+(v[\w.\-+]+)`)
	var deps []dep
	for _, line := range strings.Split(string(b), "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) == 3 {
			deps = append(deps, dep{
				Name:      m[1],
				Version:   strings.TrimPrefix(m[2], "v"),
				Ecosystem: "Go",
				DepFile:   path,
			})
		}
	}
	return deps
}

func parseRequirementsTxt(path string) []dep {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// Match: package==1.2.3 or package>=1.0,<2.0
	re := regexp.MustCompile(`^([A-Za-z0-9_\-\.]+)[=><~!]{1,2}([\d][^\s,;#]*)`)
	var deps []dep
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if len(m) == 3 {
			deps = append(deps, dep{
				Name:      m[1],
				Version:   cleanVersion(m[2]),
				Ecosystem: "PyPI",
				DepFile:   path,
			})
		}
	}
	return deps
}

func parseCargoToml(path string) []dep {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// Match: name = "version" or name = { version = "..." }
	re := regexp.MustCompile(`^([a-z0-9_\-]+)\s*=\s*"([\d][^"]*)"`)
	var deps []dep
	inDeps := false
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "[dependencies]") || strings.HasPrefix(line, "[dev-dependencies]") {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
		}
		if !inDeps {
			continue
		}
		m := re.FindStringSubmatch(strings.TrimSpace(line))
		if len(m) == 3 {
			deps = append(deps, dep{
				Name:      m[1],
				Version:   cleanVersion(m[2]),
				Ecosystem: "crates.io",
				DepFile:   path,
			})
		}
	}
	return deps
}

// cleanVersion strips leading semver range operators (^, ~, >=, >, etc.)
func cleanVersion(v string) string {
	return strings.TrimLeft(strings.TrimSpace(v), "^~>=<!")
}

// ---------------------------------------------------------------------------
// OSV batch query
// ---------------------------------------------------------------------------

// osvVulnFound is a flattened result ready to fire.
type osvVulnFound struct {
	ID       string
	Aliases  []string // includes CVE-xxxx IDs
	Summary  string
	Severity string // "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | ""
	Pkg      dep
}

func (d *DepMonitor) queryOSV(ctx context.Context, deps []dep) ([]osvVulnFound, error) {
	type osvQuery struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		Version string `json:"version"`
	}
	type batchReq struct {
		Queries []osvQuery `json:"queries"`
	}

	queries := make([]osvQuery, 0, len(deps))
	for _, d := range deps {
		if d.Version == "" {
			continue
		}
		var q osvQuery
		q.Package.Name = d.Name
		q.Package.Ecosystem = d.Ecosystem
		q.Version = d.Version
		queries = append(queries, q)
	}
	if len(queries) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(batchReq{Queries: queries})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, osvBatchURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osv: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("osv: HTTP %d: %s", resp.StatusCode, raw)
	}

	var batchResp struct {
		Results []struct {
			Vulns []struct {
				ID      string   `json:"id"`
				Aliases []string `json:"aliases"`
				Summary string   `json:"summary"`
				Severity []struct {
					Type  string `json:"type"`
					Score string `json:"score"`
				} `json:"severity"`
				DatabaseSpecific map[string]any `json:"database_specific"`
			} `json:"vulns"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return nil, fmt.Errorf("osv: decode: %w", err)
	}

	var found []osvVulnFound
	for i, result := range batchResp.Results {
		if i >= len(deps) {
			break
		}
		pkg := deps[i]
		for _, v := range result.Vulns {
			sev := extractSeverity(v.DatabaseSpecific, v.Severity)
			found = append(found, osvVulnFound{
				ID:       v.ID,
				Aliases:  v.Aliases,
				Summary:  v.Summary,
				Severity: sev,
				Pkg:      pkg,
			})
		}
	}
	return found, nil
}

// extractSeverity pulls a human-readable severity string from OSV response fields.
func extractSeverity(dbSpecific map[string]any, severity []struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}) string {
	// database_specific.severity is the most reliable field (GitHub Advisory)
	if dbSpecific != nil {
		if s, ok := dbSpecific["severity"].(string); ok && s != "" {
			return strings.ToUpper(s)
		}
	}
	// Fall back to CVSS vector: look for impact letters in the score string
	for _, s := range severity {
		score := s.Score
		if strings.Contains(score, "/C:H") || strings.Contains(score, "/I:H") || strings.Contains(score, "/A:H") {
			if strings.Contains(score, "AV:N") {
				return "CRITICAL"
			}
			return "HIGH"
		}
		if strings.Contains(score, "/C:L") || strings.Contains(score, "/I:L") || strings.Contains(score, "/A:L") {
			return "MEDIUM"
		}
	}
	return "HIGH" // assume high if we can't tell — better to over-alert than miss
}

// ---------------------------------------------------------------------------
// Fire observation
// ---------------------------------------------------------------------------

func (d *DepMonitor) fireVuln(ctx context.Context, v osvVulnFound) {
	// Pick the best ID to display: prefer CVE-xxxx over GHSA-xxxx
	displayID := v.ID
	for _, alias := range v.Aliases {
		if strings.HasPrefix(alias, "CVE-") {
			displayID = alias
			break
		}
	}

	// Dedup: don't re-fire a CVE already reported this session or last session
	d.mu.Lock()
	if d.reportedCVEs[displayID] {
		d.mu.Unlock()
		return
	}
	d.reportedCVEs[displayID] = true
	// Persist updated set so we survive restarts
	ids := make([]string, 0, len(d.reportedCVEs))
	for id := range d.reportedCVEs {
		ids = append(ids, id)
	}
	_ = d.db.SetBaseline(depReportedKey, strings.Join(ids, ","))
	d.mu.Unlock()

	confidence := severityConfidence(v.Severity)
	summary := v.Summary
	if summary == "" {
		summary = "No description available."
	}

	text := fmt.Sprintf(
		"%s in %s %s (%s ecosystem): %s  Fix: update to a patched version.",
		displayID, v.Pkg.Name, v.Pkg.Version, v.Pkg.Ecosystem, summary,
	)

	fp := v.Pkg.DepFile
	obs := schema.Observation{
		Id:         uuid.New().String(),
		Kind:       schema.ObservationKindTechnical,
		Category:   "security",
		Confidence: confidence,
		Text:       text,
		FilePath:   &fp,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	// Broadcast immediately — bypasses all time gates
	d.server.BroadcastObservation(obs)

	// Persist to store
	_ = d.db.InsertObservation(store.Observation{
		ID:         obs.Id,
		Kind:       obs.Kind,
		Category:   obs.Category,
		Confidence: obs.Confidence,
		Text:       obs.Text,
		FilePath:   v.Pkg.DepFile,
		Timestamp:  time.Now().UTC(),
	})

	// Log to error log
	if d.el != nil {
		_ = d.el.CVE(fmt.Sprintf("%s: %s", displayID, summary),
			errorlog.WithFile(v.Pkg.DepFile, 0),
		)
	}

	log.Printf("deps: %s %s %s — %s", v.Severity, displayID, v.Pkg.Name, summary)
}

func severityConfidence(sev string) float64 {
	switch strings.ToUpper(sev) {
	case "CRITICAL":
		return 0.99
	case "HIGH":
		return 0.95
	case "MEDIUM":
		return 0.80
	default:
		return 0.70
	}
}
