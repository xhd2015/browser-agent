package browseragent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const briefUsage = `Usage: browser-agent <command> [flags]

Commands:
  serve       Start the browser-agent control server
  session     Session side-commands: info|eval|run|logs|screenshot|cdp
  skill       Show/list/install the embedded agent skill
  install-chrome-extension   Extract embedded Chrome extension

Run 'browser-agent --help' for full help.
`

const fullHelp = `Usage: browser-agent <command> [flags]

browser-agent is the operator CLI for the browser-agent control plane.

Commands:
  serve [flags]              Start control server (default addr 127.0.0.1:43761)
  session <cmd> [flags]      Nested session side-commands (require session id)
    session info [flags]               Print session snapshot JSON
    session eval [flags] <expr>        POST an eval job and print the result
    session run [flags] <path.js>      Read a JS file and POST a run job
    session logs [flags]               POST a logs job (optional --limit N)
    session screenshot [flags]         POST a screenshot job; optional -o file.png
    session cdp [flags] <Method> [json]
                                       POST a raw CDP job (method + optional params JSON)
  install-chrome-extension   Extract embedded extension and print Load unpacked help
  skill --list|--show|--install …
                             Embedded agent skill (see: browser-agent skill --help)

Global flags:
  -h, --help                 Show this help

serve flags:
  --addr <host:port>         Listen address (default: 127.0.0.1:43761)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-agent)
  --session-id <id>          Fixed session id
  --no-open-chrome           Do not launch Chrome
  --no-agent-run             Do not launch agent-run

session info / eval / run / logs / screenshot / cdp flags:
  --session-id <id>          Session id (or env BROWSER_AGENT_SESSION_ID)
  --addr <url|host:port>     Control server base (default: http://127.0.0.1:43761)

screenshot flags:
  -o, --output <file.png>    Write decoded PNG (from result base64) to path

logs flags:
  --limit <N>                Optional max log entries
  --level <level>            Optional log level filter

Session resolution: --session-id flag, else BROWSER_AGENT_SESSION_ID env.
`

// HandleCLI dispatches browser-agent CLI args (after the binary name).
// When env != nil, session id is resolved only from that map (no process env
// fallback). When env == nil, process env may be used.
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if args == nil {
		args = []string{}
	}

	// bare → brief usage + error
	if len(args) == 0 {
		_, _ = io.WriteString(stdout, briefUsage)
		if !strings.HasSuffix(briefUsage, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return fmt.Errorf("missing command; see usage")
	}

	// help
	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "serve":
		return cliServe(rest, env, stdout, stderr)
	case "session":
		return cliSession(rest, env, stdout, stderr)
	case "install-chrome-extension":
		return cliInstallExt(rest, env, stdout, stderr)
	case "skill":
		return cliSkill(rest, env, stdout, stderr)
	case "-h", "--help":
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	default:
		// Flat side-commands (info/eval/…) are not handlers after the nested refactor.
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("unknown command %q; try serve, session, install-chrome-extension, or skill", cmd)
	}
}

// cliSession dispatches nested session side-commands:
// session info|eval|run|logs|screenshot|cdp …
func cliSession(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("session requires a subcommand: info|eval|run|logs|screenshot|cdp")
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "info":
		return cliInfo(rest, env, stdout, stderr)
	case "eval":
		return cliEval(rest, env, stdout, stderr)
	case "run":
		return cliRun(rest, env, stdout, stderr)
	case "logs":
		return cliLogs(rest, env, stdout, stderr)
	case "screenshot":
		return cliScreenshot(rest, env, stdout, stderr)
	case "cdp":
		return cliCDP(rest, env, stdout, stderr)
	case "-h", "--help":
		_, _ = io.WriteString(stdout, fullHelp)
		if !strings.HasSuffix(fullHelp, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
		return nil
	default:
		_, _ = io.WriteString(stderr, briefUsage)
		return fmt.Errorf("unknown session subcommand %q; try info, eval, run, logs, screenshot, or cdp", sub)
	}
}

func cliInstallExt(args []string, env map[string]string, stdout, stderr io.Writer) error {
	baseDir := flagString(args, "--base-dir")
	return InstallChromeExtension(stdout, baseDir)
}

