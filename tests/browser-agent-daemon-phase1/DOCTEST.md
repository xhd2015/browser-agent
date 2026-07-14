# browser-agent daemon Phase 1 — foundations (session id, daemon meta, session dir, process alive)

Pure package helpers in `github.com/xhd2015/browser-agent/browseragent` with **zero**
dependency on HTTP, WebSocket, or CLI changes. Used by later phases (registry,
serve, session new, status).

| Surface | What is under test |
|---------|-------------------|
| Session ID validate | `ValidateSessionID` — pattern `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$` |
| Session ID generate | `GenerateSessionID` — `sess-` + 6 random `[a-z0-9]` |
| Daemon meta | `WriteDaemonMeta`, `ReadDaemonMeta`, `RemoveDaemonMeta` on `server.json` |
| Session dir | `SessionDirPath`, `SessionDirExists` under `{baseDir}/sessions/{id}` |
| Process alive | `IsProcessAlive` — portable existence check via signal 0 or equivalent |

**No HTTP.** **No listen socket.** **No CLI.** Disk tests use `t.TempDir()`.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Later phases** (registry, serve, CLI) will use **Daemon Foundations** helpers
from package `browseragent`:

### Session ID rules

- Valid pattern: `^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`
- Reject: empty, leading non-alnum, `/`, `..`, length > 64, invalid chars
- Auto-generate (later CLI): `sess-<6 lowercase alnum>` e.g. `sess-k3m9x2`

```text
ValidateSessionID(id) error     # nil when valid
GenerateSessionID() string      # sess- + 6 [a-z0-9]
```

### Daemon discovery file (`{BaseDir}/server.json`)

```text
type DaemonMeta {
  PID       int
  Addr      string    // host:port
  BaseURL   string    // http://host:port
  BaseDir   string    // absolute path preferred
  StartedAt time.Time // RFC3339 in JSON
}

WriteDaemonMeta(path, meta) error   # atomic write JSON + trailing newline
ReadDaemonMeta(path) (DaemonMeta, error)
RemoveDaemonMeta(path) error        # nil if already absent
```

### Session directory layout

```text
SessionDirPath(baseDir, sessionID) string   # {baseDir}/sessions/{id}
SessionDirExists(baseDir, sessionID) bool   # true when dir exists on disk
```

### Process liveness

```text
IsProcessAlive(pid int) bool   # true when process exists (signal 0 or equivalent)
```

**Test Client** in this tree calls package APIs directly with temp dirs and
fixed inputs. No servers started.

## Decision Tree

```
browser-agent-daemon-phase1
├── validate-session-id/                    [ValidateSessionID]
│   ├── valid/
│   │   ├── simple/                             "my-flow" → nil
│   │   ├── with-dots/                          "a.b.c" → nil
│   │   └── max-length-64/                      64-char valid id → nil
│   └── invalid/
│       ├── empty/                              "" → error
│       ├── leading-dash/                       "-bad" → error
│       ├── contains-slash/                     "a/b" → error
│       ├── contains-dotdot/                    "a..b" → error
│       └── too-long/                           65 chars → error
├── generate-session-id/                    [GenerateSessionID]
│   └── format/                                 ^sess-[a-z0-9]{6}$; two calls differ
├── daemon-meta/                            [Write/Read/Remove DaemonMeta]
│   ├── write-read-roundtrip/                   atomic JSON roundtrip preserves fields
│   ├── remove-missing-ok/                      Remove absent file → nil
│   └── read-missing-error/                     Read missing → error
├── session-dir/                            [SessionDirPath + SessionDirExists]
│   ├── exists-true/                            MkdirAll → true
│   └── exists-false/                           absent → false
└── process-alive/                          [IsProcessAlive]
    ├── self-pid/                               os.Getpid() → true
    └── dead-pid/                               unlikely large pid → false
```

### Parameter significance (high → low)

1. **Surface / Mode** — validate vs generate vs daemon-meta vs session-dir vs
   process-alive (different APIs).
2. **Within validate** — valid vs invalid (outcome polarity).
3. **Within invalid** — which rejection rule fires.
4. **Within daemon-meta** — write/read vs remove vs read-missing.
5. **Within session-dir** — dir present vs absent.
6. **Within process-alive** — live vs dead pid.

## Test Index

