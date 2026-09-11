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
	// Prefer real binary so --new-window joins the existing Chrome session
	// (avoids bare open -a URL tab reuse and open -a --args no-op).
	// macOS: if binary missing, open -na --args as last resort.
	return launchBrowserArgs(ResolveChromeBinary, "Google Chrome", "Chrome", "chrome", args)
}

// openChrome launches/reuses default-profile Chrome in a new window and
// navigates to sessionURL exactly once. Never passes --user-data-dir or
// --load-extension (see BuildChromeArgs).
//
// Attach/register retries belong on the session page and extension (content
// script + SW), not here. CLI must not re-open the URL (that creates duplicate
// /go tabs).
func openChrome(sessionURL, extensionPath string) error {
	sessionURL = strings.TrimSpace(sessionURL)
	if sessionURL == "" {
		return fmt.Errorf("open chrome: empty session URL")
	}
	args := BuildChromeArgs(sessionURL, extensionPath)
	return launchChromeWithArgs(args)
}

// openExtensionWakeup opens chrome-extension://<id>/wakeup.html to start the
// MV3 service worker (same proven path as the toolbar popup). Prefers a
// background tab in the existing window (no --new-window focus steal); wakeup.js
// heals /go tabs and closes the tab quickly.
func openExtensionWakeup(extensionInstallPath string) error {
	id := ChromeUnpackedExtensionID(extensionInstallPath)
	if id == "" {
		return fmt.Errorf("open extension wakeup: empty extension id (path %q)", extensionInstallPath)
	}
	url := "chrome-extension://" + id + "/wakeup.html"
	if runtime.GOOS == "darwin" {
		if err := openChromeBackgroundTabDarwin(url); err == nil {
			return nil
		}
		// Fall through to argv tab open if AppleScript fails (Chrome not running, etc.).
	}
	// Existing Chrome session: bare URL opens a tab (avoid --new-window).
	return launchChromeWithArgs([]string{url})
}

// openChromeBackgroundTabDarwin opens url as a new tab then restores the previous
// active tab so the operator keeps focus. Does not activate Chrome.
func openChromeBackgroundTabDarwin(pageURL string) error {
	pageURL = strings.TrimSpace(pageURL)
	if pageURL == "" {
		return fmt.Errorf("open chrome background tab: empty url")
	}
	// AppleScript string: escape backslashes and quotes.
	escaped := strings.ReplaceAll(pageURL, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	script := `
tell application "Google Chrome"
  if (count of windows) = 0 then
    make new window
  end if
  tell front window
    set prevIndex to active tab index
    make new tab with properties {URL:"` + escaped + `"}
    try
      set active tab index to prevIndex
    end try
  end tell
end tell`
	cmd := exec.Command("osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript background tab: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}
