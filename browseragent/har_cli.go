package browseragent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const harHelp = `Usage: browser-agent session har <command> [flags]

Commands:
  start [session-id]       Start HAR capture for all user HTTP(S) tabs in the session window
  end [session-id]         End capture and export one HAR per tab plus manifest.json

Session resolution: positional session-id, --session-id, or BROWSER_AGENT_SESSION_ID.
HAR capture is currently supported only for Chrome.
`

const harStartHelp = `Usage: browser-agent session har start [session-id] [flags]

Starts one Chrome HAR capture covering all current and newly opened user HTTP(S)
tabs in the Browser Agent session window. The session control page is excluded.

Flags:
  --session-id <id>        Session id (or positional id / BROWSER_AGENT_SESSION_ID)
  --timeout <duration>     Job timeout (default 60s)
  --base-dir, --host, --server-port
`

const harEndHelp = `Usage: browser-agent session har end [session-id] [flags]

Stops the active Chrome HAR capture and exports a private directory containing
manifest.json and tab-{id}-{title-slug}.har for every enrolled tab.

Flags:
  -o, --output-dir <dir>   Destination directory (default /tmp/browser-agent-<session-slug>/)
  --session-id <id>        Session id (or positional id / BROWSER_AGENT_SESSION_ID)
  --timeout <duration>     Job timeout (default 60s)
  --base-dir, --host, --server-port
`

type harArtifactDescriptor struct {
	TabID int   `json:"tab_id"`
	Size  int64 `json:"size"`
}

type harExportTab struct {
	TabID        int      `json:"tab_id"`
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	State        string   `json:"state"`
	RequestCount int      `json:"request_count"`
	ErrorCount   int      `json:"error_count"`
	Errors       []string `json:"errors,omitempty"`
	File         string   `json:"file"`
}

type harEndPayload struct {
	CaptureID     string                  `json:"capture_id"`
	ArtifactToken string                  `json:"artifact_token"`
	StartedAt     string                  `json:"started_at"`
	EndedAt       string                  `json:"ended_at"`
	Partial       bool                    `json:"partial"`
	Warnings      []string                `json:"warnings"`
	Tabs          []harExportTab          `json:"tabs"`
	Artifacts     []harArtifactDescriptor `json:"artifacts"`
}

type harManifest struct {
	SchemaVersion int            `json:"schema_version"`
	SessionID     string         `json:"session_id"`
	CaptureID     string         `json:"capture_id"`
	Browser       string         `json:"browser"`
	StartedAt     string         `json:"started_at"`
	EndedAt       string         `json:"ended_at"`
	Partial       bool           `json:"partial"`
	Warnings      []string       `json:"warnings,omitempty"`
	TabCount      int            `json:"tab_count"`
	RequestCount  int            `json:"request_count"`
	ErrorCount    int            `json:"error_count"`
	Tabs          []harExportTab `json:"tabs"`
}

func cliSessionHAR(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		_, _ = io.WriteString(stdout, harHelp)
		return nil
	}
	switch args[0] {
	case "start":
		return cliHARStart(args[1:], env, stdout)
	case "end":
		return cliHAREnd(args[1:], env, stdout)
	default:
		return fmt.Errorf("unknown session har subcommand %q; try start or end", args[0])
	}
}

func resolveHARSession(args []string, env map[string]string) (string, error) {
	positional := strings.TrimSpace(takePositional(args, 0))
	flagValue, flagSet := flagStringSet(args, "--session-id")
	envValue, envSet := "", false
	if env != nil {
		envValue, envSet = env["BROWSER_AGENT_SESSION_ID"]
	} else {
		envValue, envSet = os.LookupEnv("BROWSER_AGENT_SESSION_ID")
	}
	if positional != "" {
		if flagSet && strings.TrimSpace(flagValue) != positional {
			return "", errors.New("session id was provided both positionally and with --session-id")
		}
		if !flagSet && envSet && strings.TrimSpace(envValue) != "" && strings.TrimSpace(envValue) != positional {
			return "", errors.New("session id was provided both positionally and through BROWSER_AGENT_SESSION_ID")
		}
		if err := ValidateSessionID(positional); err != nil {
			return "", err
		}
		return positional, nil
	}
	return ResolveSessionID(flagValue, flagSet, envValue, envSet)
}

