package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DaemonStatus is a read-only snapshot of daemon discovery state and sessions.
type DaemonStatus struct {
	Running   bool
	PID       int
	Addr      string
	BaseURL   string
	BaseDir   string
	Uptime    time.Duration
	StartedAt time.Time
	Sessions  []sessionSnapshot

	DaemonVersion    string
	ExtensionVersion string
	ExtensionMD5     string
	ExtensionPath    string
}

// QueryDaemonStatus reads {baseDir}/server.json and probes the control plane when
// the recorded pid is alive. Missing meta or a dead pid yields Running=false without
// mutating server.json.
func QueryDaemonStatus(baseDir string) (DaemonStatus, error) {
	st := DaemonStatus{BaseDir: baseDir}
	metaPath := filepath.Join(baseDir, "server.json")
	meta, ok, err := readDaemonMetaIfPresent(metaPath)
	if err != nil {
		return DaemonStatus{}, err
	}
	if !ok {
		return populateDaemonStatusExtension(&st)
	}

	st.PID = meta.PID
	st.Addr = meta.Addr
	st.BaseURL = meta.BaseURL
	st.StartedAt = meta.StartedAt
	if meta.BaseDir != "" {
		st.BaseDir = meta.BaseDir
	}

	if !IsProcessAlive(meta.PID) {
		return populateDaemonStatusExtension(&st)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(meta.BaseURL), "/")
	if baseURL == "" {
		addr := strings.TrimSpace(meta.Addr)
		if addr == "" {
			return populateDaemonStatusExtension(&st)
		}
		baseURL = "http://" + addr
		st.BaseURL = baseURL
	}

	if !daemonHealthOK(baseURL) {
		return populateDaemonStatusExtension(&st)
	}

	sessions, err := fetchDaemonSessions(baseURL)
	if err != nil {
		return populateDaemonStatusExtension(&st)
	}

	st.Running = true
	st.Sessions = sessions
	if !st.StartedAt.IsZero() {
		st.Uptime = time.Since(st.StartedAt)
		if st.Uptime < 0 {
			st.Uptime = 0
		}
	}
	st.DaemonVersion = resolveStatusDaemonVersion(meta, baseURL)
	return populateDaemonStatusExtension(&st)
}

// FormatDaemonStatus writes operator-facing status output to w.
func FormatDaemonStatus(w io.Writer, st DaemonStatus) error {
	if w == nil {
		w = io.Discard
	}

	if _, err := fmt.Fprintln(w, "browser-agent daemon status"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	if st.Running {
		if _, err := fmt.Fprintf(w, "Status:   running\n"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "PID:      %d\n", st.PID); err != nil {
			return err
		}
		if st.Addr != "" {
			if _, err := fmt.Fprintf(w, "Addr:     %s\n", st.Addr); err != nil {
				return err
			}
		}
		if st.BaseURL != "" {
			if _, err := fmt.Fprintf(w, "Base URL: %s\n", st.BaseURL); err != nil {
				return err
			}
		}
		if st.BaseDir != "" {
			if _, err := fmt.Fprintf(w, "Base dir: %s\n", st.BaseDir); err != nil {
				return err
			}
		}
		if !st.StartedAt.IsZero() {
			if _, err := fmt.Fprintf(w, "Started:  %s\n", st.StartedAt.Format(time.RFC3339)); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "Uptime:   %s\n", formatDaemonUptime(st.Uptime)); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if st.DaemonVersion != "" {
			if _, err := fmt.Fprintf(w, "Version:  %s\n", st.DaemonVersion); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if err := writeDaemonStatusExtensionBlock(w, st); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Sessions (%d)\n", len(st.Sessions)); err != nil {
			return err
		}
		return writeDaemonStatusSessionsTable(w, st.Sessions)
	}

	if _, err := fmt.Fprintf(w, "Status:   not running\n"); err != nil {
		return err
	}
	if st.BaseDir != "" {
		if _, err := fmt.Fprintf(w, "Base dir: %s\n", st.BaseDir); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := writeDaemonStatusExtensionBlock(w, st); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Sessions"); err != nil {
		return err
	}
	_, err := fmt.Fprintln(w, "(none — daemon not running)")
	return err
}

func populateDaemonStatusExtension(st *DaemonStatus) (DaemonStatus, error) {
	if st == nil {
		return DaemonStatus{}, fmt.Errorf("populateDaemonStatusExtension: nil status")
	}
	version, md5, path, err := resolveStatusExtension(st.Sessions)
	if err != nil {
		return *st, err
	}
	st.ExtensionVersion = version
	st.ExtensionMD5 = md5
	st.ExtensionPath = path
	return *st, nil
}

func resolveStatusDaemonVersion(meta DaemonMeta, baseURL string) string {
	v := strings.TrimSpace(meta.DaemonVersion)
	if v == "" {
		v = fetchDaemonVersion(baseURL)
	}
	return EffectiveDaemonVersion(v)
}

func resolveStatusExtension(sessions []sessionSnapshot) (version, md5, path string, err error) {
	for _, snap := range sessions {
		be := snap.BundledExtension
		if strings.TrimSpace(be.Path) != "" {
			path = strings.TrimSpace(be.Path)
			version = strings.TrimSpace(be.Version)
			md5 = strings.ToLower(strings.TrimSpace(be.MD5))
			if md5 == "" {
				if sum, sumErr := ReadBundleSumFromDir(path); sumErr == nil {
					md5 = strings.ToLower(strings.TrimSpace(sum.MD5))
					if version == "" {
						version = strings.TrimSpace(sum.Version)
					}
				}
			}
			return version, md5, path, nil
		}
	}

	extPath, extVer, extErr := EnsureCanonicalExtension()
	if extErr != nil {
		return "", "", "", extErr
	}
	path = extPath
	version = strings.TrimSpace(extVer)
	if sum, sumErr := ReadBundleSumFromDir(extPath); sumErr == nil {
		if strings.TrimSpace(sum.Version) != "" {
			version = strings.TrimSpace(sum.Version)
		}
		md5 = strings.ToLower(strings.TrimSpace(sum.MD5))
	}
	if md5 == "" {
		if s, e := EnsureExtensionBundleSum(extPath, version); e == nil {
			version = strings.TrimSpace(s.Version)
			md5 = strings.ToLower(strings.TrimSpace(s.MD5))
		}
	}
	return version, md5, path, nil
}

func writeDaemonStatusExtensionBlock(w io.Writer, st DaemonStatus) error {
	if strings.TrimSpace(st.ExtensionPath) == "" {
		return nil
	}
	if _, err := fmt.Fprintln(w, "Extension (embedded):"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  version  %s\n", st.ExtensionVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  md5      %s\n", st.ExtensionMD5); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "  path     %s\n", st.ExtensionPath)
	return err
}

