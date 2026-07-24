package browseragent

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Mini session-page fixture fingerprints (browseragent/fixtures + staged poller).
// Full install must never treat these as a production React SPA.
const fixtureSessionPageAssetMarker = "fixture session-page asset"

// SessionPageDirIsMiniFixture reports whether dir looks like the mini/test
// session-page tree (poller shim or hydrate test fixture), not a vite SPA.
func SessionPageDirIsMiniFixture(dir string) bool {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return true
	}
	// Explicit fixture markers in any JS under assets/.
	assetsDir := filepath.Join(dir, "assets")
	if st, err := os.Stat(assetsDir); err == nil && st.IsDir() {
		var fixtureHit bool
		var totalJS int64
		_ = filepath.WalkDir(assetsDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			if !strings.HasSuffix(strings.ToLower(d.Name()), ".js") {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			totalJS += info.Size()
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return nil
			}
			s := string(data)
			if strings.Contains(s, fixtureSessionPageAssetMarker) ||
				strings.Contains(s, `data-fixture`) ||
				strings.Contains(s, "session-page-complete") {
				fixtureHit = true
			}
			return nil
		})
		if fixtureHit {
			return true
		}
		// Real vite multi-chunk apps are far larger than the ~300B poller.
		// A single tiny assets/*.js with empty #root shell is still a mini embed.
		if totalJS > 0 && totalJS < 2048 {
			indexBody, _ := os.ReadFile(filepath.Join(dir, "index.html"))
			if len(indexBody) == 0 {
				indexBody, _ = os.ReadFile(filepath.Join(dir, "session-page.html"))
			}
			low := strings.ToLower(string(indexBody))
			// Production vite builds use type="module" and often hashed assets;
			// mini fixtures use a plain <script src="/assets/session-page.js">.
			if !strings.Contains(low, `type="module"`) && !strings.Contains(low, `type='module'`) {
				return true
			}
		}
	}

	// index-only / tiny shell with install details and no module entry.
	indexBody, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		indexBody, err = os.ReadFile(filepath.Join(dir, "session-page.html"))
	}
	if err == nil {
		s := string(indexBody)
		if strings.Contains(s, "browser-agent session (fixture)") {
			return true
		}
		// Known mini shell: only install <details> + /assets/session-page.js, no module.
		low := strings.ToLower(s)
		if strings.Contains(low, "data-browser-agent-install") &&
			strings.Contains(low, "/assets/session-page.js") &&
			!strings.Contains(low, `type="module"`) &&
			!strings.Contains(low, `type='module'`) &&
			len(s) < 2500 {
			return true
		}
	}
	return false
}
