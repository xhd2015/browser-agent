// Pack three release asset archives from on-disk embeds for hydrate downloads.
// Names from browseragent.AssetReleaseNames(version).
// Default: pack only. --upload opt-in wraps gh (create release if missing,
// gh release upload --clobber if exists). COPYFILE_DISABLE / exclude ._* .
//
// CLI (from module root):
//
//	go run ./script/github/release-assets [flags]
//
// Flags:
//
//	--out DIR         Output directory for archives; when omitted, create via
//	                  os.MkdirTemp("", "browser-agent-release-assets-*")
//	                  and print "out: <abs-path>" on stdout (preferred last-line
//	                  token). Do not require deleting the temp dir on success.
//	                  Help: default is temp dir — NOT "required for pack".
//	--version VER     Version string; normalize like AssetReleaseNames
//	                  (default: browseragent.ClientVersion() / VERSION.txt)
//	--upload          Opt-in GitHub upload via gh
//	-h, --help        Print usage including --upload and temp --out default;
//	                  exit 0; trailing \n
//
// Pack sources (relative to module / process cwd = module root in tests):
//
//	browseragent/embedded/session-page  -> browser-agent_{ver}_session-page.tar.gz
//	browseragent/embedded/extension     -> browser-agent_{ver}_extension.tar.gz
//	browsertrace/embedded/extension     -> browser-trace_{ver}_extension.tar.gz
//
// Without --upload: write the three archives under --out (or temp default) and exit 0.
// Archive basenames MUST match browseragent.AssetReleaseNames(version) exactly.
// Prefer non-empty valid .tar.gz; exclude AppleDouble ._* entries.
// No real GitHub required for pack/help.
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/browser-agent/browseragent"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Avoid macOS resource-fork AppleDouble noise in archives / copy trees.
	_ = os.Setenv("COPYFILE_DISABLE", "1")

	var (
		outDir  string
		version string
		upload  bool
		help    bool
	)

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			help = true
		case a == "--upload":
			upload = true
		case a == "--out":
			if i+1 >= len(args) {
				return fmt.Errorf("--out requires a directory argument")
			}
			i++
			outDir = args[i]
		case strings.HasPrefix(a, "--out="):
			outDir = strings.TrimPrefix(a, "--out=")
		case a == "--version":
			if i+1 >= len(args) {
				return fmt.Errorf("--version requires a value")
			}
			i++
			version = args[i]
		case strings.HasPrefix(a, "--version="):
			version = strings.TrimPrefix(a, "--version=")
		case strings.HasPrefix(a, "-"):
			return fmt.Errorf("unrecognized flag %s (try --help)", a)
		default:
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}

	if help {
		printHelp(os.Stdout)
		return nil
	}

	if version == "" {
		version = browseragent.ClientVersion()
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("--version is required (or set browseragent VERSION.txt)")
	}

	names := browseragent.AssetReleaseNames(version)
	if len(names) != 3 {
		return fmt.Errorf("AssetReleaseNames(%q) returned %d names want 3: %v", version, len(names), names)
	}

	if outDir == "" {
		tmp, err := os.MkdirTemp("", "browser-agent-release-assets-*")
		if err != nil {
			return fmt.Errorf("mkdir temp --out: %w", err)
		}
		outDir = tmp
	} else {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("mkdir --out: %w", err)
		}
	}
	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("resolve --out absolute path: %w", err)
	}
	outDir = absOut

	// Resolve module root as process cwd (tests run from ModuleRoot).
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	// Source dirs aligned with AssetReleaseNames order:
	//   0 browser-agent_*_session-page.tar.gz
	//   1 browser-agent_*_extension.tar.gz
	//   2 browser-trace_*_extension.tar.gz
	sources := []string{
		filepath.Join(root, "browseragent", "embedded", "session-page"),
		filepath.Join(root, "browseragent", "embedded", "extension"),
		filepath.Join(root, "browsertrace", "embedded", "extension"),
	}

	var archives []string
	for i, src := range sources {
		st, err := os.Stat(src)
		if err != nil || !st.IsDir() {
			return fmt.Errorf("pack source missing or not a dir: %s: %v", src, err)
		}
		dest := filepath.Join(outDir, names[i])
		if err := writeTarGz(src, dest); err != nil {
			return fmt.Errorf("pack %s: %w", names[i], err)
		}
		info, err := os.Stat(dest)
		if err != nil {
			return fmt.Errorf("stat packed archive %s: %w", dest, err)
		}
		if info.Size() <= 0 {
			return fmt.Errorf("packed archive empty: %s", dest)
		}
		fmt.Printf("wrote %s (%d bytes)\n", dest, info.Size())
		archives = append(archives, dest)
	}

	// Preferred operator token: last-line out: <abs-path>.
	fmt.Printf("out: %s\n", outDir)

	if !upload {
		return nil
	}
	return uploadWithGH(version, archives)
}

