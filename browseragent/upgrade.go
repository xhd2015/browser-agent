package browseragent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultReattachWait is how long EnsureDaemon waits after upgrade for waitList
// sessions to show extension.connected again.
const DefaultReattachWait = 30 * time.Second

// DefaultPrepareSlackMS is extra pre-kill sleep after prepare delay_ms.
const DefaultPrepareSlackMS = 200

// DefaultReattachPollInterval is the default poll cadence for WaitSessionsReconnected.
const DefaultReattachPollInterval = 200 * time.Millisecond

// SessionConnView is the minimal connectivity view for upgrade wait-list helpers.
// Maps from session snapshot extension.connected + session_id.
type SessionConnView struct {
	SessionID string
	Connected bool
}

// ConnectedIDsFromSnapshots returns session ids with Connected==true in input order.
// Nil/empty input → empty slice. Skips blank SessionID.
func ConnectedIDsFromSnapshots(sessions []SessionConnView) []string {
	if len(sessions) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(sessions))
	for _, s := range sessions {
		if !s.Connected {
			continue
		}
		id := strings.TrimSpace(s.SessionID)
		if id == "" {
			continue
		}
		out = append(out, id)
	}
	return out
}

// WaitSessionsReconnected polls fetch until every waitList id is Connected, or timeout.
//   - timeout <= 0 → default DefaultReattachWait (30s)
//   - pollInterval <= 0 → DefaultReattachPollInterval (200ms)
//   - empty waitList → return immediately (reattached=[], stillWaiting=[])
//   - fetch errors are retryable until timeout (do not hard-fail on first error)
//   - reattached: waitList ids Connected at the last successful poll by deadline
//   - stillWaiting: waitList ids not Connected at deadline
//
// Order of returned slices follows waitList order.
func WaitSessionsReconnected(
	waitList []string,
	fetch func() ([]SessionConnView, error),
	timeout time.Duration,
	pollInterval time.Duration,
) (reattached, stillWaiting []string) {
	reattached = []string{}
	stillWaiting = []string{}
	if len(waitList) == 0 {
		return reattached, stillWaiting
	}
	if timeout <= 0 {
		timeout = DefaultReattachWait
	}
	if pollInterval <= 0 {
		pollInterval = DefaultReattachPollInterval
	}
	if fetch == nil {
		return []string{}, append([]string{}, waitList...)
	}

	deadline := time.Now().Add(timeout)
	var connected map[string]bool
	haveSnapshot := false

	for {
		views, err := fetch()
		if err == nil {
			connected = make(map[string]bool, len(views))
			for _, v := range views {
				if v.Connected {
					id := strings.TrimSpace(v.SessionID)
					if id != "" {
						connected[id] = true
					}
				}
			}
			haveSnapshot = true
			all := true
			for _, id := range waitList {
				if !connected[id] {
					all = false
					break
				}
			}
			if all {
				out := make([]string, 0, len(waitList))
				for _, id := range waitList {
					out = append(out, id)
				}
				return out, []string{}
			}
		}

		if !time.Now().Before(deadline) {
			break
		}
		time.Sleep(pollInterval)
	}

	if !haveSnapshot {
		return []string{}, append([]string{}, waitList...)
	}
	for _, id := range waitList {
		if connected[id] {
			reattached = append(reattached, id)
		} else {
			stillWaiting = append(stillWaiting, id)
		}
	}
	return reattached, stillWaiting
}

func ensureDaemonClientVersion(cfg EnsureDaemonConfig) string {
	if v := strings.TrimSpace(cfg.ClientVersion); v != "" {
		return v
	}
	return ClientVersion()
}

func writeUpgradeWarning(w io.Writer, format string, args ...any) {
	if w == nil {
		w = io.Discard
	}
	fmt.Fprintf(w, "browser-agent: warning: "+format+"\n", args...)
}

func connectedSessionIDs(sessions []sessionSnapshot) []string {
	views := sessionSnapshotsToConnViews(sessions)
	return ConnectedIDsFromSnapshots(views)
}

func disconnectedSessionIDs(sessions []sessionSnapshot) []string {
	var ids []string
	for _, s := range sessions {
		if !s.Extension.Connected && strings.TrimSpace(s.SessionID) != "" {
			ids = append(ids, s.SessionID)
		}
	}
	return ids
}

func sessionSnapshotsToConnViews(sessions []sessionSnapshot) []SessionConnView {
	out := make([]SessionConnView, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, SessionConnView{
			SessionID: s.SessionID,
			Connected: s.Extension.Connected,
		})
	}
	return out
}

