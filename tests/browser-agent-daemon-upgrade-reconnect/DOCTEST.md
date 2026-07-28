# browser-agent daemon upgrade reconnect — Phase 3 (EnsureDaemon prepare + wait ≤30s)

Classic TDD for **Phase 3 of 3** of the reconnect-upgrade plan:

When **client > daemon**, `EnsureDaemon` must **upgrade even with extension-connected
sessions**, after preparing reconnect and waiting for reattach:

1. Resolve **waitList** = connected session ids (from GET sessions / snapshot fetch)
2. **POST prepare-upgrade** (`delay_ms=1000`) — **best-effort** (old daemons may 404)
3. Sleep `delay_ms + slack` (~200ms)
4. Kill daemon **without** `removeSessionDirs` (Phase 1)
5. Spawn new daemon + `waitDaemonReady`
6. Wait up to **30s** for waitList ids to show `extension.connected` again
7. Stderr: `upgraded daemon vOLD → vNEW`; reattached list and/or still-waiting warning

**Out of scope:** VERSION.txt unify; real browser e2e; prepare_reconnect WS shape
(Phase 2 tree `browser-agent-prepare-reconnect`).

| Surface | What is under test |
|---------|-------------------|
| Connected IDs | Pure `ConnectedIDsFromSnapshots` (waitList builder) |
| Wait reattach | Pure `WaitSessionsReconnected` with **mock fetch** (no HTTP) |
| EnsureDaemon | Injectable hooks: prepare-upgrade, sleep, kill/spawn, session poll, stderr |

**No real Chrome/Firefox.** Pure leaves need no network. EnsureDaemon leaves use a
minimal **httptest** health plane + injected hooks so tests avoid real process
lifecycle where possible.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Client CLI** embeds a **Client Version**. **Ensure Daemon** probes a live
**Control Plane** (health + `server.json`). When **client > daemon**, an upgrade
is required.

**Wait List** is the set of session ids whose extension is **connected** before
kill. Those extensions should receive **prepare_reconnect** (Phase 2) via
**POST /v1/admin/prepare-upgrade**, close after `delay_ms`, then reattach to the
replacement daemon.

**Prepare Upgrade** is **best-effort**: if the HTTP call fails (old daemon without
the admin route), upgrade continues.

**Kill + Spawn** replaces the process without wiping session dirs (Phase 1). After
the new daemon is healthy, **Wait Sessions Reconnected** polls session snapshots
until every waitList id is connected again, or until **Reattach Wait** (default
**30s**) elapses.

**Stderr** informs the operator:

- `upgraded daemon vOLD → vNEW` on successful version bump
- reattached session list when waitList members reconnect
- **warning** for ids still waiting after the reattach deadline

**Connected no longer blocks upgrade** (removes prior Q1 reuse-when-connected).

**Test Client** exercises pure helpers with in-memory snapshots and mock fetch
sequences; EnsureDaemon leaves inject PrepareUpgradeFn / FetchSessionsFn / SleepFn /
KillFn / SpawnFn against a fake health server + `server.json`.

```text
# pure waitList
ConnectedIDsFromSnapshots([{id, connected}, ...]) -> [connected ids in order]

# pure reattach wait
WaitSessionsReconnected(waitList, fetch, timeout, pollInterval)
  -> reattached[], stillWaiting[]
  fetch errors retry until timeout

# EnsureDaemon client > daemon (hooks)
fetch sessions -> waitList
PrepareUpgradeFn(baseURL, delay_ms=1000)  # best-effort
SleepFn(delay_ms + slack_ms)
KillFn(meta)   # no removeSessionDirs
SpawnFn()
waitDaemonReady
WaitSessionsReconnected(waitList, ...)  # default timeout 30s
stderr: upgraded daemon vOLD → vNEW
stderr: reattached ... / warning still waiting ...
```

## Decision Tree

