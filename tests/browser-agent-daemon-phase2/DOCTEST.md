# browser-agent daemon Phase 2 — SessionRegistry (create, get, list, exists)

In-memory multi-session registry in `github.com/xhd2015/browser-agent/browseragent`.
Session creation writes per-session artifacts (`meta.json`, `SYSTEM.md`) under
`{baseDir}/sessions/{id}/`. Rejects duplicates when id is already in the registry
**or** the session directory exists on disk (crash recovery).

**No HTTP.** **No listen socket.** **No CLI.** Pure package API tests with temp
`BaseDir` and a fixed addr string like `127.0.0.1:43761`.

| Surface | What is under test |
|---------|-------------------|
| Create | `NewSessionRegistry`, `Create` — artifacts, registration, error paths |
| Get | `Get` — lookup live session by id |
| List | `List` — all snapshots sorted by session id |
| Exists | `Exists` — registry **or** on-disk session dir |

## Version

0.0.2

# DSN (Domain Specific Notion)

**SessionRegistry** holds a thread-safe map of live `session` values keyed by
session id. **Test Client** constructs a registry with a base directory and
control addr (host:port for URLs in meta), then exercises registry operations.

### Registry lifecycle

```text
NewSessionRegistry(baseDir, addr) *SessionRegistry
```

### Create session

```text
Create(id) (*CreateSessionResult, error)
  -> ValidateSessionID(id)
  -> Exists(id) guard (registry OR disk dir)
  -> mkdir {baseDir}/sessions/{id}
  -> write SYSTEM.md via FormatSystemPrompt(id)
  -> write meta.json (session_id, addr, base_url, session_url, system_prompt_path, product, control_port)
  -> register *session in map
  -> return CreateSessionResult paths + session_url
```

Duplicate paths return `ErrSessionExists` (checkable with `errors.Is`).

### Lookup and enumeration

```text
Get(id) (*session, bool)     # live registry only
List() []sessionSnapshot     # stable sorted order by session id
Exists(id) bool              # registry OR SessionDirExists(baseDir, id)
```

**Test Client** uses temp dirs, fixed addr `127.0.0.1:43761`, and direct package
imports. No servers started.

## Decision Tree

```
browser-agent-daemon-phase2
├── create/                              [Create + artifacts]
│   ├── ok/
│   │   ├── writes-meta/                     meta.json fields + paths
│   │   └── writes-system-md/                SYSTEM.md content = FormatSystemPrompt
│   └── error/
│       ├── invalid-id/                      ValidateSessionID error (not ErrSessionExists)
│       ├── duplicate-in-registry/           second Create → ErrSessionExists
│       └── disk-dir-exists/                 pre-seeded dir → ErrSessionExists
├── get/                                 [Get]
│   ├── found/                               after Create → ok true
│   └── not-found/                           absent id → ok false
├── list/                                [List]
│   ├── empty/                               fresh registry → len 0
│   └── two-sessions-sorted/                 Create b then a → sorted [a, b]
└── exists/                              [Exists]
    ├── registry-only/                       after Create → true
    ├── disk-only/                           dir on disk only → true
    └── neither/                             absent → false
```

### Parameter significance (high → low)

1. **Operation** — create vs get vs list vs exists (different APIs).
2. **Within create** — success vs error outcome.
3. **Within create/error** — which rejection path (validation, registry dup, disk dup).
4. **Within create/ok** — which artifact is asserted (meta vs SYSTEM.md).
5. **Within get/list/exists** — presence vs absence (or sort order for list).

## Test Index

