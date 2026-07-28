## Expected

- After Run (BuildFirefox + Bundle UseFixture under temp roots), **firefox embed** exists:
  - `FirefoxEmbedDir` non-empty absolute path containing `embedded/extension-firefox`
  - `FirefoxManifestOK` true — `manifest.json` readable
- Manifest should look like a Firefox package when content is staged from fixtures/public:
  - prefer `browser_specific_settings` / `gecko` / non-empty `"version"`
- BuildFirefoxExtensionShell may succeed (`BuildShellDir` set) but **does not** satisfy
  this leaf alone — embed path is required.

## Side Effects

- Writes only under temp BundleRoot / ShellRoot.

## Errors

- Missing `browseragent/embedded/extension-firefox/manifest.json` fails Phase 1 E1.

## Exit Code

- Not asserted (package API).

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if strings.TrimSpace(resp.FirefoxEmbedDir) == "" {
		t.Fatal("FirefoxEmbedDir is empty; expected browseragent/embedded/extension-firefox under temp Root")
	}
	if !filepath.IsAbs(resp.FirefoxEmbedDir) {
		t.Fatalf("FirefoxEmbedDir must be absolute; got %q", resp.FirefoxEmbedDir)
	}
	norm := filepath.ToSlash(resp.FirefoxEmbedDir)
	if !strings.Contains(norm, "embedded/extension-firefox") {
		t.Fatalf("FirefoxEmbedDir should contain embedded/extension-firefox; got %q", resp.FirefoxEmbedDir)
	}
	if !resp.FirefoxManifestOK {
		t.Fatalf("firefox embed manifest.json missing at %q (StageAPIUsed=%q BundleErr=%q BuildShellDir=%q); implement StageFirefoxExtensionEmbed or Bundle dual-stage",
			resp.FirefoxManifestPath, resp.StageAPIUsed, resp.BundleErr, resp.BuildShellDir)
	}
	text := resp.FirefoxManifestText
	low := strings.ToLower(text)
	// Soft Firefox identity: gecko id or Browser Agent Firefox naming when from real/fixture sources.
	hasGecko := strings.Contains(low, "gecko") || strings.Contains(low, "browser_specific_settings")
	hasVersion := strings.Contains(low, `"version"`) || strings.Contains(low, `"version" :`)
	if !hasVersion {
		// JSON version field required for extract/embed identity.
		if !strings.Contains(text, "version") {
			t.Fatalf("staged firefox manifest must include version; text=%s", truncate(text, 400))
		}
	}
	if !hasGecko {
		// Prefer gecko when staging Firefox sources; allow chrome-shaped interim only if
		// path is extension-firefox (still require non-empty manifest above).
		t.Logf("warning: staged firefox manifest lacks gecko/browser_specific_settings (may be interim): %s",
			truncate(text, 200))
	}
}
```