func cliServe(args []string, env map[string]string, stdout, stderr io.Writer) error {
	// Manual flag parse (no os.Exit).
	addr := flagString(args, "--addr")
	baseDir := flagString(args, "--base-dir")
	sessionID := flagString(args, "--session-id")
	noOpenChrome := flagBool(args, "--no-open-chrome")
	noAgentRun := flagBool(args, "--no-agent-run")

	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}

	if sessionID == "" {
		// serve can generate, but Run requires SessionID — pick one.
		sessionID = fmt.Sprintf("sess-%d", time.Now().UnixNano()%1e12)
	}
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		baseDir = home + "/.tmp/browser-agent"
	}

	cfg := Config{
		Addr:         addr,
		BaseDir:      baseDir,
		SessionID:    sessionID,
		NoOpenChrome: noOpenChrome,
		NoAgentRun:   noAgentRun,
		Stdout:       stdout,
		Stderr:       stderr,
	}
	ctx := context.Background()
	// Note: real CLI would use signal.NotifyContext; package API blocks until cancel.
	// For binary main, main wraps with signal context via Run directly or cancel.
	// Here we use a context that never cancels unless process ends — binary uses Run.
	_, err := Run(ctx, cfg)
	return err
}

func cliInfo(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	base := normalizeAddr(flagString(args, "--addr"))
	u := strings.TrimRight(base, "/") + "/v1/session?session=" + url.QueryEscape(sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("info request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("info failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	// Build control + optional browser inventory view.
	outObj := map[string]any{}
	if json.Valid(body) {
		if err := json.Unmarshal(body, &outObj); err != nil {
			// Fall back to raw pretty body below.
			outObj = nil
		}
	} else {
		outObj = nil
	}

	if outObj != nil {
		connected := extensionConnectedFromSnapshot(outObj)
		if connected {
			browser, berr := fetchInfoJobBrowser(base, sessionID, 15000)
			if berr != nil {
				outObj["browser"] = nil
				outObj["browser_error"] = berr.Error()
			} else {
				outObj["browser"] = browser
				delete(outObj, "browser_error")
			}
		} else {
			// Never fabricate tabs when extension is disconnected.
			outObj["browser"] = nil
			outObj["browser_error"] = "extension not connected"
		}
		pretty, err := json.MarshalIndent(outObj, "", "  ")
		if err != nil {
			return err
		}
		out := string(pretty)
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		_, err = io.WriteString(stdout, out)
		return err
	}

	// Non-JSON control response: pass through pretty if possible.
	var buf bytes.Buffer
	if json.Valid(body) {
		_ = json.Indent(&buf, body, "", "  ")
		if buf.Len() == 0 {
			buf.Write(body)
		}
	} else {
		buf.Write(body)
	}
	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err = io.WriteString(stdout, out)
	return err
}

// extensionConnectedFromSnapshot reads extension.connected (preferred) or
// top-level connected from a GET /v1/session JSON object.
func extensionConnectedFromSnapshot(m map[string]any) bool {
	if m == nil {
		return false
	}
	if ext, ok := m["extension"].(map[string]any); ok {
		if c, ok := ext["connected"].(bool); ok {
			return c
		}
	}
	if c, ok := m["connected"].(bool); ok {
		return c
	}
	return false
}

// fetchInfoJobBrowser POSTs an info job and returns a browser map suitable for
// merging into session info stdout (tabs, version, features when present).
// On failure returns a non-nil error; caller should still print control data.
func fetchInfoJobBrowser(base, sessionID string, timeoutMS int64) (map[string]any, error) {
	if timeoutMS <= 0 {
		timeoutMS = 15000
	}
	u := strings.TrimRight(base, "/") + "/v1/jobs"
	payload := map[string]any{
		"session_id": sessionID,
		"type":       JobTypeInfo,
		"params":     map[string]any{},
		"timeout_ms": timeoutMS,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	httpTO := time.Duration(timeoutMS)*time.Millisecond + 5*time.Second
	if httpTO < 20*time.Second {
		httpTO = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpTO)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("info job request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("info job failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("info job: invalid JSON: %w", err)
	}
	if ok, _ := result["ok"].(bool); !ok {
		errMsg, _ := result["error"].(string)
		if errMsg == "" {
			errMsg = "info job failed"
		}
		return nil, fmt.Errorf("%s", errMsg)
	}
	// Prefer result.data as browser payload; otherwise use whole result sans meta.
	browser := map[string]any{}
	if data, ok := result["data"].(map[string]any); ok && data != nil {
		for k, v := range data {
			browser[k] = v
		}
	} else {
		// Fall back: surface tabs/version/features if top-level.
		for _, k := range []string{"tabs", "version", "features"} {
			if v, ok := result[k]; ok {
				browser[k] = v
			}
		}
	}
	return browser, nil
}

func cliEval(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	expr := takePositional(args, 0)
	if strings.TrimSpace(expr) == "" {
		return fmt.Errorf("eval requires an expression argument")
	}
	return postJobAndPrint(args, sessionID, JobTypeEval, map[string]any{
		"expression": expr,
		"expr":       expr,
	}, 15000, stdout, nil)
}

func cliRun(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	path := takePositional(args, 0)
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("run requires a .js file path argument")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		// try absolute from cwd-relative
		abs, aerr := filepath.Abs(path)
		if aerr == nil {
			data, err = os.ReadFile(abs)
			if err == nil {
				path = abs
			}
		}
		if err != nil {
			return fmt.Errorf("run: read %s: %w", path, err)
		}
	}
	source := string(data)
	return postJobAndPrint(args, sessionID, JobTypeRun, map[string]any{
		"source":     source,
		"expression": source,
		"path":       path,
	}, 30000, stdout, nil)
}

func cliLogs(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	params := map[string]any{}
	if lim := flagString(args, "--limit"); lim != "" {
		if n, err := strconv.Atoi(lim); err == nil {
			params["limit"] = n
		} else {
			params["limit"] = lim
		}
	}
	if level := flagString(args, "--level"); level != "" {
		params["level"] = level
	}
	return postJobAndPrint(args, sessionID, JobTypeLogs, params, 15000, stdout, nil)
}

func cliScreenshot(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	params := map[string]any{
		"format":    "png",
		"full_page": false,
	}
	outPath := flagString(args, "-o")
	if outPath == "" {
		outPath = flagString(args, "--output")
	}
	return postJobAndPrint(args, sessionID, JobTypeScreenshot, params, 20000, stdout, func(result map[string]any) error {
		if outPath == "" {
			return nil
		}
		b64 := extractBase64(result)
		if b64 == "" {
			return nil
		}
		// Strip data-url prefix if present.
		if i := strings.Index(b64, ","); i >= 0 && strings.Contains(b64[:i], "base64") {
			b64 = b64[i+1:]
		}
		raw, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			// try raw std without padding issues — still best-effort
			raw, err = base64.RawStdEncoding.DecodeString(b64)
			if err != nil {
				return fmt.Errorf("screenshot: decode base64 for -o: %w", err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil && filepath.Dir(outPath) != "." {
			return err
		}
		return os.WriteFile(outPath, raw, 0o644)
	})
}

func cliCDP(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, fullHelp)
		return nil
	}
	sessionID, err := resolveCLISession(args, env)
	if err != nil {
		return err
	}
	method := takePositional(args, 0)
	if strings.TrimSpace(method) == "" {
		return fmt.Errorf("cdp requires a Method argument (e.g. Page.navigate)")
	}
	paramsJSON := takePositional(args, 1)
	params := map[string]any{
		"method": method,
	}
	if strings.TrimSpace(paramsJSON) != "" {
		var nested any
		if err := json.Unmarshal([]byte(paramsJSON), &nested); err != nil {
			return fmt.Errorf("cdp: invalid json params: %w", err)
		}
		params["params"] = nested
	}
	return postJobAndPrint(args, sessionID, JobTypeCDP, params, 30000, stdout, nil)
}

// postJobAndPrint POSTs /v1/jobs and writes the JSON result to stdout with trailing \n.
// afterOK is optional post-processing of a successful JSON result (e.g. write screenshot file).
func postJobAndPrint(args []string, sessionID, jobType string, params map[string]any, timeoutMS int64, stdout io.Writer, afterOK func(map[string]any) error) error {
	base := normalizeAddr(flagString(args, "--addr"))
	u := strings.TrimRight(base, "/") + "/v1/jobs"

	if params == nil {
		params = map[string]any{}
	}
	payload := map[string]any{
		"session_id": sessionID,
		"type":       jobType,
		"params":     params,
		"timeout_ms": timeoutMS,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// HTTP client timeout slightly above job wait.
	httpTO := time.Duration(timeoutMS)*time.Millisecond + 5*time.Second
	if httpTO < 20*time.Second {
		httpTO = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpTO)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s request failed: %w", jobType, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s failed: status %d: %s", jobType, res.StatusCode, strings.TrimSpace(string(body)))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err == nil {
		if ok, _ := result["ok"].(bool); !ok {
			if errMsg, _ := result["error"].(string); errMsg != "" {
				return fmt.Errorf("%s job failed: %s", jobType, errMsg)
			}
		}
		if afterOK != nil {
			if err := afterOK(result); err != nil {
				return err
			}
		}
		enc, _ := json.Marshal(result)
		out := string(enc)
		if !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		_, err = io.WriteString(stdout, out)
		return err
	}
	out := string(body)
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	_, err = io.WriteString(stdout, out)
	return err
}

func extractBase64(result map[string]any) string {
	if result == nil {
		return ""
	}
	if s, ok := result["base64"].(string); ok && s != "" {
		return s
	}
	if data, ok := result["data"].(map[string]any); ok {
		if s, ok := data["base64"].(string); ok {
			return s
		}
		if s, ok := data["data"].(string); ok {
			return s
		}
	}
	return ""
}

func resolveCLISession(args []string, env map[string]string) (string, error) {
	flagVal, flagSet := flagStringSet(args, "--session-id")
	envVal, envSet := "", false
	if env != nil {
		// Injectable env: only use keys present in the map; missing = unset.
		if v, ok := env["BROWSER_AGENT_SESSION_ID"]; ok {
			envVal = v
			envSet = true
		}
	} else {
		// Process env fallback when env map is nil.
		if v, ok := os.LookupEnv("BROWSER_AGENT_SESSION_ID"); ok {
			envVal = v
			envSet = true
		}
	}
	return ResolveSessionID(flagVal, flagSet, envVal, envSet)
}

func normalizeAddr(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "http://127.0.0.1:43761"
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	// host:port form
	return "http://" + addr
}

func flagString(args []string, name string) string {
	v, _ := flagStringSet(args, name)
	return v
}

func flagStringSet(args []string, name string) (string, bool) {
	eqPrefix := name + "="
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == name {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", true
		}
		if strings.HasPrefix(a, eqPrefix) {
			return strings.TrimPrefix(a, eqPrefix), true
		}
	}
	return "", false
}

func flagBool(args []string, name string) bool {
	for _, a := range args {
		if a == name || a == name+"=true" {
			return true
		}
		if a == name+"=false" {
			return false
		}
	}
	return false
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}

// takePositional returns the n-th non-flag positional argument (0-based).
func takePositional(args []string, n int) string {
	skipNext := false
	idx := 0
	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}
		a := args[i]
		if a == "--session-id" || a == "--addr" || a == "--base-dir" ||
			a == "--limit" || a == "--level" || a == "--output" || a == "-o" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "--") {
			// --flag=value or bare boolean
			continue
		}
		// Short flags with value: -o path handled above when a == "-o".
		// Bare -o=file style:
		if strings.HasPrefix(a, "-o=") {
			continue
		}
		if strings.HasPrefix(a, "-") && a != "-" {
			// unknown short flag without value consumption
			continue
		}
		if idx == n {
			return a
		}
		idx++
	}
	return ""
}

// takeEvalExpr returns the first non-flag positional argument after "eval".
// Kept for backward compatibility with callers.
func takeEvalExpr(args []string) string {
	return takePositional(args, 0)
}
