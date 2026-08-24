package browseragent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func testHARBytes(t *testing.T, tabID int) []byte {
	t.Helper()
	value := map[string]any{
		"log": map[string]any{
			"version": "1.2",
			"creator": map[string]string{"name": "test", "version": "1"},
			"entries": []any{map[string]any{
				"startedDateTime": "2026-08-21T00:00:00Z",
				"time":            0,
				"request": map[string]any{
					"method": "GET", "url": "https://example.test/", "headers": []any{}, "queryString": []any{}, "cookies": []any{}, "headersSize": -1, "bodySize": -1,
				},
				"response": map[string]any{
					"status": 200, "statusText": "OK", "httpVersion": "h2", "headers": []any{}, "cookies": []any{}, "content": map[string]any{"size": 2, "mimeType": "text/plain", "text": "ok"}, "redirectURL": "", "headersSize": -1, "bodySize": 2,
				},
				"cache": map[string]any{}, "timings": map[string]any{"blocked": -1, "dns": -1, "ssl": -1, "connect": -1, "send": 0, "wait": 0, "receive": 0}, "_tabId": tabID,
			}},
		},
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func newHARTestServer(t *testing.T) (*SessionRegistry, *session, *httptest.Server) {
	t.Helper()
	registry := NewSessionRegistry(t.TempDir(), "127.0.0.1:43761")
	if _, err := registry.Create("sess-test"); err != nil {
		t.Fatal(err)
	}
	sess, ok := registry.Get("sess-test")
	if !ok {
		t.Fatal("session not found")
	}
	server := httptest.NewServer(NewRegistryControlHandler(registry))
	t.Cleanup(server.Close)
	return registry, sess, server
}

func TestHARArtifactUploadDownloadAndCleanup(t *testing.T) {
	_, sess, server := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	endpoint := server.URL + "/v1/har/artifact?session=sess-test&capture=" + capture.ID + "&token=" + capture.Token + "&tab_id=17"
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(testHARBytes(t, 17)))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d", res.StatusCode)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("PUT Cache-Control = %q", got)
	}
	info, err := os.Stat(capture.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
		t.Fatalf("capture dir mode = %o", info.Mode().Perm())
	}

	get, err := http.Get(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(get.Body)
	_ = get.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if get.StatusCode != http.StatusOK || !bytes.Equal(body, testHARBytes(t, 17)) {
		t.Fatalf("GET status/body mismatch: %d %s", get.StatusCode, body)
	}
	if got := get.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("GET Cache-Control = %q", got)
	}

	duplicate, _ := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(testHARBytes(t, 17)))
	duplicateRes, err := http.DefaultClient.Do(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	_ = duplicateRes.Body.Close()
	if duplicateRes.StatusCode != http.StatusOK {
		t.Fatalf("idempotent duplicate status = %d, want 200", duplicateRes.StatusCode)
	}

	conflict, _ := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(testHARBytes(t, 18)))
	conflictRes, err := http.DefaultClient.Do(conflict)
	if err != nil {
		t.Fatal(err)
	}
	_ = conflictRes.Body.Close()
	if conflictRes.StatusCode != http.StatusBadRequest {
		t.Fatalf("conflicting duplicate status = %d, want 400", conflictRes.StatusCode)
	}

	badToken := strings.Replace(endpoint, "token="+capture.Token, "token=wrong", 1)
	bad, _ := http.NewRequest(http.MethodPut, badToken, bytes.NewReader(testHARBytes(t, 18)))
	badRes, err := http.DefaultClient.Do(bad)
	if err != nil {
		t.Fatal(err)
	}
	_ = badRes.Body.Close()
	if badRes.StatusCode != http.StatusForbidden {
		t.Fatalf("bad token status = %d, want 403", badRes.StatusCode)
	}

	cleanupURL := server.URL + "/v1/har/capture?session=sess-test&capture=" + capture.ID + "&token=" + capture.Token
	cleanup, _ := http.NewRequest(http.MethodDelete, cleanupURL, nil)
	cleanupRes, err := http.DefaultClient.Do(cleanup)
	if err != nil {
		t.Fatal(err)
	}
	_ = cleanupRes.Body.Close()
	if cleanupRes.StatusCode != http.StatusConflict {
		t.Fatalf("active cleanup status = %d, want 409", cleanupRes.StatusCode)
	}
	endResult := JobResult{OK: true, Data: map[string]any{"capture_id": capture.ID}}
	sess.finishHAREnd(capture, &endResult)
	cleanup, _ = http.NewRequest(http.MethodDelete, cleanupURL, nil)
	cleanupRes, err = http.DefaultClient.Do(cleanup)
	if err != nil {
		t.Fatal(err)
	}
	_ = cleanupRes.Body.Close()
	if cleanupRes.StatusCode != http.StatusOK {
		t.Fatalf("cleanup status = %d", cleanupRes.StatusCode)
	}
	if _, err := os.Stat(capture.Dir); !os.IsNotExist(err) {
		t.Fatalf("capture dir still exists: %v", err)
	}
}

