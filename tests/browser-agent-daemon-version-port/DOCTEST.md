# browser-agent daemon version + fixed port + host/port flags

Greenfield feature: embedded `VERSION.txt`, semver upgrade in `EnsureDaemon` /
`serve --kill-existing`, extended `GET /v1/health` + `server.json`, flag migration
(`--host`, `--port`, `--server-port`), foreign-port fail-hard, TTY `serve --stop`
confirm. **Classic TDD RED** — current code uses `pickEphemeralAddr`, minimal health
`{"ok":true}`, `--addr` flags, no version compare.

| Surface | What is under test |
|---------|-------------------|
| Version | `ClientVersion`, loose `CompareVersion`, pre-release ignored |
| Health | Extended `/v1/health` + `daemon_version` in `server.json` |
| Port | Fixed default `127.0.0.1:43761`, port-in-use fail, no ephemeral `:0` |
| Foreign port | Non-daemon listener → fail hard + hint |
| EnsureDaemon upgrade | Q1/Q2/Q4–Q6/Q11/Q12 version paths |
| Kill-existing | `--kill-existing` always kills (Q10), disk cleanup (Q14) |
| Serve stop | TTY `[Y/n]` confirm (Q15) |
| Flags | `--host`/`--port`/`--server-port` help + `server.json` resolve |
| Regression | phase8 spawn-when-down with explicit `--port` |

Depends on Phases 1–8 (`RunDaemon`, `EnsureDaemon`, `SessionRegistry`, fake extension
patterns from phase4). Implementer updates sibling trees documented in root `SETUP.md`.

## Version

0.0.2

# DSN (Domain Specific Notion)

**Client CLI** embeds **Client Version** from `VERSION.txt` via `go:embed`. Session
commands and `serve` share **Control Port** default **43761** on **127.0.0.1**.

**Daemon Host** (`RunDaemon`) binds a fixed control address, writes **server.json**
(pid, addr, base_dir, **daemon_version**), serves **GET /v1/health** with
`ok`, `product`, `daemon_version`, `base_dir`.

**Ensure Daemon** (`EnsureDaemon`) probes health + version. When **client > daemon**
and no **extension-connected** sessions → kill + respawn. When connected ≥1 → **warn
stderr**, reuse old daemon (Q1). When **client < daemon** → warn, reuse (Q5). Missing
daemon version treated as **0.0.0** (Q6). Foreign listener on host:port → **fail hard**
(Q3).

**Fake Extension** dials `GET /v1/ws?session=<id>`, sends `hello` → session
`extension.connected=true`. Only connected sessions block upgrade.

**Foreign Listener** occupies the control port with non-browser-agent HTTP before
`serve` or `EnsureDaemon`.

**Operator serve --stop** on TTY prints connected session summary and `Stop daemon? [Y/n]`;
non-TTY warns and stops immediately when connected.

```text
EnsureDaemon: health + CompareVersion(client, daemon)
  client > daemon + connected>=1 -> warn, reuse (no kill)
  client > daemon + connected==0 -> kill + respawn + orphan dir cleanup
  client == daemon             -> reuse
  client < daemon              -> warn stderr, reuse

serve --kill-existing -> always kill (overrides Q1) + warn + RemoveAll session dirs
foreign port on 127.0.0.1:N -> exit != 0 + hint --server-port / --port
```

## Decision Tree

```
browser-agent-daemon-version-port/
├── version/                           [VERSION.txt + compare + embed]
│   ├── client-version-readable/
│   ├── compare-newer/
│   ├── compare-equal/
│   ├── compare-older/
│   ├── missing-daemon-version/
│   └── prerelease-ignored/
├── health/                            [GET /v1/health extended]
│   ├── fields-present/
│   └── server-json-daemon-version/
├── port/                              [fixed port bind]
│   ├── serve-default-addr/
│   ├── serve-port-in-use/
│   ├── serve-custom-host-port/
│   └── no-ephemeral-spawn/
├── foreign-port/                      [Q3 fail hard]
│   ├── serve-foreign-listener/
│   └── ensure-daemon-foreign/
├── ensure-daemon-upgrade/             [EnsureDaemon version path]
│   ├── reuse-equal-version/
│   ├── warn-older-client/
│   ├── blocked-connected-warn-reuse/
│   ├── upgrade-no-connected/
│   ├── upgrade-warn-orphan-dirs/
│   ├── kill-fails-hard/
│   └── respawn-fails-hard/
├── kill-existing/                     [serve --kill-existing Q10]
│   ├── always-kills-connected/
│   ├── warn-disconnected-orphans/
│   └── no-version-check/
├── serve-stop/                        [Q15 TTY confirm]
│   ├── tty-default-yes/
│   ├── tty-n-aborts/
│   └── non-tty-warn-connected/
├── flags/                             [CLI flag migration]
│   ├── serve-help-host-port/
│   ├── session-help-server-port/
│   └── resolve-server-json-port/
└── regression/                        [sibling parity]
    └── phase8-ensure-daemon-spawn/
```

## Test Index

