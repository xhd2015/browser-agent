// Inspect script for LOOP session-info-addr-mismatch.
// bug-repro mode (default): exits non-zero when serve --status lists a session
// but session info without --addr returns 404.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	bin := os.Getenv("BROWSER_AGENT_BIN")
	if bin == "" {
		bin = "browser-agent"
	}

	statusOut, err := run(bin, "serve", "--status")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: serve --status: %v\nstderr: %s\n", err, statusOut.stderr)
		os.Exit(2)
	}

	if !strings.Contains(statusOut.stdout, "Status:   running") {
		fmt.Fprintf(os.Stderr, "FAIL: precondition — daemon not running\n%s\n", statusOut.stdout)
		os.Exit(2)
	}

	baseURL := extractLine(statusOut.stdout, `Base URL:\s*(\S+)`)
	sessionID := extractFirstSessionID(statusOut.stdout)
	if baseURL == "" || sessionID == "" {
		fmt.Fprintf(os.Stderr, "FAIL: precondition — could not parse Base URL or session from status\n%s\n", statusOut.stdout)
		os.Exit(2)
	}

	infoOut, err := run(bin, "session", "info", "--session-id", sessionID)
	if err == nil {
		fmt.Printf("VERIFY: session info succeeded without --addr for %s\n", sessionID)
		os.Exit(0)
	}

	combined := infoOut.stdout + infoOut.stderr + err.Error()
	if strings.Contains(combined, "session not found") || strings.Contains(combined, "status 404") {
		fmt.Printf("REPRO: serve --status lists %s at %s but session info (no --addr) fails\n", sessionID, baseURL)
		fmt.Printf("REPRO: session info error: %s\n", strings.TrimSpace(combined))
		fmt.Printf("REPRO: hint: session info works with --addr %s\n", strings.TrimPrefix(baseURL, "http://"))
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "FAIL: unexpected session info error: %v\n%s\n", err, combined)
	os.Exit(2)
}

type cmdOut struct {
	stdout string
	stderr string
}

func run(bin string, args ...string) (cmdOut, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return cmdOut{stdout: stdout.String(), stderr: stderr.String()}, err
}

func extractLine(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func extractFirstSessionID(status string) string {
	lines := strings.Split(status, "\n")
	pastHeader := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "Session ID\tPhase" || strings.HasPrefix(line, "Session ID") && strings.Contains(line, "Phase") {
			pastHeader = true
			continue
		}
		if !pastHeader || line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 && strings.HasPrefix(fields[0], "sess-") {
			return fields[0]
		}
	}
	return ""
}