func printHelp(w io.Writer) {
	// Trailing newline required by help leaf.
	fmt.Fprint(w, `Usage: go run ./script/github/release-assets [flags]

Pack three release hydrate archives from on-disk embeds. Basenames match
browseragent.AssetReleaseNames(version).

Flags:
  --out DIR         Output directory for .tar.gz archives (default: temp dir
                    via MkdirTemp browser-agent-release-assets-*; prints out: path)
  --version VER     Version string (default: embedded ClientVersion / VERSION.txt)
  --upload          Opt-in: wrap gh to create the release tag if missing and
                    upload archives with --clobber
  -h, --help        Show this help

Sources (relative to module root / process cwd):
  browseragent/embedded/session-page
  browseragent/embedded/extension
  browsertrace/embedded/extension

Default is pack-only (no network). When --out is omitted, packs into a temp dir
and prints out: <abs-path>. COPYFILE_DISABLE is set; AppleDouble ._* entries are
excluded from archives.
`)
}

// writeTarGz packs the contents of srcDir into destPath as gzip-compressed tar.
// Paths inside the archive are relative to srcDir (no outer directory prefix).
// Skips AppleDouble ._* files and .DS_Store.
func writeTarGz(srcDir, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		base := filepath.Base(path)
		// Exclude macOS resource-fork / Finder noise.
		if strings.HasPrefix(base, "._") || base == ".DS_Store" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(rel)

		if info.IsDir() {
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = name + "/"
			return tw.WriteHeader(hdr)
		}
		if !info.Mode().IsRegular() {
			// Skip symlinks and specials for hydrate trees.
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = name
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		rf, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, rf)
		closeErr := rf.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	if err := gz.Close(); err != nil {
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}
	return f.Close()
}

// uploadWithGH creates the GitHub release for the version tag if missing, then
// uploads archives with clobber. Requires `gh` on PATH and authenticated repo.
func uploadWithGH(version string, archives []string) error {
	tag := strings.TrimSpace(version)
	if tag == "" {
		return fmt.Errorf("upload: empty version tag")
	}
	// Normalize to v-prefixed tag like AssetReleaseNames / cache keys.
	if !strings.HasPrefix(strings.ToLower(tag), "v") {
		tag = "v" + tag
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("upload requires gh on PATH: %w", err)
	}

	// Create release if missing.
	view := exec.Command("gh", "release", "view", tag)
	if err := view.Run(); err != nil {
		create := exec.Command("gh", "release", "create", tag, "--title", tag, "--notes", "Hydrate asset archives for "+tag)
		create.Stdout = os.Stdout
		create.Stderr = os.Stderr
		if err := create.Run(); err != nil {
			return fmt.Errorf("gh release create %s: %w", tag, err)
		}
		fmt.Printf("created release %s\n", tag)
	}

	args := []string{"release", "upload", tag, "--clobber"}
	args = append(args, archives...)
	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh release upload %s: %w", tag, err)
	}
	fmt.Printf("uploaded %d archives to release %s\n", len(archives), tag)
	return nil
}