| Leaf | Scenario |
|------|----------|
| `version/client-version-readable` | `ClientVersion()` non-empty, matches embed |
| `version/compare-newer` | `0.2.0 > 0.1.0` |
| `version/compare-equal` | `0.2.0 == 0.2.0` |
| `version/compare-older` | `0.1.0 < 0.2.0` |
| `version/missing-daemon-version` | empty → `0.0.0`; client newer |
| `version/prerelease-ignored` | `0.2.0-beta` orders below `0.2.0` |
| `health/fields-present` | health JSON has product, daemon_version, base_dir |
| `health/server-json-daemon-version` | `RunDaemon` writes `daemon_version` |
| `port/serve-default-addr` | help/constants document `127.0.0.1:43761` via `--host`/`--port` |
| `port/serve-port-in-use` | second bind same port → fail hard |
| `port/serve-custom-host-port` | `--host` + `--port <free>` binds correctly |
| `port/no-ephemeral-spawn` | empty addr spawn uses default port not `:0` |
| `foreign-port/serve-foreign-listener` | foreign HTTP → serve exit ≠ 0 + hint |
| `foreign-port/ensure-daemon-foreign` | session path detects foreign → fail ≠ 0 |
| `ensure-daemon-upgrade/reuse-equal-version` | equal versions → reuse, no kill |
| `ensure-daemon-upgrade/warn-older-client` | client older → warn stderr, reuse |
| `ensure-daemon-upgrade/blocked-connected-warn-reuse` | connected blocks upgrade → warn, reuse |
| `ensure-daemon-upgrade/upgrade-no-connected` | newer client, 0 connected → kill+respawn |
| `ensure-daemon-upgrade/upgrade-warn-orphan-dirs` | disconnected → warn ids + RemoveAll |
| `ensure-daemon-upgrade/kill-fails-hard` | kill error → EnsureDaemon fails (Q11) |
| `ensure-daemon-upgrade/respawn-fails-hard` | respawn unhealthy → error (Q12) |
| `kill-existing/always-kills-connected` | connected + `--kill-existing` still kills |
| `kill-existing/warn-disconnected-orphans` | disconnected → warn + RemoveAll |
| `kill-existing/no-version-check` | equal version + flag → still kills |
| `serve-stop/tty-default-yes` | TTY + empty stdin → stops |
| `serve-stop/tty-n-aborts` | TTY + `n` → daemon stays |
| `serve-stop/non-tty-warn-connected` | !TTY + connected → warn + stop |
| `flags/serve-help-host-port` | serve help documents `--host`, `--port` |
| `flags/session-help-server-port` | session help documents `--host`, `--server-port` |
| `flags/resolve-server-json-port` | omit port → reads `server.json` addr |
| `regression/phase8-ensure-daemon-spawn` | spawn-when-down with explicit `--port` |

**Leaf count: 31**

## How to Run

```sh
doctest vet ./tests/browser-agent-daemon-version-port
doctest test ./tests/browser-agent-daemon-version-port   # expect RED (Classic TDD)
# After implementer GREEN:
doctest test ./tests/browser-agent-daemon-phase8/...
doctest test ./tests/browser-agent-serve-stop/...
doctest test ./tests/browser-agent-session-addr-resolve/...
doctest test ./...
```

**Implementer inject hooks** (referenced by `Run`; add before GREEN):

- `browseragent/inject/versionhooks.go` — `ClientVersionOverride func() string`
- `browseragent/inject/terminalhooks.go` — `IsTerminalFn func(io.Reader) bool`
- `browseragent/version.go` — `ClientVersion()`, `CompareVersion(a,b string) int`,
  `EffectiveDaemonVersion(s string) string`
- `DaemonConfig.DaemonVersion`, `EnsureDaemonConfig.ClientVersion`, `KillFn`

