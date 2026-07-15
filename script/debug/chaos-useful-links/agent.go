package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// AgentCLI shells out to browser-agent session side-commands.
type AgentCLI struct {
	Bin       string
	SessionID string
	Timeout   time.Duration
}

func (a *AgentCLI) run(ctx context.Context, args ...string) (stdout, stderr string, err error) {
	if a.Bin == "" {
		return "", "", fmt.Errorf("browser-agent binary not set")
	}
	if a.Timeout <= 0 {
		a.Timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, a.Timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, a.Bin, args...)
	cmd.Env = append(os.Environ(), "BROWSER_AGENT_SESSION_ID="+a.SessionID)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()
	if cctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, fmt.Errorf("timeout after %s: %w", a.Timeout, cctx.Err())
	}
	if runErr != nil {
		msg := strings.TrimSpace(stderr)
		if msg == "" {
			msg = runErr.Error()
		}
		return stdout, stderr, fmt.Errorf("%s", msg)
	}
	return stdout, stderr, nil
}

func (a *AgentCLI) sessionNew(ctx context.Context) (sessionID string, raw string, err error) {
	// session new does not need prior session id.
	if a.Timeout <= 0 {
		a.Timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, a.Timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, a.Bin, "session", "new")
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", outBuf.String() + errBuf.String(), fmt.Errorf("session new: %v: %s", err, strings.TrimSpace(errBuf.String()))
	}
	raw = outBuf.String()
	// Parse "session-id: sess-xxx" or export line.
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "session-id:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "session-id:")), raw, nil
		}
		if strings.HasPrefix(line, "export BROWSER_AGENT_SESSION_ID=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "export BROWSER_AGENT_SESSION_ID=")), raw, nil
		}
	}
	return "", raw, fmt.Errorf("session new: could not parse session id from output")
}

func (a *AgentCLI) infoJSON(ctx context.Context) (map[string]any, string, string, error) {
	stdout, stderr, err := a.run(ctx, "session", "info", "--session-id", a.SessionID, "--json")
	if err != nil {
		return nil, stdout, stderr, err
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(stdout), &m); err != nil {
		return nil, stdout, stderr, fmt.Errorf("parse session info json: %w", err)
	}
	return m, stdout, stderr, nil
}

func (a *AgentCLI) createTab(ctx context.Context, url string) (tabID int64, stdout, stderr string, err error) {
	args := []string{"session", "create-tab", "--session-id", a.SessionID}
	if strings.TrimSpace(url) != "" {
		args = append(args, url)
	}
	stdout, stderr, err = a.run(ctx, args...)
	if err != nil {
		return 0, stdout, stderr, err
	}
	tabID = extractTabID(stdout)
	if tabID == 0 {
		return 0, stdout, stderr, fmt.Errorf("create-tab: missing tab_id in result: %s", strings.TrimSpace(stdout))
	}
	return tabID, stdout, stderr, nil
}

func (a *AgentCLI) eval(ctx context.Context, tabID int64, expr string) (stdout, stderr string, err error) {
	args := []string{"session", "eval", "--session-id", a.SessionID}
	if tabID > 0 {
		args = append(args, "--tab-id", strconv.FormatInt(tabID, 10))
	}
	args = append(args, expr)
	return a.run(ctx, args...)
}

func (a *AgentCLI) screenshot(ctx context.Context, tabID int64, outPath string) (stdout, stderr string, err error) {
	args := []string{"session", "screenshot", "--session-id", a.SessionID}
	if tabID > 0 {
		args = append(args, "--tab-id", strconv.FormatInt(tabID, 10))
	}
	if outPath != "" {
		args = append(args, "-o", outPath)
	}
	return a.run(ctx, args...)
}

func (a *AgentCLI) logs(ctx context.Context, tabID int64) (stdout, stderr string, err error) {
	args := []string{"session", "logs", "--session-id", a.SessionID, "--limit", "20"}
	if tabID > 0 {
		args = append(args, "--tab-id", strconv.FormatInt(tabID, 10))
	}
	return a.run(ctx, args...)
}