| Leaf | Scenario |
|------|----------|
| `validate-session-id/valid/simple` | Valid id `my-flow` → nil error |
| `validate-session-id/valid/with-dots` | Valid id `a.b.c` → nil error |
| `validate-session-id/valid/max-length-64` | 64-char valid id → nil error |
| `validate-session-id/invalid/empty` | Empty id → non-nil error |
| `validate-session-id/invalid/leading-dash` | Leading `-` → non-nil error |
| `validate-session-id/invalid/contains-slash` | Contains `/` → non-nil error |
| `validate-session-id/invalid/contains-dotdot` | Contains `..` → non-nil error |
| `validate-session-id/invalid/too-long` | Length 65 → non-nil error |
| `generate-session-id/format` | Matches `^sess-[a-z0-9]{6}$`; two calls differ |
| `daemon-meta/write-read-roundtrip` | Write/read preserves pid, addr, base_url, base_dir, started_at |
| `daemon-meta/remove-missing-ok` | Remove on absent path → nil |
| `daemon-meta/read-missing-error` | Read missing file → error |
| `session-dir/exists-true` | After MkdirAll, `SessionDirExists` true; path matches |
| `session-dir/exists-false` | Absent dir → `SessionDirExists` false |
| `process-alive/self-pid` | `IsProcessAlive(os.Getpid())` true |
| `process-alive/dead-pid` | Very large unlikely pid → false |

**Leaf count: 16**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-phase1
doctest test ./tests/browser-agent-daemon-phase1
# or:
cd tests/browser-agent-daemon-phase1 && doctest vet . && doctest test -v .
```

### Implementer contract (authoritative for GREEN)

```text
type DaemonMeta struct {
    PID       int       `json:"pid"`
    Addr      string    `json:"addr"`
    BaseURL   string    `json:"base_url"`
    BaseDir   string    `json:"base_dir"`
    StartedAt time.Time `json:"started_at"`
}

func ValidateSessionID(id string) error
func GenerateSessionID() string
func WriteDaemonMeta(path string, meta DaemonMeta) error   // atomic; JSON + trailing newline
func ReadDaemonMeta(path string) (DaemonMeta, error)
func RemoveDaemonMeta(path string) error                  // nil if already absent
func SessionDirPath(baseDir, sessionID string) string
func SessionDirExists(baseDir, sessionID string) bool
func IsProcessAlive(pid int) bool
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level API surface under test.
const (
	ModeValidateSessionID = "validate-session-id"
	ModeGenerateSessionID = "generate-session-id"
	ModeDaemonMeta        = "daemon-meta"
	ModeSessionDir        = "session-dir"
	ModeProcessAlive      = "process-alive"
)

// DaemonMetaOp — daemon-meta sub-operations.
const (
	DaemonMetaWriteReadRoundtrip = "write-read-roundtrip"
	DaemonMetaRemoveMissingOK    = "remove-missing-ok"
	DaemonMetaReadMissingError   = "read-missing-error"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	// ModuleRoot is workspace module directory.
	ModuleRoot string

	// BaseDir is temp parent for disk tests (set by Setup).
	BaseDir string

	// --- validate-session-id ---
	SessionID string

	// --- daemon-meta ---
	DaemonMetaOp string
	MetaPath     string
	Meta         browseragent.DaemonMeta

	// --- session-dir ---
	SessionDirID     string
	CreateSessionDir bool

	// --- process-alive ---
	PID int
}

