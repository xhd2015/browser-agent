package browseragent

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ResolveSessionID picks a session id for CLI/side-commands.
// Priority: flag (when flagSet) → non-empty env (when envSet) → error mentioning both.
func ResolveSessionID(flagValue string, flagSet bool, envValue string, envSet bool) (string, error) {
	if flagSet {
		return flagValue, nil
	}
	if envSet {
		if v := strings.TrimSpace(envValue); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("session id required: pass --session-id or set BROWSER_AGENT_SESSION_ID")
}

// ResolveControlBaseURLWithHostPort picks the control plane base URL for session side-commands.
// When host and port are both omitted (empty host, port 0), reads {baseDir}/server.json.
func ResolveControlBaseURLWithHostPort(baseDir, host string, port int) (string, error) {
	if strings.TrimSpace(host) != "" || port != 0 {
		addr := ComposeControlAddr(host, port)
		return "http://" + addr, nil
	}
	return ResolveControlBaseURL(baseDir, "")
}

// ResolveControlBaseURL picks the control plane base URL for session side-commands.
// Resolution order: explicit addrFlag → live daemon in {baseDir}/server.json → default :43761.
func ResolveControlBaseURL(baseDir, addrFlag string) (string, error) {
	if strings.TrimSpace(addrFlag) != "" {
		return normalizeAddr(addrFlag), nil
	}

	if strings.TrimSpace(baseDir) == "" {
		baseDir = defaultCLIBaseDir()
	}

	metaPath := filepath.Join(baseDir, "server.json")
	meta, ok, err := readDaemonMetaIfPresent(metaPath)
	if err != nil {
		return "", err
	}
	if !ok {
		return normalizeAddr(""), nil
	}

	if meta.PID > 0 && !IsProcessAlive(meta.PID) {
		return normalizeAddr(""), nil
	}

	baseURL := daemonMetaBaseURL(meta)
	if baseURL == "" {
		return normalizeAddr(""), nil
	}

	if !daemonHealthOK(baseURL) {
		return normalizeAddr(""), nil
	}

	return baseURL, nil
}

// defaultCLIBaseDir returns the default session parent directory (~/.tmp/browser-agent).
func defaultCLIBaseDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	return filepath.Join(home, ".tmp", "browser-agent")
}

// resolveCLIBaseDir reads --base-dir from args or returns the default parent directory.
func resolveCLIBaseDir(args []string) string {
	baseDir := flagString(args, "--base-dir")
	if baseDir == "" {
		return defaultCLIBaseDir()
	}
	return baseDir
}

// resolveCLIControlBase resolves session command control URL from --host/--server-port,
// legacy --addr, and server.json fallback.
func resolveCLIControlBase(args []string) (string, error) {
	baseDir := resolveCLIBaseDir(args)
	if addr := flagString(args, "--addr"); strings.TrimSpace(addr) != "" {
		return ResolveControlBaseURL(baseDir, addr)
	}
	host := flagString(args, "--host")
	port := flagInt(args, "--server-port")
	return ResolveControlBaseURLWithHostPort(baseDir, host, port)
}

func flagInt(args []string, name string) int {
	v, ok := flagStringSet(args, name)
	if !ok {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0
	}
	return n
}

func parseHostPortFromAddr(addr string) (host string, port int) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return DefaultControlHost, DefaultControlPort
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		addr = strings.TrimPrefix(strings.TrimPrefix(addr, "https://"), "http://")
	}
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return DefaultControlHost, DefaultControlPort
	}
	port, _ = strconv.Atoi(p)
	if port == 0 {
		port = DefaultControlPort
	}
	return h, port
}