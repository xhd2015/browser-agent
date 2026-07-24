package inject

import (
	"sync"
	"sync/atomic"
)

// ManagedChromeHooks records injectable managed-chrome hooks for CLI doctests.
type ManagedChromeHooks struct {
	LaunchFn func(args []string) error
}

// ManagedChromeTestHooks is a compatibility mirror of the active CLI hooks.
// Prefer OpenManagedChromeConfig.LaunchFn for package-API tests.
// For HandleCLI leaves, install only via WithManagedChromeHooks.
//
// Direct assignment is racy under t.Parallel(); use WithManagedChromeHooks.
var ManagedChromeTestHooks *ManagedChromeHooks

// managedChromeHooks holds the active CLI LaunchFn adapter.
// Reads are lock-free (atomic). Writers serialize via managedChromeInstallMu
// for the full set → use → restore window so parallel leaves do not cross-wire.
var managedChromeHooks atomic.Pointer[ManagedChromeHooks]

var managedChromeInstallMu sync.Mutex

// WithManagedChromeHooks installs h for the duration of fn, restoring the prior
// value afterward. Serializes installers for the whole critical section
// (set → fn → restore). fn should be short (e.g. one HandleCLI call).
//
// Do not nest WithManagedChromeHooks on the same goroutine (deadlock).
// LaunchFn snapshots (ManagedChromeLaunchFn) are lock-free and safe to call
// from within fn.
func WithManagedChromeHooks(h *ManagedChromeHooks, fn func() error) error {
	managedChromeInstallMu.Lock()
	defer managedChromeInstallMu.Unlock()

	prev := managedChromeHooks.Swap(h)
	ManagedChromeTestHooks = h
	defer func() {
		managedChromeHooks.Store(prev)
		ManagedChromeTestHooks = prev
	}()
	return fn()
}

// ManagedChromeLaunchFn returns a snapshot of the active LaunchFn (may be nil).
// Lock-free so it is safe to call while WithManagedChromeHooks holds the install mutex.
func ManagedChromeLaunchFn() func(args []string) error {
	h := managedChromeHooks.Load()
	if h == nil {
		return nil
	}
	return h.LaunchFn
}
