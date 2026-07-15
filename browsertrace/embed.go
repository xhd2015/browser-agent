package browsertrace

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// embeddedExtension is the staged MV3 tree under embedded/extension/.
// Git tracks only placeholder.txt; full trees come from script/browser-trace/bundle
// (or install) and are gitignored. Incomplete embeds hydrate at runtime
// (docs/assets-hydrate.md).
//
//go:embed embedded/extension/**
var embeddedExtension embed.FS

const embeddedExtensionRoot = "embedded/extension"

// ExtractEmbeddedExtension writes the embedded MV3 package under
// {baseDir}/extension/{version}/ and returns the absolute install path and
// version string from manifest.json. Idempotent for the same version.
func ExtractEmbeddedExtension(baseDir string) (installPath, version string, err error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", "", fmt.Errorf("baseDir is required")
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve baseDir: %w", err)
	}

	maniBytes, err := embeddedExtension.ReadFile(embeddedExtensionRoot + "/manifest.json")
	if err != nil {
		return "", "", fmt.Errorf("read embedded manifest.json: %w", err)
	}
	var mani struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(maniBytes, &mani); err != nil {
		return "", "", fmt.Errorf("parse embedded manifest.json: %w", err)
	}
	version = strings.TrimSpace(mani.Version)
	if version == "" {
		return "", "", fmt.Errorf("embedded manifest.json has empty version")
	}

	dest := filepath.Join(absBase, "extension", version)
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return "", "", fmt.Errorf("create extract dir: %w", err)
	}

	// Walk embed FS and copy every file under embedded/extension.
	err = fs.WalkDir(embeddedExtension, embeddedExtensionRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(embeddedExtensionRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Embed paths use forward slashes; normalize for OS.
		rel = filepath.FromSlash(rel)
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := embeddedExtension.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		// Always rewrite so re-extract refreshes files while keeping path stable.
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return "", "", fmt.Errorf("extract embedded extension: %w", err)
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", "", err
	}
	return absDest, version, nil
}

// InstallChromeExtension extracts the embedded extension and writes user-facing
// Load unpacked instructions to w. Output ends with a trailing newline.
func InstallChromeExtension(w io.Writer, baseDir string) error {
	if w == nil {
		w = io.Discard
	}
	if strings.TrimSpace(baseDir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.TempDir()
		}
		baseDir = filepath.Join(home, ".tmp", "browser-trace")
	}
	path, version, err := ExtractEmbeddedExtension(baseDir)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, `Chrome extension extracted for browser-trace (version %s).

Install / load the unpacked extension:

  1. Open chrome://extensions
  2. Enable Developer mode (top-right toggle)
  3. Click Load unpacked
  4. Select this folder:

     %s

After loading, keep the session page open so the extension can connect.
`, version, path)
	return err
}

// BuildChromeLaunchArgs returns Chrome argv (without the binary name) for a
// best-effort launch: new window, load-extension, session URL.
// Does not include --user-data-dir (uses the default profile).
func BuildChromeLaunchArgs(sessionURL, extensionPath string) []string {
	args := []string{"--new-window"}
	if strings.TrimSpace(extensionPath) != "" {
		args = append(args, "--load-extension="+extensionPath)
	}
	if strings.TrimSpace(sessionURL) != "" {
		args = append(args, sessionURL)
	}
	return args
}