| Leaf | Scenario |
|------|----------|
| `create/ok/writes-meta` | Create writes meta.json with expected discovery fields |
| `create/ok/writes-system-md` | Create writes SYSTEM.md matching FormatSystemPrompt |
| `create/error/invalid-id` | Invalid id → validation error, not ErrSessionExists |
| `create/error/duplicate-in-registry` | Second Create same id → ErrSessionExists |
| `create/error/disk-dir-exists` | Pre-created session dir → ErrSessionExists |
| `get/found` | After Create, Get returns ok true |
| `get/not-found` | Get unknown id → ok false |
| `list/empty` | New registry List returns empty |
| `list/two-sessions-sorted` | Two sessions listed sorted by id ascending |
| `exists/registry-only` | Created session → Exists true |
| `exists/disk-only` | Dir without registry entry → Exists true |
| `exists/neither` | No registry, no dir → Exists false |

**Leaf count: 12**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase2
doctest test ./tests/browser-agent-daemon-phase2
# or:
cd tests/browser-agent-daemon-phase2 && doctest vet . && doctest test -v .
```

### Implementer contract (authoritative for GREEN)

```text
type SessionRegistry struct { /* unexported */ }

type CreateSessionResult struct {
    SessionID  string
    SessionDir string // absolute preferred
    MetaPath   string
    SystemPath string
    SessionURL string // http://{addr}/go?session={id}
}

var ErrSessionExists error // sentinel; errors.Is for duplicates

func NewSessionRegistry(baseDir, addr string) *SessionRegistry
func (r *SessionRegistry) Create(id string) (*CreateSessionResult, error)
func (r *SessionRegistry) Get(id string) (*session, bool)
func (r *SessionRegistry) List() []sessionSnapshot
func (r *SessionRegistry) Exists(id string) bool
```

```go
import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — registry operation under test.
const (
	ModeCreate = "create"
	ModeGet    = "get"
	ModeList   = "list"
	ModeExists = "exists"
)

// CreateCase — create sub-scenarios.
const (
	CreateCaseOK               = "ok"
	CreateCaseInvalidID        = "invalid-id"
	CreateCaseDuplicateInReg   = "duplicate-in-registry"
	CreateCaseDiskDirExists    = "disk-dir-exists"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	// ModuleRoot is workspace module directory.
	ModuleRoot string

	// BaseDir is temp parent for disk tests (set by Setup).
	BaseDir string

	// Addr is host:port passed to NewSessionRegistry (default 127.0.0.1:43761).
	Addr string

	// --- create ---
	CreateCase           string
	SessionID            string
	SeedSessionDirBefore bool

	// --- get ---
	GetSessionID string

	// --- list ---
	ListSessionIDs []string

	// --- exists ---
	ExistsSessionID     string
	ExistsSeedDiskOnly  bool
	ExistsPreCreate     bool
}

