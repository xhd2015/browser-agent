package browseragent

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

func TestAppendAndReadSessionLog(t *testing.T) {
	base := t.TempDir()
	id := "sess-log1"
	if err := os.MkdirAll(SessionDirPath(base, id), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := AppendSessionLog(base, id, "info", "session_created", map[string]any{
		"control_port": 43761,
	}); err != nil {
		t.Fatal(err)
	}
	if err := AppendSessionLog(base, id, "warn", "attach", map[string]any{
		"stage":      "register_sent",
		"attempts":   3,
		"last_error": "register_no_response",
	}); err != nil {
		t.Fatal(err)
	}
	path := SessionLogPath(base, id)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines=%d data=%s", len(lines), data)
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &rec); err != nil {
		t.Fatal(err)
	}
	if rec["msg"] != "attach" || rec["stage"] != "register_sent" {
		t.Fatalf("rec=%v", rec)
	}
	pretty := FormatSessionLogPretty(lines[1])
	if !strings.Contains(pretty, "attach") || !strings.Contains(pretty, "attempts=3") {
		t.Fatalf("pretty=%q", pretty)
	}
	tail, err := ReadSessionLogLines(base, id, 1)
	if err != nil || len(tail) != 1 {
		t.Fatalf("tail=%v err=%v", tail, err)
	}
}

func TestApplyAttachEventWritesLog(t *testing.T) {
	base := t.TempDir()
	id := "sess-attlog"
	_ = os.MkdirAll(SessionDirPath(base, id), 0o755)
	s := newSession(id, base)
	s.applyAttachEvent(AttachEvent{
		SessionID:        id,
		Stage:            AttachStageRegisterSent,
		RegisterAttempts: 2,
		LastError:        "register_rejected",
	})
	lines, err := ReadSessionLogLines(base, id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) < 1 {
		t.Fatal("expected attach log line")
	}
	if !strings.Contains(lines[len(lines)-1], `"msg":"attach"`) {
		t.Fatalf("line=%s", lines[len(lines)-1])
	}
	if !strings.Contains(lines[len(lines)-1], "register_rejected") {
		t.Fatalf("line=%s", lines[len(lines)-1])
	}
}

func TestCliSessionLogPrettyAndJSON(t *testing.T) {
	base := t.TempDir()
	id := "sess-cli1"
	_ = os.MkdirAll(SessionDirPath(base, id), 0o755)
	_ = AppendSessionLog(base, id, "info", "wait_start", map[string]any{"base_ms": 15000})
	var stdout, stderr strings.Builder
	err := cliSessionLog([]string{"--session-id", id, "--base-dir", base}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "path") || !strings.Contains(out, "wait_start") {
		t.Fatalf("stdout=%s", out)
	}
	stdout.Reset()
	err = cliSessionLog([]string{"--session-id", id, "--base-dir", base, "--json"}, nil, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"msg":"wait_start"`) {
		t.Fatalf("json stdout=%s", stdout.String())
	}
}

func TestCliSessionLogMissing(t *testing.T) {
	err := cliSessionLog([]string{"--session-id", "sess-nope", "--base-dir", t.TempDir()}, nil, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "session not found") {
		t.Fatalf("err=%v", err)
	}
}
