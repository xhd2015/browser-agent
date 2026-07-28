# browser-agent daemon session restore — Phase 1 (disk rehydrate + no wipe on upgrade)

Classic TDD for **session restore from disk** and **upgrade that keeps session dirs**.

Today `RunDaemon` starts an empty `SessionRegistry` and the EnsureDaemon upgrade path
calls `removeSessionDirs` for disconnected "orphans". Phase 1:

1. **`RestoreSessionsFromDisk`** (or equivalent) rehydrates `{baseDir}/sessions/*/meta.json`
   into the in-memory registry as `waiting_extension` after `NewSessionRegistry`.
2. **Normal upgrade** (`ensureDaemonKillAndRespawn`) must **not** delete orphan session dirs.
3. **`Create` still fails** when the session already exists live or on disk (restore uses a
   register-from-disk path, not `Create`).

**Package API preferred.** Disk fixtures use `t.TempDir()`. One leaf wires `RunDaemon` +
`GET /v1/sessions`. No prepare_reconnect, no 30s reattach wait (Phase 2/3).

| Surface | What is under test |
|---------|-------------------|
| Restore from disk | `RestoreSessionsFromDisk(*SessionRegistry)` — scan, register, skip corrupt |
| Create still rejects | `Create` → `ErrSessionExists` for restored and disk-only ids |
| Upgrade keeps dirs | `ensureDaemonKillAndRespawn` / upgrade path does not `removeSessionDirs` |
| RunDaemon wiring | After start, live list includes disk-seeded sessions |

## Version

0.0.2

# DSN (Domain Specific Notion)

**SessionRegistry** holds live sessions under a **BaseDir**. On disk each session lives at
`{BaseDir}/sessions/{session_id}/` with `meta.json` (and optionally other artifacts).

**Restore Sessions From Disk** walks `sessions/*/`, validates the directory name as a session
id, requires readable `meta.json`, and **registers** an in-memory session in phase
`waiting_extension` without rewriting disk files (not `Create`).

**Upgrade Path** (`EnsureDaemon` when client version > daemon, zero extension-connected
sessions) kills and respawns the daemon. Disconnected session ids may still be **warned** as
orphans, but their **directories must remain** so the new process can restore them.

**RunDaemon** constructs a registry then restores from disk before serving HTTP, so a restarted
daemon reappears with prior session ids in `GET /v1/sessions`.

**Test Client** seeds temp BaseDirs with fixture dirs/meta, calls package APIs (and one
`RunDaemon` wiring leaf), and asserts registry membership, phase, Create rejection, and dir
preservation.

```text
# restore
seed {baseDir}/sessions/{id}/meta.json
NewSessionRegistry(baseDir, addr)
RestoreSessionsFromDisk(registry) error   # best-effort skip bad entries
  -> List/Get includes id; phase=waiting_extension; extension.connected=false

# create still exclusive
Create(id) after restore or disk-only dir -> ErrSessionExists

# upgrade keeps dirs
ensureDaemonKillAndRespawn(..., orphanIDs)
  -> SessionDirExists(baseDir, id) still true; marker file intact

# RunDaemon wiring
seed disk session -> RunDaemon -> GET /v1/sessions includes id
```

## Decision Tree

```
browser-agent-daemon-session-restore
├── restore-from-disk/                         [RestoreSessionsFromDisk]
│   ├── happy/
│   │   ├── single-valid-meta/                     one sess-*/meta.json → registered waiting_extension
│   │   └── multi-sorted/                          two ids → List sorted by session_id
│   ├── skip-invalid/                          [best-effort continue]
│   │   ├── bad-session-id-dirname/                invalid dirname → skip; empty registry
│   │   ├── missing-meta/                          dir without meta.json → skip
│   │   ├── corrupt-meta/                          unreadable JSON → skip
│   │   └── mixed-valid-and-skip/                  valid + bad → only valid; err nil
│   └── empty/
│       ├── no-sessions-dir/                       absent sessions/ → empty; nil
│       └── empty-sessions-dir/                    empty sessions/ → empty; nil
├── create-still-rejects/                      [Create path unchanged]
│   ├── after-restore-in-registry/                 restore then Create same id → ErrSessionExists
│   └── disk-only-no-restore/                      seed dir only → Create ErrSessionExists
├── upgrade-keeps-dirs/                        [no removeSessionDirs on upgrade]
│   ├── kill-respawn-orphans-remain/               kill+respawn with orphan ids → dirs + marker kept
│   └── multi-orphan-ids-remain/                   two orphan ids → both dirs kept
└── run-daemon-wires-restore/                  [RunDaemon calls restore]
    └── list-includes-disk/                        seed meta → RunDaemon → GET /v1/sessions has id
```

