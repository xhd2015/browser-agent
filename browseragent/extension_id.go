package browseragent

import (
	"crypto/sha256"
	"path/filepath"
	"strings"
)

// ChromeUnpackedExtensionID returns the 32-char a–p Chrome extension id for an
// unpacked Load-unpacked directory path (no trailing slash).
//
// Chrome derives the id as: SHA-256(absolute path) first 16 bytes, each nibble
// mapped 0–15 → 'a'–'p'. Used so the session page can chrome.runtime.sendMessage
// (externally_connectable) when content scripts miss cold-start injection.
func ChromeUnpackedExtensionID(installPath string) string {
	p := filepath.Clean(strings.TrimSpace(installPath))
	if p == "" || p == "." {
		return ""
	}
	// Prefer absolute; relative paths produce different ids than Chrome uses.
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	// Chrome uses the path without a trailing separator.
	p = strings.TrimRight(p, string(filepath.Separator))
	sum := sha256.Sum256([]byte(p))
	var b strings.Builder
	b.Grow(32)
	for i := 0; i < 16; i++ {
		b.WriteByte('a' + (sum[i] >> 4))
		b.WriteByte('a' + (sum[i] & 0x0f))
	}
	return b.String()
}
