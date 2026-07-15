package browseragent

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Asset resolve source labels returned by ResolveSessionPage.
const (
	AssetSourceEmbed = "embed"
	AssetSourceCache = "cache"
)

// ResolveSessionPage prefers a complete embedFS (source "embed"); otherwise
// EnsureAsset(browser-agent, version, session-page) and reads the index from
// the cache dir (source "cache").
func ResolveSessionPage(embedFS fs.FS, version string, cfg AssetDownloadConfig) (html, source string, err error) {
	cfg = resolveDownloadConfig(cfg)
	if EmbedCompleteFS(embedFS, AssetKindSessionPage) {
		return ResolveSessionPageIndexFS(embedFS)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dir, err := EnsureAsset(ctx, ProductName, version, AssetKindSessionPage, cfg)
	if err != nil {
		return "", "", err
	}
	body, err := readSessionPageIndexFS(os.DirFS(dir))
	if err != nil {
		return "", "", fmt.Errorf("session-page cache incomplete: %w", err)
	}
	if strings.TrimSpace(body) == "" {
		return "", "", fmt.Errorf("session-page cache incomplete: empty index")
	}
	return body, AssetSourceCache, nil
}

// ResolveExtensionDir prefers a complete embedFS: materializes it under baseDir
// and returns that install path. If embed is incomplete, EnsureAsset for
// extension and returns the cache dir as installPath.
// ver is the normalized version string used for the ensure/cache key (or
// manifest version when materializing a complete embed).
func ResolveExtensionDir(embedFS fs.FS, baseDir, version string, cfg AssetDownloadConfig) (installPath, ver string, err error) {
	cfg = resolveDownloadConfig(cfg)
	ver = normalizeCacheVersion(version)

	if EmbedCompleteFS(embedFS, AssetKindExtension) {
		// Prefer manifest version for on-disk layout when present.
		if mv := readManifestVersionFS(embedFS); mv != "" {
			ver = mv
		}
		if strings.TrimSpace(baseDir) == "" {
			return "", ver, fmt.Errorf("ResolveExtensionDir: baseDir is required when materializing embed")
		}
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			return "", ver, fmt.Errorf("ResolveExtensionDir: resolve baseDir: %w", err)
		}
		dest := filepath.Join(absBase, ver)
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return "", ver, fmt.Errorf("ResolveExtensionDir: mkdir: %w", err)
		}
		if err := copyFSTree(embedFS, dest); err != nil {
			return "", ver, fmt.Errorf("ResolveExtensionDir: materialize embed: %w", err)
		}
		if !EmbedCompleteFS(os.DirFS(dest), AssetKindExtension) {
			return "", ver, fmt.Errorf("ResolveExtensionDir: materialize incomplete for extension")
		}
		absDest, err := filepath.Abs(dest)
		if err != nil {
			return dest, ver, nil
		}
		return absDest, ver, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dir, err := EnsureAsset(ctx, ProductName, version, AssetKindExtension, cfg)
	if err != nil {
		return "", ver, err
	}
	return dir, ver, nil
}

// resolveDownloadConfig fills BaseURL from BROWSER_AGENT_ASSET_BASE_URL when empty.
func resolveDownloadConfig(cfg AssetDownloadConfig) AssetDownloadConfig {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		if v := strings.TrimSpace(os.Getenv("BROWSER_AGENT_ASSET_BASE_URL")); v != "" {
			cfg.BaseURL = v
		}
	}
	return cfg
}

func readManifestVersionFS(fsys fs.FS) string {
	if fsys == nil {
		return ""
	}
	data, err := fs.ReadFile(fsys, "manifest.json")
	if err != nil {
		return ""
	}
	var mani struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &mani); err != nil {
		return ""
	}
	return strings.TrimSpace(mani.Version)
}
