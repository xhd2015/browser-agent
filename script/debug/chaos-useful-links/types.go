package main

import "time"

// Seed is one chaos corpus URL.
type Seed struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Kind   string   `json:"kind"`
	Env    string   `json:"env"`
	Market string   `json:"market"`
	Title  string   `json:"title"`
	Tags   []string `json:"tags"`
}

// Corpus is the on-disk seed file shape.
type Corpus struct {
	Source  string `json:"source"`
	Service string `json:"service"`
	Risk    string `json:"risk"`
	Seeds   []Seed `json:"seeds"`
}

// OpName is a dice-selectable chaos operation.
type OpName string

const (
	opOpenSeed       OpName = "open_seed"
	opNavigateSeed   OpName = "navigate_seed"
	opEvalIdentity   OpName = "eval_identity"
	opScreenshot     OpName = "screenshot"
	opSessionInfo    OpName = "session_info"
	opLogs           OpName = "logs"
	opBackgroundEval OpName = "background_eval"
	opRacePair       OpName = "race_pair"
)

// ResultClass is the classification of one op outcome.
type ResultClass string

const (
	classOKLoaded    ResultClass = "OK_LOADED"
	classOKAuthWall  ResultClass = "OK_AUTH_WALL"
	classOKSlow      ResultClass = "OK_SLOW"
	classOKInfo      ResultClass = "OK_INFO"
	classFailRouting ResultClass = "FAIL_ROUTING"
	classFailTimeout ResultClass = "FAIL_TIMEOUT"
	classFailDisconnect ResultClass = "FAIL_DISCONNECT"
	classFailAttach  ResultClass = "FAIL_ATTACH"
	classFailCrash   ResultClass = "FAIL_CRASH"
	classFailOther   ResultClass = "FAIL_OTHER"
)

// Severity for issues.
type Severity string

const (
	sevP0 Severity = "P0"
	sevP1 Severity = "P1"
	sevP2 Severity = "P2"
	sevP3 Severity = "P3"
)

// TabState tracks a harness-opened content tab (not the /go session page).
type TabState struct {
	TabID  int64  `json:"tab_id"`
	SeedID string `json:"seed_id,omitempty"`
	URL    string `json:"url,omitempty"`
}

// PlannedOp is one dice decision before execution.
type PlannedOp struct {
	Step   int    `json:"step"`
	Op     OpName `json:"op"`
	SeedID string `json:"seed_id,omitempty"`
	TabID  int64  `json:"tab_id,omitempty"`
}

// StepRecord is one executed step.
type StepRecord struct {
	Step      int            `json:"step"`
	Op        OpName         `json:"op"`
	SeedID    string         `json:"seed_id,omitempty"`
	TabID     int64          `json:"tab_id,omitempty"`
	Class     ResultClass    `json:"class"`
	DurationMS int64         `json:"duration_ms"`
	Error     string         `json:"error,omitempty"`
	EvalHref  string         `json:"eval_href,omitempty"`
	EvalTitle string         `json:"eval_title,omitempty"`
	Evidence  map[string]string `json:"evidence,omitempty"`
	Stdout    string         `json:"stdout,omitempty"`
	Stderr    string         `json:"stderr,omitempty"`
}

// Issue is a durable failure report.
type Issue struct {
	IssueID     string         `json:"issue_id"`
	Severity    Severity       `json:"severity"`
	Category    ResultClass    `json:"category"`
	SeedID      string         `json:"seed_id,omitempty"`
	URL         string         `json:"url,omitempty"`
	OpSequence  []string       `json:"op_sequence"`
	Expected    string         `json:"expected"`
	Actual      string         `json:"actual"`
	Step        int            `json:"step"`
	Evidence    map[string]string `json:"evidence,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

// RunRecord is the full run artifact.
type RunRecord struct {
	RunID     string            `json:"run_id"`
	SessionID string            `json:"session_id"`
	SeedRNG   int64             `json:"seed"`
	MaxOps    int               `json:"max_ops"`
	MaxTabs   int               `json:"max_tabs"`
	Corpus    string            `json:"corpus"`
	Chrome    string            `json:"chrome"`
	DryRun    bool              `json:"dry_run"`
	StartedAt time.Time         `json:"started_at"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	Counts    map[string]int    `json:"counts"`
	Steps     []StepRecord      `json:"steps"`
	Issues    []Issue           `json:"issues"`
	TabsAtEnd []TabState        `json:"tabs_at_end,omitempty"`
}
