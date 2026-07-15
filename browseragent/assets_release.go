package browseragent

import "fmt"

// AssetReleaseNames returns release archive basenames for a version, matching
// EnsureAsset download names: {product}_v{version}_{kind}.tar.gz
//
// For version "v0.2.0" / "0.2.0" the set includes at least:
//
//	browser-agent_v0.2.0_session-page.tar.gz
//	browser-agent_v0.2.0_extension.tar.gz
//	browser-trace_v0.2.0_extension.tar.gz
func AssetReleaseNames(version string) []string {
	v := normalizeCacheVersion(version)
	if v == "" || v == "v" {
		return nil
	}
	return []string{
		fmt.Sprintf("%s_%s_%s.tar.gz", ProductName, v, AssetKindSessionPage),
		fmt.Sprintf("%s_%s_%s.tar.gz", ProductName, v, AssetKindExtension),
		fmt.Sprintf("browser-trace_%s_%s.tar.gz", v, AssetKindExtension),
	}
}
