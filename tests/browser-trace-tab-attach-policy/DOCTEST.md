# browser-trace — tab attach policy (capturable URL + attempt attach gates)

Exercises the **pure tab-attach policy** used while browser-trace is **recording**
in a pinned window. New tabs often start with an empty or `chrome://` URL; attach
must **not** permanently give up — when the tab later navigates to a capturable
URL, the agent should attempt debugger attach.

This tree covers two pure Go helpers (extension JS mirrors the same rules):

1. **`IsCapturableTabURL(url string) bool`** — whether a tab URL is eligible for
   `chrome.debugger` attach (skip empty, chrome internals, extensions, about:blank).
2. **`ShouldAttemptAttach(recording, windowMatch, alreadyAttached, url) bool`** —
   conjunction of session gates + capturable URL.

**No Chrome, no control server, no filesystem.** Leaves only call the pure
helpers exported from package `browsertrace`.

This tree is **separate** from:

- `./tests/browser-trace-exclude-clear-preview/` — `ShouldCaptureURL` (do **not**
  store control-host *request* traffic). Attach policy intentionally **allows**
  attaching the debugger to the control page (`http://127.0.0.1:43759/go`).
- `./tests/browser-trace/` — sealed lifecycle / HAR complete
- `./tests/browser-trace-session-page/` — `/v1/session` + `/go` status UI

Product rules relevant here:

| Rule | Detail |
|------|--------|
| Non-capturable | empty, `chrome://…`, `chrome-extension://…`, `devtools://…`, `about:blank` |
| Capturable | `http(s)://…` including product control page `http://127.0.0.1:43759/go` |
| Attempt attach | `recording && windowMatch && !alreadyAttached && IsCapturableTabURL(url)` |
| Events (product) | `tabs.onCreated` / `tabs.onUpdated` call the same gate (wiring out of scope) |

## Version

0.0.2

# DSN (Domain Specific Notion)

**User** records with **browser-trace** in a pinned Chrome window. The
**Extension Agent** must attach the debugger to tabs that will carry app traffic.

**Tab lifecycle problem**: on `tabs.onCreated` the URL is often empty or
`chrome://newtab/`. Attach at that moment fails or is skipped. When the tab
later navigates to `https://…` (`tabs.onUpdated` with url or status complete),
policy must allow a **new** attach attempt (no permanent give-up).

**Attach Policy** (pure helpers under test):

```
URL
  -> IsCapturableTabURL
       false: empty | chrome:// | chrome-extension:// | devtools:// | about:blank
       true:  http(s)://… (including control page http://127.0.0.1:43759/…)

recording, windowMatch, alreadyAttached, URL
  -> ShouldAttemptAttach
       true  iff recording && windowMatch && !alreadyAttached && IsCapturableTabURL(URL)
       false otherwise
```

**Participants**:

- **Recording session** — `recording` flag; only while true may attach.
- **Pinned window** — `windowMatch` is true when `tab.windowId == targetWindowId`.
- **Attached set** — `alreadyAttached` when tab id is already in `attachedTabs`
  (avoid double attach).
- **Tab URL** — current `tab.url` (may be empty at create time).

**Out of scope for this tree**: actual `chrome.debugger.attach`, Network.enable,
HAR buffer, control-server HTTP, and whether *request capture* stores control
host traffic (`ShouldCaptureURL` is a different helper).

**Test Client** builds `Request` fixtures in Setup and calls both pure helpers
via root `Run` — no Chrome process.

## Decision Tree

```
browser-trace-tab-attach-policy
├── capturable-url/                         [URL class → IsCapturable; gates default open]
│   ├── reject/                               expect Capturable=false; Attempt=false
│   │   ├── empty/                              "" (create-time blank)
│   │   ├── chrome-newtab/                      chrome://newtab/
│   │   ├── chrome-extension/                   chrome-extension://…/page.html
│   │   ├── devtools/                           devtools://devtools/…
│   │   └── about-blank/                        about:blank (prefer false until real nav)
│   └── allow/                                expect Capturable=true; Attempt=true
│       ├── https-app/                          https://app.example.com/… (#4)
│       └── control-page/                       http://127.0.0.1:43759/go (#7 attach OK)
└── attach-gates/                           [URL fixed capturable; one gate blocks]
    ├── not-recording/                          recording=false → Attempt=false (#6)
    ├── wrong-window/                           windowMatch=false → Attempt=false (#6)
    └── already-attached/                       alreadyAttached=true → Attempt=false (#5)
```

### Parameter significance (high → low)

1. **URL capturability class** — non-capturable schemes/empty vs capturable http(s)
   (drives both helpers; foundation of skip-list alignment with `attachAllTabsInWindow`).
2. **Attach session gates** — recording / window match / already attached
   (only meaningful once URL is capturable; each gate independently blocks attempt).
3. **Concrete URL instance** — which non-capturable scheme or which allow URL
   (empty vs chrome vs extension vs devtools vs about:blank; https app vs control page).

## Test Index