```
browser-agent-daemon-upgrade-reconnect
├── connected-ids/                              [pure ConnectedIDsFromSnapshots]
│   ├── empty/                                      nil/empty snapshots → []
│   ├── only-connected/                             all connected → all ids
│   ├── mix-connected-disconnected/                 only connected subset
│   └── all-disconnected/                           none connected → []
├── wait-reattach/                              [WaitSessionsReconnected + mock fetch]
│   ├── empty-waitlist/                             waitList=[] → immediate empty
│   ├── all-reattach-immediate/                     first poll all connected
│   ├── all-reattach-after-polls/                   delayed reattach across polls
│   ├── partial-timeout/                            some never reconnect
│   └── fetch-error-retries/                        transient fetch err then ok
└── ensure-daemon-upgrade/                      [EnsureDaemon hooks orchestration]
    ├── allow-connected-upgrade/                    connected ≥1 still kill+spawn
    ├── prepare-upgrade-before-kill/                prepare then sleep then kill order
    ├── prepare-upgrade-best-effort/                prepare err does not abort
    ├── upgraded-line-and-reattach-ok/              stderr upgraded + reattached
    └── reattach-timeout-warn/                      stderr still-waiting warning
```

### Parameter significance (high → low)

1. **Surface** — pure waitList extract vs pure reattach wait vs EnsureDaemon orchestration.
2. **Within connected-ids** — connectivity composition of the snapshot list.
3. **Within wait-reattach** — poll outcome (empty / immediate / delayed / partial / fetch err).
4. **Within ensure-daemon-upgrade** — orchestration concern (allow connected, prepare order,
   best-effort prepare, success stderr, timeout warn).

## Test Index

| Leaf | Scenario |
|------|----------|
| `connected-ids/empty` | Empty/nil snapshots → empty id list |
| `connected-ids/only-connected` | Two connected → both ids in input order |
| `connected-ids/mix-connected-disconnected` | Connected A + disconnected B → `[A]` only |
| `connected-ids/all-disconnected` | Only disconnected → empty |
| `wait-reattach/empty-waitlist` | Empty waitList → reattached=[], stillWaiting=[], no fetch required |
| `wait-reattach/all-reattach-immediate` | First fetch all connected → reattached=waitList |
| `wait-reattach/all-reattach-after-polls` | Connects on 3rd poll → full reattach; poll count ≥3 |
| `wait-reattach/partial-timeout` | One of two never reconnects → stillWaiting has that id |
| `wait-reattach/fetch-error-retries` | First fetch errors, later ok → still succeeds |
| `ensure-daemon-upgrade/allow-connected-upgrade` | client>daemon + connected → KillFn+SpawnFn; no "cannot upgrade" reuse |
| `ensure-daemon-upgrade/prepare-upgrade-before-kill` | PrepareUpgradeFn(delay 1000) before Sleep before Kill |
| `ensure-daemon-upgrade/prepare-upgrade-best-effort` | PrepareUpgradeFn error → still kill+spawn+upgraded |
| `ensure-daemon-upgrade/upgraded-line-and-reattach-ok` | stderr has `upgraded daemon vOLD → vNEW` + reattached ids |
| `ensure-daemon-upgrade/reattach-timeout-warn` | short ReattachWait; still-waiting warning; EnsureDaemon ok |

**Leaf count: 14**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-upgrade-reconnect
doctest test ./tests/browser-agent-daemon-upgrade-reconnect   # expect RED (Classic TDD)
```

### Implementer contract (authoritative for GREEN)

```text
// --- Pure helpers (preferred exported names) ---

// SessionConnView is the minimal connectivity view for upgrade wait-list helpers.
// Maps from session snapshot extension.connected + session_id.
type SessionConnView struct {
    SessionID string
    Connected bool
}

// ConnectedIDsFromSnapshots returns session ids with Connected==true in input order.
// Nil/empty input → empty non-nil slice (or empty slice). Skips blank SessionID.
func ConnectedIDsFromSnapshots(sessions []SessionConnView) []string

// WaitSessionsReconnected polls fetch until every waitList id is Connected, or timeout.
// - timeout <= 0 → default 30s (DefaultReattachWait)
// - pollInterval <= 0 → a small default (e.g. 200ms); tests pass explicit short intervals
// - empty waitList → return immediately (reattached=[], stillWaiting=[])
// - fetch errors are retryable until timeout (do not hard-fail on first error)
// - reattached: waitList ids observed Connected at least once by deadline (prefer final set)
// - stillWaiting: waitList ids not Connected at deadline
// Order of returned slices: stable (waitList order or sorted — tests accept set equality).
func WaitSessionsReconnected(
    waitList []string,
    fetch func() ([]SessionConnView, error),
    timeout time.Duration,
    pollInterval time.Duration,
) (reattached, stillWaiting []string)