func writeDaemonStatusSessionsTable(w io.Writer, sessions []sessionSnapshot) error {
	if sessions == nil {
		sessions = []sessionSnapshot{}
	}

	colID := "Session ID"
	colPhase := "Phase"
	colConn := "Connected"

	idW, phaseW, connW := len(colID), len(colPhase), len(colConn)
	for _, snap := range sessions {
		if n := len(snap.SessionID); n > idW {
			idW = n
		}
		if n := len(snap.Phase); n > phaseW {
			phaseW = n
		}
		conn := formatDaemonStatusConnected(snap.Extension.Connected)
		if n := len(conn); n > connW {
			connW = n
		}
	}
	idW += 2
	phaseW += 2
	connW += 2

	if _, err := fmt.Fprintf(w, "%-*s %-*s %-*s\n", idW, colID, phaseW, colPhase, connW, colConn); err != nil {
		return err
	}
	for _, snap := range sessions {
		conn := formatDaemonStatusConnected(snap.Extension.Connected)
		if _, err := fmt.Fprintf(w, "%-*s %-*s %-*s\n", idW, snap.SessionID, phaseW, snap.Phase, connW, conn); err != nil {
			return err
		}
	}
	return nil
}

func formatDaemonStatusConnected(connected bool) string {
	if connected {
		return "yes"
	}
	return "no"
}

func readDaemonMetaIfPresent(path string) (DaemonMeta, bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return DaemonMeta{}, false, nil
		}
		return DaemonMeta{}, false, fmt.Errorf("stat daemon meta: %w", err)
	}
	meta, err := ReadDaemonMeta(path)
	if err != nil {
		return DaemonMeta{}, false, err
	}
	return meta, true, nil
}

func fetchDaemonSessions(baseURL string) ([]sessionSnapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/sessions", nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /v1/sessions status=%d", res.StatusCode)
	}

	var list []sessionSnapshot
	if err := json.Unmarshal(body, &list); err == nil {
		if list == nil {
			return []sessionSnapshot{}, nil
		}
		return list, nil
	}

	var wrap map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrap); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}
	raw, ok := wrap["sessions"]
	if !ok {
		return []sessionSnapshot{}, nil
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}
	if list == nil {
		return []sessionSnapshot{}, nil
	}
	return list, nil
}

func formatDaemonUptime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		return d.String()
	}
	return d.Round(time.Second).String()
}