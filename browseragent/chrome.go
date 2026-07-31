package browseragent

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// launchChromeWithArgs launches Chrome with pre-built argv (managed or system).
// Consults ManagedChromeTestHooks.LaunchFn when set (CLI doctest argv recording).
// Prefer OpenManagedChromeConfig.LaunchFn / SessionNewConfig.OpenChromeFn in tests.
func launchChromeWithArgs(args []string) error {
	if fn := inj.ManagedChromeLaunchFn(); fn != nil {
		return fn(args)
	}
	return startChromeProcess(args)
}

func startChromeProcess(args []string) error {
	switch runtime.GOOS {
	case "darwin":
		// Do NOT use `open -n` (force new instance). Reuse the running profile so
		// Load-unpacked extensions stay available.
		cmdArgs := append([]string{"-a", "Google Chrome", "--args"}, args...)
		cmd := exec.Command("open", cmdArgs...)
		return cmd.Start()
	case "linux":
		for _, bin := range []string{"google-chrome", "google-chrome-stable", "chromium", "chromium-browser"} {
			if path, err := exec.LookPath(bin); err == nil {
				cmd := exec.Command(path, args...)
				return cmd.Start()
			}
		}
		return fmt.Errorf("chrome/chromium not found on PATH")
	case "windows":
		cmdArgs := append([]string{"/c", "start", "chrome"}, args...)
		cmd := exec.Command("cmd", cmdArgs...)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported OS for chrome launch: %s", runtime.GOOS)
	}
}

// openChrome launches/reuses default-profile Chrome and navigates to sessionURL
// exactly once. Never passes --user-data-dir.
//
// Attach/register retries belong on the session page and extension (content
// script + SW), not here. CLI must not re-open the URL (that creates duplicate
// /go tabs).
func openChrome(sessionURL, extensionPath string) error {
	if runtime.GOOS == "darwin" {
		return openChromeDarwin(sessionURL, extensionPath)
	}
	args := BuildChromeArgs(sessionURL, extensionPath)
	return launchChromeWithArgs(args)
}

// openChromeDarwin opens the session URL once in the user's real Chrome profile.
func openChromeDarwin(sessionURL, extensionPath string) error {
	sessionURL = strings.TrimSpace(sessionURL)
	if sessionURL == "" {
		return fmt.Errorf("open chrome: empty session URL")
	}

	// Single navigation: open -a reuses the running browser (or starts it once).
	// No delayed re-open, no multi-tab retry loop.
	if err := exec.Command("open", "-a", "Google Chrome", sessionURL).Start(); err != nil {
		// Fallback: argv launch once (still without -n).
		args := BuildChromeArgs(sessionURL, extensionPath)
		if err2 := launchChromeWithArgs(args); err2 != nil {
			return fmt.Errorf("open chrome: %v (fallback: %v)", err, err2)
		}
	}
	return nil
}