| Leaf | Scenario (requirement #) |
|------|--------------------------|
| `capturable-url/reject/empty` | (#1) empty URL → Capturable false; Attempt false |
| `capturable-url/reject/chrome-newtab` | (#2) `chrome://newtab/` → not capturable; Attempt false |
| `capturable-url/reject/chrome-extension` | (#3) `chrome-extension://…` → not capturable; Attempt false |
| `capturable-url/reject/devtools` | (skip-list) `devtools://…` → not capturable; Attempt false |
| `capturable-url/reject/about-blank` | (product prefer) `about:blank` → not capturable; Attempt false |
| `capturable-url/allow/https-app` | (#4) HTTPS app URL + gates open → Capturable true; Attempt true |
| `capturable-url/allow/control-page` | (#7) control page loopback → Capturable true; Attempt true |
| `attach-gates/not-recording` | (#6) capturable URL, not recording → Attempt false |
| `attach-gates/wrong-window` | (#6) capturable URL, wrong window → Attempt false |
| `attach-gates/already-attached` | (#5) capturable URL, already attached → Attempt false |

## How to Run

```sh
cd tests/browser-trace-tab-attach-policy
doctest vet .
doctest test -v .
# or from repo root:
doctest vet ./tests/browser-trace-tab-attach-policy
doctest test ./tests/browser-trace-tab-attach-policy
# related pure-policy regression (different helper — control traffic exclude):
doctest test ./tests/browser-trace-exclude-clear-preview
```

Requires package `github.com/xhd2015/browser-agent/browsertrace` exporting:

```go
// IsCapturableTabURL reports whether a tab's URL is eligible for debugger attach.
// false for: empty/whitespace, chrome://, chrome-extension://, devtools://, about:blank
// true for:  http:// and https:// (including product control page on 127.0.0.1:43759)
func IsCapturableTabURL(url string) bool

// ShouldAttemptAttach reports whether the agent should call attach for this tab now.
// true iff recording && windowMatch && !alreadyAttached && IsCapturableTabURL(url)
func ShouldAttemptAttach(recording, windowMatch, alreadyAttached bool, url string) bool
```

Leaves fail to compile / run red until the implementer adds these helpers
(TDD red → green). Extension JS in `Chrome-Ext-Capture-API` (and embedded copy)
should mirror the same rules on `onCreated` / `onUpdated`.

### Expected pure API (implementer)

| Helper | Contract |
|--------|----------|
| `IsCapturableTabURL("")` | false |
| `IsCapturableTabURL("chrome://newtab/")` | false |
| `IsCapturableTabURL("chrome-extension://id/path")` | false |
| `IsCapturableTabURL("devtools://devtools/bundled/…")` | false |
| `IsCapturableTabURL("about:blank")` | false |
| `IsCapturableTabURL("https://app.example.com/app")` | true |
| `IsCapturableTabURL("http://127.0.0.1:43759/go")` | true |
| `ShouldAttemptAttach(true, true, false, capturableURL)` | true |
| `ShouldAttemptAttach(false, true, false, capturableURL)` | false |
| `ShouldAttemptAttach(true, false, false, capturableURL)` | false |
| `ShouldAttemptAttach(true, true, true, capturableURL)` | false |
| `ShouldAttemptAttach(true, true, false, nonCapturableURL)` | false |

**Distinction from `ShouldCaptureURL`:** control page traffic is **not** stored
in the capture buffer (`ShouldCaptureURL` → false), but the control tab **may**
be debugger-attached (`IsCapturableTabURL` → true) so multi-tab attach stays
consistent with `attachAllTabsInWindow` (which only skips chrome/extension/devtools).

```go
import (
	"testing"

	"github.com/xhd2015/browser-agent/browsertrace"
)

// Request is narrowed root→leaf by Setup functions.
// Defaults (root SETUP): Recording=true, WindowMatch=true, AlreadyAttached=false
// so capturable-url leaves exercise URL class with gates open.
type Request struct {
	// URL is the tab URL passed to both pure helpers.
	// Empty string is a valid input (create-time blank tab) — do not treat as missing.
	URL string

	// Session gates for ShouldAttemptAttach.
	Recording       bool
	WindowMatch     bool
	AlreadyAttached bool

	// Expected outcomes (set by grouping/leaf Setup).
	WantCapturable bool
	WantAttempt    bool

	// AssertCapturable / AssertAttempt control which fields Assert must check.
	// Root defaults both true; leaves may leave them true (assert both).
	AssertCapturable bool
	AssertAttempt    bool
}

// Response holds pure policy results.
type Response struct {
	Capturable bool
	Attempt    bool
}

// Run invokes both pure helpers. No I/O, no Chrome.
func Run(t *testing.T, req *Request) (*Response, error) {
	t.Helper()
	// URL may be empty; that is intentional for reject/empty.
	capturable := browsertrace.IsCapturableTabURL(req.URL)
	attempt := browsertrace.ShouldAttemptAttach(
		req.Recording,
		req.WindowMatch,
		req.AlreadyAttached,
		req.URL,
	)
	return &Response{
		Capturable: capturable,
		Attempt:    attempt,
	}, nil
}
```
