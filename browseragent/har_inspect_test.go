package browseragent

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func harInspectFixture(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join("testdata", "har-inspect", rel)
}

func TestCLIHARInspectHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleCLI([]string{"har", "--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "inspect") {
		t.Fatalf("har help = %q", stdout.String())
	}
	stdout.Reset()
	if err := HandleCLI([]string{"har", "inspect", "--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"summary", "paths", "entries", "show"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("inspect help missing %q: %q", want, stdout.String())
		}
	}
	for _, args := range [][]string{
		{"har", "inspect", "summary", "--help"},
		{"har", "inspect", "paths", "--help"},
		{"har", "inspect", "entries", "--help"},
		{"har", "inspect", "show", "--help"},
	} {
		stdout.Reset()
		if err := HandleCLI(args, nil, &stdout, &stderr); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !strings.Contains(stdout.String(), "Usage:") {
			t.Fatalf("%v help = %q", args, stdout.String())
		}
	}
}

func TestHARInspectSummaryDir(t *testing.T) {
	dir := harInspectFixture(t, "export-ok")
	var stdout, stderr bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "summary", dir, "--host", "app.example.com"}, map[string]string{"NO_COLOR": "1"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	if !strings.Contains(out, "sess-test") || !strings.Contains(out, "partial     false") {
		t.Fatalf("summary = %q", out)
	}
	if !strings.Contains(out, "tab-1-app.har") || !strings.Contains(out, "tab-2-quiet.har") {
		t.Fatalf("missing tabs: %q", out)
	}
	if stderr.Len() != 0 {
		t.Fatalf("unexpected stderr %q", stderr.String())
	}
}

func TestHARInspectSummarySingleHAR(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	var stdout, stderr bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "summary", path, "--json"}, map[string]string{"NO_COLOR": "1"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["manifest"] != nil {
		t.Fatalf("manifest = %#v", payload["manifest"])
	}
	cov := payload["body_coverage"].(map[string]any)
	if int(cov["entries"].(float64)) < 1 {
		t.Fatalf("coverage = %#v", cov)
	}
}

func TestHARInspectSummaryMissingManifest(t *testing.T) {
	dir := harInspectFixture(t, "export-no-manifest")
	err := HandleCLI([]string{"har", "inspect", "summary", dir}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "manifest.json") {
		t.Fatalf("err = %v", err)
	}
}

func TestHARInspectPathsFiltersNoise(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "paths", path, "--host", "app.example.com", "--json"}, nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Paths []harPathCount `json:"paths"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	joined := ""
	for _, p := range payload.Paths {
		joined += p.Path + "\n"
		if strings.HasSuffix(p.Path, ".js") || strings.Contains(p.Path, "google-analytics") {
			t.Fatalf("noise path present: %#v", payload.Paths)
		}
	}
	if !strings.Contains(joined, "/api/v1/items/create") {
		t.Fatalf("paths = %#v", payload.Paths)
	}
}

func TestHARInspectEntriesMethodFilter(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "entries", path, "--host", "app.example.com", "--method", "POST", "--json"}, nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Entries []harInspectRow `json:"entries"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Entries) == 0 {
		t.Fatal("expected POST entries")
	}
	for _, e := range payload.Entries {
		if !strings.EqualFold(e.Method, "POST") {
			t.Fatalf("entry = %#v", e)
		}
	}
}

func TestHARInspectShowMatchRedactsAuth(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "show", path, "--match", "/api/v1/items/create", "--json"}, nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	raw := stdout.String()
	if strings.Contains(raw, "Bearer eyJ") {
		t.Fatalf("auth not redacted: %s", raw)
	}
	if !strings.Contains(raw, "<REDACTED>") && !strings.Contains(raw, `\u003cREDACTED\u003e`) {
		t.Fatalf("missing redaction marker: %s", raw)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	req := payload["request"].(map[string]any)
	headers := req["headers"].(map[string]any)
	auth := headers["authorization"].(string)
	if auth != "<REDACTED>" {
		t.Fatalf("authorization = %q", auth)
	}
	body := req["body"].(map[string]any)
	if body["token"] != "<REDACTED>" && body["token"] != "<REDACTED_JWT>" {
		// token field name triggers redactJSONValue key rule
		if s, ok := body["token"].(string); !ok || !strings.Contains(s, "REDACTED") {
			t.Fatalf("token body = %#v", body["token"])
		}
	}
	resp := payload["response"].(map[string]any)
	respBody := resp["body"].(map[string]any)
	result := respBody["result"].(map[string]any)
	if int(result["id"].(float64)) != 42 {
		t.Fatalf("result = %#v", result)
	}
}

func TestHARInspectShowNoMatch(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	err := HandleCLI([]string{"har", "inspect", "show", path, "--match", "/does-not-exist"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "no entries matched") {
		t.Fatalf("err = %v", err)
	}
}

func TestHARInspectShowMultipleMatch(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	err := HandleCLI([]string{"har", "inspect", "show", path, "--match", "/api/v1/items"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "entries matched") {
		t.Fatalf("err = %v", err)
	}
}

func TestHARInspectSummaryPartialWarns(t *testing.T) {
	dir := harInspectFixture(t, "export-partial")
	var stdout, stderr bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "summary", dir}, map[string]string{"NO_COLOR": "1"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "warning:") || !strings.Contains(stderr.String(), "partial") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestHARInspectShowBase64JSON(t *testing.T) {
	path := harInspectFixture(t, "single-api.har")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "show", path, "--match", "/api/v1/binary", "--json"}, nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	resp := payload["response"].(map[string]any)
	body := resp["body"].(map[string]any)
	if body["code"].(float64) != 0 {
		t.Fatalf("body = %#v", body)
	}
}

func TestHARInspectShowIndexRequiresTabIDForDir(t *testing.T) {
	dir := harInspectFixture(t, "export-ok")
	err := HandleCLI([]string{"har", "inspect", "show", dir, "--index", "3"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--tab-id") {
		t.Fatalf("err = %v", err)
	}
}

func TestHARInspectShowIndexWithTabID(t *testing.T) {
	dir := harInspectFixture(t, "export-ok")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"har", "inspect", "show", dir, "--tab-id", "1", "--index", "4", "--json"}, nil, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if int(payload["index"].(float64)) != 4 {
		t.Fatalf("payload = %#v", payload)
	}
	if !strings.Contains(payload["url"].(string), "/api/v1/items/create") {
		t.Fatalf("url = %#v", payload["url"])
	}
}

func TestHARInspectRootHelpListsInFullHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"--help"}, nil, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "har inspect") {
		t.Fatalf("full help missing har inspect: %q", stdout.String())
	}
}

func TestIsHARNoiseURL(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://app.example.com/api/x", false},
		{"https://cdn.example.com/app.js", true},
		{"https://www.google-analytics.com/g/collect", true},
		{"https://app.example.com/tags/web-performance/x", true},
	}
	for _, tc := range cases {
		if got := isHARNoiseURL(tc.url); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.url, got, tc.want)
		}
	}
}

func TestLoadHARInspectSourceRejectsNonHARFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := loadHARInspectSource(path)
	if err == nil {
		t.Fatal("expected error")
	}
}
