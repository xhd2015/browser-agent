package browseragent

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openChrome launches default-profile Chrome in a new window with
// optional --load-extension. Never passes --user-data-dir.
// Best-effort production path; tests inject OpenChromeFn instead.
func openChrome(sessionURL, extensionPath string) error {
	args := BuildChromeArgs(sessionURL, extensionPath)
	switch runtime.GOOS {
	case "darwin":
		// open -na "Google Chrome" --args <chrome-args...>
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
