## Expected

- Process exit code **0**.
- Stdout (preferred) includes a line `out: <abs-path>` naming the script-created
  output directory; fallback “packing into …” path is acceptable if `out:` is absent
  (Run parses either into `req.OutDir` / listing).
- Parsed out path is non-empty, absolute, and is a directory.
- Soft: path contains `browser-agent-release-assets` (MkdirTemp pattern hint).
- Under that directory: exactly **3** non-empty `.tar.gz` archives whose basenames
  match `browseragent.AssetReleaseNames(version)` for `v0.2.0`.

## Side Effects

- Writes under the script temp dir only (not under the test’s own `t.TempDir` unless
  the implementer chose that path).
- Does not call GitHub / `gh`.
- Temp dir need not be deleted on success.

## Errors

- Non-zero exit, missing/unparseable out path, missing/empty archives, or basename
  mismatch fails.

## Exit Code

- 0.

```go
import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertExitZero(t, resp)

	// Prefer explicit out: line on stdout; Run should already have filled OutDir.
	parsed, ok := parseOutDirFromStdout(resp.Stdout)
	if !ok {
		parsed, ok = parseOutDirFromStdout(resp.Stdout + "\n" + resp.Stderr)
	}
	outDir := req.OutDir
	if ok {
		outDir = parsed
	}
	if strings.TrimSpace(outDir) == "" {
		t.Fatalf("could not determine pack out dir from stdout (prefer 'out: <path>'); stdout=%s stderr=%s",
			truncate(resp.Stdout, 500), truncate(resp.Stderr, 300))
	}
	if !filepath.IsAbs(outDir) {
		t.Fatalf("out dir must be absolute; got %q", outDir)
	}
	st, statErr := os.Stat(outDir)
	if statErr != nil || !st.IsDir() {
		t.Fatalf("out dir %q not a directory: %v", outDir, statErr)
	}
	// Soft contract with os.MkdirTemp("", "browser-agent-release-assets-*").
	if !strings.Contains(outDir, "browser-agent-release-assets") {
		t.Fatalf("out dir should come from MkdirTemp pattern browser-agent-release-assets-*; got %q", outDir)
	}

	// Ensure listing reflects the discovered out dir (Run may have already listed).
	if len(resp.OutBasenames) == 0 {
		resp.OutBasenames, resp.OutSizes = listOutArchives(t, outDir)
	}

	want := expectedReleaseNames(req.Version)
	if len(want) != 3 {
		t.Fatalf("AssetReleaseNames(%q) len=%d want 3: %v", req.Version, len(want), want)
	}

	got := append([]string(nil), resp.OutBasenames...)
	sort.Strings(got)
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)

	if len(got) != 3 {
		t.Fatalf("out archive count=%d want 3 under %s; got=%v stdout=%s stderr=%s",
			len(got), outDir, got, truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
	for i := range wantSorted {
		if got[i] != wantSorted[i] {
			t.Fatalf("out basenames mismatch under %s:\n  got  %v\n  want %v", outDir, got, wantSorted)
		}
	}
	for _, name := range want {
		sz, ok := resp.OutSizes[name]
		if !ok {
			t.Fatalf("missing size for %q under %s; sizes=%v", name, outDir, resp.OutSizes)
		}
		if sz <= 0 {
			t.Fatalf("archive %q size=%d want > 0 (dir %s)", name, sz, outDir)
		}
	}

	// Preferred operator token: out: <path> on stdout (not only stderr).
	if !strings.Contains(resp.Stdout, "out:") {
		// Soft-fail only if we already accepted packing-into fallback via Run.
		// Hard-require the preferred token so implementers print out: <path>.
		t.Fatalf("stdout should include preferred token 'out: <abs-path>'; stdout=%s",
			truncate(resp.Stdout, 500))
	}
}
```