func fetchSessionSnapshot(base, sessionID string) (sessionSnapshot, error) {
	var snapshot sessionSnapshot
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u := strings.TrimRight(base, "/") + "/v1/session?session=" + url.QueryEscape(sessionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return snapshot, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return snapshot, fmt.Errorf("session info request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
		return snapshot, fmt.Errorf("session info failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(res.Body).Decode(&snapshot); err != nil {
		return snapshot, fmt.Errorf("session info returned invalid JSON: %w", err)
	}
	return snapshot, nil
}

func requireChromeHARSession(base, sessionID string) error {
	snapshot, err := fetchSessionSnapshot(base, sessionID)
	if err != nil {
		return err
	}
	if sessionSnapIsFirefox(snapshot) {
		return fmt.Errorf("HAR capture is currently supported only for Chrome; session %s uses Firefox", sessionID)
	}
	if snapshot.Extension.Connected {
		hasHAR := false
		for _, feature := range snapshot.Extension.Features {
			if feature == "har-capture" {
				hasHAR = true
				break
			}
		}
		if !hasHAR {
			return errors.New("connected Chrome extension does not support HAR capture; reinstall or reload the current Browser Agent extension")
		}
	}
	return nil
}

func postHARJob(args []string, base, sessionID, jobType string) (JobResult, error) {
	var result JobResult
	timeoutMS := jobTimeoutMS(args, 60000)
	payload := jobsRequest{
		SessionID: sessionID,
		Type:      jobType,
		Params:    map[string]any{},
		TimeoutMS: timeoutMS,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	httpTimeout := time.Duration(timeoutMS)*time.Millisecond + 10*time.Second
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/v1/jobs", bytes.NewReader(raw))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("%s request failed: %w", jobType, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return result, err
	}
	if res.StatusCode != http.StatusOK {
		return result, fmt.Errorf("%s failed: status %d: %s", jobType, res.StatusCode, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("%s returned invalid JSON: %w", jobType, err)
	}
	if !result.OK {
		message := strings.TrimSpace(result.Error)
		if message == "" {
			message = jobType + " job failed"
		}
		return result, errors.New(message)
	}
	return result, nil
}

func cliHARStart(args []string, env map[string]string, stdout io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, harStartHelp)
		return nil
	}
	sessionID, err := resolveHARSession(args, env)
	if err != nil {
		return err
	}
	base, err := resolveCLIControlBase(args)
	if err != nil {
		return err
	}
	if err := requireChromeHARSession(base, sessionID); err != nil {
		return err
	}
	result, err := postHARJob(args, base, sessionID, JobTypeHARStart)
	if err != nil {
		return fmt.Errorf("start HAR capture: %w", err)
	}
	return writeJSONLine(stdout, result)
}

func cliHAREnd(args []string, env map[string]string, stdout io.Writer) error {
	if hasHelpFlag(args) {
		_, _ = io.WriteString(stdout, harEndHelp)
		return nil
	}
	sessionID, err := resolveHARSession(args, env)
	if err != nil {
		return err
	}
	base, err := resolveCLIControlBase(args)
	if err != nil {
		return err
	}
	if err := requireChromeHARSession(base, sessionID); err != nil {
		return err
	}
	outputDir := flagString(args, "--output-dir")
	if outputDir == "" {
		outputDir = flagString(args, "-o")
	}
	isDefault := strings.TrimSpace(outputDir) == ""
	if isDefault {
		outputDir = defaultHAROutputDir(sessionID)
	}
	if err := preflightHAROutput(outputDir, sessionID, isDefault); err != nil {
		return fmt.Errorf("prepare HAR output: %w", err)
	}
	result, err := postHARJob(args, base, sessionID, JobTypeHAREnd)
	if err != nil {
		return fmt.Errorf("end HAR capture: %w", err)
	}
	var payload harEndPayload
	rawData, err := json.Marshal(result.Data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(rawData, &payload); err != nil {
		return fmt.Errorf("end HAR capture returned invalid artifact metadata: %w", err)
	}
	if payload.CaptureID == "" || payload.ArtifactToken == "" {
		return errors.New("end HAR capture returned incomplete artifact metadata")
	}
	manifest, err := exportHARCapture(base, sessionID, outputDir, isDefault, payload)
	if err != nil {
		return err
	}
	if err := cleanupRemoteHARCapture(base, sessionID, payload.CaptureID, payload.ArtifactToken); err != nil {
		return fmt.Errorf("HAR export published but daemon cleanup failed: %w", err)
	}
	result.Data = map[string]any{
		"type":          JobTypeHAREnd,
		"capture_id":    payload.CaptureID,
		"output_dir":    outputDir,
		"manifest_path": filepath.Join(outputDir, "manifest.json"),
		"partial":       manifest.Partial,
		"tab_count":     manifest.TabCount,
		"request_count": manifest.RequestCount,
		"error_count":   manifest.ErrorCount,
	}
	return writeJSONLine(stdout, result)
}

func writeJSONLine(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(value)
}

var unsafeSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func harSlug(value, fallback string, max int) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var ascii strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			ascii.WriteRune(r)
		} else {
			ascii.WriteByte('-')
		}
	}
	slug := strings.Trim(unsafeSlugChars.ReplaceAllString(ascii.String(), "-"), "-")
	if slug == "" {
		slug = fallback
	}
	if max > 0 && len(slug) > max {
		slug = strings.TrimRight(slug[:max], "-")
	}
	if slug == "" {
		return fallback
	}
	return slug
}