// Response holds outcomes for all modes.
type Response struct {
	// create
	CreateResult             *browseragent.CreateSessionResult
	CreateErr                error
	SecondCreateErr          error
	CreateErrIsSessionExists bool

	// meta.json (create ok / writes-meta)
	MetaSessionID        string
	MetaAddr             string
	MetaBaseURL          string
	MetaSessionURL       string
	MetaSystemPromptPath string
	MetaProduct          string
	MetaControlPort      int
	MetaFileExists       bool

	// SYSTEM.md (create ok / writes-system-md)
	SystemMDExists  bool
	SystemMDContent string

	// get
	GetOK bool

	// list
	ListSessionIDs []string

	// exists
	Exists bool

	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	}
	ensureBaseDir(t, req)
	ensureAddr(t, req)

	switch req.Mode {
	case ModeCreate:
		return runCreate(t, req)
	case ModeGet:
		return runGet(t, req)
	case ModeList:
		return runList(t, req)
	case ModeExists:
		return runExists(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runCreate(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CreateCase == "" {
		t.Fatal("CreateCase must be set by leaf Setup")
	}
	if req.SessionID == "" && req.CreateCase != CreateCaseInvalidID {
		t.Fatal("SessionID must be set for create cases except invalid-id")
	}

	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	resp := &Response{ExitCode: 0}

	if req.SeedSessionDirBefore {
		dir := browseragent.SessionDirPath(req.BaseDir, req.SessionID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	switch req.CreateCase {
	case CreateCaseOK:
		result, err := reg.Create(req.SessionID)
		resp.CreateResult = result
		resp.CreateErr = err
		if err != nil {
			return resp, nil
		}
		fillArtifactFields(t, req, resp)
		return resp, nil

	case CreateCaseInvalidID:
		_, err := reg.Create(req.SessionID)
		resp.CreateErr = err
		resp.CreateErrIsSessionExists = errors.Is(err, browseragent.ErrSessionExists)
		return resp, nil

	case CreateCaseDuplicateInReg:
		if _, err := reg.Create(req.SessionID); err != nil {
			return nil, fmt.Errorf("first Create: %w", err)
		}
		_, err := reg.Create(req.SessionID)
		resp.SecondCreateErr = err
		resp.CreateErrIsSessionExists = errors.Is(err, browseragent.ErrSessionExists)
		return resp, nil

	case CreateCaseDiskDirExists:
		_, err := reg.Create(req.SessionID)
		resp.CreateErr = err
		resp.CreateErrIsSessionExists = errors.Is(err, browseragent.ErrSessionExists)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown CreateCase %q", req.CreateCase)
	}
}

func runGet(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.GetSessionID == "" {
		t.Fatal("GetSessionID must be set by leaf Setup")
	}
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	if req.SessionID != "" {
		if _, err := reg.Create(req.SessionID); err != nil {
			return nil, fmt.Errorf("pre-create for get: %w", err)
		}
	}
	_, ok := reg.Get(req.GetSessionID)
	return &Response{GetOK: ok, ExitCode: 0}, nil
}

func runList(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	for _, id := range req.ListSessionIDs {
		if _, err := reg.Create(id); err != nil {
			return nil, fmt.Errorf("pre-create %q: %w", id, err)
		}
	}
	snaps := reg.List()
	ids := make([]string, 0, len(snaps))
	for _, s := range snaps {
		ids = append(ids, s.SessionID)
	}
	return &Response{ListSessionIDs: ids, ExitCode: 0}, nil
}

func runExists(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ExistsSessionID == "" {
		t.Fatal("ExistsSessionID must be set by leaf Setup")
	}
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	if req.ExistsSeedDiskOnly {
		dir := browseragent.SessionDirPath(req.BaseDir, req.ExistsSessionID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	if req.ExistsPreCreate {
		if _, err := reg.Create(req.ExistsSessionID); err != nil {
			return nil, fmt.Errorf("pre-create for exists: %w", err)
		}
	}
	exists := reg.Exists(req.ExistsSessionID)
	return &Response{Exists: exists, ExitCode: 0}, nil
}

func fillArtifactFields(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if resp.CreateResult == nil {
		return
	}
	sessionDir := browseragent.SessionDirPath(req.BaseDir, req.SessionID)
	metaPath := filepath.Join(sessionDir, "meta.json")
	sysPath := filepath.Join(sessionDir, "SYSTEM.md")

	if _, err := os.Stat(metaPath); err == nil {
		resp.MetaFileExists = true
		data, err := os.ReadFile(metaPath)
		if err != nil {
			t.Fatalf("read meta.json: %v", err)
		}
		var meta map[string]any
		if err := json.Unmarshal(data, &meta); err != nil {
			t.Fatalf("parse meta.json: %v", err)
		}
		resp.MetaSessionID = stringField(meta, "session_id")
		resp.MetaAddr = stringField(meta, "addr")
		resp.MetaBaseURL = stringField(meta, "base_url")
		resp.MetaSessionURL = stringField(meta, "session_url")
		resp.MetaSystemPromptPath = stringField(meta, "system_prompt_path")
		resp.MetaProduct = stringField(meta, "product")
		resp.MetaControlPort = intField(meta, "control_port")
	}

	if data, err := os.ReadFile(sysPath); err == nil {
		resp.SystemMDExists = true
		resp.SystemMDContent = string(data)
	}
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func intField(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

func sortedStrings(ss []string) []string {
	out := append([]string(nil), ss...)
	sort.Strings(out)
	return out
}
```