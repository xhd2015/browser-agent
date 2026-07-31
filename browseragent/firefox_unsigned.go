package browseragent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FirefoxUnsignedPackage is the dist/ layout produced for web-ext / local load.
type FirefoxUnsignedPackage struct {
	Version     string
	BuildDir    string // Firefox-Ext-Browser-Agent/build
	StageDir    string // dist/firefox-package
	UnsignedXPI string // dist/browser-agent-firefox-<ver>.xpi (empty if zip failed)
}

// StageFirefoxUnsignedPackage copies a prepared Firefox build into
// dist/firefox-package and zips dist/browser-agent-firefox-<version>.xpi.
// buildDir should come from PrepareFirefoxExtensionBuild (or an equivalent build/).
func StageFirefoxUnsignedPackage(root, buildDir, version string) (*FirefoxUnsignedPackage, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	buildDir = strings.TrimSpace(buildDir)
	if buildDir == "" {
		return nil, fmt.Errorf("buildDir is required")
	}
	if st, err := os.Stat(filepath.Join(buildDir, "manifest.json")); err != nil || st.IsDir() {
		return nil, fmt.Errorf("firefox build missing manifest.json under %s", buildDir)
	}
	version = strings.TrimSpace(version)
	if version == "" {
		version = strings.TrimSpace(ReadProductVersion(absRoot))
	}
	if version == "" {
		version = strings.TrimSpace(ClientVersion())
	}
	if version == "" {
		return nil, fmt.Errorf("product version empty (set VERSION.txt)")
	}

	stagePath := filepath.Join(absRoot, "dist", "firefox-package")
	if err := stageDir(buildDir, stagePath); err != nil {
		return nil, fmt.Errorf("stage dist/firefox-package: %w", err)
	}

	unsignedXPI := filepath.Join(absRoot, "dist", fmt.Sprintf("browser-agent-firefox-%s.xpi", version))
	if err := zipFirefoxPackageDir(stagePath, unsignedXPI); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write unsigned xpi: %v\n", err)
		unsignedXPI = ""
	}

	return &FirefoxUnsignedPackage{
		Version:     version,
		BuildDir:    buildDir,
		StageDir:    stagePath,
		UnsignedXPI: unsignedXPI,
	}, nil
}

// BundleFirefoxUnsignedPackage runs PrepareFirefoxExtensionBuild then
// StageFirefoxUnsignedPackage. Caller should run script/generate first if
// VERSION sinks need refresh.
func BundleFirefoxUnsignedPackage(root string) (*FirefoxUnsignedPackage, error) {
	buildDir, version, err := PrepareFirefoxExtensionBuild(root)
	if err != nil {
		return nil, err
	}
	return StageFirefoxUnsignedPackage(root, buildDir, version)
}

func zipFirefoxPackageDir(srcDir, zipPath string) error {
	absZip, err := filepath.Abs(zipPath)
	if err != nil {
		return err
	}
	_ = os.Remove(absZip)
	if err := os.MkdirAll(filepath.Dir(absZip), 0o755); err != nil {
		return err
	}
	// System zip (same as former firefox/sign helper).
	cmd := exec.Command("zip", "-r", "-FS", absZip, ".")
	cmd.Dir = srcDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
