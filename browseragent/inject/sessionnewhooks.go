package inject

import (
	"sync"
	"sync/atomic"
)

// SessionNewHooks records injectable session-new hooks for CLI doctests.
type SessionNewHooks struct {
	OpenChromeFn    func(sessionURL, extensionInstallPath string) error
	OpenFirefoxFn   func(sessionURL string) error
	AgentRunProbeFn func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
}

// SessionNewTestHooks is a compatibility mirror of the active CLI hooks.
// Prefer SessionNewConfig.OpenChromeFn for package-API tests.
// For HandleCLI("session","new",…) leaves, install only via WithSessionNewHooks.
//
// Direct assignment is racy under t.Parallel(); use WithSessionNewHooks.
var SessionNewTestHooks *SessionNewHooks

var sessionNewHooks atomic.Pointer[SessionNewHooks]

var sessionNewInstallMu sync.Mutex

// WithSessionNewHooks installs h for the duration of fn, restoring the prior
// value afterward. Serializes installers for the whole critical section.
//
// Do not nest WithSessionNewHooks on the same goroutine (deadlock).
// Snapshot helpers are lock-free and safe to call from within fn.
func WithSessionNewHooks(h *SessionNewHooks, fn func() error) error {
	sessionNewInstallMu.Lock()
	defer sessionNewInstallMu.Unlock()

	prev := sessionNewHooks.Swap(h)
	SessionNewTestHooks = h
	defer func() {
		sessionNewHooks.Store(prev)
		SessionNewTestHooks = prev
	}()
	return fn()
}

// SessionNewOpenChromeFn returns a snapshot of OpenChromeFn (may be nil).
func SessionNewOpenChromeFn() func(sessionURL, extensionInstallPath string) error {
	h := sessionNewHooks.Load()
	if h == nil {
		return nil
	}
	return h.OpenChromeFn
}

// SessionNewOpenFirefoxFn returns a snapshot of OpenFirefoxFn (may be nil).
func SessionNewOpenFirefoxFn() func(sessionURL string) error {
	h := sessionNewHooks.Load()
	if h == nil {
		return nil
	}
	return h.OpenFirefoxFn
}

// SessionNewAgentRunProbeFn returns a snapshot of AgentRunProbeFn (may be nil).
func SessionNewAgentRunProbeFn() func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error {
	h := sessionNewHooks.Load()
	if h == nil {
		return nil
	}
	return h.AgentRunProbeFn
}