func disconnectedIDsFromViews(views []SessionConnView) []string {
	var ids []string
	for _, v := range views {
		if !v.Connected {
			id := strings.TrimSpace(v.SessionID)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func formatSessionList(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return strings.Join(ids, ", ")
}

func upgradeWarnConnected(w io.Writer, ids []string, daemonVer, clientVer string) {
	n := len(ids)
	label := "sessions"
	if n == 1 {
		label = "session"
	}
	writeUpgradeWarning(w, "cannot upgrade daemon (%d extension-connected %s: %s); reusing daemon v%s (client v%s)",
		n, label, formatSessionList(ids), EffectiveDaemonVersion(daemonVer), clientVer)
}

func upgradeWarnOrphans(w io.Writer, ids []string) {
	writeUpgradeWarning(w, "upgrading daemon; disconnected sessions will be orphaned: %s", formatSessionList(ids))
}

func warnOlderClient(w io.Writer, clientVer, daemonVer string) {
	writeUpgradeWarning(w, "client v%s older than daemon v%s; reusing", clientVer, EffectiveDaemonVersion(daemonVer))
}

func warnKillExistingConnected(w io.Writer, ids []string) {
	n := len(ids)
	label := "sessions"
	if n == 1 {
		label = "session"
	}
	writeUpgradeWarning(w, "killing daemon (%d extension-connected %s: %s)", n, label, formatSessionList(ids))
}

func removeSessionDirs(baseDir string, ids []string) {
	for _, id := range ids {
		_ = os.RemoveAll(SessionDirPath(baseDir, id))
	}
}

func removeAllSessionDirs(baseDir string) {
	sessionsRoot := strings.TrimRight(baseDir, string(os.PathSeparator)) + string(os.PathSeparator) + "sessions"
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			_ = os.RemoveAll(strings.TrimRight(sessionsRoot, string(os.PathSeparator)) + string(os.PathSeparator) + e.Name())
		}
	}
}

func resolvePrepareDelayMS(cfg EnsureDaemonConfig) (delayMS, slackMS int) {
	delayMS = cfg.PrepareDelayMS
	slackMS = cfg.PrepareSlackMS
	if delayMS <= 0 {
		delayMS = DefaultPrepareReconnectDelayMS
		// Product zero-value config: apply default slack with default delay.
		if slackMS == 0 {
			slackMS = DefaultPrepareSlackMS
		}
	}
	if slackMS < 0 {
		slackMS = DefaultPrepareSlackMS
	}
	return delayMS, slackMS
}

func sleepFnOrDefault(cfg EnsureDaemonConfig) func(time.Duration) {
	if cfg.SleepFn != nil {
		return cfg.SleepFn
	}
	return time.Sleep
}

func fetchSessionConnViews(cfg EnsureDaemonConfig, baseURL string) ([]SessionConnView, error) {
	if cfg.FetchSessionsFn != nil {
		return cfg.FetchSessionsFn(baseURL)
	}
	sessions, err := fetchDaemonSessions(baseURL)
	if err != nil {
		return nil, err
	}
	return sessionSnapshotsToConnViews(sessions), nil
}

// defaultPrepareUpgradePOST POSTs /v1/admin/prepare-upgrade (best-effort caller ignores errors).
func defaultPrepareUpgradePOST(baseURL string, delayMS int) error {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return fmt.Errorf("empty baseURL")
	}
	payload, err := json.Marshal(map[string]any{
		"delay_ms": delayMS,
		"reason":   "daemon-upgrade",
	})
	if err != nil {
		return err
	}
	ctxTO := 5 * time.Second
	client := &http.Client{Timeout: ctxTO}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/admin/prepare-upgrade", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("prepare-upgrade status=%d", res.StatusCode)
	}
	return nil
}

func reportUpgradeStderr(w io.Writer, oldVer, newVer string, reattached, stillWaiting []string) {
	if w == nil {
		w = io.Discard
	}
	oldVer = EffectiveDaemonVersion(oldVer)
	newVer = EffectiveDaemonVersion(newVer)
	fmt.Fprintf(w, "browser-agent: upgraded daemon v%s → v%s\n", oldVer, newVer)
	if len(reattached) > 0 {
		fmt.Fprintf(w, "browser-agent: reattached sessions: %s\n", formatSessionList(reattached))
	}
	if len(stillWaiting) > 0 {
		writeUpgradeWarning(w, "still waiting for extension reattach: %s", formatSessionList(stillWaiting))
	}
}

// ensureDaemonUpgradeClientNewer runs the client>daemon upgrade path:
// prepare-upgrade (best-effort) → sleep → kill+spawn (no dir wipe) → wait ready →
// wait reattach ≤ ReattachWait → stderr upgraded/reattached/still-waiting.
// Connected sessions no longer block upgrade (Phase 3).
func ensureDaemonUpgradeClientNewer(cfg EnsureDaemonConfig, meta DaemonMeta, baseURL, daemonVer, clientVer string, wantAddr string, readyTimeout time.Duration, stderr io.Writer) (DaemonMeta, error) {
	views, _ := fetchSessionConnViews(cfg, baseURL)
	waitList := ConnectedIDsFromSnapshots(views)
	orphans := disconnectedIDsFromViews(views)
	if len(orphans) > 0 {
		upgradeWarnOrphans(stderr, orphans)
	}

	delayMS, slackMS := resolvePrepareDelayMS(cfg)

	prepareFn := cfg.PrepareUpgradeFn
	if prepareFn == nil {
		prepareFn = defaultPrepareUpgradePOST
	}
	// Best-effort: old daemons may lack the admin route.
	_ = prepareFn(baseURL, delayMS)

	sleep := sleepFnOrDefault(cfg)
	sleep(time.Duration(delayMS+slackMS) * time.Millisecond)

	if err := ensureDaemonKillAndRespawn(cfg, meta, stderr, orphans); err != nil {
		return DaemonMeta{}, err
	}

	newMeta, err := waitDaemonReady(cfg, wantAddr, readyTimeout)
	if err != nil {
		return DaemonMeta{}, err
	}

	reattachWait := cfg.ReattachWait
	if reattachWait <= 0 {
		reattachWait = DefaultReattachWait
	}
	newBaseURL := daemonMetaBaseURL(newMeta)
	if newBaseURL == "" {
		newBaseURL = baseURL
	}
	fetch := func() ([]SessionConnView, error) {
		return fetchSessionConnViews(cfg, newBaseURL)
	}
	reattached, stillWaiting := WaitSessionsReconnected(waitList, fetch, reattachWait, 0)

	newVer := strings.TrimSpace(newMeta.DaemonVersion)
	if newVer == "" {
		newVer = fetchDaemonVersion(newBaseURL)
		newMeta.DaemonVersion = newVer
	}
	if strings.TrimSpace(newVer) == "" {
		newVer = clientVer
	}
	reportUpgradeStderr(stderr, daemonVer, newVer, reattached, stillWaiting)
	return newMeta, nil
}
