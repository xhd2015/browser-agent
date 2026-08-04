package browseragent

import (
	"fmt"
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
// navigates to sessionURL exactly once. Never passes --user-data-dir.
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