func defaultHAROutputDir(sessionID string) string {
	return filepath.Join("/tmp", "browser-agent-"+harSlug(sessionID, "session", 64))
}

func existingOwnedHAROutput(path, sessionID string) bool {
	raw, err := os.ReadFile(filepath.Join(path, "manifest.json"))
	if err != nil {
		return false
	}
	var owner struct {
		SchemaVersion int    `json:"schema_version"`
		SessionID     string `json:"session_id"`
	}
	return json.Unmarshal(raw, &owner) == nil && owner.SchemaVersion == 1 && owner.SessionID == sessionID
}

func prepareHAROutput(path, sessionID string, isDefault bool) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("HAR output path exists and is not a directory: %s", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if existingOwnedHAROutput(path, sessionID) {
		return nil
	}
	return fmt.Errorf("refusing to replace non-empty HAR output directory: %s", path)
}

func preflightHAROutput(path, sessionID string, isDefault bool) error {
	path = filepath.Clean(path)
	if err := prepareHAROutput(path, sessionID, isDefault); err != nil {
		return err
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return err
	}
	probe, err := os.MkdirTemp(parent, ".browser-agent-har-preflight-")
	if err != nil {
		return err
	}
	return os.RemoveAll(probe)
}

func artifactURL(base, sessionID, captureID, token string, tabID int) string {
	values := url.Values{}
	values.Set("session", sessionID)
	values.Set("capture", captureID)
	values.Set("token", token)
	values.Set("tab_id", fmt.Sprintf("%d", tabID))
	return strings.TrimRight(base, "/") + "/v1/har/artifact?" + values.Encode()
}

func downloadHARArtifact(base, sessionID string, payload harEndPayload, artifact harArtifactDescriptor, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, artifactURL(base, sessionID, payload.CaptureID, payload.ArtifactToken, artifact.TabID), nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
		return fmt.Errorf("download HAR for tab %d failed: status %d: %s", artifact.TabID, res.StatusCode, strings.TrimSpace(string(body)))
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(file, io.LimitReader(res.Body, maxHARArtifactBytes+1))
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if written > maxHARArtifactBytes {
		return fmt.Errorf("HAR artifact for tab %d exceeds %d bytes", artifact.TabID, maxHARArtifactBytes)
	}
	if artifact.Size > 0 && written != artifact.Size {
		return fmt.Errorf("HAR artifact size mismatch for tab %d: expected %d bytes, received %d", artifact.TabID, artifact.Size, written)
	}
	return validateHARFile(path)
}

