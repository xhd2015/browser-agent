package browseragent

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AssetCacheRoot returns the local asset cache root directory.
//
//	if XDG_CACHE_HOME set and non-empty:
//	  filepath.Join(XDG_CACHE_HOME, "browser-agent", "asset-cache")
//	else:
//	  filepath.Join(home, ".cache", "browser-agent", "asset-cache")
func AssetCacheRoot() string {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); xdg != "" {
		return filepath.Join(xdg, "browser-agent", "asset-cache")
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, ".cache", "browser-agent", "asset-cache")
}

// AssetCacheDir joins root with product / version / kind.
// version is stored with a leading "v" (e.g. "0.2.0" → "v0.2.0"; "v0.2.0" kept).
// Layout: {AssetCacheRoot}/{product}/{version}/{kind}
func AssetCacheDir(product, version, kind string) string {
	return filepath.Join(AssetCacheRoot(), strings.TrimSpace(product), normalizeCacheVersion(version), strings.TrimSpace(kind))
}

// WriteAssetCache copies src tree into the cache key dir and returns the absolute dir written.
// Prefers atomic temp-sibling write + rename.
func WriteAssetCache(product, version, kind string, src fs.FS) (dir string, err error) {
	if src == nil {
		return "", fmt.Errorf("WriteAssetCache: src is nil")
	}
	product = strings.TrimSpace(product)
	kind = strings.TrimSpace(kind)
	if product == "" || kind == "" {
		return "", fmt.Errorf("WriteAssetCache: product and kind are required")
	}
	version = normalizeCacheVersion(version)
	if version == "" || version == "v" {
		return "", fmt.Errorf("WriteAssetCache: version is required")
	}

	dest := AssetCacheDir(product, version, kind)
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("WriteAssetCache: mkdir parent: %w", err)
	}

	// Stage into a temp sibling under parent, then rename into place.
	tmp, err := os.MkdirTemp(parent, ".asset-cache-"+kind+"-*")
	if err != nil {
		return "", fmt.Errorf("WriteAssetCache: mkdir temp: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(tmp)
		}
	}()

	if err := copyFSTree(src, tmp); err != nil {
		return "", fmt.Errorf("WriteAssetCache: copy: %w", err)
	}

	// Replace existing dest if present.
	if err := os.RemoveAll(dest); err != nil {
		return "", fmt.Errorf("WriteAssetCache: remove old dest: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		// Cross-device rename can fail; fall back to copy+remove.
		if err2 := os.MkdirAll(dest, 0o755); err2 != nil {
			return "", fmt.Errorf("WriteAssetCache: rename/mkdir dest: %v / %w", err, err2)
		}
		if err2 := copyDirContents(tmp, dest); err2 != nil {
			return "", fmt.Errorf("WriteAssetCache: fallback copy: %w", err2)
		}
		_ = os.RemoveAll(tmp)
	}
	cleanup = false

	abs, err := filepath.Abs(dest)
	if err != nil {
		return dest, nil
	}
	return abs, nil
}

// OpenAssetCache opens the cache dir for the key if present and complete
// (same rules as EmbedCompleteFS on that kind).
// ok=false when missing or incomplete (err may be nil on clean miss).
func OpenAssetCache(product, version, kind string) (fsys fs.FS, dir string, ok bool, err error) {
	dir = AssetCacheDir(product, version, kind)
	st, statErr := os.Stat(dir)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, dir, false, nil
		}
		return nil, dir, false, statErr
	}
	if !st.IsDir() {
		return nil, dir, false, nil
	}
	fsys = os.DirFS(dir)
	if !EmbedCompleteFS(fsys, strings.TrimSpace(kind)) {
		return nil, dir, false, nil
	}
	return fsys, dir, true, nil
}

// CacheComplete reports whether the cache key has a complete tree
// (EmbedCompleteFS rules on DirFS). Missing dir → false.
func CacheComplete(product, version, kind string) bool {
	dir := AssetCacheDir(product, version, kind)
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return false
	}
	return EmbedCompleteFS(os.DirFS(dir), strings.TrimSpace(kind))
}

func normalizeCacheVersion(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(v), "v") && len(v) > 1 {
		// Keep as provided when already v-prefixed (tests use "v0.2.0").
		return v
	}
	return "v" + v
}

// copyFSTree walks src and writes files under destDir preserving relative paths.
func copyFSTree(src fs.FS, destDir string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == "." {
			return nil
		}
		// fs.FS paths use slash separators.
		target := filepath.Join(destDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFSFile(src, path, target)
	})
}

func copyFSFile(src fs.FS, name, dest string) error {
	in, err := src.Open(name)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// copyDirContents copies all entries from srcDir into destDir (already exists).
func copyDirContents(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyOSFile(path, target)
	})
}

func copyOSFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
