package browseragent

import (
	"fmt"
	"os/exec"
	"runtime"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

// launchChromeWithArgs launches Chrome with pre-built argv (managed or system).
// Consults ManagedChromeTestHooks.LaunchFn when set (doctest argv recording).
func launchChromeWithArgs(args []string) error {
	if inj.ManagedChromeTestHooks != nil && inj.ManagedChromeTestHooks.LaunchFn != nil {
		return inj.ManagedChromeTestHooks.LaunchFn(args)
	}
	return startChromeProcess(args)
}

func startChromeProcess(args []string) error {
	switch runtime.GOOS {
	case "darwin":
		cmdArgs := append([]string{"-na", "Google Chrome", "--args"}, args...)
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

// openChrome launches default-profile Chrome in a new window with
// optional --load-extension. Never passes --user-data-dir.
// Best-effort production path; tests inject OpenChromeFn instead.
func openChrome(sessionURL, extensionPath string) error {
	args := BuildChromeArgs(sessionURL, extensionPath)
	return launchChromeWithArgs(args)
}