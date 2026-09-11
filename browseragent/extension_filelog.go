package browseragent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Extension log files under the daemon base dir (unified SW — not per-session).
const (
	ExtensionLogChromeFile  = "extension-chrome.log.jsonl"
	ExtensionLogFirefoxFile = "extension-firefox.log.jsonl"
)

// NormalizeExtensionLogBrowser returns "chrome" or "firefox".
func NormalizeExtensionLogBrowser(browser string) string {
	switch strings.ToLower(strings.TrimSpace(browser)) {
	case "firefox", "ff":
		return "firefox"
	default:
		return "chrome"
	}
}

// ExtensionLogPath returns {baseDir}/extension-{chrome|firefox}.log.jsonl.
func ExtensionLogPath(baseDir, browser string) string {
	name := ExtensionLogChromeFile
	if NormalizeExtensionLogBrowser(browser) == "firefox" {
		name = ExtensionLogFirefoxFile
	}
	return filepath.Join(baseDir, name)
}

// ExtensionLogEvent is POST /v1/ext/log body (and JSONL record shape).
type ExtensionLogEvent struct {
	Browser   string `json:"browser,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Level     string `json:"level,omitempty"`
	Msg       string `json:"msg"`
	Detail    any    `json:"detail,omitempty"`
	TS        string `json:"ts,omitempty"`
	Source    string `json:"source,omitempty"`
}

// AppendExtensionLog appends one JSON line to the browser-scoped extension log.
func AppendExtensionLog(baseDir string, ev ExtensionLogEvent) error {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		return fmt.Errorf("extension log: empty baseDir")
	}
	msg := strings.TrimSpace(ev.Msg)
	if msg == "" {
		return fmt.Errorf("extension log: empty msg")
	}
	browser := NormalizeExtensionLogBrowser(ev.Browser)
	level := strings.TrimSpace(ev.Level)
	if level == "" {
		level = "info"
	}
	ts := strings.TrimSpace(ev.TS)
	if ts == "" {
		ts = time.Now().Format(time.RFC3339Nano)
	}
	source := strings.TrimSpace(ev.Source)
	if source == "" {
		source = "extension"
	}
	rec := map[string]any{
		"ts":      ts,
		"level":   level,
		"msg":     msg,
		"browser": browser,
		"source":  source,
	}
	if sid := strings.TrimSpace(ev.SessionID); sid != "" {
		rec["session_id"] = sid
	}
	if ev.Detail != nil {
		rec["detail"] = ev.Detail
	}
	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("extension log: marshal: %w", err)
	}
	path := ExtensionLogPath(baseDir, browser)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

// ReadExtensionLogLines reads the browser extension log. tail>0 keeps last N lines.
func ReadExtensionLogLines(baseDir, browser string, tail int) ([]string, error) {
	path := ExtensionLogPath(baseDir, browser)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	raw := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	lines := make([]string, 0, len(raw))
	for _, ln := range raw {
		if strings.TrimSpace(ln) == "" {
			continue
		}
		lines = append(lines, ln)
	}
	if tail > 0 && len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return lines, nil
}
