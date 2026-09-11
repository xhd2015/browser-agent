package browseragent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendExtensionLog_ByBrowser(t *testing.T) {
	base := t.TempDir()
	if err := AppendExtensionLog(base, ExtensionLogEvent{
		Browser: "chrome",
		Level:   "warn",
		Msg:     "job fail",
		Detail:  map[string]any{"job_id": "job-1"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := AppendExtensionLog(base, ExtensionLogEvent{
		Browser:   "firefox",
		Level:     "info",
		Msg:       "poll ok",
		SessionID: "sess-ff",
	}); err != nil {
		t.Fatal(err)
	}
	chromePath := ExtensionLogPath(base, "chrome")
	ffPath := ExtensionLogPath(base, "firefox")
	if filepath.Base(chromePath) != ExtensionLogChromeFile {
		t.Fatalf("chrome path=%s", chromePath)
	}
	if filepath.Base(ffPath) != ExtensionLogFirefoxFile {
		t.Fatalf("ff path=%s", ffPath)
	}
	cLines, err := ReadExtensionLogLines(base, "chrome", 0)
	if err != nil || len(cLines) != 1 {
		t.Fatalf("chrome lines=%v err=%v", cLines, err)
	}
	if !strings.Contains(cLines[0], `"msg":"job fail"`) || !strings.Contains(cLines[0], `"browser":"chrome"`) {
		t.Fatalf("line=%s", cLines[0])
	}
	fLines, err := ReadExtensionLogLines(base, "firefox", 0)
	if err != nil || len(fLines) != 1 {
		t.Fatalf("ff lines=%v err=%v", fLines, err)
	}
	if !strings.Contains(fLines[0], "sess-ff") {
		t.Fatalf("line=%s", fLines[0])
	}
}

func TestHandleExtLog_HTTP(t *testing.T) {
	base := t.TempDir()
	reg := NewSessionRegistry(base, "127.0.0.1:43761")
	h := NewRegistryControlHandlerConfig(reg, RegistryHandlerConfig{BaseDir: base})
	srv := httptest.NewServer(h)
	defer srv.Close()

	body := `{"browser":"chrome","level":"error","msg":"debugger attach fail","session_id":"sess-x","detail":{"tab_id":1}}`
	resp, err := http.Post(srv.URL+"/v1/ext/log", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d body=%s", resp.StatusCode, b)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out["ok"] != true {
		t.Fatalf("out=%v", out)
	}
	data, err := os.ReadFile(ExtensionLogPath(base, "chrome"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "debugger attach fail") {
		t.Fatalf("file=%s", data)
	}
}

func TestCliExtensionLog(t *testing.T) {
	base := t.TempDir()
	_ = AppendExtensionLog(base, ExtensionLogEvent{Browser: "chrome", Msg: "heal start", Level: "info"})
	var stdout strings.Builder
	err := cliExtensionLog([]string{"--base-dir", base, "--browser", "chrome"}, nil, &stdout, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "heal start") {
		t.Fatalf("stdout=%s", stdout.String())
	}
}
