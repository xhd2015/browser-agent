package browseragent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Required extension permissions for production browser-agent MV3 packages.
var requiredExtensionPermissions = []string{
	"debugger",
	"tabs",
	"alarms",
	"storage",
}

// ValidateExtensionManifestJSON checks that manifest JSON bytes satisfy the
// production permission contract: debugger, tabs, alarms, storage; control
// host coverage for port 43761 (127.0.0.1 and/or localhost); and broad page
// host access (<all_urls> or equivalent).
func ValidateExtensionManifestJSON(data []byte) error {
	if len(strings.TrimSpace(string(data))) == 0 {
		return fmt.Errorf("extension manifest: empty JSON")
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("extension manifest: invalid JSON: %w", err)
	}

	perms := stringSliceField(m, "permissions")
	permSet := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		permSet[p] = struct{}{}
	}
	var missing []string
	for _, need := range requiredExtensionPermissions {
		if _, ok := permSet[need]; !ok {
			missing = append(missing, need)
		}
	}
	if len(missing) > 0 {
		// Prefer mentioning the first missing permission by name so callers
		// and tests can match "debugger" / "tabs" specifically.
		return fmt.Errorf("extension manifest: missing required permission %q (need debugger, tabs, alarms, storage; missing: %s)",
			missing[0], strings.Join(missing, ", "))
	}

	hosts := stringSliceField(m, "host_permissions")
	// Also accept host patterns mistakenly placed under permissions (lenient).
	allHostLike := append([]string{}, hosts...)
	for _, p := range perms {
		if looksLikeHostPattern(p) {
			allHostLike = append(allHostLike, p)
		}
	}

	if !hasControlHost43761(allHostLike) {
		return fmt.Errorf("extension manifest: missing host permission for control plane port 43761 (need http://127.0.0.1:43761/* and/or http://localhost:43761/*)")
	}
	if !hasBroadHostAccess(allHostLike) {
		return fmt.Errorf("extension manifest: missing broad host access (need <all_urls> or equivalent such as *://*/*)")
	}
	return nil
}

// ValidateExtensionManifestPath reads path and validates with
// ValidateExtensionManifestJSON.
func ValidateExtensionManifestPath(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("extension manifest: read %s: %w", path, err)
	}
	return ValidateExtensionManifestJSON(data)
}

func stringSliceField(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	raw, ok := m[key]
	if !ok || raw == nil {
		return nil
	}
	switch t := raw.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, v := range t {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return append([]string(nil), t...)
	default:
		return nil
	}
}

func looksLikeHostPattern(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if s == "<all_urls>" {
		return true
	}
	return strings.Contains(s, "://") || strings.HasPrefix(s, "*")
}

func hasControlHost43761(hosts []string) bool {
	for _, h := range hosts {
		hl := strings.ToLower(strings.TrimSpace(h))
		if !strings.Contains(hl, "43761") {
			continue
		}
		// Accept 127.0.0.1 or localhost on port 43761 (any path suffix / scheme).
		if strings.Contains(hl, "127.0.0.1") || strings.Contains(hl, "localhost") {
			return true
		}
	}
	return false
}

func hasBroadHostAccess(hosts []string) bool {
	hasHTTPStar, hasHTTPSStar := false, false
	for _, h := range hosts {
		hl := strings.TrimSpace(h)
		switch hl {
		case "<all_urls>", "*://*/*", "*://*/":
			return true
		case "http://*/*", "http://*/":
			hasHTTPStar = true
		case "https://*/*", "https://*/":
			hasHTTPSStar = true
		}
		// Also accept match patterns that cover all hosts for both schemes.
		low := strings.ToLower(hl)
		if low == "<all_urls>" || strings.HasPrefix(low, "*://*") {
			return true
		}
	}
	return hasHTTPStar && hasHTTPSStar
}
