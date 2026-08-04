package browseragent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Env vars for operator override of browser binaries (session new / open).
const (
	EnvChromePath  = "BROWSER_AGENT_CHROME_PATH"
	EnvFirefoxPath = "BROWSER_AGENT_FIREFOX_PATH"
)

// browserBinaryResolveOpts customizes resolve for production and tests.
type browserBinaryResolveOpts struct {
	// Getenv defaults to os.Getenv.
	Getenv func(key string) string
	// LookPath defaults to exec.LookPath.
	LookPath func(file string) (string, error)
	// Home, when non-empty, is used instead of os.UserHomeDir (parallel-safe tests).
	Home string
	// PathOK reports whether a candidate path is a usable binary.
	// Defaults to fileExistsRegular (exists, not a directory).
	PathOK func(path string) bool
	// GOOS overrides runtime.GOOS for tests (empty = real GOOS).
	GOOS string
}

func defaultGetenv(key string) string {
	return os.Getenv(key)
}

func defaultLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func fileExistsRegular(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func (o browserBinaryResolveOpts) getenv(key string) string {
	if o.Getenv != nil {
		return o.Getenv(key)
	}
	return defaultGetenv(key)
}

func (o browserBinaryResolveOpts) lookPath(file string) (string, error) {
	if o.LookPath != nil {
		return o.LookPath(file)
	}
	return defaultLookPath(file)
}

func (o browserBinaryResolveOpts) pathOK(path string) bool {
	if o.PathOK != nil {
		return o.PathOK(path)
	}
	return fileExistsRegular(path)
}

func (o browserBinaryResolveOpts) homeDir() string {
	if strings.TrimSpace(o.Home) != "" {
		return o.Home
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func (o browserBinaryResolveOpts) goos() string {
	if strings.TrimSpace(o.GOOS) != "" {
		return o.GOOS
	}
	return runtime.GOOS
}

// ResolveChromeBinary finds the Google Chrome / Chromium executable.
// Order: BROWSER_AGENT_CHROME_PATH → PATH candidates → known install paths
// (macOS: ~/Applications before /Applications).
func ResolveChromeBinary() (string, error) {
	return resolveChromeBinary(browserBinaryResolveOpts{})
}

func resolveChromeBinary(o browserBinaryResolveOpts) (string, error) {
	return resolveBrowserBinary(o, browserBinarySpec{
		envKey:    EnvChromePath,
		label:     "Chrome",
		pathNames: []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser", "chrome"},
		darwinRel: "Google Chrome.app/Contents/MacOS/Google Chrome",
		windowsRels: []string{
			`Google\Chrome\Application\chrome.exe`,
		},
	})
}

// ResolveFirefoxBinary finds the Firefox executable.
// Order: BROWSER_AGENT_FIREFOX_PATH → PATH candidates → known install paths
// (macOS: ~/Applications before /Applications).
func ResolveFirefoxBinary() (string, error) {
	return resolveFirefoxBinary(browserBinaryResolveOpts{})
}

func resolveFirefoxBinary(o browserBinaryResolveOpts) (string, error) {
	return resolveBrowserBinary(o, browserBinarySpec{
		envKey:    EnvFirefoxPath,
		label:     "Firefox",
		pathNames: []string{"firefox", "firefox-esr", "firefox-bin"},
		darwinRel: "Firefox.app/Contents/MacOS/firefox",
		windowsRels: []string{
			`Mozilla Firefox\firefox.exe`,
		},
	})
}

type browserBinarySpec struct {
	envKey      string
	label       string
	pathNames   []string
	darwinRel   string   // under Applications/
	windowsRels []string // under ProgramFiles etc.
}

func resolveBrowserBinary(o browserBinaryResolveOpts, spec browserBinarySpec) (string, error) {
	if p := strings.TrimSpace(o.getenv(spec.envKey)); p != "" {
		if o.pathOK(p) {
			return p, nil
		}
		return "", fmt.Errorf("%s binary from %s not found: %s", spec.label, spec.envKey, p)
	}

	for _, name := range spec.pathNames {
		if path, err := o.lookPath(name); err == nil && strings.TrimSpace(path) != "" && o.pathOK(path) {
			return path, nil
		}
	}

	for _, p := range knownBrowserInstallPaths(o, spec) {
		if o.pathOK(p) {
			return p, nil
		}
	}

	return "", fmt.Errorf("%s binary not found (set %s or install the browser)", spec.label, spec.envKey)
}

// knownBrowserInstallPaths returns candidate absolute paths.
// macOS: ~/Applications first, then /Applications.
func knownBrowserInstallPaths(o browserBinaryResolveOpts, spec browserBinarySpec) []string {
	switch o.goos() {
	case "darwin":
		rel := strings.TrimSpace(spec.darwinRel)
		if rel == "" {
			return nil
		}
		var out []string
		if home := o.homeDir(); home != "" {
			out = append(out, filepath.Join(home, "Applications", rel))
		}
		out = append(out, filepath.Join("/Applications", rel))
		return out
	case "windows":
		return windowsBrowserInstallPaths(o, spec.windowsRels)
	default:
		return nil
	}
}

func windowsBrowserInstallPaths(o browserBinaryResolveOpts, rels []string) []string {
	var roots []string
	for _, key := range []string{"PROGRAMFILES", "PROGRAMFILES(X86)", "LOCALAPPDATA"} {
		if v := strings.TrimSpace(o.getenv(key)); v != "" {
			roots = append(roots, v)
		}
	}
	var out []string
	for _, root := range roots {
		for _, rel := range rels {
			// Windows paths use backslash in rel; filepath.Join still works.
			out = append(out, filepath.Join(root, rel))
		}
	}
	return out
}

// launchBrowserArgs runs argv against a resolved binary.
// macOS fallback when no binary: open -na <appName> --args <argv>.
// Windows fallback: cmd /c start <startName> <argv>.
func launchBrowserArgs(resolve func() (string, error), appName, productLabel, windowsStartName string, args []string) error {
	bin, err := resolve()
	if err == nil && strings.TrimSpace(bin) != "" {
		cmd := exec.Command(bin, args...)
		return cmd.Start()
	}

	switch runtime.GOOS {
	case "darwin":
		fmt.Fprintf(os.Stderr, "browser-agent: warning: %s binary not found; falling back to open -na (may start a new instance)\n", productLabel)
		cmdArgs := append([]string{"-na", appName, "--args"}, args...)
		cmd := exec.Command("open", cmdArgs...)
		return cmd.Start()
	case "windows":
		if strings.TrimSpace(windowsStartName) == "" {
			windowsStartName = strings.ToLower(productLabel)
		}
		fmt.Fprintf(os.Stderr, "browser-agent: warning: %s binary not found; falling back to start %s\n", productLabel, windowsStartName)
		cmdArgs := append([]string{"/c", "start", windowsStartName}, args...)
		cmd := exec.Command("cmd", cmdArgs...)
		return cmd.Start()
	default:
		if err != nil {
			return err
		}
		return fmt.Errorf("%s binary not found", productLabel)
	}
}
