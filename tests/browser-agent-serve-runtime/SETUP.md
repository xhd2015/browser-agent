# Scenario

**Feature**: browser-agent serve runtime — meta/SYSTEM, launch hooks, agent-run args, extension WS source

```
# Serve with injectable launch hooks (no real Chrome / agent-run)
Test Client -> browseragent.Run(Config{OpenChromeFn, AgentRunFn, No*}) -> Control Server
Serve Runtime -> extract extension -> write SYSTEM.md + meta.json
Serve Runtime -> OpenChromeFn(sessionURL, extPath)   # unless NoOpenChrome
Serve Runtime -> AgentRunFn(sessionID, SYSTEM.md, ws, env)  # unless NoAgentRun

# Pure argv — prefixed session-id + --env control + always --no-submit
Test Client -> BuildAgentRunArgs(control, prompt, workspace) -> []string

# Extension sources (filesystem)
Test Client -> read Chrome-Ext-Browser-Agent/** + browseragent/embedded/extension/**
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Tree root is `tests/browser-agent-serve-runtime/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- Package `browseragent` will export (TDD red until implemented):
  - `Config.OpenChromeFn`, `Config.AgentRunFn`, optional `Config.WorkspaceDir`
  - `BuildAgentRunArgs(sessionID, promptOrSystemPath, workspaceDir string) []string`
  - serve writes `meta.json` + extracts extension; honors launch flags/hooks
  - optional `extension_install_path` on `GET /v1/session`
- Each server leaf uses isolated temp `BaseDir` and free loopback `Addr`.
- No real Chrome; no real agent-run; hooks only record.
- Sealed `tests/browser-agent/` and `tests/browser-agent-cli-react/` must not be modified.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate a unique temp `BaseDir` for leaves that extract or serve.
3. Default `SessionID`, short `ReadyTimeout`, short `HookSettle`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Parallel-safe: temp dirs + free ports per leaf.
- Shared helpers below available to all descendant Assert/Setup packages.
- Prefer package APIs over building `cmd/browser-agent` binary.