func TestHARArtifactRejectsMalformedJSON(t *testing.T) {
	_, sess, server := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	endpoint := server.URL + "/v1/har/artifact?session=sess-test&capture=" + capture.ID + "&token=" + capture.Token + "&tab_id=1"
	req, _ := http.NewRequest(http.MethodPut, endpoint, strings.NewReader("not-json"))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestHARCaptureSingleActive(t *testing.T) {
	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.beginHARCapture(); err == nil || !strings.Contains(err.Error(), "already active") {
		t.Fatalf("second begin error = %v", err)
	}
	sess.abortHARCapture(capture)
	if _, err := sess.beginHARCapture(); err != nil {
		t.Fatalf("begin after cleanup: %v", err)
	}
}

func TestExportHARCaptureWritesPrivatePerTabFiles(t *testing.T) {
	_, sess, server := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := capture.storeArtifact(17, bytes.NewReader(testHARBytes(t, 17))); err != nil {
		t.Fatal(err)
	}
	payload := harEndPayload{
		CaptureID: capture.ID, ArtifactToken: capture.Token,
		StartedAt: "2026-08-21T00:00:00Z", EndedAt: "2026-08-21T00:01:00Z",
		Tabs:      []harExportTab{{TabID: 17, Title: "Orders / Current", URL: "https://example.test/orders", State: "closed", RequestCount: 1}},
		Artifacts: []harArtifactDescriptor{{TabID: 17, Size: int64(len(testHARBytes(t, 17)))}},
	}
	output := filepath.Join(t.TempDir(), "capture")
	manifest, err := exportHARCapture(server.URL, "sess-test", output, false, payload)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.TabCount != 1 || manifest.RequestCount != 1 {
		t.Fatalf("manifest counts = %+v", manifest)
	}
	harPath := filepath.Join(output, "tab-17-orders-current.har")
	if err := validateHARFile(harPath); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		dirInfo, _ := os.Stat(output)
		fileInfo, _ := os.Stat(harPath)
		manifestInfo, _ := os.Stat(filepath.Join(output, "manifest.json"))
		if dirInfo.Mode().Perm() != 0o700 || fileInfo.Mode().Perm() != 0o600 || manifestInfo.Mode().Perm() != 0o600 {
			t.Fatalf("modes dir=%o har=%o manifest=%o", dirInfo.Mode().Perm(), fileInfo.Mode().Perm(), manifestInfo.Mode().Perm())
		}
	}
}