// Response holds outcomes for all modes.
type Response struct {
	// validate
	ValidateErr     error
	ValidateErrText string

	// generate
	GeneratedID1 string
	GeneratedID2 string

	// daemon-meta
	WriteErr    error
	ReadErr     error
	RemoveErr   error
	ReadMeta    browseragent.DaemonMeta
	ReadRawJSON []byte

	// session-dir
	SessionDirPath   string
	SessionDirExists bool

	// process-alive
	ProcessAlive bool

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
	switch req.Mode {
	case ModeValidateSessionID:
		return runValidateSessionID(t, req)
	case ModeGenerateSessionID:
		return runGenerateSessionID(t, req)
	case ModeDaemonMeta:
		return runDaemonMeta(t, req)
	case ModeSessionDir:
		return runSessionDir(t, req)
	case ModeProcessAlive:
		return runProcessAlive(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runValidateSessionID(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	err := browseragent.ValidateSessionID(req.SessionID)
	resp := &Response{
		ValidateErr:     err,
		ValidateErrText: errString(err),
		ExitCode:        0,
	}
	return resp, nil
}

func runGenerateSessionID(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	id1 := browseragent.GenerateSessionID()
	id2 := browseragent.GenerateSessionID()
	return &Response{
		GeneratedID1: id1,
		GeneratedID2: id2,
		ExitCode:     0,
	}, nil
}

func runDaemonMeta(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.DaemonMetaOp == "" {
		t.Fatal("DaemonMetaOp must be set by leaf Setup")
	}
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	metaPath := req.MetaPath
	if metaPath == "" {
		metaPath = filepath.Join(req.BaseDir, "server.json")
	}
	resp := &Response{ExitCode: 0}

	switch req.DaemonMetaOp {
	case DaemonMetaWriteReadRoundtrip:
		meta := req.Meta
		if meta.StartedAt.IsZero() {
			meta = sampleDaemonMeta(req.BaseDir)
		}
		resp.WriteErr = browseragent.WriteDaemonMeta(metaPath, meta)
		if resp.WriteErr != nil {
			return resp, nil
		}
		readMeta, readErr := browseragent.ReadDaemonMeta(metaPath)
		resp.ReadErr = readErr
		resp.ReadMeta = readMeta
		if readErr == nil {
			if raw, err := os.ReadFile(metaPath); err == nil {
				resp.ReadRawJSON = raw
			}
		}
		return resp, nil

	case DaemonMetaRemoveMissingOK:
		missing := filepath.Join(req.BaseDir, "no-such-server.json")
		resp.RemoveErr = browseragent.RemoveDaemonMeta(missing)
		return resp, nil

	case DaemonMetaReadMissingError:
		missing := filepath.Join(req.BaseDir, "missing-server.json")
		_, readErr := browseragent.ReadDaemonMeta(missing)
		resp.ReadErr = readErr
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown DaemonMetaOp %q", req.DaemonMetaOp)
	}
}

func runSessionDir(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.BaseDir == "" {
		t.Fatal("BaseDir must be set by Setup")
	}
	sid := req.SessionDirID
	if sid == "" {
		sid = "my-flow"
	}
	path := browseragent.SessionDirPath(req.BaseDir, sid)
	if req.CreateSessionDir {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}
	}
	exists := browseragent.SessionDirExists(req.BaseDir, sid)
	return &Response{
		SessionDirPath:   path,
		SessionDirExists: exists,
		ExitCode:         0,
	}, nil
}

func runProcessAlive(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	pid := req.PID
	if pid == 0 {
		pid = os.Getpid()
	}
	alive := browseragent.IsProcessAlive(pid)
	return &Response{
		ProcessAlive: alive,
		ExitCode:     0,
	}, nil
}

func sampleDaemonMeta(baseDir string) browseragent.DaemonMeta {
	return browseragent.DaemonMeta{
		PID:       os.Getpid(),
		Addr:      "127.0.0.1:43761",
		BaseURL:   "http://127.0.0.1:43761",
		BaseDir:   baseDir,
		StartedAt: time.Date(2026, 3, 14, 12, 30, 0, 0, time.UTC),
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// sessionIDFormat matches GenerateSessionID output.
var sessionIDFormat = regexp.MustCompile(`^sess-[a-z0-9]{6}$`)

func assertValidSessionIDFormat(t *testing.T, id string) {
	t.Helper()
	if !sessionIDFormat.MatchString(id) {
		t.Fatalf("id %q does not match ^sess-[a-z0-9]{6}$", id)
	}
}

func daemonMetaFieldsEqual(t *testing.T, got, want browseragent.DaemonMeta) {
	t.Helper()
	if got.PID != want.PID {
		t.Fatalf("PID=%d want %d", got.PID, want.PID)
	}
	if got.Addr != want.Addr {
		t.Fatalf("Addr=%q want %q", got.Addr, want.Addr)
	}
	if got.BaseURL != want.BaseURL {
		t.Fatalf("BaseURL=%q want %q", got.BaseURL, want.BaseURL)
	}
	if got.BaseDir != want.BaseDir {
		t.Fatalf("BaseDir=%q want %q", got.BaseDir, want.BaseDir)
	}
	if !got.StartedAt.Equal(want.StartedAt) {
		t.Fatalf("StartedAt=%v want %v", got.StartedAt, want.StartedAt)
	}
}
```