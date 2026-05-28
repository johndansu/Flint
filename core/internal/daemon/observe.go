package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"flint/core/internal/ai"
	"flint/core/internal/ipc"
	"flint/core/internal/store"
	schema "flint/core/pkg/schema"

	"github.com/google/uuid"
)

const observeSystem = `You are a senior software engineer watching a developer's file saves in real time.
A file was recently saved after a period of stillness. Decide if there is one thing worth saying.

Rules:
- Only speak if you have something genuinely useful — silence is correct most of the time
- One observation only. Specific. Actionable. No lectures.
- Prefer security, correctness, and hidden complexity over style
- If nothing is worth saying, return {"worth_saying": false}

Return JSON only — no prose, no markdown fences:
{
  "worth_saying": true,
  "kind": "technical" | "win" | "question" | "taste",
  "category": "security" | "performance" | "correctness" | "complexity" | "testing" | "debt" | "other",
  "confidence": 0.0-1.0,
  "text": "One specific observation about what was saved",
  "agent_prompt": "Optional: one instruction for an AI agent to fix this. Omit if nothing to fix."
}`

// Pipeline generates observations from file events and broadcasts them.
type Pipeline struct {
	db     *store.DB
	server *ipc.Server
	client *ai.Client
}

// NewPipeline creates a Pipeline.
func NewPipeline(db *store.DB, server *ipc.Server, client *ai.Client) *Pipeline {
	return &Pipeline{db: db, server: server, client: client}
}

// Run is called by the nudge engine when time gates pass.
// It reads the file, checks footguns, optionally calls the AI, then broadcasts.
func (p *Pipeline) Run(ctx context.Context, filePath, sessionType string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return // file may have been deleted — not an error
	}
	contentStr := string(content)

	// Footgun check is free — always run it first.
	// We dispatch only the highest-severity hit: one observation per pipeline run.
	hits := CheckFootguns(filePath, contentStr)
	if len(hits) > 0 {
		top := highestSeverity(hits)
		obs := p.obsFromFootgun(top, sessionType)
		p.dispatch(obs)
		_, _ = p.db.Raw().ExecContext(ctx,
			`INSERT INTO footgun_hits (footgun_id, file_path, line_number, timestamp)
			 VALUES (?,?,?,?)`,
			top.Entry.ID, filePath, top.LineNumber,
			time.Now().UTC().Format(time.RFC3339),
		)
		return
	}

	// No footgun matched — call AI for a technical observation
	if p.client == nil {
		return
	}
	obs, err := p.obsFromAI(ctx, filePath, contentStr, sessionType)
	if err != nil {
		log.Printf("observe: AI call failed: %v", err)
		return
	}
	if obs == nil {
		return // AI decided nothing worth saying
	}
	p.dispatch(*obs)
}

// highestSeverity returns the most severe hit from a non-empty slice.
func highestSeverity(hits []FootgunHit) FootgunHit {
	order := map[string]int{"critical": 3, "high": 2, "medium": 1, "low": 0}
	best := hits[0]
	for _, h := range hits[1:] {
		if order[h.Entry.Severity] > order[best.Entry.Severity] {
			best = h
		}
	}
	return best
}

// ---------------------------------------------------------------------------
// Observation builders
// ---------------------------------------------------------------------------

func (p *Pipeline) obsFromFootgun(hit FootgunHit, sessionType string) schema.Observation {
	id := uuid.New().String()
	ts := time.Now().UTC().Format(time.RFC3339)

	confidence := severityToConfidence(hit.Entry.Severity)

	var explainWhy *string
	if confidence < 0.82 {
		why := fmt.Sprintf("Regex match in %s at line %d — verify before acting on this",
			hit.Entry.Framework, hit.LineNumber)
		explainWhy = &why
	}

	agentPrompt := SubstitutePrompt(hit.Entry.AgentPrompt, hit.FilePath, hit.LineNumber)
	lineNum := hit.LineNumber
	footgunID := hit.Entry.ID

	return schema.Observation{
		Id:          id,
		Kind:        "technical",
		Category:    hit.Entry.Category,
		Confidence:  confidence,
		Text:        hit.Entry.Observation,
		AgentPrompt: &agentPrompt,
		ExplainWhy:  explainWhy,
		Timestamp:   ts,
		FilePath:    &hit.FilePath,
		LineNumber:  &lineNum,
		SessionType: &sessionType,
		FootgunId:   &footgunID,
	}
}

