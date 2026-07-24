package browseragent

import (
	"os"
	"sync"
)

// processEnvMu serializes process-wide env mutations.
// Workspace doctest leaves run under t.Parallel() and must not use t.Setenv;
// callers that temporarily change HOME / XDG_CACHE_HOME / etc. must hold this
// mutex for the entire critical section (set → use → restore).
var processEnvMu sync.Mutex

// WithProcessEnv applies env via os.Setenv for the duration of fn, restoring
// prior values afterward. Holds processEnvMu for the whole critical section so
// EnsureCanonicalExtension / AssetCacheRoot callers outside isolation wait.
//
// IMPORTANT: fn must not call public APIs that re-lock processEnvMu
// (EnsureCanonicalExtension, AssetCacheRoot, nested WithProcessEnv). Prefer
// unlocked helpers from the same package, or set env then call work that only
// reads HOME/XDG once the lock is already held via LockProcessEnv.
//
// For CLI assets ensure/status, use unlocked cache helpers under this lock.
func WithProcessEnv(env map[string]string, fn func() error) error {
	processEnvMu.Lock()
	defer processEnvMu.Unlock()

	if len(env) == 0 {
		return fn()
	}

	type prev struct {
		val string
		ok  bool
	}
	saved := make(map[string]prev, len(env))
	for k, v := range env {
		old, ok := os.LookupEnv(k)
		saved[k] = prev{val: old, ok: ok}
		if err := os.Setenv(k, v); err != nil {
			for sk, sp := range saved {
				if !sp.ok {
					_ = os.Unsetenv(sk)
				} else {
					_ = os.Setenv(sk, sp.val)
				}
			}
			return err
		}
	}
	defer func() {
		for k, p := range saved {
			if !p.ok {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, p.val)
			}
		}
	}()
	return fn()
}

// ApplyProcessEnvLocked sets keys from env while processEnvMu is already held.
// Used by CLI helpers that set env then call EnsureAsset in the same section.
// Prefer WithProcessEnv for most call sites.
func applyProcessEnvLocked(env map[string]string) (restore func()) {
	if len(env) == 0 {
		return func() {}
	}
	type prev struct {
		val string
		ok  bool
	}
	saved := make(map[string]prev, len(env))
	for k, v := range env {
		old, ok := os.LookupEnv(k)
		saved[k] = prev{val: old, ok: ok}
		_ = os.Setenv(k, v)
	}
	return func() {
		for k, p := range saved {
			if !p.ok {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, p.val)
			}
		}
	}
}

// LockProcessEnv / UnlockProcessEnv expose the mutex for long critical sections
// that mix Setenv with multiple API calls (e.g. Setup-less Run bodies).
func LockProcessEnv()   { processEnvMu.Lock() }
func UnlockProcessEnv() { processEnvMu.Unlock() }