func exportHARCapture(base, sessionID, outputDir string, isDefault bool, payload harEndPayload) (harManifest, error) {
	var manifest harManifest
	if len(payload.Artifacts) > maxHARArtifacts {
		return manifest, fmt.Errorf("HAR capture exceeds %d artifacts", maxHARArtifacts)
	}
	var totalBytes int64
	seenArtifacts := make(map[int]struct{}, len(payload.Artifacts))
	for _, artifact := range payload.Artifacts {
		if artifact.TabID <= 0 || artifact.Size <= 0 || artifact.Size > maxHARArtifactBytes {
			return manifest, fmt.Errorf("invalid HAR artifact metadata for tab %d", artifact.TabID)
		}
		if _, exists := seenArtifacts[artifact.TabID]; exists {
			return manifest, fmt.Errorf("duplicate HAR artifact metadata for tab %d", artifact.TabID)
		}
		seenArtifacts[artifact.TabID] = struct{}{}
		if artifact.Size > maxHARTotalBytes-totalBytes {
			return manifest, fmt.Errorf("HAR capture exceeds %d total bytes", maxHARTotalBytes)
		}
		totalBytes += artifact.Size
	}
	outputDir = filepath.Clean(outputDir)
	if err := preflightHAROutput(outputDir, sessionID, isDefault); err != nil {
		return manifest, err
	}
	parent := filepath.Dir(outputDir)
	staging, err := os.MkdirTemp(parent, ".browser-agent-har-staging-")
	if err != nil {
		return manifest, err
	}
	defer os.RemoveAll(staging)
	if err := os.Chmod(staging, 0o700); err != nil {
		return manifest, err
	}

	artifacts := make(map[int]harArtifactDescriptor, len(payload.Artifacts))
	for _, artifact := range payload.Artifacts {
		artifacts[artifact.TabID] = artifact
	}
	sort.Slice(payload.Tabs, func(i, j int) bool { return payload.Tabs[i].TabID < payload.Tabs[j].TabID })
	for i := range payload.Tabs {
		tab := &payload.Tabs[i]
		artifact, ok := artifacts[tab.TabID]
		if !ok {
			return manifest, fmt.Errorf("missing HAR artifact for tab %d", tab.TabID)
		}
		tab.File = fmt.Sprintf("tab-%d-%s.har", tab.TabID, harSlug(tab.Title, "untitled", 60))
		if err := downloadHARArtifact(base, sessionID, payload, artifact, filepath.Join(staging, tab.File)); err != nil {
			return manifest, err
		}
	}
	if len(payload.Tabs) != len(payload.Artifacts) {
		return manifest, fmt.Errorf("HAR tab/artifact count mismatch: %d tabs, %d artifacts", len(payload.Tabs), len(payload.Artifacts))
	}

	manifest = harManifest{
		SchemaVersion: 1,
		SessionID:     sessionID,
		CaptureID:     payload.CaptureID,
		Browser:       "chrome",
		StartedAt:     payload.StartedAt,
		EndedAt:       payload.EndedAt,
		Partial:       payload.Partial,
		Warnings:      payload.Warnings,
		TabCount:      len(payload.Tabs),
		Tabs:          payload.Tabs,
	}
	for _, tab := range manifest.Tabs {
		manifest.RequestCount += tab.RequestCount
		manifest.ErrorCount += tab.ErrorCount
		if tab.ErrorCount > 0 {
			manifest.Partial = true
		}
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return manifest, err
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(filepath.Join(staging, "manifest.json"), manifestBytes, 0o600); err != nil {
		return manifest, err
	}
	if err := publishHAROutput(staging, outputDir, sessionID, isDefault); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func publishHAROutput(staging, outputDir, sessionID string, isDefault bool) error {
	_, statErr := os.Stat(outputDir)
	if errors.Is(statErr, os.ErrNotExist) {
		return os.Rename(staging, outputDir)
	}
	if statErr != nil {
		return statErr
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	if len(entries) > 0 && !existingOwnedHAROutput(outputDir, sessionID) {
		return fmt.Errorf("refusing to replace non-empty HAR output directory: %s", outputDir)
	}
	backup := outputDir + ".previous-" + fmt.Sprintf("%d", time.Now().UnixNano())
	if err := os.Rename(outputDir, backup); err != nil {
		return err
	}
	if err := os.Rename(staging, outputDir); err != nil {
		_ = os.Rename(backup, outputDir)
		return err
	}
	return os.RemoveAll(backup)
}

func cleanupRemoteHARCapture(base, sessionID, captureID, token string) error {
	values := url.Values{}
	values.Set("session", sessionID)
	values.Set("capture", captureID)
	values.Set("token", token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, strings.TrimRight(base, "/")+"/v1/har/capture?"+values.Encode(), nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if readErr != nil {
		return readErr
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var response struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid cleanup response: %w", err)
	}
	if !response.OK {
		return errors.New("daemon did not confirm HAR cleanup")
	}
	return nil
}