func (a *AgentCLI) cdpNavigate(ctx context.Context, tabID int64, url string) (stdout, stderr string, err error) {
	params, _ := json.Marshal(map[string]string{"url": url})
	args := []string{"session", "cdp", "--session-id", a.SessionID}
	if tabID > 0 {
		args = append(args, "--tab-id", strconv.FormatInt(tabID, 10))
	}
	args = append(args, "Page.navigate", string(params))
	return a.run(ctx, args...)
}

func extractTabID(stdout string) int64 {
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &m); err != nil {
		return 0
	}
	// Common shapes: {ok, data:{tab_id}}, {tab_id}, {data:{tab_id}}
	if id := jsonInt64(m["tab_id"]); id > 0 {
		return id
	}
	if data, ok := m["data"].(map[string]any); ok {
		if id := jsonInt64(data["tab_id"]); id > 0 {
			return id
		}
		if id := jsonInt64(data["id"]); id > 0 {
			return id
		}
	}
	if result, ok := m["result"].(map[string]any); ok {
		if id := jsonInt64(result["tab_id"]); id > 0 {
			return id
		}
	}
	return 0
}

func jsonInt64(v any) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
}

// parseEvalIdentity extracts href/title from eval job JSON stdout.
func parseEvalIdentity(stdout string) (href, title string, ready string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &m); err != nil {
		// Maybe bare JSON value.
		var bare map[string]any
		if err2 := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &bare); err2 == nil {
			href, _ = bare["href"].(string)
			title, _ = bare["title"].(string)
			ready, _ = bare["ready"].(string)
		}
		return href, title, ready
	}
	// Prefer data / result / value nesting used by job results.
	candidates := []any{m["data"], m["result"], m["value"], m}
	for _, c := range candidates {
		href, title, ready = pickIdentity(c)
		if href != "" || title != "" {
			return href, title, ready
		}
	}
	return "", "", ""
}

func pickIdentity(v any) (href, title, ready string) {
	switch t := v.(type) {
	case map[string]any:
		if h, ok := t["href"].(string); ok {
			href = h
		}
		if h, ok := t["url"].(string); ok && href == "" {
			href = h
		}
		if ti, ok := t["title"].(string); ok {
			title = ti
		}
		if r, ok := t["ready"].(string); ok {
			ready = r
		}
		// Nested value from Runtime.evaluate style
		if href == "" {
			if inner, ok := t["value"].(map[string]any); ok {
				return pickIdentity(inner)
			}
			if inner, ok := t["result"].(map[string]any); ok {
				return pickIdentity(inner)
			}
			if data, ok := t["data"].(map[string]any); ok {
				return pickIdentity(data)
			}
		}
	case string:
		// Sometimes value is a JSON string.
		var inner map[string]any
		if json.Unmarshal([]byte(t), &inner) == nil {
			return pickIdentity(inner)
		}
	}
	return href, title, ready
}

func extensionConnected(info map[string]any) bool {
	if info == nil {
		return false
	}
	// extension.connected or status fields
	if ext, ok := info["extension"].(map[string]any); ok {
		if c, ok := ext["connected"].(bool); ok {
			return c
		}
	}
	if c, ok := info["extension_connected"].(bool); ok {
		return c
	}
	return false
}

func tabsFromInfo(info map[string]any) []map[string]any {
	if info == nil {
		return nil
	}
	var raw []any
	if t, ok := info["tabs"].([]any); ok {
		raw = t
	} else if browser, ok := info["browser"].(map[string]any); ok {
		if t, ok := browser["tabs"].([]any); ok {
			raw = t
		}
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func activeTabIDFromInfo(info map[string]any) int64 {
	for _, t := range tabsFromInfo(info) {
		active, _ := t["active"].(bool)
		if !active {
			continue
		}
		if id := jsonInt64(t["id"]); id > 0 {
			return id
		}
		if id := jsonInt64(t["tab_id"]); id > 0 {
			return id
		}
	}
	return 0
}

func roleOfTab(t map[string]any) string {
	if r, ok := t["role"].(string); ok {
		return r
	}
	return ""
}

func urlOfTab(t map[string]any) string {
	if u, ok := t["url"].(string); ok {
		return u
	}
	return ""
}
