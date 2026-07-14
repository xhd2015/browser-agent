# Scenario

**Feature**: pure tab-attach policy for browser-trace (capturable URL + attempt gates)

```
# Tab create often has empty / chrome:// URL → skip attach, do not give up
# Later navigate to https://… → ShouldAttemptAttach may become true

# Pure helpers (no Chrome)
Test Client -> browsertrace.IsCapturableTabURL(url)
          -> true only for http(s) tab URLs (control page included)
Test Client -> browsertrace.ShouldAttemptAttach(recording, windowMatch, alreadyAttached, url)
          -> recording && windowMatch && !alreadyAttached && IsCapturableTabURL(url)
```

## Preconditions

- Module path `github.com/xhd2015/browser-agent` is the workspace root.
- Package `browsertrace` exports:
  - `IsCapturableTabURL(url string) bool`
  - `ShouldAttemptAttach(recording, windowMatch, alreadyAttached bool, url string) bool`
- No Chrome process, no control server, no filesystem side effects.
- Default session gates (gates open): `Recording=true`, `WindowMatch=true`,
  `AlreadyAttached=false` so URL-class leaves isolate capturability.
- Gate leaves flip exactly one blocking gate with a fixed capturable URL.
- Empty `URL` is a first-class input (not a missing field).

## Steps

1. Default `Recording = true`.
2. Default `WindowMatch = true`.
3. Default `AlreadyAttached = false`.
4. Default `AssertCapturable = true` and `AssertAttempt = true`.
5. Leave `URL`, `WantCapturable`, `WantAttempt` for grouping/leaf Setup.

## Context

- Parallel-safe: pure functions, no shared mutable process state.
- Helpers below are available to all descendant Setup/Assert packages.
- `ShouldCaptureURL` (exclude control *request* traffic) is **not** under test
  here; control page attach is intentionally allowed.

```go
import (
	"testing"
)

// CapturableFixture is a stable https URL used by attach-gates leaves.
const CapturableFixture = "https://app.example.com/app/weekly"

// ControlPageFixture is the product control session page (attach allowed).
const ControlPageFixture = "http://127.0.0.1:43759/go"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Gates open by default — capturable-url leaves only vary URL.
	req.Recording = true
	req.WindowMatch = true
	req.AlreadyAttached = false
	req.AssertCapturable = true
	req.AssertAttempt = true
	return nil
}

// assertNoRunError fails if Run returned an error (helpers must not panic/err).
func assertNoRunError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("attach policy helpers should not error; got %v", err)
	}
}

// assertPolicy checks Capturable and/or Attempt against Request wants.
func assertPolicy(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if req.AssertCapturable {
		if resp.Capturable != req.WantCapturable {
			t.Fatalf("IsCapturableTabURL(%q) = %v, want %v",
				req.URL, resp.Capturable, req.WantCapturable)
		}
	}
	if req.AssertAttempt {
		if resp.Attempt != req.WantAttempt {
			t.Fatalf("ShouldAttemptAttach(recording=%v, windowMatch=%v, alreadyAttached=%v, url=%q) = %v, want %v",
				req.Recording, req.WindowMatch, req.AlreadyAttached, req.URL,
				resp.Attempt, req.WantAttempt)
		}
	}
	// Consistency: Attempt implies Capturable (when both asserted).
	if req.AssertCapturable && req.AssertAttempt && resp.Attempt && !resp.Capturable {
		t.Fatalf("Attempt=true but Capturable=false for url=%q (invariant broken)", req.URL)
	}
}
```
