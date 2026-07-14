package browseragent

import (
	"fmt"
	"io"
	"os"
	"strings"
)

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
	var ids []string
	for _, s := range sessions {
		if s.Extension.Connected {
			ids = append(ids, s.SessionID)
		}
	}
	return ids
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