### Parameter significance (high → low)

1. **Surface** — restore vs create-reject vs upgrade-keep vs RunDaemon wiring.
2. **Within restore** — happy register vs skip-invalid vs empty layout.
3. **Within skip** — which invalid condition is skipped.
4. **Within create-reject** — restored (live) vs disk-only.
5. **Within upgrade** — single vs multi orphan id list.
6. **Within happy/empty** — cardinality / layout variant.

## Test Index

| Leaf | Scenario |
|------|----------|
| `restore-from-disk/happy/single-valid-meta` | Seed `sess-alpha` meta → restore → Get/List ok; phase `waiting_extension` |
| `restore-from-disk/happy/multi-sorted` | Seed `sess-bbb` + `sess-aaa` → List ids sorted ascending |
| `restore-from-disk/skip-invalid/bad-session-id-dirname` | Dir `-badid` → not registered; restore err nil |
| `restore-from-disk/skip-invalid/missing-meta` | Dir without meta.json → not registered |
| `restore-from-disk/skip-invalid/corrupt-meta` | Corrupt meta.json → not registered |
| `restore-from-disk/skip-invalid/mixed-valid-and-skip` | Valid + corrupt → only valid; err nil |
| `restore-from-disk/empty/no-sessions-dir` | No `sessions/` → List empty; err nil |
| `restore-from-disk/empty/empty-sessions-dir` | Empty `sessions/` → List empty; err nil |
| `create-still-rejects/after-restore-in-registry` | After restore, Create same id → `ErrSessionExists` |
| `create-still-rejects/disk-only-no-restore` | Disk dir only, no restore → Create → `ErrSessionExists` |
| `upgrade-keeps-dirs/kill-respawn-orphans-remain` | `ensureDaemonKillAndRespawn` with orphan ids keeps dir + marker |
| `upgrade-keeps-dirs/multi-orphan-ids-remain` | Two orphan ids → both dirs remain |
| `run-daemon-wires-restore/list-includes-disk` | Seed disk → `RunDaemon` → `GET /v1/sessions` includes id |

**Leaf count: 13**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-session-restore
doctest test ./tests/browser-agent-daemon-session-restore   # expect RED (Classic TDD)
```

### Implementer contract (authoritative for GREEN)

```text
// RestoreSessionsFromDisk scans registry baseDir/sessions/*/.
// Valid session id dirname + readable meta.json → register in-memory session
// as phase waiting_extension (extension.connected=false). Does NOT call Create
// and must not fail when the session directory already exists.
// Best-effort: skip invalid dirname, missing meta, corrupt meta; continue.
// Missing sessions/ directory is not an error (empty restore).
func RestoreSessionsFromDisk(r *SessionRegistry) error

// RunDaemon: after NewSessionRegistry(cfg.BaseDir, addr), call RestoreSessionsFromDisk
// before serving HTTP.

// ensureDaemonKillAndRespawn: for normal upgrade, do NOT call removeSessionDirs
// (orphan session directories are kept for restore on respawn).
// Tests call TestExported_ensureDaemonKillAndRespawn when the helper stays unexported:
func TestExported_ensureDaemonKillAndRespawn(cfg EnsureDaemonConfig, meta DaemonMeta, stderr io.Writer, orphanIDs []string) error