// If helpers stay unexported, export TestExported_ wrappers with the same signatures
// (mirroring TestExported_ensureDaemonKillAndRespawn).

// --- EnsureDaemonConfig hooks / fields ---

// EnsureDaemonConfig gains (names may match closely):
//   PrepareUpgradeFn func(baseURL string, delayMS int) error
//     - when non-nil, called instead of HTTP POST /v1/admin/prepare-upgrade
//     - when nil, default HTTP POST with delay_ms (default 1000) + reason daemon-upgrade
//     - errors are ignored (best-effort); upgrade continues
//   FetchSessionsFn func(baseURL string) ([]SessionConnView, error)
//     - when non-nil, used for waitList + reattach polling instead of GET /v1/sessions
//   SleepFn func(d time.Duration)
//     - when non-nil, used for prepare delay+slack and wait poll sleeps; default time.Sleep
//   ReattachWait time.Duration
//     - max wait after waitDaemonReady for waitList reattach; default 30s when <=0
//   PrepareDelayMS int
//     - delay_ms for prepare-upgrade + pre-kill sleep base; default 1000 when <=0
//   PrepareSlackMS int
//     - extra pre-kill sleep after delay; default 200 when <0? use 200 when unset (0 may mean 0;
//       product default: if both delay defaults applied, sleep 1000+200; tests set explicit values)

// EnsureDaemon when client > daemon:
//   1. sessions := FetchSessionsFn|GET; waitList := ConnectedIDsFromSnapshots(sessions)
//   2. PrepareUpgradeFn|POST (best-effort); do NOT abort on error
//   3. SleepFn(PrepareDelay + PrepareSlack)
//   4. ensureDaemonKillAndRespawn (no removeSessionDirs) — EVEN IF waitList non-empty
//      (REMOVE the connected-block reuse path / upgradeWarnConnected early return)
//   5. waitDaemonReady
//   6. WaitSessionsReconnected(waitList, ..., ReattachWait, ...)
//   7. stderr: "upgraded daemon v{old} → v{new}" (substring match; arrow →)
//   8. if reattached non-empty: stderr mentions reattached + ids
//   9. if stillWaiting non-empty: stderr warning mentions still waiting / waiting + ids
//   EnsureDaemon returns the new meta even when stillWaiting non-empty (soft warn).

// Stderr substrings tests look for (case-insensitive where noted):
//   "upgraded daemon" + oldVer + newVer  (and "→" or "->")
//   "reattach" (success path)
//   "still waiting" OR ("waiting" AND "reattach") for timeout warn
// Must NOT print "cannot upgrade" reuse block when connected (allow-connected leaf).

// Sibling note: tests/browser-agent-daemon-version-port/ensure-daemon-upgrade/blocked-connected-warn-reuse
// expects the OLD Q1 block. Update that leaf when greening Phase 3 product behavior.
```

```go
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// Mode — top-level surface under test.
const (
	ModeConnectedIDs         = "connected-ids"
	ModeWaitReattach         = "wait-reattach"
	ModeEnsureDaemonUpgrade  = "ensure-daemon-upgrade"
)

// ConnectedIDsCase — pure ConnectedIDsFromSnapshots composition.
const (
	ConnectedIDsEmpty                 = "empty"
	ConnectedIDsOnlyConnected         = "only-connected"
	ConnectedIDsMixConnectedDisconnected = "mix-connected-disconnected"
	ConnectedIDsAllDisconnected       = "all-disconnected"
)

// WaitReattachCase — WaitSessionsReconnected poll outcomes.
const (
	WaitReattachEmptyWaitlist       = "empty-waitlist"
	WaitReattachAllImmediate        = "all-reattach-immediate"
	WaitReattachAllAfterPolls       = "all-reattach-after-polls"
	WaitReattachPartialTimeout      = "partial-timeout"
	WaitReattachFetchErrorRetries   = "fetch-error-retries"
)

