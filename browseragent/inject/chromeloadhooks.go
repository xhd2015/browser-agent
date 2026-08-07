package inject

import (
	"context"
	"sync"
	"sync/atomic"
)

// ChromeLoadOpts is a minimal opts mirror so inject does not import the chrome
// package (keeps inject free of heavy deps for compile graphs). Callers cast via
// browseragent wiring.
//
// Prefer WithChromeLoadHooks for tests; do not assign globals under t.Parallel.
type ChromeLoadHooks struct {
	// LoadFn is invoked instead of the default chrome.LoadUnpacked when non-nil.
	// Signature uses any to avoid an inject → chrome import; browseragent passes
	// chrome.LoadUnpackedOpts / Result through its own adapter.
	LoadFn func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error
}

var chromeLoadHooks atomic.Pointer[ChromeLoadHooks]

var chromeLoadInstallMu sync.Mutex

// WithChromeLoadHooks installs h for the duration of fn.
func WithChromeLoadHooks(h *ChromeLoadHooks, fn func() error) error {
	chromeLoadInstallMu.Lock()
	defer chromeLoadInstallMu.Unlock()

	prev := chromeLoadHooks.Swap(h)
	defer func() {
		chromeLoadHooks.Store(prev)
	}()
	return fn()
}

// ChromeLoadFn returns the injected LoadFn, or nil.
func ChromeLoadFn() func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
	h := chromeLoadHooks.Load()
	if h == nil {
		return nil
	}
	return h.LoadFn
}
