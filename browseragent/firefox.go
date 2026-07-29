package browseragent

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// BuildFirefoxOpenArgs returns Firefox argv (without the binary name) for a
// best-effort launch with the session URL only.
//
// Must NOT include managed-profile flags (--load-extension, --user-data-dir):
// Firefox temporary add-ons are loaded manually via about:debugging.
func BuildFirefoxOpenArgs(sessionURL string) []string {
	var args []string
	if strings.TrimSpace(sessionURL) != "" {
		args = append(args, sessionURL)
	}
	return args
}

// openFirefox launches default-profile Firefox with the session URL.
// Best-effort production path; tests inject OpenFirefoxFn instead.
func openFirefox(sessionURL string) error {
	args := BuildFirefoxOpenArgs(sessionURL)
	return launchFirefoxWithArgs(args)
}

// openFirefoxXPI opens a signed .xpi in Firefox so the browser prompts to install
// the permanent add-on. Best-effort; path must be an absolute filesystem path.
func openFirefoxXPI(xpiPath string) error {
	xpiPath = strings.TrimSpace(xpiPath)
	if xpiPath == "" {
		return fmt.Errorf("xpi path is required")
	}
	switch runtime.GOOS {
	case "darwin":
		// open -a Firefox file.xpi → Firefox shows the install prompt.
		cmd := exec.Command("open", "-a", "Firefox", xpiPath)
		return cmd.Start()
	case "linux":
		for _, bin := range []string{"firefox", "firefox-esr", "firefox-bin"} {
			if path, err := exec.LookPath(bin); err == nil {
				cmd := exec.Command(path, xpiPath)
				return cmd.Start()
			}
		}
		return fmt.Errorf("firefox not found on PATH")
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "firefox", xpiPath)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported OS for firefox xpi open: %s", runtime.GOOS)
	}
}

func launchFirefoxWithArgs(args []string) error {
	switch runtime.GOOS {
	case "darwin":
		cmdArgs := append([]string{"-na", "Firefox", "--args"}, args...)
		cmd := exec.Command("open", cmdArgs...)
		return cmd.Start()
	case "linux":
		for _, bin := range []string{"firefox", "firefox-esr", "firefox-bin"} {
			if path, err := exec.LookPath(bin); err == nil {
				cmd := exec.Command(path, args...)
				return cmd.Start()
			}
		}
		return fmt.Errorf("firefox not found on PATH")
	case "windows":
		cmdArgs := append([]string{"/c", "start", "firefox"}, args...)
		cmd := exec.Command("cmd", cmdArgs...)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported OS for firefox launch: %s", runtime.GOOS)
	}
}