// EnsureUpgradeCase — EnsureDaemon orchestration concerns.
const (
	EnsureUpgradeAllowConnected     = "allow-connected-upgrade"
	EnsureUpgradePrepareBeforeKill  = "prepare-upgrade-before-kill"
	EnsureUpgradePrepareBestEffort  = "prepare-upgrade-best-effort"
	EnsureUpgradeLineAndReattachOK  = "upgraded-line-and-reattach-ok"
	EnsureUpgradeReattachTimeoutWarn = "reattach-timeout-warn"
)

// ConnSnap is the test-side view fed into pure helpers / mock fetch.
type ConnSnap struct {
	SessionID string
	Connected bool
}

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Addr       string

	// connected-ids
	ConnectedIDsCase string
	Snapshots        []ConnSnap

	// wait-reattach
	WaitReattachCase string
	WaitList         []string
	// MockFetchPlan encodes sequential fetch results for WaitSessionsReconnected.
	// Each entry is one poll. Empty ConnectedIDs + FetchErr → error that poll.
	MockFetchPlan []MockFetchStep
	WaitTimeout   time.Duration
	PollInterval  time.Duration

	// ensure-daemon-upgrade
	EnsureUpgradeCase string
	ClientVersion     string
	DaemonVersion     string
	// ConnectedSessionIDs appear as connected on initial waitList fetch (and may reattach).
	ConnectedSessionIDs []string
	// ReattachSessionIDs become connected during post-spawn reattach polling.
	// If empty and not timeout case, defaults to ConnectedSessionIDs (full reattach).
	ReattachSessionIDs []string
	// PrepareUpgradeErr when non-empty makes PrepareUpgradeFn return that error.
	PrepareUpgradeErr string
	// ReattachWait for EnsureDaemon; tests use short values.
	ReattachWait time.Duration
	// PrepareDelayMS / PrepareSlackMS explicit for sleep-order assertions.
	PrepareDelayMS int
	PrepareSlackMS int
	// WaitTimeout for waitDaemonReady.
	ReadyTimeout time.Duration
}

// MockFetchStep is one poll outcome for wait-reattach leaves.
type MockFetchStep struct {
	// Sessions returned this poll (nil + FetchErr for error).
	Sessions []ConnSnap
	// FetchErr when non-empty: this poll returns an error with that text.
	FetchErr string
}