```go
import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-rt-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.HookSettle == 0 {
		req.HookSettle = 120 * time.Millisecond
	}
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertHTTPStatus(t *testing.T, resp *Response, want int) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != want {
		t.Fatalf("HTTP status = %d, want %d; content-type=%q body=%s",
			resp.StatusCode, want, resp.ContentType, truncate(resp.BodyString, 400))
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertSystemMDRecipes(t *testing.T, text, sessionID string) {
	t.Helper()
	if strings.TrimSpace(text) == "" {
		t.Fatal("SYSTEM.md text empty")
	}
	// Nested complete refactor: do NOT embed concrete control session id.
	if sessionID != "" && strings.Contains(text, sessionID) {
		t.Fatalf("SYSTEM.md must not embed control session id %q; text=%s", sessionID, truncate(text, 400))
	}
	needles := []string{
		"browser-agent session info",
		"browser-agent session eval",
		"browser-agent session run",
		"browser-agent session logs",
		"browser-agent session screenshot",
		"BROWSER_AGENT_SESSION_ID",
	}
	for _, n := range needles {
		if !strings.Contains(text, n) {
			t.Fatalf("SYSTEM.md missing recipe/marker %q; text=%s", n, truncate(text, 600))
		}
	}
}

func assertMetaCore(t *testing.T, resp *Response, sessionID string) {
	t.Helper()
	if strings.TrimSpace(resp.MetaJSON) == "" {
		t.Fatalf("meta.json missing or empty at %s", resp.MetaPath)
	}
	if resp.Meta == nil {
		t.Fatalf("meta.json did not parse; raw=%s", truncate(resp.MetaJSON, 400))
	}
	sid := stringField(resp.Meta, "session_id", "sessionId")
	if sid == "" {
		t.Fatalf("meta.json missing session_id; raw=%s", truncate(resp.MetaJSON, 400))
	}
	if sessionID != "" && sid != sessionID {
		t.Fatalf("meta session_id=%q, want %q", sid, sessionID)
	}
	product := stringField(resp.Meta, "product")
	if product != "" && product != "browser-agent" {
		t.Fatalf("meta product=%q, want browser-agent", product)
	}
	if product == "" {
		// product required by contract
		t.Fatalf("meta.json missing product=browser-agent; raw=%s", truncate(resp.MetaJSON, 400))
	}
	baseURL := stringField(resp.Meta, "base_url", "baseUrl")
	sessionURL := stringField(resp.Meta, "session_url", "sessionUrl")
	if baseURL == "" && sessionURL == "" {
		t.Fatalf("meta.json needs base_url and/or session_url; raw=%s", truncate(resp.MetaJSON, 400))
	}
	if sessionURL != "" {
		if !strings.Contains(sessionURL, "/go") {
			t.Fatalf("session_url should contain /go; got %q", sessionURL)
		}
		if sessionID != "" && !strings.Contains(sessionURL, sessionID) {
			t.Fatalf("session_url should contain session id %q; got %q", sessionID, sessionURL)
		}
	}
	if baseURL != "" {
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			t.Fatalf("base_url should be absolute URL; got %q", baseURL)
		}
	}
}

func assertAgentRunArgsCore(t *testing.T, args []string, sessionID string, wantDir bool) {
	t.Helper()
	if len(args) == 0 {
		t.Fatal("BuildAgentRunArgs returned empty argv")
	}
	joined := strings.Join(args, " ")
	// run subcommand
	hasRun := false
	for _, a := range args {
		if a == "run" {
			hasRun = true
			break
		}
	}
	if !hasRun {
		t.Fatalf("argv missing run; args=%v", args)
	}

	// agent-run id = browser-agent-sess-<control> (idempotent if already prefixed)
	const agentPrefix = "browser-agent-sess-"
	wantAgentID := sessionID
	if sessionID != "" && !strings.HasPrefix(sessionID, agentPrefix) {
		wantAgentID = agentPrefix + sessionID
	}
	sidVal := ""
	for i, a := range args {
		if strings.HasPrefix(a, "--session-id=") {
			sidVal = strings.TrimPrefix(a, "--session-id=")
			break
		}
		if a == "--session-id" && i+1 < len(args) {
			sidVal = args[i+1]
			break
		}
	}
	if sidVal == "" {
		t.Fatalf("argv missing --session-id; args=%v", args)
	}
	if wantAgentID != "" && sidVal != wantAgentID {
		t.Fatalf("--session-id=%q, want agent-run id %q; args=%v", sidVal, wantAgentID, args)
	}

	// --env BROWSER_AGENT_SESSION_ID=<control>
	envVal := ""
	for i, a := range args {
		if a == "--env" && i+1 < len(args) {
			v := args[i+1]
			if strings.HasPrefix(v, "BROWSER_AGENT_SESSION_ID=") {
				envVal = strings.TrimPrefix(v, "BROWSER_AGENT_SESSION_ID=")
				break
			}
			if v == "BROWSER_AGENT_SESSION_ID" && i+2 < len(args) {
				envVal = args[i+2]
				break
			}
		}
		if strings.HasPrefix(a, "--env=") {
			v := strings.TrimPrefix(a, "--env=")
			if strings.HasPrefix(v, "BROWSER_AGENT_SESSION_ID=") {
				envVal = strings.TrimPrefix(v, "BROWSER_AGENT_SESSION_ID=")
				break
			}
		}
	}
	if sessionID != "" {
		if envVal == "" {
			t.Fatalf("argv missing --env BROWSER_AGENT_SESSION_ID=<control>; args=%v", args)
		}
		if envVal != sessionID {
			t.Fatalf("--env BROWSER_AGENT_SESSION_ID=%q, want control %q; args=%v", envVal, sessionID, args)
		}
	}

	// agent-runner=grok-tty
	if !strings.Contains(joined, "grok-tty") {
		t.Fatalf("argv missing grok-tty agent-runner; args=%v", args)
	}
	if !argvHasToken(args, "--auto-send-or-resume") && !strings.Contains(joined, "auto-send-or-resume") {
		t.Fatalf("argv missing --auto-send-or-resume; args=%v", args)
	}
	if !argvHasToken(args, "--new-terminal") && !strings.Contains(joined, "new-terminal") {
		t.Fatalf("argv missing --new-terminal; args=%v", args)
	}
	// --no-submit is ALWAYS required so the first prompt remains draft (no auto-submit)
	if !argvHasToken(args, "--no-submit") {
		// also accept --no-submit=… forms if ever used
		foundNoSubmit := false
		for _, a := range args {
			if a == "--no-submit" || strings.HasPrefix(a, "--no-submit=") {
				foundNoSubmit = true
				break
			}
		}
		if !foundNoSubmit {
			t.Fatalf("argv missing --no-submit (always required for draft open); args=%v", args)
		}
	}
	if !argvHasToken(args, "--open") && !strings.Contains(joined, "--open") {
		// --open may appear as standalone flag
		foundOpen := false
		for _, a := range args {
			if a == "--open" {
				foundOpen = true
				break
			}
		}
		if !foundOpen {
			t.Fatalf("argv missing --open; args=%v", args)
		}
	}
	hasDir := argvHasDirFlag(args)
	if wantDir && !hasDir {
		t.Fatalf("expected --dir when workspace set; args=%v", args)
	}
	if !wantDir && hasDir {
		t.Fatalf("expected no --dir when workspace empty; args=%v", args)
	}
}

func assertWSAgentTokens(t *testing.T, text string, label string) {
	t.Helper()
	if strings.TrimSpace(text) == "" {
		t.Fatalf("%s source empty", label)
	}
	// Must mention WS control path or ws:// scheme
	if !strings.Contains(text, "/v1/ws") && !strings.Contains(text, "ws://") {
		t.Fatalf("%s must reference /v1/ws or ws://; text=%s", label, truncate(text, 500))
	}
	low := strings.ToLower(text)
	if !strings.Contains(low, "hello") {
		t.Fatalf("%s must send/handle hello; text=%s", label, truncate(text, 500))
	}
	// job and result protocol tokens
	if !strings.Contains(low, "job") {
		t.Fatalf("%s must handle job messages; text=%s", label, truncate(text, 500))
	}
	if !strings.Contains(low, "result") {
		t.Fatalf("%s must send result messages; text=%s", label, truncate(text, 500))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// silence unused in some leaves
var (
	_ = http.StatusOK
)
```