```go
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/browser-agent/browseragent"
	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// --- Modes (top-level split) ---

const (
	ModeVersion             = "version"
	ModeHealth              = "health"
	ModePort                = "port"
	ModeForeignPort         = "foreign-port"
	ModeEnsureDaemonUpgrade = "ensure-daemon-upgrade"
	ModeKillExisting        = "kill-existing"
	ModeServeStop           = "serve-stop"
	ModeFlags               = "flags"
	ModeRegression          = "regression"
)

// VersionOp
const (
	VersionOpClientReadable     = "client-version-readable"
	VersionOpCompareNewer       = "compare-newer"
	VersionOpCompareEqual       = "compare-equal"
	VersionOpCompareOlder       = "compare-older"
	VersionOpMissingDaemonVer   = "missing-daemon-version"
	VersionOpPrereleaseIgnored  = "prerelease-ignored"
)

// HealthOp
const (
	HealthOpFieldsPresent         = "fields-present"
	HealthOpServerJSONDaemonVer   = "server-json-daemon-version"
)

// PortOp
const (
	PortOpServeDefaultAddr    = "serve-default-addr"
	PortOpServePortInUse      = "serve-port-in-use"
	PortOpServeCustomHostPort = "serve-custom-host-port"
	PortOpNoEphemeralSpawn    = "no-ephemeral-spawn"
)

// ForeignPortOp
const (
	ForeignPortOpServe         = "serve-foreign-listener"
	ForeignPortOpEnsureDaemon  = "ensure-daemon-foreign"
)

// UpgradeOp
const (
	UpgradeOpReuseEqual           = "reuse-equal-version"
	UpgradeOpWarnOlderClient      = "warn-older-client"
	UpgradeOpBlockedConnected     = "blocked-connected-warn-reuse"
	UpgradeOpUpgradeNoConnected   = "upgrade-no-connected"
	UpgradeOpUpgradeWarnOrphans     = "upgrade-warn-orphan-dirs"
	UpgradeOpKillFailsHard          = "kill-fails-hard"
	UpgradeOpRespawnFailsHard       = "respawn-fails-hard"
)

// KillExistingOp
const (
	KillExistingOpAlwaysKillsConnected = "always-kills-connected"
	KillExistingOpWarnOrphans          = "warn-disconnected-orphans"
	KillExistingOpNoVersionCheck       = "no-version-check"
)

// ServeStopOp
const (
	ServeStopOpTTYDefaultYes      = "tty-default-yes"
	ServeStopOpTTYNAborts         = "tty-n-aborts"
	ServeStopOpNonTTYWarnConnected = "non-tty-warn-connected"
)

// FlagsOp
const (
	FlagsOpServeHelpHostPort      = "serve-help-host-port"
	FlagsOpSessionHelpServerPort  = "session-help-server-port"
	FlagsOpResolveServerJSONPort  = "resolve-server-json-port"
)

// RegressionOp
const (
	RegressionOpPhase8Spawn = "phase8-ensure-daemon-spawn"
)

// Request is narrowed root→leaf by Setup functions.
type Request struct {
	Mode string

	ModuleRoot string
	BaseDir    string
	Host       string
	Port       int
	Addr       string

	VersionOp      string
	HealthOp       string
	PortOp         string
	ForeignPortOp  string
	UpgradeOp      string
	KillExistingOp string
	ServeStopOp    string
	FlagsOp        string
	RegressionOp   string

	ClientVersion string
	DaemonVersion string
	CompareA      string
	CompareB      string

	SessionIDA string
	SessionIDB string
	OrphanID   string

	HelloVersion  string
	HelloFeatures []string

	CLIArgs []string
	Env     map[string]string
	Stdin   string
	IsTTY   bool

	KillFnFails   bool
	SpawnUnhealthy bool

	ReadyTimeout time.Duration
	ShutdownWait time.Duration
}

// Response holds outcomes for all modes.
type Response struct {
	BaseURL string
	Addr    string
	Port    int

	CompareResult int
	ClientVersion string
	EmbeddedVer   string

	HealthJSON map[string]any
	HealthOK   bool

	Meta            browseragent.DaemonMeta
	MetaExists      bool
	MetaDaemonVer   string
	OldPID          int
	NewPID          int
	KillFnCalled    bool
	SpawnFnCalled   bool
	SpawnAddrUsed   string

	Stderr string
	Stdout string
	CLIErr string
	ExitCode int

	HelpText string

	ExtensionConnected bool
	SessionCreated     bool
	SessionDirExists   bool
	OrphanDirExists    bool

	ForeignHintSeen bool
	PortInUseSeen   bool

	TTYConfirmSeen bool
	DaemonStopped  bool
	DaemonStillUp  bool

	ResolveBaseURL string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.Mode == "" {
		t.Fatal("Mode must be set by grouping/leaf Setup")
	}
	switch req.Mode {
	case ModeVersion:
		return runVersionMode(t, req)
	case ModeHealth:
		return runHealthMode(t, req)
	case ModePort:
		return runPortMode(t, req)
	case ModeForeignPort:
		return runForeignPortMode(t, req)
	case ModeEnsureDaemonUpgrade:
		return runEnsureDaemonUpgradeMode(t, req)
	case ModeKillExisting:
		return runKillExistingMode(t, req)
	case ModeServeStop:
		return runServeStopMode(t, req)
	case ModeFlags:
		return runFlagsMode(t, req)
	case ModeRegression:
		return runRegressionMode(t, req)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
}

func runVersionMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.VersionOp == "" {
		t.Fatal("VersionOp must be set by leaf Setup")
	}
	resp := &Response{}
	switch req.VersionOp {
	case VersionOpClientReadable:
		ver := browseragent.ClientVersion()
		resp.ClientVersion = ver
		resp.EmbeddedVer = readVERSIONTxt(req.ModuleRoot)
		return resp, nil
	case VersionOpCompareNewer:
		resp.CompareResult = browseragent.CompareVersion(req.CompareA, req.CompareB)
		return resp, nil
	case VersionOpCompareEqual:
		resp.CompareResult = browseragent.CompareVersion(req.CompareA, req.CompareB)
		return resp, nil
	case VersionOpCompareOlder:
		resp.CompareResult = browseragent.CompareVersion(req.CompareA, req.CompareB)
		return resp, nil
	case VersionOpMissingDaemonVer:
		norm := browseragent.EffectiveDaemonVersion(req.CompareA)
		resp.CompareResult = browseragent.CompareVersion(req.ClientVersion, norm)
		resp.ClientVersion = norm
		return resp, nil
	case VersionOpPrereleaseIgnored:
		resp.CompareResult = browseragent.CompareVersion(req.CompareA, req.CompareB)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown VersionOp %q", req.VersionOp)
	}
}

func runHealthMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.HealthOp == "" {
		t.Fatal("HealthOp must be set by leaf Setup")
	}
	port := req.Port
	if port == 0 {
		port = pickFreePort(t)
	}
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	srv, cleanup, err := startDaemonAt(t, req, addr, req.DaemonVersion)
	if err != nil {
		return resp, err
	}
	defer cleanup()

	resp.BaseURL = srv.BaseURL
	raw, err := fetchHealthJSON(srv.BaseURL)
	if err != nil {
		return resp, err
	}
	resp.HealthJSON = raw
	resp.HealthOK = true

	meta, err := browseragent.ReadDaemonMeta(daemonMetaPath(req.BaseDir))
	if err != nil {
		return resp, err
	}
	resp.Meta = meta
	resp.MetaExists = true

	switch req.HealthOp {
	case HealthOpFieldsPresent:
		return resp, nil
	case HealthOpServerJSONDaemonVer:
		resp.MetaDaemonVer = metaDaemonVersion(meta)
		return resp, nil
	default:
		return nil, fmt.Errorf("unknown HealthOp %q", req.HealthOp)
	}
}

func runPortMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.PortOp == "" {
		t.Fatal("PortOp must be set by leaf Setup")
	}
	resp := &Response{}

	switch req.PortOp {
	case PortOpServeDefaultAddr:
		var stdout bytes.Buffer
		cliErr := browseragent.HandleCLI([]string{"serve", "--help"}, req.Env, &stdout, &bytes.Buffer{})
		resp.HelpText = stdout.String()
		if cliErr != nil {
			resp.CLIErr = cliErr.Error()
		}
		resp.Addr = browseragent.DefaultAddr
		return resp, nil

	case PortOpServePortInUse:
		port := pickFreePort(t)
		addr := loopbackAddr(port)
		resp.Addr = addr
		srv1, cleanup1, err := startDaemonAt(t, req, addr, req.DaemonVersion)
		if err != nil {
			return resp, err
		}
		defer cleanup1()
		resp.BaseURL = srv1.BaseURL

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		cfg := browseragent.DaemonConfig{Addr: addr, BaseDir: req.BaseDir + "-dup", Stdout: io.Discard, Stderr: io.Discard}
		_, err = browseragent.RunDaemon(ctx, cfg)
		resp.CLIErr = ""
		if err != nil {
			resp.CLIErr = err.Error()
			low := strings.ToLower(err.Error())
			resp.PortInUseSeen = strings.Contains(low, "in use") || strings.Contains(low, "bind") || strings.Contains(low, "listen")
		}
		return resp, nil

	case PortOpServeCustomHostPort:
		port := pickFreePort(t)
		host := "127.0.0.1"
		if req.Host != "" {
			host = req.Host
		}
		resp.Port = port
		resp.Addr = net.JoinHostPort(host, strconv.Itoa(port))

		ctx, cancel := context.WithCancel(context.Background())
		var stderr bytes.Buffer
		args := []string{
			"serve",
			"--host", host,
			"--port", strconv.Itoa(port),
			"--base-dir", req.BaseDir,
			"--no-open-chrome",
			"--no-agent-run",
		}
		done := make(chan error, 1)
		go func() {
			done <- browseragent.ServeWithContext(ctx, args, req.Env, io.Discard, &stderr)
		}()
		t.Cleanup(cancel)
		baseURL := "http://" + resp.Addr
		if err := waitHealthOK(baseURL, req.ReadyTimeout); err != nil {
			resp.Stderr = stderr.String()
			return resp, err
		}
		resp.BaseURL = baseURL
		resp.HealthOK = true
		return resp, nil

	case PortOpNoEphemeralSpawn:
		spawnAddr := ""
		spawnCalled := false
		cfg := browseragent.EnsureDaemonConfig{
			BaseDir:     req.BaseDir,
			Addr:        "",
			ClientVersion: req.ClientVersion,
			WaitTimeout: req.ReadyTimeout,
			SpawnFn: func() error {
				spawnCalled = true
				picked := browseragent.DefaultAddr
				spawnAddr = picked
				ctx := context.Background()
				go func() {
					_, _ = browseragent.RunDaemon(ctx, browseragent.DaemonConfig{
						Addr:          picked,
						BaseDir:       req.BaseDir,
						DaemonVersion: req.DaemonVersion,
						Stdout:        io.Discard,
						Stderr:        io.Discard,
					})
				}()
				return nil
			},
		}
		meta, err := browseragent.EnsureDaemon(cfg)
		resp.SpawnFnCalled = spawnCalled
		resp.SpawnAddrUsed = spawnAddr
		if err != nil {
			resp.CLIErr = err.Error()
			return resp, err
		}
		resp.Meta = meta
		resp.Addr = meta.Addr
		if meta.Addr != "" {
			resp.BaseURL = "http://" + meta.Addr
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown PortOp %q", req.PortOp)
	}
}

func runForeignPortMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ForeignPortOp == "" {
		t.Fatal("ForeignPortOp must be set by leaf Setup")
	}
	port := pickFreePort(t)
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	foreignCleanup := startForeignHTTP(t, addr)
	defer foreignCleanup()

	switch req.ForeignPortOp {
	case ForeignPortOpServe:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var stderr bytes.Buffer
		args := []string{
			"serve",
			"--host", "127.0.0.1",
			"--port", strconv.Itoa(port),
			"--base-dir", req.BaseDir,
			"--no-open-chrome",
			"--no-agent-run",
		}
		err := browseragent.ServeWithContext(ctx, args, req.Env, io.Discard, &stderr)
		resp.Stderr = stderr.String()
		if err != nil {
			resp.CLIErr = err.Error()
			resp.ExitCode = 1
			low := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
			resp.ForeignHintSeen = strings.Contains(low, "not browser-agent") ||
				strings.Contains(low, "control port") ||
				strings.Contains(low, "--server-port") ||
				strings.Contains(low, "--port")
		}
		return resp, nil

	case ForeignPortOpEnsureDaemon:
		var stderr bytes.Buffer
		cfg := browseragent.SessionNewConfig{
			BaseDir:  req.BaseDir,
			Addr:     addr,
			SessionID: req.SessionIDA,
			NoOpenChrome: true,
			Stdout:   io.Discard,
			Stderr:   &stderr,
		}
		err := browseragent.SessionNew(cfg)
		resp.Stderr = stderr.String()
		if err != nil {
			resp.CLIErr = err.Error()
			resp.ExitCode = 1
			low := strings.ToLower(resp.Stderr + "\n" + resp.CLIErr)
			resp.ForeignHintSeen = strings.Contains(low, "not browser-agent") ||
				strings.Contains(low, "control port")
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown ForeignPortOp %q", req.ForeignPortOp)
	}
}

func runEnsureDaemonUpgradeMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.UpgradeOp == "" {
		t.Fatal("UpgradeOp must be set by leaf Setup")
	}
	port := pickFreePort(t)
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	var stderr bytes.Buffer
	daemonVer := req.DaemonVersion
	clientVer := req.ClientVersion
	if req.UpgradeOp == UpgradeOpReuseEqual {
		daemonVer = clientVer
	}
	if req.UpgradeOp == UpgradeOpWarnOlderClient {
		clientVer = "0.1.0"
		daemonVer = "0.2.0"
	}

	srv, cleanup, err := startDaemonAt(t, req, addr, daemonVer)
	if err != nil {
		return resp, err
	}
	resp.OldPID = srv.PID
	resp.BaseURL = srv.BaseURL

	switch req.UpgradeOp {
	case UpgradeOpBlockedConnected:
		if err := postCreateSessionHTTP(srv.BaseURL, req.SessionIDA); err != nil {
			cleanup()
			return resp, err
		}
		ext, err := dialFakeExtension(srv.BaseURL, req.SessionIDA, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			cleanup()
			return resp, err
		}
		if err := ext.SendHello(); err != nil {
			ext.Close()
			cleanup()
			return resp, err
		}
		go ext.Loop()
		t.Cleanup(ext.Close)
		time.Sleep(50 * time.Millisecond)
		resp.ExtensionConnected = true
	case UpgradeOpUpgradeWarnOrphans:
		if err := postCreateSessionHTTP(srv.BaseURL, req.OrphanID); err != nil {
			cleanup()
			return resp, err
		}
	}

	if req.UpgradeOp == UpgradeOpKillFailsHard {
		killCalled := false
		cfg := browseragent.EnsureDaemonConfig{
			BaseDir:       req.BaseDir,
			Addr:          addr,
			ClientVersion: clientVer,
			WaitTimeout:   req.ReadyTimeout,
			Stderr:        &stderr,
			KillFn: func(meta browseragent.DaemonMeta) error {
				killCalled = true
				return fmt.Errorf("injected kill failure")
			},
		}
		withClientVersion(clientVer, func() {
			_, err = browseragent.EnsureDaemon(cfg)
		})
		resp.KillFnCalled = killCalled
		cleanup()
		if err != nil {
			resp.CLIErr = err.Error()
		}
		return resp, err
	}

	if req.UpgradeOp == UpgradeOpRespawnFailsHard {
		spawnCalled := false
		cfg := browseragent.EnsureDaemonConfig{
			BaseDir:       req.BaseDir,
			Addr:          addr,
			ClientVersion: clientVer,
			WaitTimeout:   2 * time.Second,
			Stderr:        &stderr,
			SpawnFn: func() error {
				spawnCalled = true
				return nil
			},
		}
		withClientVersion(clientVer, func() {
			_, err = browseragent.EnsureDaemon(cfg)
		})
		resp.SpawnFnCalled = spawnCalled
		cleanup()
		if err != nil {
			resp.CLIErr = err.Error()
		}
		return resp, err
	}

	killCalled := false
	spawnCalled := false
	var newDaemonCancel context.CancelFunc

	cfg := browseragent.EnsureDaemonConfig{
		BaseDir:       req.BaseDir,
		Addr:          addr,
		ClientVersion: clientVer,
		WaitTimeout:   req.ReadyTimeout,
		Stderr:        &stderr,
		KillFn: func(meta browseragent.DaemonMeta) error {
			killCalled = true
			return browseragent.KillExistingDaemon(req.BaseDir, req.ShutdownWait)
		},
		SpawnFn: func() error {
			spawnCalled = true
			ctx, cancel := context.WithCancel(context.Background())
			newDaemonCancel = cancel
			go func() {
				_, _ = browseragent.RunDaemon(ctx, browseragent.DaemonConfig{
					Addr:          addr,
					BaseDir:       req.BaseDir,
					DaemonVersion: clientVer,
					Stdout:        io.Discard,
					Stderr:        io.Discard,
				})
			}()
			return nil
		},
	}

	var ensureErr error
	withClientVersion(clientVer, func() {
		if req.UpgradeOp == UpgradeOpBlockedConnected {
			var snOut bytes.Buffer
			ensureErr = browseragent.SessionNew(browseragent.SessionNewConfig{
				BaseDir:      req.BaseDir,
				Addr:         addr,
				SessionID:    req.SessionIDB,
				NoOpenChrome: true,
				Stdout:       &snOut,
				Stderr:       &stderr,
			})
			resp.SessionCreated = ensureErr == nil
		} else {
			meta, e := browseragent.EnsureDaemon(cfg)
			ensureErr = e
			if e == nil {
				resp.Meta = meta
				resp.NewPID = meta.PID
			}
		}
	})
	resp.KillFnCalled = killCalled
	resp.SpawnFnCalled = spawnCalled
	resp.Stderr = stderr.String()
	if ensureErr != nil {
		resp.CLIErr = ensureErr.Error()
		cleanup()
		return resp, ensureErr
	}

	if newDaemonCancel != nil {
		t.Cleanup(newDaemonCancel)
	} else {
		t.Cleanup(cleanup)
	}

	if req.UpgradeOp == UpgradeOpUpgradeWarnOrphans {
		resp.OrphanDirExists = dirExists(sessionDirPath(req.BaseDir, req.OrphanID))
	}

	healthRaw, _ := fetchHealthJSON(resp.BaseURL)
	if dv, ok := healthRaw["daemon_version"].(string); ok {
		resp.MetaDaemonVer = dv
	}

	return resp, nil
}

func runKillExistingMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.KillExistingOp == "" {
		t.Fatal("KillExistingOp must be set by leaf Setup")
	}
	port := pickFreePort(t)
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	daemonVer := req.DaemonVersion
	if req.KillExistingOp == KillExistingOpNoVersionCheck {
		daemonVer = req.ClientVersion
	}

	srv, cleanup, err := startDaemonAt(t, req, addr, daemonVer)
	if err != nil {
		return resp, err
	}
	defer cleanup()
	resp.BaseURL = srv.BaseURL
	resp.OldPID = srv.PID

	switch req.KillExistingOp {
	case KillExistingOpAlwaysKillsConnected:
		if err := postCreateSessionHTTP(srv.BaseURL, req.SessionIDA); err != nil {
			return resp, err
		}
		ext, err := dialFakeExtension(srv.BaseURL, req.SessionIDA, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			return resp, err
		}
		if err := ext.SendHello(); err != nil {
			ext.Close()
			return resp, err
		}
		go ext.Loop()
		defer ext.Close()
		time.Sleep(50 * time.Millisecond)
		resp.ExtensionConnected = true
	case KillExistingOpWarnOrphans:
		if err := postCreateSessionHTTP(srv.BaseURL, req.OrphanID); err != nil {
			return resp, err
		}
	}

	var stdout, stderr bytes.Buffer
	args := []string{
		"serve",
		"--kill-existing",
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--base-dir", req.BaseDir,
		"--no-open-chrome",
		"--no-agent-run",
	}
	done := make(chan error, 1)
	go func() {
		done <- browseragent.ServeWithContext(context.Background(), args, req.Env, &stdout, &stderr)
	}()

	if err := waitHealthDown(resp.BaseURL, req.ShutdownWait); err != nil {
		resp.Stderr = stderr.String()
		return resp, fmt.Errorf("first daemon still up: %w", err)
	}

	secondURL := "http://" + addr
	if err := waitHealthOK(secondURL, 5*time.Second); err != nil {
		resp.Stderr = stderr.String()
		return resp, fmt.Errorf("replacement daemon not healthy: %w", err)
	}
	_ = browseragent.ShutdownDaemon(secondURL, 5*time.Second)

	select {
	case err := <-done:
		if err != nil {
			resp.CLIErr = err.Error()
		}
	case <-time.After(req.ShutdownWait):
	}

	resp.Stderr = stderr.String()
	resp.Stdout = stdout.String()
	resp.DaemonStopped = true

	if req.KillExistingOp == KillExistingOpWarnOrphans {
		resp.OrphanDirExists = dirExists(sessionDirPath(req.BaseDir, req.OrphanID))
	}
	return resp, nil
}

func runServeStopMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.ServeStopOp == "" {
		t.Fatal("ServeStopOp must be set by leaf Setup")
	}
	port := pickFreePort(t)
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	srv, cleanup, err := startDaemonAt(t, req, addr, req.DaemonVersion)
	if err != nil {
		return resp, err
	}
	resp.BaseURL = srv.BaseURL

	if req.ServeStopOp == ServeStopOpNonTTYWarnConnected || req.ServeStopOp == ServeStopOpTTYDefaultYes || req.ServeStopOp == ServeStopOpTTYNAborts {
		if err := postCreateSessionHTTP(srv.BaseURL, req.SessionIDA); err != nil {
			cleanup()
			return resp, err
		}
		ext, err := dialFakeExtension(srv.BaseURL, req.SessionIDA, req.HelloVersion, req.HelloFeatures)
		if err != nil {
			cleanup()
			return resp, err
		}
		if err := ext.SendHello(); err != nil {
			ext.Close()
			cleanup()
			return resp, err
		}
		go ext.Loop()
		defer ext.Close()
		time.Sleep(50 * time.Millisecond)
		resp.ExtensionConnected = true
	}

	prevTTY := inj.IsTerminalFn
	if req.IsTTY {
		inj.IsTerminalFn = func(r io.Reader) bool { return true }
	} else if req.ServeStopOp == ServeStopOpNonTTYWarnConnected {
		inj.IsTerminalFn = func(r io.Reader) bool { return false }
	}
	defer func() { inj.IsTerminalFn = prevTTY }()

	var stdout, stderr bytes.Buffer
	args := []string{
		"serve",
		"--stop",
		"--base-dir", req.BaseDir,
		"--host", "127.0.0.1",
		"--port", strconv.Itoa(port),
	}
	cliErr := browseragent.HandleCLIWithStdin(args, req.Env, strings.NewReader(req.Stdin), &stdout, &stderr)
	resp.Stdout = stdout.String()
	resp.Stderr = stderr.String()
	if cliErr != nil {
		resp.CLIErr = cliErr.Error()
	}

	low := strings.ToLower(resp.Stderr)
	resp.TTYConfirmSeen = strings.Contains(low, "[y/n]") || strings.Contains(low, "stop daemon")

	stillUp := healthOK(resp.BaseURL)
	resp.DaemonStillUp = stillUp
	resp.DaemonStopped = !stillUp
	resp.MetaExists = fileExists(daemonMetaPath(req.BaseDir))

	if req.ServeStopOp != ServeStopOpTTYNAborts {
		cleanup()
	}
	return resp, nil
}

func runFlagsMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.FlagsOp == "" {
		t.Fatal("FlagsOp must be set by leaf Setup")
	}
	resp := &Response{}

	switch req.FlagsOp {
	case FlagsOpServeHelpHostPort:
		var stdout bytes.Buffer
		err := browseragent.HandleCLI([]string{"serve", "--help"}, req.Env, &stdout, &bytes.Buffer{})
		resp.HelpText = stdout.String()
		if err != nil {
			resp.CLIErr = err.Error()
		}
		return resp, nil

	case FlagsOpSessionHelpServerPort:
		var stdout bytes.Buffer
		err := browseragent.HandleCLI([]string{"session", "new", "--help"}, req.Env, &stdout, &bytes.Buffer{})
		resp.HelpText = stdout.String()
		if err != nil {
			resp.CLIErr = err.Error()
		}
		return resp, nil

	case FlagsOpResolveServerJSONPort:
		port := pickFreePort(t)
		addr := loopbackAddr(port)
		_, cleanup, err := startDaemonAt(t, req, addr, req.DaemonVersion)
		if err != nil {
			return resp, err
		}
		defer cleanup()
		resp.Addr = addr
		baseURL, err := browseragent.ResolveControlBaseURLWithHostPort(req.BaseDir, "", 0)
		if err != nil {
			resp.CLIErr = err.Error()
			return resp, err
		}
		resp.ResolveBaseURL = baseURL
		return resp, nil

	default:
		return nil, fmt.Errorf("unknown FlagsOp %q", req.FlagsOp)
	}
}

func runRegressionMode(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	if req.RegressionOp == "" {
		t.Fatal("RegressionOp must be set by leaf Setup")
	}
	port := pickFreePort(t)
	addr := loopbackAddr(port)
	resp := &Response{Addr: addr, Port: port}

	spawnCalled := false
	var daemonCancel context.CancelFunc
	cfg := browseragent.EnsureDaemonConfig{
		BaseDir:     req.BaseDir,
		Addr:        addr,
		WaitTimeout: req.ReadyTimeout,
		SpawnFn: func() error {
			spawnCalled = true
			ctx, cancel := context.WithCancel(context.Background())
			daemonCancel = cancel
			go func() {
				_, _ = browseragent.RunDaemon(ctx, browseragent.DaemonConfig{
					Addr:    addr,
					BaseDir: req.BaseDir,
					Stdout:  io.Discard,
					Stderr:  io.Discard,
				})
			}()
			return nil
		},
	}
	meta, err := browseragent.EnsureDaemon(cfg)
	resp.SpawnFnCalled = spawnCalled
	if err != nil {
		resp.CLIErr = err.Error()
		return resp, err
	}
	resp.Meta = meta
	resp.BaseURL = "http://" + addr
	if daemonCancel != nil {
		t.Cleanup(daemonCancel)
	}
	return resp, nil
}

// --- daemon + fake extension harness ---

type daemonFixture struct {
	BaseURL string
	Addr    string
	PID     int
	cancel  context.CancelFunc
}

func startDaemonAt(t *testing.T, req *Request, addr, daemonVersion string) (*daemonFixture, func(), error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.DaemonConfig{
		Addr:          addr,
		BaseDir:       req.BaseDir,
		DaemonVersion: daemonVersion,
		Stdout:        io.Discard,
		Stderr:        io.Discard,
	}
	done := make(chan error, 1)
	go func() {
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()
	baseURL := "http://" + addr
	if err := waitHealthOK(baseURL, req.ReadyTimeout); err != nil {
		cancel()
		<-done
		return nil, nil, err
	}
	meta, _ := browseragent.ReadDaemonMeta(daemonMetaPath(req.BaseDir))
	fix := &daemonFixture{
		BaseURL: baseURL,
		Addr:    addr,
		PID:     meta.PID,
		cancel:  cancel,
	}
	cleanup := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
	return fix, cleanup, nil
}

func startForeignHTTP(t *testing.T, addr string) func() {
	t.Helper()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("foreign listen %s: %v", addr, err)
	}
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("foreign-not-browser-agent"))
		}),
	}
	go func() { _ = srv.Serve(ln) }()
	return func() {
		_ = srv.Close()
		_ = ln.Close()
	}
}

func withClientVersion(ver string, fn func()) {
	prev := inj.ClientVersionOverride
	inj.ClientVersionOverride = func() string { return ver }
	defer func() { inj.ClientVersionOverride = prev }()
	fn()
}

func readVERSIONTxt(moduleRoot string) string {
	data, err := os.ReadFile(filepath.Join(moduleRoot, "VERSION.txt"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func postCreateSessionHTTP(baseURL, sessionID string) error {
	body := map[string]string{"session_id": sessionID}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	u := strings.TrimRight(baseURL, "/") + "/v1/sessions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST /v1/sessions status=%d body=%s", res.StatusCode, strings.TrimSpace(string(out)))
	}
	return nil
}

func fetchHealthJSON(baseURL string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/v1/health", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func metaDaemonVersion(meta browseragent.DaemonMeta) string {
	b, _ := json.Marshal(meta)
	var raw map[string]any
	_ = json.Unmarshal(b, &raw)
	if v, ok := raw["daemon_version"].(string); ok {
		return v
	}
	return ""
}

func waitHealthOK(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if healthOK(baseURL) {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("health not ok at %s within %v", baseURL, timeout)
}

func waitHealthDown(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !healthOK(baseURL) {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("health still up at %s", baseURL)
}

func healthOK(baseURL string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(baseURL, "/")+"/v1/health", nil)
	if err != nil {
		return false
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
	return res.StatusCode == http.StatusOK
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// fake extension (phase4 pattern)

type wsEnvelope struct {
	V       int            `json:"v"`
	Type    string         `json:"type"`
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

type fakeExtension struct {
	conn     *websocket.Conn
	version  string
	features []string
	OnJob    func(wsEnvelope)
	mu       sync.Mutex
	closed   bool
}

func dialFakeExtension(baseURL, sessionID, version string, features []string) (*fakeExtension, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	u.Scheme = "ws"
	u.Path = "/v1/ws"
	if sessionID != "" {
		q := u.Query()
		q.Set("session", sessionID)
		u.RawQuery = q.Encode()
	}
	dialer := websocket.Dialer{HandshakeTimeout: 3 * time.Second}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	return &fakeExtension{conn: conn, version: version, features: features}, nil
}

func (f *fakeExtension) SendHello() error {
	env := wsEnvelope{
		V:    1,
		Type: "hello",
		ID:   fmt.Sprintf("hello-%d", time.Now().UnixNano()),
		Payload: map[string]any{
			"version":  f.version,
			"features": f.features,
		},
	}
	return f.conn.WriteJSON(env)
}

func (f *fakeExtension) Loop() {
	for {
		f.mu.Lock()
		closed := f.closed
		f.mu.Unlock()
		if closed {
			return
		}
		var env wsEnvelope
		if err := f.conn.ReadJSON(&env); err != nil {
			return
		}
		if env.Type == "job" && f.OnJob != nil {
			f.OnJob(env)
		}
	}
}

func (f *fakeExtension) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	_ = f.conn.Close()
}

var (
	_ = httptest.NewServer
	_ = sync.Mutex{}
)
```