// aiObsResponse is the JSON shape the AI returns.
type aiObsResponse struct {
	WorthSaying bool    `json:"worth_saying"`
	Kind        string  `json:"kind"`
	Category    string  `json:"category"`
	Confidence  float64 `json:"confidence"`
	Text        string  `json:"text"`
	AgentPrompt string  `json:"agent_prompt"`
}

func (p *Pipeline) obsFromAI(ctx context.Context, filePath, content, sessionType string) (*schema.Observation, error) {
	if len(content) > 8000 {
		content = content[:8000] + "\n... (truncated)"
	}

	userMsg := fmt.Sprintf("File: %s\n\n```\n%s\n```", filePath, content)

	raw, err := p.client.Complete(ctx, observeSystem, []ai.Message{
		{Role: "user", Content: userMsg},
	})
	if err != nil {
		return nil, err
	}

	raw = extractJSON(raw)

	var resp aiObsResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("parse AI response: %w", err)
	}

	if !resp.WorthSaying {
		return nil, nil
	}

	id := uuid.New().String()
	ts := time.Now().UTC().Format(time.RFC3339)

	var explainWhy *string
	if resp.Confidence < 0.82 {
		why := "Confidence below threshold — verify before acting on this"
		explainWhy = &why
	}

	var agentPrompt *string
	if resp.AgentPrompt != "" {
		agentPrompt = &resp.AgentPrompt
	}

	return &schema.Observation{
		Id:          id,
		Kind:        resp.Kind,
		Category:    resp.Category,
		Confidence:  resp.Confidence,
		Text:        resp.Text,
		AgentPrompt: agentPrompt,
		ExplainWhy:  explainWhy,
		Timestamp:   ts,
		FilePath:    &filePath,
		SessionType: &sessionType,
	}, nil
}

// extractJSON finds the outermost JSON object in s.
// Handles cases where the AI wraps the response in prose or markdown fences.
func extractJSON(s string) string {
	first := strings.IndexByte(s, '{')
	last := strings.LastIndexByte(s, '}')
	if first == -1 || last == -1 || last < first {
		return s // return as-is and let Unmarshal produce the error
	}
	return s[first : last+1]
}

// ---------------------------------------------------------------------------
// Dispatch: broadcast over IPC and persist to store
// ---------------------------------------------------------------------------

func (p *Pipeline) dispatch(obs schema.Observation) {
	p.server.BroadcastObservation(obs)

	// Parse the observation's own timestamp instead of creating a new time.Now().
	// The observation timestamp is set at creation — using time.Now() here would
	// record the store write time (potentially seconds later) not the observation time.
	ts, err := time.Parse(time.RFC3339, obs.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	storeObs := store.Observation{
		ID:         obs.Id,
		Kind:       obs.Kind,
		Category:   obs.Category,
		Confidence: obs.Confidence,
		Text:       obs.Text,
		Timestamp:  ts,
	}
	if obs.AgentPrompt != nil { storeObs.AgentPrompt = *obs.AgentPrompt }
	if obs.ExplainWhy  != nil { storeObs.ExplainWhy  = *obs.ExplainWhy  }
	if obs.FilePath    != nil { storeObs.FilePath     = *obs.FilePath    }
	if obs.LineNumber  != nil { storeObs.LineNumber   = *obs.LineNumber  }
	if obs.SessionType != nil { storeObs.SessionType  = *obs.SessionType }
	if obs.FootgunId   != nil { storeObs.FootgunID    = *obs.FootgunId   }

	if err := p.db.InsertObservation(storeObs); err != nil {
		log.Printf("observe: store: %v", err)
	}
}
