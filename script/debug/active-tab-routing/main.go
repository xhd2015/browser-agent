// Inspect script for LOOP active-tab-routing.
// bug-repro: exits non-zero when session info shows user tab active but eval
// runs on the session control page tab instead.
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

const sessionID = "sess-loop-active-tab"

type assertLine struct {
	Assert           string `json:"assert"`
	OK               bool   `json:"ok"`
	SessionID        string `json:"session_id"`
	UserTabActive    bool   `json:"user_tab_active"`
	ActiveTabURL     string `json:"active_tab_url"`
	EvalURL          string `json:"eval_url"`
	EvalOnSessionPage bool  `json:"eval_on_session_page"`
	BugPresent       bool   `json:"bug_present"`
}

func main() {
	mode := "inspect"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	moduleRoot, err := moduleRootDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: module root: %v\n", err)
		os.Exit(2)
	}

	evidencePath := filepath.Join(moduleRoot, "script/debug/active-tab-routing/out/observation/playwright.json")
	scriptPath := filepath.Join(moduleRoot, "script/debug/active-tab-routing/testdata/active-tab-routing.js")

	switch mode {
	case "trigger":
		if err := runTrigger(moduleRoot, scriptPath, evidencePath); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: trigger: %v\n", err)
			os.Exit(2)
		}
		fmt.Printf("TRIGGER: wrote %s\n", evidencePath)
		return
	case "inspect":
		if _, err := os.Stat(evidencePath); os.IsNotExist(err) {
			if err := runTrigger(moduleRoot, scriptPath, evidencePath); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: trigger before inspect: %v\n", err)
				os.Exit(2)
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "FAIL: unknown mode %q (use trigger|inspect)\n", mode)
		os.Exit(2)
	}

	data, err := os.ReadFile(evidencePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: read evidence: %v\n", err)
		os.Exit(2)
	}

	lines, err := parseAssertLines(string(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: parse evidence: %v\n", err)
		os.Exit(2)
	}

	extLine, ok := findAssert(lines, "extension_connected")
	if !ok || !extLine.OK {
		fmt.Fprintf(os.Stderr, "FAIL: precondition — extension did not connect\n")
		os.Exit(2)
	}

	routeLine, ok := findAssert(lines, "active_tab_routing")
	if !ok {
		fmt.Fprintf(os.Stderr, "FAIL: active_tab_routing assert line missing\nstdout:\n%s\n", string(data))
		os.Exit(2)
	}

	if routeLine.OK {
		fmt.Printf("PASS: eval ran on active user tab (%s)\n", routeLine.EvalURL)
		fmt.Printf("VERIFY: eval ran on active user tab (%s)\n", routeLine.EvalURL)
		os.Exit(0)
	}

	if routeLine.UserTabActive && routeLine.EvalOnSessionPage {
		fmt.Printf("REPRO: session info active tab is user page (%s)\n", routeLine.ActiveTabURL)
		fmt.Printf("REPRO: eval executed on session control page (%s)\n", routeLine.EvalURL)
		fmt.Printf("REPRO: pickTargetTabIdForSession pins to /go?session= tab, not active tab in window\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "FAIL: unexpected routing state user_tab_active=%v eval_url=%q\n",
		routeLine.UserTabActive, routeLine.EvalURL)
	os.Exit(2)
}

func runTrigger(moduleRoot, scriptPath, evidencePath string) error {
	pwBin, err := exec.LookPath("playwright-debug")
	if err != nil {
		return fmt.Errorf("playwright-debug not on PATH: %w", err)
	}

	baseDir, err := os.MkdirTemp("", "ba-loop-active-tab-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(baseDir)

	baseURL, cleanup, err := startDaemon(baseDir)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := createSession(baseURL, sessionID); err != nil {
		return err
	}

	extDir, _, err := browseragent.ExtractEmbeddedExtension(baseDir)
	if err != nil {
		return fmt.Errorf("ExtractEmbeddedExtension: %w", err)
	}
	if !filepath.IsAbs(extDir) {
		extDir, err = filepath.Abs(extDir)
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	args := []string{
		"--extension", extDir,
		"--headed",
		"run", scriptPath,
		baseURL,
		sessionID,
	}
	cmd := exec.CommandContext(ctx, pwBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	combined := map[string]any{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": cmd.ProcessState.ExitCode(),
		"base_url":  baseURL,
		"session_id": sessionID,
	}
	if runErr != nil {
		combined["run_error"] = runErr.Error()
	}

	if err := os.MkdirAll(filepath.Dir(evidencePath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(evidencePath, stdout.Bytes(), 0o644); err != nil {
		return err
	}
	metaPath := strings.TrimSuffix(evidencePath, ".json") + ".meta.json"
	metaRaw, _ := json.MarshalIndent(combined, "", "  ")
	if err := os.WriteFile(metaPath, metaRaw, 0o644); err != nil {
		return err
	}
	return nil
}

func startDaemon(baseDir string) (string, func(), error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cfg := browseragent.DaemonConfig{
		Addr:    addr,
		BaseDir: baseDir,
		Stdout:  io.Discard,
		Stderr:  io.Discard,
	}
	done := make(chan error, 1)
	go func() {
		_, err := browseragent.RunDaemon(ctx, cfg)
		done <- err
	}()

	baseURL := "http://" + addr
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(http.MethodGet, baseURL+"/v1/health", nil)
		if err == nil {
			res, err := http.DefaultClient.Do(req)
			if err == nil {
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
				if res.StatusCode == http.StatusOK {
					cleanup := func() {
						cancel()
						select {
						case <-done:
						case <-time.After(5 * time.Second):
						}
					}
					return baseURL, cleanup, nil
				}
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	cancel()
	<-done
	return "", nil, fmt.Errorf("daemon not healthy at %s", baseURL)
}

func createSession(baseURL, sid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	body := fmt.Sprintf(`{"session_id":%q}`, sid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/sessions", strings.NewReader(body))
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

func parseAssertLines(stdout string) ([]assertLine, error) {
	var lines []assertLine
	sc := bufio.NewScanner(strings.NewReader(stdout))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] != '{' {
			continue
		}
		var al assertLine
		if err := json.Unmarshal([]byte(line), &al); err != nil {
			continue
		}
		if al.Assert == "" {
			continue
		}
		lines = append(lines, al)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("no assert JSON lines in stdout")
	}
	return lines, sc.Err()
}

func findAssert(lines []assertLine, name string) (assertLine, bool) {
	for _, l := range lines {
		if l.Assert == name {
			return l, true
		}
	}
	return assertLine{}, false
}

func moduleRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd, nil
		}
		dir = parent
	}
}