# Scenario

**Feature**: extension identity — bundle-sum parse/hash/write + serve match + orange TTY

```
# Pure helpers
Test Client -> ParseBundleSumJS(bundle-sum.js bytes) -> BundleSum | error
Test Client -> ComputeExtensionContentMD5(dir) -> md5 (excludes bundle-sum.js)
Test Client -> WriteBundleSumJS(dir, version, md5) -> dir/bundle-sum.js
Test Client -> ColorOrangeIfTTY(s, isTTY) -> orange ANSI or plain

# Serve identity + hello match
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun)
Serve Runtime -> extract extension -> ensure bundle-sum.js -> log embedded version+md5
Serve Runtime -> meta.json {extension_version, extension_md5}
Fake Extension -> WS hello {version, features, bundle_md5?}
GET /v1/session -> bundled_extension + extension_match
```

## Preconditions

- Package `github.com/xhd2015/browser-agent/browseragent` is importable.
- Tree root is `tests/browser-agent-extension-identity/`; **ModuleRoot** =
  `filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))`.
- RED until implementer exports:
  - `ParseBundleSumJS`, `WriteBundleSumJS`, `ComputeExtensionContentMD5`
  - `ColorOrangeIfTTY` (and serve/session fields for match)
- No real Chrome / agent-run; fake WS only for session-match hello leaves.
- Each server leaf uses isolated temp `BaseDir` and free loopback `Addr`.

## Steps

1. Resolve `ModuleRoot` from `DOCTEST_ROOT`.
2. Allocate unique temp `BaseDir` for leaves that extract or serve.
3. Default `SessionID`, short `ReadyTimeout`, short `HelloSettle`.
4. Leave `Mode` and surface-specific fields for grouping/leaf Setup.

## Context

- Spec version **0.0.2**.
- Parallel-safe: temp dirs + free ports per leaf.
- Shared helpers below available to all descendant Assert/Setup packages.
- Prefer package APIs over building `cmd/browser-agent` binary.
- `extension_match` enum:
  `not_connected | ok | version_mismatch | md5_mismatch | md5_unknown`.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ModuleRoot = filepath.Clean(filepath.Join(DOCTEST_ROOT, "..", ".."))
	dir := t.TempDir()
	req.BaseDir = filepath.Join(dir, "browser-agent-base")
	if err := os.MkdirAll(req.BaseDir, 0o755); err != nil {
		return err
	}
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("ba-extid-%d", time.Now().UnixNano()%1e12)
	}
	if req.ReadyTimeout == 0 {
		req.ReadyTimeout = 5 * time.Second
	}
	if req.HelloSettle == 0 {
		req.HelloSettle = 100 * time.Millisecond
	}
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}

func assertNoRunErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run transport/setup error: %v", err)
	}
}

func assertExitZero(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.ExitCode != 0 {
		t.Fatalf("ExitCode=%d, want 0; ParseErr=%q ErrText=%q stderr=%s",
			resp.ExitCode, resp.ParseErr, resp.ErrText, truncate(resp.Stderr, 400))
	}
}

func assertHTTP200(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != 200 {
		t.Fatalf("HTTP status=%d, want 200; body=%s", resp.StatusCode, truncate(resp.BodyString, 400))
	}
}

func assertHexMD5(t *testing.T, s string, label string) {
	t.Helper()
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		t.Fatalf("%s empty", label)
	}
	ok, _ := regexp.MatchString(`^[0-9a-f]{32}$`, s)
	if !ok {
		t.Fatalf("%s=%q, want 32 lowercase hex chars", label, s)
	}
}

func assertContainsAll(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			t.Fatalf("expected text to contain %q; got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertContainsFold(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	low := strings.ToLower(haystack)
	for _, n := range needles {
		if !strings.Contains(low, strings.ToLower(n)) {
			t.Fatalf("expected text to contain %q (fold); got:\n%s", n, truncate(haystack, 800))
		}
	}
}

func assertBundledExtensionPresent(t *testing.T, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.BundledVersion) == "" && strings.TrimSpace(resp.MetaExtVersion) == "" {
		t.Fatalf("bundled_extension.version / meta extension_version missing; body=%s meta=%s",
			truncate(resp.BodyString, 400), truncate(resp.MetaJSON, 300))
	}
	md5 := resp.BundledMD5
	if md5 == "" {
		md5 = resp.MetaExtMD5
	}
	if md5 != "" {
		assertHexMD5(t, md5, "bundled md5")
	} else {
		// Prefer non-empty md5 once identity is wired; allow empty only if
		// body also lacks bundled_extension entirely (then fail).
		if resp.Raw != nil {
			if _, ok := resp.Raw["bundled_extension"]; !ok {
				t.Fatalf("bundled_extension object missing in /v1/session; body=%s",
					truncate(resp.BodyString, 500))
			}
		}
		t.Fatalf("bundled_extension.md5 empty; body=%s", truncate(resp.BodyString, 500))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func validBundleSumFixture(version, md5hex string) []byte {
	return []byte(fmt.Sprintf(
		"// browser-agent bundle-sum — generated; do not edit\nvar BROWSER_AGENT_BUNDLE_VERSION = %q;\nvar BROWSER_AGENT_BUNDLE_MD5 = %q;\n",
		version, md5hex,
	))
}
```
