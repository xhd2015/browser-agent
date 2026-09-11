package browseragent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionLogFileName is the append-only JSONL debug log under each session dir.
const SessionLogFileName = "log.jsonl"

// SessionLogPath returns sessions/<id>/log.jsonl.
func SessionLogPath(baseDir, sessionID string) string {
	return filepath.Join(SessionDirPath(baseDir, sessionID), SessionLogFileName)
}

// AppendSessionLog appends one JSON object as a line to the session log.jsonl.
// Best-effort: missing dir or empty ids are no-ops (nil error) so callers can
// fire-and-forget from hot paths. Fields must be JSON-serializable.
func AppendSessionLog(baseDir, sessionID, level, msg string, fields map[string]any) error {
	baseDir = strings.TrimSpace(baseDir)
	sessionID = strings.TrimSpace(sessionID)
	if baseDir == "" || sessionID == "" {
		return nil
	}
	level = strings.TrimSpace(level)
	if level == "" {
		level = "info"
	}
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return fmt.Errorf("session log: empty msg")
	}

	rec := map[string]any{
		"ts":    time.Now().Format(time.RFC3339Nano),
		"level": level,
		"msg":   msg,
	}
	for k, v := range fields {
		k = strings.TrimSpace(k)
		if k == "" || k == "ts" || k == "level" || k == "msg" {
			continue
		}
		rec[k] = v
	}

	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("session log: marshal: %w", err)
	}
	path := SessionLogPath(baseDir, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("session log: mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("session log: open: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("session log: write: %w", err)
	}
	return nil
}

// sessionLog is a fire-and-forget helper on a live session.
func (s *session) sessionLog(level, msg string, fields map[string]any) {
	if s == nil {
		return
	}
	_ = AppendSessionLog(s.baseDir, s.id, level, msg, fields)
}

// ReadSessionLogLines reads log.jsonl. If tail > 0, only the last tail lines
// are returned. Missing file → empty slice, nil error.
func ReadSessionLogLines(baseDir, sessionID string, tail int) ([]string, error) {
	path := SessionLogPath(baseDir, sessionID)
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

// FormatSessionLogPretty turns one JSONL object into a short human line.
func FormatSessionLogPretty(line string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return line
	}
	ts, _ := m["ts"].(string)
	level, _ := m["level"].(string)
	msg, _ := m["msg"].(string)
	if ts == "" {
		ts = "-"
	}
	if level == "" {
		level = "info"
	}
	if msg == "" {
		msg = "?"
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if k == "ts" || k == "level" || k == "msg" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, m[k]))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("%s %s %s", ts, level, msg)
	}
	return fmt.Sprintf("%s %s %s %s", ts, level, msg, strings.Join(parts, " "))
}