// Create still returns ErrSessionExists when id is in registry OR session dir exists.
```

**Sibling note:** `tests/browser-agent-daemon-version-port/ensure-daemon-upgrade/upgrade-warn-orphan-dirs`
currently expects orphan dirs **removed**. Implementer should update that leaf (and any
kill-existing orphan-wipe leaves that conflict with product policy) when greening this tree.
Phase 1 product rule: **keep session dirs across upgrade**.

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeRestoreFromDisk     = "restore-from-disk"
	ModeCreateStillRejects  = "create-still-rejects"
	ModeUpgradeKeepsDirs    = "upgrade-keeps-dirs"
	ModeRunDaemonWires      = "run-daemon-wires-restore"
)

// RestoreCase — restore-from-disk sub-scenarios.
const (
	RestoreCaseSingleValidMeta   = "single-valid-meta"
	RestoreCaseMultiSorted       = "multi-sorted"
	RestoreCaseBadSessionIDDir   = "bad-session-id-dirname"
	RestoreCaseMissingMeta       = "missing-meta"
	RestoreCaseCorruptMeta       = "corrupt-meta"
	RestoreCaseMixedValidAndSkip = "mixed-valid-and-skip"
	RestoreCaseNoSessionsDir     = "no-sessions-dir"
	RestoreCaseEmptySessionsDir  = "empty-sessions-dir"
)

// CreateRejectCase — create-still-rejects sub-scenarios.
const (
	CreateRejectAfterRestore = "after-restore-in-registry"
	CreateRejectDiskOnly     = "disk-only-no-restore"
)

// UpgradeKeepCase — upgrade-keeps-dirs sub-scenarios.
const (
	UpgradeKeepSingleOrphan = "kill-respawn-orphans-remain"
	UpgradeKeepMultiOrphan  = "multi-orphan-ids-remain"
)

// RunDaemonWireCase — run-daemon-wires-restore sub-scenarios.
const (
	RunDaemonListIncludesDisk = "list-includes-disk"
)

// SeedSession describes a fixture under {BaseDir}/sessions/{ID}/.
type SeedSession struct {
	ID           string
	// WriteMeta: write meta.json (default true when MetaJSON empty and not Corrupt/NoMeta).
	WriteMeta bool
	// MetaJSON overrides default meta body; empty → minimal valid meta when WriteMeta.
	MetaJSON string
	// CorruptMeta writes non-JSON meta.json.
	CorruptMeta bool
	// NoMeta skips meta.json entirely.
	NoMeta bool
	// Marker writes preserved.txt with this content (upgrade leaves).
	Marker string
}

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// restore-from-disk
	RestoreCase string
	Seeds       []SeedSession

	// create-still-rejects
	CreateRejectCase string
	CreateSessionID  string

	// upgrade-keeps-dirs
	UpgradeKeepCase string
	OrphanIDs       []string

	// run-daemon-wires-restore
	RunDaemonWireCase string
	ReadyTimeout      time.Duration
}

// ListEntry is a registry list snapshot subset for assertions.
type ListEntry struct {
	SessionID          string
	Phase              string
	ExtensionConnected bool
}

// Response holds outcomes for all modes.
type Response struct {
	RestoreErr error

	ListEntries []ListEntry
	ListIDs     []string
	GetOK       map[string]bool

	CreateErr                error
	CreateErrIsSessionExists bool

	// upgrade
	UpgradeErr      error
	OrphanDirExists map[string]bool
	MarkerIntact    map[string]bool

	// run-daemon
	BaseURL          string
	HTTPSessionIDs   []string
	HTTPStatusCode   int
	HTTPBody         string

	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	if req.ModuleRoot == "" {
		req.ModuleRoot = filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", ".."))
	}
	ensureBaseDir(t, req)
	ensureAddr(t, req)

	switch req.Mode {
	case ModeRestoreFromDisk:
		return runRestoreFromDisk(t, req)
	case ModeCreateStillRejects:
		return runCreateStillRejects(t, req)
	case ModeUpgradeKeepsDirs:
		return runUpgradeKeepsDirs(t, req)
	case ModeRunDaemonWires:
		return runRunDaemonWires(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runRestoreFromDisk(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.RestoreCase == "" {
		t.Fatal("RestoreCase must be set by leaf Setup")
	}
	if err := seedSessions(t, req.BaseDir, req.Addr, req.Seeds); err != nil {
		return nil, err
	}
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	resp := &Response{ExitCode: 0, GetOK: map[string]bool{}}
	resp.RestoreErr = browseragent.RestoreSessionsFromDisk(reg)
	fillListAndGet(t, reg, resp, collectSeedIDs(req.Seeds))
	return resp, nil
}

func runCreateStillRejects(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.CreateRejectCase == "" {
		t.Fatal("CreateRejectCase must be set by leaf Setup")
	}
	if req.CreateSessionID == "" {
		t.Fatal("CreateSessionID must be set by leaf Setup")
	}
	if err := seedSessions(t, req.BaseDir, req.Addr, req.Seeds); err != nil {
		return nil, err
	}
	reg := browseragent.NewSessionRegistry(req.BaseDir, req.Addr)
	resp := &Response{ExitCode: 0, GetOK: map[string]bool{}}

	switch req.CreateRejectCase {
	case CreateRejectAfterRestore:
		resp.RestoreErr = browseragent.RestoreSessionsFromDisk(reg)
		if resp.RestoreErr != nil {
			return resp, nil
		}
		_, err := reg.Create(req.CreateSessionID)
		resp.CreateErr = err
		resp.CreateErrIsSessionExists = errors.Is(err, browseragent.ErrSessionExists)
		fillListAndGet(t, reg, resp, []string{req.CreateSessionID})
		return resp, nil

	case CreateRejectDiskOnly:
		// No restore — Create must still reject existing session dir.
		_, err := reg.Create(req.CreateSessionID)
		resp.CreateErr = err
		resp.CreateErrIsSessionExists = errors.Is(err, browseragent.ErrSessionExists)
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown CreateRejectCase %q", req.CreateRejectCase)
	}
}

func runUpgradeKeepsDirs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.UpgradeKeepCase == "" {
		t.Fatal("UpgradeKeepCase must be set by leaf Setup")
	}
	if len(req.OrphanIDs) == 0 {
		t.Fatal("OrphanIDs must be set by leaf Setup")
	}
	if err := seedSessions(t, req.BaseDir, req.Addr, req.Seeds); err != nil {
		return nil, err
	}

	// Confirm markers/dirs exist before upgrade helper.
	for _, id := range req.OrphanIDs {
		if !browseragent.SessionDirExists(req.BaseDir, id) {
			return nil, fmt.Errorf("precondition: session dir missing for %s", id)
		}
	}

	meta := browseragent.DaemonMeta{
		PID:     os.Getpid(),
		Addr:    req.Addr,
		BaseURL: "http://" + req.Addr,
		BaseDir: req.BaseDir,
	}
	var stderr bytes.Buffer
	cfg := browseragent.EnsureDaemonConfig{
		BaseDir: req.BaseDir,
		Addr:    req.Addr,
		Stderr:  &stderr,
		KillFn: func(m browseragent.DaemonMeta) error {
			return nil // no real daemon; exercise wipe policy only
		},
		SpawnFn: func() error {
			return nil
		},
	}

	resp := &Response{
		ExitCode:        0,
		OrphanDirExists: map[string]bool{},
		MarkerIntact:    map[string]bool{},
	}
	// Prefer TestExported_ wrapper when ensureDaemonKillAndRespawn stays unexported.
	resp.UpgradeErr = browseragent.TestExported_ensureDaemonKillAndRespawn(cfg, meta, &stderr, req.OrphanIDs)

	for _, id := range req.OrphanIDs {
		resp.OrphanDirExists[id] = browseragent.SessionDirExists(req.BaseDir, id)
		markerPath := filepath.Join(browseragent.SessionDirPath(req.BaseDir, id), "preserved.txt")
		if data, err := os.ReadFile(markerPath); err == nil {
			resp.MarkerIntact[id] = strings.TrimSpace(string(data)) == "keep-me"
		} else {
			resp.MarkerIntact[id] = false
		}
	}
	return resp, nil
}

func runRunDaemonWires(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.RunDaemonWireCase == "" {
		t.Fatal("RunDaemonWireCase must be set by leaf Setup")
	}
	if err := seedSessions(t, req.BaseDir, req.Addr, req.Seeds); err != nil {
		return nil, err
	}

	timeout := req.ReadyTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	errCh := make(chan error, 1)
	go func() {
		_, runErr := browseragent.RunDaemon(ctx, browseragent.DaemonConfig{
			Addr:    addr,
			BaseDir: req.BaseDir,
			Stdout:  io.Discard,
			Stderr:  io.Discard,
		})
		errCh <- runErr
	}()

	baseURL := "http://" + addr
	if err := waitHealthOK(baseURL, timeout); err != nil {
		cancel()
		return &Response{ExitCode: 1, BaseURL: baseURL}, err
	}

	resp := &Response{ExitCode: 0, BaseURL: baseURL}
	ids, status, body, err := fetchSessionIDs(baseURL)
	resp.HTTPSessionIDs = ids
	resp.HTTPStatusCode = status
	resp.HTTPBody = body
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func fillListAndGet(t *testing.T, reg *browseragent.SessionRegistry, resp *Response, probeIDs []string) {
	t.Helper()
	snaps := reg.List()
	resp.ListEntries = make([]ListEntry, 0, len(snaps))
	resp.ListIDs = make([]string, 0, len(snaps))
	for _, s := range snaps {
		resp.ListIDs = append(resp.ListIDs, s.SessionID)
		resp.ListEntries = append(resp.ListEntries, ListEntry{
			SessionID:          s.SessionID,
			Phase:              s.Phase,
			ExtensionConnected: s.Extension.Connected,
		})
	}
	if resp.GetOK == nil {
		resp.GetOK = map[string]bool{}
	}
	for _, id := range probeIDs {
		if id == "" {
			continue
		}
		_, ok := reg.Get(id)
		resp.GetOK[id] = ok
	}
}

func collectSeedIDs(seeds []SeedSession) []string {
	ids := make([]string, 0, len(seeds))
	for _, s := range seeds {
		if s.ID != "" {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

func seedSessions(t *testing.T, baseDir, addr string, seeds []SeedSession) error {
	t.Helper()
	for _, s := range seeds {
		if s.ID == "" {
			continue
		}
		dir := browseragent.SessionDirPath(baseDir, s.ID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if s.Marker != "" {
			if err := os.WriteFile(filepath.Join(dir, "preserved.txt"), []byte(s.Marker+"\n"), 0o644); err != nil {
				return err
			}
		}
		if s.NoMeta {
			continue
		}
		metaPath := filepath.Join(dir, "meta.json")
		if s.CorruptMeta {
			if err := os.WriteFile(metaPath, []byte("{not-json"), 0o644); err != nil {
				return err
			}
			continue
		}
		writeMeta := s.WriteMeta || s.MetaJSON != "" || (!s.CorruptMeta && !s.NoMeta)
		// Default: write valid meta unless NoMeta/CorruptMeta.
		if s.NoMeta {
			writeMeta = false
		}
		if s.CorruptMeta {
			writeMeta = false
		}
		// When only ID is set (happy path), write default meta.
		if !s.NoMeta && !s.CorruptMeta {
			writeMeta = true
		}
		if !writeMeta {
			continue
		}
		body := s.MetaJSON
		if body == "" {
			body = defaultMetaJSON(s.ID, addr)
		}
		if err := os.WriteFile(metaPath, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func defaultMetaJSON(id, addr string) string {
	baseURL := "http://" + addr
	meta := map[string]any{
		"session_id":  id,
		"addr":        addr,
		"base_url":    baseURL,
		"session_url": baseURL + "/go?session=" + id,
		"product":     browseragent.ProductName,
	}
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return `{"session_id":"` + id + `"}` + "\n"
	}
	return string(b) + "\n"
}

func waitHealthOK(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(strings.TrimRight(baseURL, "/") + "/v1/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
			last = fmt.Errorf("health status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("health timeout")
	}
	return last
}

func fetchSessionIDs(baseURL string) ([]string, int, string, error) {
	resp, err := http.Get(strings.TrimRight(baseURL, "/") + "/v1/sessions")
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", err
	}
	body := string(raw)
	if resp.StatusCode != 200 {
		return nil, resp.StatusCode, body, fmt.Errorf("GET /v1/sessions status %d", resp.StatusCode)
	}
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, resp.StatusCode, body, fmt.Errorf("parse sessions JSON: %w", err)
	}
	ids := make([]string, 0, len(arr))
	for _, m := range arr {
		if id, ok := m["session_id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, resp.StatusCode, body, nil
}
```