// Response holds outcomes for all modes.
type Response struct {
	// connected-ids
	ConnectedIDs []string

	// wait-reattach
	Reattached   []string
	StillWaiting []string
	FetchCalls   int

	// ensure-daemon-upgrade
	EnsureErr          error
	Stderr             string
	KillFnCalled       bool
	SpawnFnCalled      bool
	PrepareFnCalled    bool
	PrepareDelayMSArg  int
	PrepareBaseURLArg  string
	PrepareErrReturned bool
	// CallOrder records hook labels in invocation order, e.g. "prepare", "sleep", "kill", "spawn".
	CallOrder []string
	// SleepCalls records SleepFn durations in order.
	SleepCalls []time.Duration
	Meta       browseragent.DaemonMeta
	NewPID     int
	OldPID     int

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
	switch req.Mode {
	case ModeConnectedIDs:
		return runConnectedIDs(t, req)
	case ModeWaitReattach:
		return runWaitReattach(t, req)
	case ModeEnsureDaemonUpgrade:
		return runEnsureDaemonUpgrade(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runConnectedIDs(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ConnectedIDsCase == "" {
		t.Fatal("ConnectedIDsCase must be set by leaf Setup")
	}
	views := toPackageViews(req.Snapshots)
	ids := browseragent.ConnectedIDsFromSnapshots(views)
	return &Response{ExitCode: 0, ConnectedIDs: ids}, nil
}

func runWaitReattach(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.WaitReattachCase == "" {
		t.Fatal("WaitReattachCase must be set by leaf Setup")
	}
	timeout := req.WaitTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	poll := req.PollInterval
	if poll <= 0 {
		poll = 5 * time.Millisecond
	}

	var calls int32
	plan := req.MockFetchPlan
	fetch := func() ([]browseragent.SessionConnView, error) {
		n := int(atomic.AddInt32(&calls, 1))
		idx := n - 1
		if len(plan) == 0 {
			return []browseragent.SessionConnView{}, nil
		}
		if idx >= len(plan) {
			// Repeat last step for further polls.
			idx = len(plan) - 1
		}
		step := plan[idx]
		if step.FetchErr != "" {
			return nil, fmt.Errorf("%s", step.FetchErr)
		}
		return toPackageViews(step.Sessions), nil
	}

	reattached, stillWaiting := browseragent.WaitSessionsReconnected(req.WaitList, fetch, timeout, poll)
	return &Response{
		ExitCode:     0,
		Reattached:   reattached,
		StillWaiting: stillWaiting,
		FetchCalls:   int(atomic.LoadInt32(&calls)),
	}, nil
}

func runEnsureDaemonUpgrade(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.EnsureUpgradeCase == "" {
		t.Fatal("EnsureUpgradeCase must be set by leaf Setup")
	}
	ensureBaseDir(t, req)

	clientVer := req.ClientVersion
	if clientVer == "" {
		clientVer = "0.2.0"
	}
	daemonVer := req.DaemonVersion
	if daemonVer == "" {
		daemonVer = "0.1.0"
	}
	if req.PrepareDelayMS <= 0 {
		req.PrepareDelayMS = 1000
	}
	// PrepareSlackMS: 0 is a valid explicit value for tests that only care about delay;
	// default 200 when leaf left it at sentinel -1 is awkward — use default 200 if not set
	// by leaf: leaves set explicitly. Root default applied in grouping Setup is 200.
	if req.PrepareSlackMS < 0 {
		req.PrepareSlackMS = 200
	}
	readyTO := req.ReadyTimeout
	if readyTO <= 0 {
		readyTO = 2 * time.Second
	}
	reattachWait := req.ReattachWait
	if reattachWait <= 0 && req.EnsureUpgradeCase != EnsureUpgradeReattachTimeoutWarn {
		// Success-path leaves use a short but positive wait; timeout leaf sets explicitly.
		reattachWait = 200 * time.Millisecond
	}

	// Mutable health version for waitDaemonReady after spawn.
	var healthVer atomic.Value
	healthVer.Store(daemonVer)

	// Fake control plane: health only (sessions via FetchSessionsFn hook).
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		v, _ := healthVer.Load().(string)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"product":        browseragent.ProductName,
			"daemon_version": v,
			"base_dir":       req.BaseDir,
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// Derive addr from httptest URL host:port.
	u := strings.TrimPrefix(srv.URL, "http://")
	req.Addr = u
	baseURL := srv.URL

	oldPID := os.Getpid()
	metaPath := filepath.Join(req.BaseDir, "server.json")
	if err := browseragent.WriteDaemonMeta(metaPath, browseragent.DaemonMeta{
		PID:           oldPID,
		Addr:          req.Addr,
		BaseURL:       baseURL,
		BaseDir:       req.BaseDir,
		StartedAt:     time.Now().UTC(),
		DaemonVersion: daemonVer,
	}); err != nil {
		return nil, err
	}

	connected := append([]string(nil), req.ConnectedSessionIDs...)
	reattachIDs := append([]string(nil), req.ReattachSessionIDs...)
	if len(reattachIDs) == 0 && req.EnsureUpgradeCase != EnsureUpgradeReattachTimeoutWarn {
		reattachIDs = append([]string(nil), connected...)
	}

	var (
		mu           sync.Mutex
		callOrder    []string
		sleepCalls   []time.Duration
		prepareCalled bool
		prepareDelay  int
		prepareBase   string
		killCalled    bool
		spawnCalled   bool
		fetchPhase    int32 // 0=pre-kill waitList, 1=post-spawn reattach polls
	)

	record := func(label string) {
		mu.Lock()
		callOrder = append(callOrder, label)
		mu.Unlock()
	}

	// After spawn, reattach fetch returns Connected for reattachIDs.
	fetchSessions := func(bu string) ([]browseragent.SessionConnView, error) {
		_ = bu
		phase := atomic.LoadInt32(&fetchPhase)
		ids := connected
		connectedFlag := phase == 0 // pre-kill: connected true for waitList
		if phase > 0 {
			// post-spawn reattach: only reattachIDs connected
			out := make([]browseragent.SessionConnView, 0, len(connected)+len(reattachIDs))
			seen := map[string]bool{}
			for _, id := range reattachIDs {
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				out = append(out, browseragent.SessionConnView{SessionID: id, Connected: true})
			}
			// Include waitList ids that have not reattached as disconnected.
			for _, id := range connected {
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				out = append(out, browseragent.SessionConnView{SessionID: id, Connected: false})
			}
			return out, nil
		}
		out := make([]browseragent.SessionConnView, 0, len(ids))
		for _, id := range ids {
			if id == "" {
				continue
			}
			out = append(out, browseragent.SessionConnView{SessionID: id, Connected: connectedFlag})
		}
		return out, nil
	}

	var stderr bytes.Buffer
	cfg := browseragent.EnsureDaemonConfig{
		BaseDir:       req.BaseDir,
		Addr:          req.Addr,
		ClientVersion: clientVer,
		WaitTimeout:   readyTO,
		Stderr:        &stderr,
		// Hook fields — implementer must wire these on EnsureDaemonConfig.
		PrepareDelayMS: req.PrepareDelayMS,
		PrepareSlackMS: req.PrepareSlackMS,
		ReattachWait:   reattachWait,
		PrepareUpgradeFn: func(bu string, delayMS int) error {
			prepareCalled = true
			prepareDelay = delayMS
			prepareBase = bu
			record("prepare")
			if req.PrepareUpgradeErr != "" {
				return fmt.Errorf("%s", req.PrepareUpgradeErr)
			}
			return nil
		},
		FetchSessionsFn: fetchSessions,
		SleepFn: func(d time.Duration) {
			mu.Lock()
			sleepCalls = append(sleepCalls, d)
			mu.Unlock()
			record("sleep")
			// No real sleep — tests stay fast.
		},
		KillFn: func(m browseragent.DaemonMeta) error {
			killCalled = true
			record("kill")
			// Mark pre-kill phase done for subsequent fetches after spawn.
			return nil
		},
		SpawnFn: func() error {
			spawnCalled = true
			record("spawn")
			atomic.StoreInt32(&fetchPhase, 1)
			healthVer.Store(clientVer)
			// Rewrite server.json so waitDaemonReady succeeds with new version.
			return browseragent.WriteDaemonMeta(metaPath, browseragent.DaemonMeta{
				PID:           oldPID, // same test process; alive
				Addr:          req.Addr,
				BaseURL:       baseURL,
				BaseDir:       req.BaseDir,
				StartedAt:     time.Now().UTC(),
				DaemonVersion: clientVer,
			})
		},
	}

	resp := &Response{
		ExitCode: 0,
		OldPID:   oldPID,
	}

	meta, err := browseragent.EnsureDaemon(cfg)
	resp.EnsureErr = err
	resp.Meta = meta
	resp.NewPID = meta.PID
	resp.Stderr = stderr.String()
	resp.KillFnCalled = killCalled
	resp.SpawnFnCalled = spawnCalled
	resp.PrepareFnCalled = prepareCalled
	resp.PrepareDelayMSArg = prepareDelay
	resp.PrepareBaseURLArg = prepareBase
	resp.PrepareErrReturned = req.PrepareUpgradeErr != ""
	mu.Lock()
	resp.CallOrder = append([]string(nil), callOrder...)
	resp.SleepCalls = append([]time.Duration(nil), sleepCalls...)
	mu.Unlock()
	if err != nil {
		resp.ExitCode = 1
		// Return resp with err so Assert can inspect hooks even on failure;
		// transport error path uses non-nil err.
		return resp, err
	}
	return resp, nil
}

func toPackageViews(snaps []ConnSnap) []browseragent.SessionConnView {
	if snaps == nil {
		return nil
	}
	out := make([]browseragent.SessionConnView, len(snaps))
	for i, s := range snaps {
		out[i] = browseragent.SessionConnView{
			SessionID: s.SessionID,
			Connected: s.Connected,
		}
	}
	return out
}
```
