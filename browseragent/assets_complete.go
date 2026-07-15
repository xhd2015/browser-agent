package browseragent

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
)

// Asset kinds accepted by EmbedCompleteFS.
const (
	AssetKindSessionPage = "session-page"
	AssetKindExtension   = "extension"
)

// scriptSrcRe captures src attribute values on <script> tags.
var scriptSrcRe = regexp.MustCompile(`(?is)<script[^>]*\ssrc\s*=\s*["']([^"']+)["']`)

// EmbedCompleteFS reports whether outstanding files exist and are non-empty
// for the given asset kind on fsys (rooted at the asset tree).
// kind: "session-page" | "extension"
func EmbedCompleteFS(fsys fs.FS, kind string) bool {
	if fsys == nil {
		return false
	}
	switch strings.TrimSpace(kind) {
	case AssetKindSessionPage:
		return sessionPageComplete(fsys)
	case AssetKindExtension:
		return extensionComplete(fsys)
	default:
		return false
	}
}

// SessionPageEmbedComplete reports whether the live //go:embed session-page tree is complete.
func SessionPageEmbedComplete() bool {
	sub, err := fs.Sub(embeddedSessionPage, embeddedSessionPageRoot)
	if err != nil {
		return false
	}
	return EmbedCompleteFS(sub, AssetKindSessionPage)
}

// ExtensionEmbedComplete reports whether the live //go:embed extension tree is complete.
func ExtensionEmbedComplete() bool {
	sub, err := fs.Sub(embeddedExtension, embeddedExtensionRoot)
	if err != nil {
		return false
	}
	return EmbedCompleteFS(sub, AssetKindExtension)
}

// ResolveSessionPageIndexFS returns index HTML when the session-page embed FS is complete.
// On success: source == "embed".
// On incomplete: non-nil error mentioning incomplete / not available (no download).
func ResolveSessionPageIndexFS(fsys fs.FS) (html, source string, err error) {
	if fsys == nil || !EmbedCompleteFS(fsys, AssetKindSessionPage) {
		return "", "", fmt.Errorf("session-page embed incomplete / not available")
	}
	body, err := readSessionPageIndexFS(fsys)
	if err != nil {
		return "", "", fmt.Errorf("session-page embed incomplete: %w", err)
	}
	if strings.TrimSpace(body) == "" {
		return "", "", fmt.Errorf("session-page embed incomplete: empty index")
	}
	return body, "embed", nil
}

// ResolveSessionPageIndex resolves the live embedded session-page index when complete.
func ResolveSessionPageIndex() (html, source string, err error) {
	sub, err := fs.Sub(embeddedSessionPage, embeddedSessionPageRoot)
	if err != nil {
		return "", "", fmt.Errorf("session-page embed incomplete / not available: %w", err)
	}
	return ResolveSessionPageIndexFS(sub)
}

func sessionPageComplete(fsys fs.FS) bool {
	indexName, indexBody, ok := readNonEmptySessionIndex(fsys)
	if !ok {
		return false
	}
	_ = indexName
	if hasNonEmptyAssetsJS(fsys) {
		return true
	}
	return hasLocalScriptSrcPresent(fsys, indexBody)
}

func extensionComplete(fsys fs.FS) bool {
	return nonEmptyFile(fsys, "manifest.json") && nonEmptyFile(fsys, "background.js")
}

func readNonEmptySessionIndex(fsys fs.FS) (name, body string, ok bool) {
	for _, candidate := range []string{"index.html", "session-page.html"} {
		data, err := fs.ReadFile(fsys, candidate)
		if err != nil || len(data) == 0 {
			continue
		}
		return candidate, string(data), true
	}
	return "", "", false
}

func readSessionPageIndexFS(fsys fs.FS) (string, error) {
	name, body, ok := readNonEmptySessionIndex(fsys)
	if !ok {
		return "", fmt.Errorf("no non-empty index.html or session-page.html")
	}
	_ = name
	return body, nil
}

func nonEmptyFile(fsys fs.FS, name string) bool {
	data, err := fs.ReadFile(fsys, name)
	return err == nil && len(data) > 0
}

// hasNonEmptyAssetsJS reports whether any non-empty *.js exists under assets/.
func hasNonEmptyAssetsJS(fsys fs.FS) bool {
	found := false
	err := fs.WalkDir(fsys, "assets", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".js") {
			return nil
		}
		if nonEmptyFile(fsys, path) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return false
	}
	return found
}

// hasLocalScriptSrcPresent checks index HTML for local <script src=...> that
// resolve to non-empty files in fsys (leading "/" stripped).
func hasLocalScriptSrcPresent(fsys fs.FS, indexHTML string) bool {
	matches := scriptSrcRe.FindAllStringSubmatch(indexHTML, -1)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		src := strings.TrimSpace(m[1])
		if src == "" || isRemoteScriptSrc(src) {
			continue
		}
		// strip query/hash and leading slash for FS lookup
		path := stripURLPathExtras(src)
		path = strings.TrimPrefix(path, "/")
		if path == "" {
			continue
		}
		if nonEmptyFile(fsys, path) {
			return true
		}
	}
	return false
}

func isRemoteScriptSrc(src string) bool {
	low := strings.ToLower(src)
	return strings.HasPrefix(low, "http://") ||
		strings.HasPrefix(low, "https://") ||
		strings.HasPrefix(low, "//") ||
		strings.HasPrefix(low, "data:") ||
		strings.HasPrefix(low, "blob:")
}

func stripURLPathExtras(src string) string {
	if i := strings.IndexAny(src, "?#"); i >= 0 {
		return src[:i]
	}
	return src
}
