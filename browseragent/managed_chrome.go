package browseragent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

const DefaultManagedChromeRoot = "~/.browser-agent/managed-chrome"

// ManagedChromeLayout resolves filesystem paths for managed Chrome.
type ManagedChromeLayout struct {
	Root          string
	DataDir       string
	ExtensionsDir string
}

// DefaultManagedChromeLayout resolves home → default managed-chrome root.
func DefaultManagedChromeLayout() (ManagedChromeLayout, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return ManagedChromeLayout{}, fmt.Errorf("resolve home dir: %w", err)
	}
	return LayoutFromRoot(filepath.Join(home, ".browser-agent", "managed-chrome")), nil
}

// LayoutFromRoot builds layout paths under an absolute root.
func LayoutFromRoot(root string) ManagedChromeLayout {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	return ManagedChromeLayout{
		Root:          absRoot,
		DataDir:       filepath.Join(absRoot, "data"),
		ExtensionsDir: filepath.Join(absRoot, "extensions"),
	}
}

// EnsureManagedExtension syncs the embedded extension under
// {ExtensionsDir}/browser-agent/{version}/ and returns the absolute install path and version.
func EnsureManagedExtension(layout ManagedChromeLayout) (path, version string, err error) {
	if strings.TrimSpace(layout.ExtensionsDir) == "" {
		return "", "", fmt.Errorf("ExtensionsDir is required")
	}
	browserAgentDir := filepath.Join(layout.ExtensionsDir, "browser-agent")
	if err := os.MkdirAll(browserAgentDir, 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir browser-agent extensions dir: %w", err)
	}
	return extractEmbeddedExtensionDirect(browserAgentDir)
}

// BuildManagedChromeArgs returns Chrome argv (without the binary name) for a
// managed profile launch: user-data-dir, load-extension, new-window, optional url.
func BuildManagedChromeArgs(dataDir, extensionPath, url string) []string {
	args := []string{
		"--user-data-dir=" + dataDir,
		"--load-extension=" + extensionPath,
		"--new-window",
	}
	if strings.TrimSpace(url) != "" {
		args = append(args, url)
	}
	return args
}

// OpenManagedChromeConfig controls managed Chrome launch.
type OpenManagedChromeConfig struct {
	Root     string // optional override
	URL      string // optional; empty = blank window
	LaunchFn func(args []string) error
	Stdout   io.Writer
	Stderr   io.Writer
}

// OpenChromeResult captures managed Chrome launch outcomes.
type OpenChromeResult struct {
	Layout        ManagedChromeLayout
	ExtensionPath string
	ExtensionVer  string
	ChromeArgs    []string
}

// OpenManagedChrome ensures layout + extension sync, builds args, and launches Chrome.
func OpenManagedChrome(cfg OpenManagedChromeConfig) (*OpenChromeResult, error) {
	var layout ManagedChromeLayout
	if strings.TrimSpace(cfg.Root) != "" {
		layout = LayoutFromRoot(cfg.Root)
	} else {
		var err error
		layout, err = DefaultManagedChromeLayout()
		if err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(layout.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}

	extPath, extVer, err := EnsureManagedExtension(layout)
	if err != nil {
		return nil, fmt.Errorf("ensure managed extension: %w", err)
	}

	args := BuildManagedChromeArgs(layout.DataDir, extPath, cfg.URL)

	stderr := cfg.Stderr
	if stderr == nil {
		stderr = io.Discard
	}
	WarnLoadExtensionIgnored(stderr, extPath)

	launchFn := cfg.LaunchFn
	if launchFn == nil && inj.ManagedChromeTestHooks != nil && inj.ManagedChromeTestHooks.LaunchFn != nil {
		launchFn = inj.ManagedChromeTestHooks.LaunchFn
	}
	if launchFn == nil {
		launchFn = launchChromeWithArgs
	}
	if err := launchFn(args); err != nil {
		return nil, fmt.Errorf("launch chrome: %w", err)
	}

	result := &OpenChromeResult{
		Layout:        layout,
		ExtensionPath: extPath,
		ExtensionVer:  extVer,
		ChromeArgs:    append([]string(nil), args...),
	}

	if cfg.Stdout != nil {
		if err := formatOpenChromeStdout(cfg.Stdout, result, cfg.URL); err != nil {
			return result, err
		}
	}

	return result, nil
}

func formatOpenChromeStdout(w io.Writer, result *OpenChromeResult, url string) error {
	if w == nil {
		w = io.Discard
	}
	if result == nil {
		return fmt.Errorf("missing open chrome result")
	}
	ver := strings.TrimSpace(result.ExtensionVer)
	if ver != "" && !strings.HasPrefix(ver, "v") {
		ver = "v" + ver
	}
	lines := []string{
		fmt.Sprintf("extension    %s  %s", ver, result.ExtensionPath),
		fmt.Sprintf("profile      managed  %s", result.Layout.DataDir),
		"chrome       launch requested (new window)",
	}
	if strings.TrimSpace(url) != "" {
		lines = append(lines, fmt.Sprintf("url          %s", url))
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, strings.TrimRight(line, "\n")); err != nil {
			return err
		}
	}
	return nil
}