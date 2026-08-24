package browseragent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLISessionHARHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleCLI([]string{"session", "har", "--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "start [session-id]") || !strings.Contains(stdout.String(), "Chrome") {
		t.Fatalf("help = %q", stdout.String())
	}

	stdout.Reset()
	if err := HandleCLI([]string{"session", "har", "end", "--help"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "--output-dir") || !strings.Contains(stdout.String(), "manifest.json") {
		t.Fatalf("end help = %q", stdout.String())
	}
}

func TestCLIHARStartAcceptsPositionalSessionID(t *testing.T) {
	jobs := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/session":
			_ = json.NewEncoder(w).Encode(sessionSnapshot{SessionID: "sess-one", Browsers: []string{"Chrome"}})
		case "/v1/jobs":
			jobs++
			var request jobsRequest
			_ = json.NewDecoder(r.Body).Decode(&request)
			if request.SessionID != "sess-one" || request.Type != JobTypeHARStart {
				t.Errorf("job request = %+v", request)
			}
			_ = json.NewEncoder(w).Encode(JobResult{JobID: "job-1", OK: true, Data: map[string]any{"capture_id": "cap", "tab_count": 2}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := HandleCLI([]string{"session", "har", "start", "sess-one", "--addr", server.URL}, map[string]string{}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if jobs != 1 || !strings.Contains(stdout.String(), `"capture_id":"cap"`) {
		t.Fatalf("jobs=%d stdout=%q", jobs, stdout.String())
	}
}

func TestCLIHAREndPreflightsBeforeSubmittingJob(t *testing.T) {
	jobs := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/session":
			_ = json.NewEncoder(w).Encode(sessionSnapshot{SessionID: "sess-one", Browsers: []string{"Chrome"}})
		case "/v1/jobs":
			jobs++
			_ = json.NewEncoder(w).Encode(JobResult{OK: true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	output := t.TempDir()
	if err := os.WriteFile(filepath.Join(output, "keep"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := HandleCLI([]string{"session", "har", "end", "sess-one", "--addr", server.URL, "-o", output}, map[string]string{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("preflight error = %v", err)
	}
	if jobs != 0 {
		t.Fatalf("submitted %d har_end jobs before output preflight", jobs)
	}
}

func TestCLIHAREndRetainsRemoteCaptureOnExportFailureAndRetries(t *testing.T) {
	jobs, deletes, downloads := 0, 0, 0
	har := testHARBytes(t, 3)
	payload := harEndPayload{
		CaptureID: "cap-retry", ArtifactToken: "secret", StartedAt: "2026-08-21T00:00:00Z", EndedAt: "2026-08-21T00:01:00Z",
		Tabs:      []harExportTab{{TabID: 3, Title: "Retry", RequestCount: 1}},
		Artifacts: []harArtifactDescriptor{{TabID: 3, Size: int64(len(har))}},
	}
	payloadRaw, _ := json.Marshal(payload)
	var payloadMap map[string]any
	_ = json.Unmarshal(payloadRaw, &payloadMap)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/session":
			_ = json.NewEncoder(w).Encode(sessionSnapshot{SessionID: "sess-one", Browsers: []string{"Chrome"}})
		case "/v1/jobs":
			jobs++
			_ = json.NewEncoder(w).Encode(JobResult{JobID: "job-end", OK: true, Data: payloadMap})
		case "/v1/har/artifact":
			downloads++
			if downloads == 1 {
				http.Error(w, "temporary download failure", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(har)
		case "/v1/har/capture":
			deletes++
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	output := filepath.Join(t.TempDir(), "capture")
	args := []string{"session", "har", "end", "sess-one", "--addr", server.URL, "-o", output}
	if err := HandleCLI(args, map[string]string{}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "download HAR") {
		t.Fatalf("first export error = %v", err)
	}
	if deletes != 0 {
		t.Fatalf("remote capture cleaned after export failure: deletes=%d", deletes)
	}
	var stdout bytes.Buffer
	if err := HandleCLI(args, map[string]string{}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("retry export: %v", err)
	}
	if jobs != 2 || deletes != 1 || !strings.Contains(stdout.String(), `"capture_id":"cap-retry"`) {
		t.Fatalf("jobs=%d deletes=%d stdout=%q", jobs, deletes, stdout.String())
	}
}

func TestCLIHAREndDoesNotReportSuccessWhenCleanupFails(t *testing.T) {
	har := testHARBytes(t, 4)
	cleanupAttempts := 0
	payload := harEndPayload{
		CaptureID: "cap-cleanup", ArtifactToken: "secret", StartedAt: "2026-08-21T00:00:00Z", EndedAt: "2026-08-21T00:01:00Z",
		Tabs:      []harExportTab{{TabID: 4, Title: "Cleanup", RequestCount: 1}},
		Artifacts: []harArtifactDescriptor{{TabID: 4, Size: int64(len(har))}},
	}
	payloadRaw, _ := json.Marshal(payload)
	var payloadMap map[string]any
	_ = json.Unmarshal(payloadRaw, &payloadMap)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/session":
			_ = json.NewEncoder(w).Encode(sessionSnapshot{SessionID: "sess-one", Browsers: []string{"Chrome"}})
		case "/v1/jobs":
			_ = json.NewEncoder(w).Encode(JobResult{OK: true, Data: payloadMap})
		case "/v1/har/artifact":
			_, _ = w.Write(har)
		case "/v1/har/capture":
			cleanupAttempts++
			if cleanupAttempts == 1 {
				http.Error(w, "spool busy", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	output := filepath.Join(t.TempDir(), "capture")
	var stdout bytes.Buffer
	err := HandleCLI([]string{"session", "har", "end", "sess-one", "--addr", server.URL, "-o", output}, map[string]string{}, &stdout, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "cleanup failed") || !strings.Contains(err.Error(), "status 500") {
		t.Fatalf("cleanup error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("reported success output: %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(output, "manifest.json")); err != nil {
		t.Fatalf("published export missing after cleanup failure: %v", err)
	}
	stdout.Reset()
	if err := HandleCLI([]string{"session", "har", "end", "sess-one", "--addr", server.URL, "-o", output}, map[string]string{}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("cleanup retry with owned custom output failed: %v", err)
	}
	if cleanupAttempts != 2 || !strings.Contains(stdout.String(), `"capture_id":"cap-cleanup"`) {
		t.Fatalf("cleanup attempts=%d stdout=%q", cleanupAttempts, stdout.String())
	}
}

func TestCLIHARRejectsFirefoxBeforeSubmittingJob(t *testing.T) {
	jobs := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/session":
			_ = json.NewEncoder(w).Encode(sessionSnapshot{SessionID: "sess-firefox", Browsers: []string{"firefox"}})
		case "/v1/jobs":
			jobs++
			_ = json.NewEncoder(w).Encode(JobResult{OK: true})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	err := HandleCLI([]string{"session", "har", "start", "sess-firefox", "--addr", server.URL}, map[string]string{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "supported only for Chrome") || !strings.Contains(err.Error(), "Firefox") {
		t.Fatalf("error = %v", err)
	}
	if jobs != 0 {
		t.Fatalf("submitted %d jobs for Firefox", jobs)
	}
}