func TestHAROutputRefusesUnrelatedNonEmptyDirectory(t *testing.T) {
	output := t.TempDir()
	if err := os.WriteFile(filepath.Join(output, "keep.txt"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prepareHAROutput(output, "sess-test", false); err == nil || !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("error = %v", err)
	}
	if raw, err := os.ReadFile(filepath.Join(output, "keep.txt")); err != nil || string(raw) != "keep" {
		t.Fatalf("unrelated file changed: %q %v", raw, err)
	}
}

func TestHAROutputAllowsOwnedCustomDirectoryForCleanupRetry(t *testing.T) {
	output := t.TempDir()
	manifest := []byte(`{"schema_version":1,"session_id":"sess-test"}`)
	if err := os.WriteFile(filepath.Join(output, "manifest.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prepareHAROutput(output, "sess-test", false); err != nil {
		t.Fatalf("owned custom output rejected: %v", err)
	}
}

func TestHARValidationRejectsIncompleteAndTrailingJSON(t *testing.T) {
	for name, raw := range map[string]string{
		"missing log":     `{"other":{}}`,
		"missing creator": `{"log":{"version":"1.2","entries":[]}}`,
		"entries object":  `{"log":{"version":"1.2","creator":{"name":"x","version":"1"},"entries":{}}}`,
		"trailing value":  `{"log":{"version":"1.2","creator":{"name":"x","version":"1"},"entries":[]}} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "capture.har")
			if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := validateHARFile(path); err == nil {
				t.Fatal("expected invalid HAR error")
			}
		})
	}
}

func TestHARCaptureArtifactLimits(t *testing.T) {
	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	capture.mu.Lock()
	capture.TotalBytes = maxHARTotalBytes
	capture.mu.Unlock()
	if _, err := capture.storeArtifact(1, bytes.NewReader(testHARBytes(t, 1))); err == nil || !strings.Contains(err.Error(), "total bytes") {
		t.Fatalf("total byte limit error = %v", err)
	}
	capture.mu.Lock()
	capture.TotalBytes = 0
	for id := 1; id <= maxHARArtifacts; id++ {
		capture.Artifacts[id] = harArtifact{TabID: id}
	}
	capture.mu.Unlock()
	if _, err := capture.storeArtifact(maxHARArtifacts+1, bytes.NewReader(testHARBytes(t, maxHARArtifacts+1))); err == nil || !strings.Contains(err.Error(), "artifacts") {
		t.Fatalf("artifact count limit error = %v", err)
	}
}

func TestEndedHARCanResumeWithoutExtension(t *testing.T) {
	registry, sess, server := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	result := JobResult{JobID: "job-original", OK: true, Data: map[string]any{
		"capture_id": capture.ID, "artifact_token": capture.Token, "tabs": []any{}, "artifacts": []any{},
	}}
	sess.finishHAREnd(capture, &result)

	body, _ := json.Marshal(jobsRequest{SessionID: "sess-test", Type: JobTypeHAREnd})
	res, err := http.Post(server.URL+"/v1/jobs", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var resumed JobResult
	if err := json.NewDecoder(res.Body).Decode(&resumed); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK || !resumed.OK || resumed.JobID != "job-original" {
		t.Fatalf("resume status/result = %d %+v", res.StatusCode, resumed)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("resume Cache-Control = %q", got)
	}
	if got := registry.List()[0].InflightJobs; got != 0 {
		t.Fatalf("resume queued %d jobs", got)
	}
}

func TestHARCleanupFailureIsReportedAndRetained(t *testing.T) {
	_, sess, server := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	result := JobResult{OK: true, Data: map[string]any{"capture_id": capture.ID}}
	sess.finishHAREnd(capture, &result)
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	capture.mu.Lock()
	capture.Dir = filepath.Join(blocker, "child")
	capture.mu.Unlock()

	u := server.URL + "/v1/har/capture?session=sess-test&capture=" + capture.ID + "&token=" + capture.Token
	req, _ := http.NewRequest(http.MethodDelete, u, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("cleanup status = %d, want 500", res.StatusCode)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("capture removed after cleanup failure: %v", err)
	}
}

func TestHARDisconnectGracePreservesCaptureAndReconnects(t *testing.T) {
	oldGrace := wsDisconnectGracePeriod
	wsDisconnectGracePeriod = 20 * time.Millisecond
	defer func() { wsDisconnectGracePeriod = oldGrace }()

	sess := newSession("sess-test", t.TempDir())
	sess.markHello("1.0.0", []string{FeatureBrowserAgent}, "")
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	first := &wsConn{}
	sess.setWS(first)
	sess.markDisconnectedAfterGrace(first)
	if !sess.isExtensionConnected() {
		t.Fatal("session disconnected before grace elapsed")
	}
	second := &wsConn{}
	sess.setWS(second)
	time.Sleep(2 * wsDisconnectGracePeriod)
	if !sess.isExtensionConnected() {
		t.Fatal("reconnect did not cancel disconnect reconciliation")
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("capture was deleted across reconnect: %v", err)
	}

	endResult := JobResult{OK: true, Data: map[string]any{"capture_id": capture.ID}}
	sess.finishHAREnd(capture, &endResult)
	sess.markDisconnectedAfterGrace(second)
	time.Sleep(2 * wsDisconnectGracePeriod)
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("ended capture was deleted on disconnect: %v", err)
	}
}

func TestReconcileHARHelloConfirmsMatchingCapture(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	if abortID := sess.reconcileHARHello("  "+capture.ID+"  ", nil); abortID != "" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if !capture.startIsConfirmed() {
		t.Fatal("matching extension capture did not confirm daemon capture")
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("matching capture was not preserved: %v", err)
	}
}

func TestReconcileHARHelloDropsMismatchedIncompleteCapture(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	spool := capture.Dir
	if abortID := sess.reconcileHARHello("extension-stale", nil); abortID != "extension-stale" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if _, err := sess.activeHARCapture(); err == nil {
		t.Fatal("mismatched daemon capture was preserved")
	}
	if _, err := os.Stat(spool); !os.IsNotExist(err) {
		t.Fatalf("mismatched capture spool remains: %v", err)
	}
}

func TestReconcileHARHelloPreservesStartBeforeEnqueue(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	if abortID := sess.reconcileHARHello("", nil); abortID != "" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("capture before start enqueue was not preserved: %v", err)
	}
}

func TestReconcileHARHelloPreservesQueuedStart(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sess.queue.Enqueue(Job{
		Type:   JobTypeHARStart,
		Params: map[string]any{"capture_id": capture.ID},
	}); err != nil {
		t.Fatal(err)
	}
	if abortID := sess.reconcileHARHello("", nil); abortID != "" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("capture for queued start was not preserved: %v", err)
	}
}

func TestReconcileHARHelloPreservesPendingEndResult(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	if abortID := sess.reconcileHARHello("", []string{"other", capture.ID}); abortID != "" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("capture for pending end result was not preserved: %v", err)
	}
	if _, err := os.Stat(capture.Dir); err != nil {
		t.Fatalf("capture spool for pending end result was removed: %v", err)
	}
}

func TestReconcileHARHelloPreservesCompletedExport(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	result := JobResult{OK: true, Data: map[string]any{"capture_id": capture.ID}}
	sess.finishHAREnd(capture, &result)

	if abortID := sess.reconcileHARHello("", nil); abortID != "" {
		t.Fatalf("abort capture id = %q", abortID)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("completed capture was not preserved: %v", err)
	}
	if abortID := sess.reconcileHARHello("extension-stale", nil); abortID != "extension-stale" {
		t.Fatalf("stale extension capture id = %q", abortID)
	}
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("completed capture was removed while requesting extension rollback: %v", err)
	}
}

func TestCompleteWithPreviousReturnsExpiredStateAtomically(t *testing.T) {
	queue := NewJobQueue()
	job, err := queue.Enqueue(Job{Type: JobTypeHAREnd})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := queue.Wait(ctx, job.ID); err == nil {
		t.Fatal("expected job expiration")
	}

	previous, err := queue.CompleteWithPrevious(job.ID, JobResult{OK: true})
	if err != nil {
		t.Fatal(err)
	}
	if previous.Status != JobStatusExpired {
		t.Fatalf("previous status = %s", previous.Status)
	}
	current, ok := queue.Get(job.ID)
	if !ok {
		t.Fatal("job disappeared")
	}
	if current.Status != JobStatusExpired || current.Result == nil || current.Result.OK {
		t.Fatalf("late completion changed expired job: %+v", current)
	}
}

func TestTimedOutHARStartRemainsReconcileable(t *testing.T) {
	oldGrace := harLateResultGracePeriod
	harLateResultGracePeriod = 10 * time.Millisecond
	defer func() { harLateResultGracePeriod = oldGrace }()

	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	job, err := sess.queue.Enqueue(Job{Type: JobTypeHARStart, Params: map[string]any{"capture_id": capture.ID}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sess.queue.Wait(ctx, job.ID); err == nil {
		t.Fatal("expected job expiration")
	}
	sess.scheduleHARReconciliation(job, capture)
	time.Sleep(2 * harLateResultGracePeriod)
	if current, err := sess.activeHARCapture(); err != nil || current != capture {
		t.Fatalf("uncertain start was discarded: %v", err)
	}
	server := &controlServer{}
	server.handleWSResult(sess, nil, wsEnvelope{Type: "result", ID: job.ID, Payload: map[string]any{"ok": true}})
	if !capture.startIsConfirmed() {
		t.Fatal("late start result did not reconcile capture")
	}
}

func TestHAREndResultCommitsBeforeAcknowledgement(t *testing.T) {
	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	if _, err := capture.storeArtifact(7, bytes.NewReader(testHARBytes(t, 7))); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.prepareHAREnd(); err != nil {
		t.Fatal(err)
	}
	job, err := sess.queue.Enqueue(Job{Type: JobTypeHAREnd, Params: map[string]any{"capture_id": capture.ID}})
	if err != nil {
		t.Fatal(err)
	}
	server := &controlServer{}
	ack := server.completeExtensionResult(sess, JobResult{
		JobID: job.ID,
		OK:    true,
		Data:  map[string]any{"tabs": []any{map[string]any{"tab_id": float64(7)}}, "capture_id": capture.ID},
	})
	if !ack {
		t.Fatal("committed HAR end result was not acknowledged")
	}
	completed, ok := sess.completedHAREnd()
	if !ok || !completed.OK {
		t.Fatalf("HAR end was not committed before acknowledgement: %+v, %v", completed, ok)
	}
	queued, ok := sess.queue.Get(job.ID)
	if !ok || queued.Result == nil || queued.Result.Data["artifact_token"] == "" {
		t.Fatalf("queue did not receive decorated HAR result: %+v", queued)
	}
	if server.completeExtensionResult(sess, JobResult{JobID: "unknown", OK: true}) {
		t.Fatal("unknown result was acknowledged")
	}
}

func TestDuplicateHARStartResultDoesNotAbortCapture(t *testing.T) {
	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	job, err := sess.queue.Enqueue(Job{Type: JobTypeHARStart, Params: map[string]any{"capture_id": capture.ID}})
	if err != nil {
		t.Fatal(err)
	}
	server := &controlServer{}
	if !server.completeExtensionResult(sess, JobResult{JobID: job.ID, OK: true}) {
		t.Fatal("start result was not acknowledged")
	}
	if !server.completeExtensionResult(sess, JobResult{JobID: job.ID, OK: false, Error: "duplicate start"}) {
		t.Fatal("duplicate result was not acknowledged")
	}
	current, err := sess.activeHARCapture()
	if err != nil || current != capture || !capture.startIsConfirmed() {
		t.Fatalf("duplicate result changed capture ownership: capture=%p err=%v confirmed=%v", current, err, capture.startIsConfirmed())
	}
}

func TestLateHAREndResultReconcilesExpiredJob(t *testing.T) {
	_, sess, _ := newHARTestServer(t)
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	if _, err := capture.storeArtifact(7, bytes.NewReader(testHARBytes(t, 7))); err != nil {
		t.Fatal(err)
	}
	if _, err := sess.prepareHAREnd(); err != nil {
		t.Fatal(err)
	}
	job, err := sess.queue.Enqueue(Job{Type: JobTypeHAREnd, Params: map[string]any{"capture_id": capture.ID}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sess.queue.Wait(ctx, job.ID); err == nil {
		t.Fatal("expected job expiration")
	}
	server := &controlServer{}
	server.handleWSResult(sess, nil, wsEnvelope{Type: "result", ID: job.ID, Payload: map[string]any{
		"ok":   true,
		"data": map[string]any{"tabs": []any{map[string]any{"tab_id": float64(7)}}, "capture_id": capture.ID},
	}})
	resumed, ok := sess.completedHAREnd()
	if !ok || !resumed.OK {
		t.Fatalf("late end result was not reconciled: %+v, %v", resumed, ok)
	}
}

func TestPermanentHAREndUploadFailureAbortsCapture(t *testing.T) {
	sess := newSession("sess-test", t.TempDir())
	capture, err := sess.beginHARCapture()
	if err != nil {
		t.Fatal(err)
	}
	sess.confirmHARStart(capture)
	spool := capture.Dir
	job, err := sess.queue.Enqueue(Job{Type: JobTypeHAREnd, Params: map[string]any{"capture_id": capture.ID}})
	if err != nil {
		t.Fatal(err)
	}
	server := &controlServer{}
	if !server.completeExtensionResult(sess, JobResult{
		JobID: job.ID,
		OK:    false,
		Data:  map[string]any{"capture_aborted": true},
	}) {
		t.Fatal("permanent failure result was not acknowledged")
	}
	if _, err := sess.activeHARCapture(); err == nil {
		t.Fatal("permanently failed capture remains active")
	}
	if _, err := os.Stat(spool); !os.IsNotExist(err) {
		t.Fatalf("permanently failed capture spool remains: %v", err)
	}
}

func TestRestoreSessionsSweepsOrphanHARSpool(t *testing.T) {
	baseDir := t.TempDir()
	registry := NewSessionRegistry(baseDir, "127.0.0.1:43761")
	if _, err := registry.Create("sess-restore"); err != nil {
		t.Fatal(err)
	}
	spool := filepath.Join(SessionDirPath(baseDir, "sess-restore"), "har-spool", "orphan")
	if err := os.MkdirAll(spool, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(spool, "secret.har"), testHARBytes(t, 1), 0o600); err != nil {
		t.Fatal(err)
	}
	restored := NewSessionRegistry(baseDir, "127.0.0.1:43761")
	if err := RestoreSessionsFromDisk(restored); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(SessionDirPath(baseDir, "sess-restore"), "har-spool")); !os.IsNotExist(err) {
		t.Fatalf("orphan spool remains: %v", err)
	}
	if _, ok := restored.Get("sess-restore"); !ok {
		t.Fatal("session was not restored")
	}
}

func TestHARSlugAndDefaultOutput(t *testing.T) {
	if got := harSlug(" Hello, 世界 / API ", "untitled", 60); got != "hello-api" {
		t.Fatalf("slug = %q", got)
	}
	if got := filepath.Base(defaultHAROutputDir("Sess_Test.1")); got != "browser-agent-sess-test-1" {
		t.Fatalf("default output = %q", got)
	}
}
