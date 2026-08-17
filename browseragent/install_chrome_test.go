package browseragent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent/inject"
	"github.com/xhd2015/dot-pkgs/go-pkgs/computer-use/macos/chrome"
	lessflags "github.com/xhd2015/less-flags"
)

func TestParseInstallChromeExtOptions(t *testing.T) {
	opts, err := parseInstallChromeExtOptions([]string{
		"--no-open",
		"--extension-dir", "/tmp/ext",
		"--dry-run",
		"--keep-old",
		"--debug-screenshot",
		"--screenshot-dir=/tmp/shots",
		"--write-json-result", "/tmp/ba-ext.json",
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.noOpen || !opts.dryRun || opts.extensionDir != "/tmp/ext" || !opts.keepOld || !opts.debugScreenshot || opts.screenshotDir != "/tmp/shots" || opts.writeJSONResult != "/tmp/ba-ext.json" {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestParseInstallChromeExtOptionsHelp(t *testing.T) {
	var buf bytes.Buffer
	_, err := parseInstallChromeExtOptions([]string{"--help"}, &buf)
	if err != lessflags.ErrHelp {
		t.Fatalf("err=%v want ErrHelp", err)
	}
	if !strings.Contains(buf.String(), "install-chrome-extension") || !strings.Contains(buf.String(), "--debug-screenshot") || !strings.Contains(buf.String(), "--write-json-result") {
		t.Fatalf("help body: %s", buf.String())
	}
}

func TestResolveChromeScreenshotDir(t *testing.T) {
	// neither → empty
	dir, err := resolveChromeScreenshotDir(false, "")
	if err != nil || dir != "" {
		t.Fatalf("neither: dir=%q err=%v", dir, err)
	}
	// --screenshot-dir wins (deterministic)
	fixed := t.TempDir() + "/shots"
	dir, err = resolveChromeScreenshotDir(true, fixed)
	if err != nil {
		t.Fatal(err)
	}
	if dir != fixed {
		t.Fatalf("dir=%q want %q", dir, fixed)
	}
	// --debug-screenshot only → temp under TempDir
	dir, err = resolveChromeScreenshotDir(true, "")
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" || !strings.Contains(dir, "browser-agent-chrome-debug-") {
		t.Fatalf("temp dir unexpected: %q", dir)
	}
}

func TestCliInstallChromeOpenAndNoOpenError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := HandleCLI([]string{"install-chrome-extension", "--open", "--no-open"}, nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for --open and --no-open")
	}
	if !strings.Contains(err.Error(), "--open") || !strings.Contains(err.Error(), "--no-open") {
		t.Fatalf("error = %v", err)
	}
}

func TestCliInstallChromeNoOpenSkipsUI(t *testing.T) {
	home := t.TempDir()
	var called bool
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			called = true
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		return HandleCLI(
			[]string{"install-chrome-extension", "--no-open"},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("UI load should not run with --no-open")
	}
}

func TestCliInstallChromeOpenCallsInject(t *testing.T) {
	home := t.TempDir()
	var called bool
	var gotDir string
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			called = true
			gotDir = extensionDir
			if dryRun || dumpTree {
				t.Errorf("unexpected dryRun=%v dumpTree=%v", dryRun, dumpTree)
			}
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		err := HandleCLI(
			[]string{"install-chrome-extension", "--open"},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
		if err != nil {
			return err
		}
		if !strings.Contains(stdout.String(), "extensions/browser-agent/") {
			t.Fatalf("stdout missing path: %s", stdout.String())
		}
		if !strings.Contains(stdout.String(), "Loaded unpacked") {
			t.Fatalf("stdout missing loaded line: %s", stdout.String())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected inject LoadFn with --open")
	}
	if gotDir == "" {
		t.Fatal("empty extension dir passed to inject")
	}
}

func TestCliInstallChromeDryRunCallsInject(t *testing.T) {
	home := t.TempDir()
	var dry bool
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			dry = dryRun
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		return HandleCLI(
			[]string{"install-chrome-extension", "--dry-run"},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !dry {
		t.Fatal("expected dryRun=true")
	}
}

func TestCliInstallChromeExtensionDirOverride(t *testing.T) {
	home := t.TempDir()
	// Create a fake unpacked dir for ExtensionDirOK when real chrome would run;
	// with inject we only check the path passed through.
	ext := t.TempDir()
	var got string
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			got = extensionDir
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		return HandleCLI(
			[]string{"install-chrome-extension", "--open", "--extension-dir", ext},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != ext {
		t.Fatalf("got dir %q want %q", got, ext)
	}
}

func TestWriteInstallChromeJSONResult(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "ba-ext.json")
	loadPath := "/tmp/ext/1.0.10"
	if err := writeInstallChromeJSONResult(outFile, loadPath); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	var got installChromeJSONResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json: %v\n%s", err, raw)
	}
	if !got.OK || got.Path != loadPath {
		t.Fatalf("got=%+v", got)
	}
	if len(got.UI) != 2 {
		t.Fatalf("ui=%+v", got.UI)
	}
	if got.UI[0].Kind != "open-url" || got.UI[0].App != chrome.DefaultAppName || got.UI[0].URL != chrome.ExtensionsURL {
		t.Fatalf("ui[0]=%+v", got.UI[0])
	}
	if got.UI[1].Kind != "reveal" || got.UI[1].Path != loadPath {
		t.Fatalf("ui[1]=%+v", got.UI[1])
	}
	if err := writeInstallChromeJSONResult("  ", loadPath); err == nil {
		t.Fatal("expected error for empty dest")
	}
}

func TestCliInstallChromeWriteJSONResult(t *testing.T) {
	if !ExtensionEmbedComplete() {
		t.Skip("chrome extension embed is placeholder-only")
	}
	home := t.TempDir()
	outFile := filepath.Join(t.TempDir(), "ba-ext.json")
	var called bool
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			called = true
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		return HandleCLI(
			[]string{"install-chrome-extension", "--write-json-result", outFile},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("UI load should not run with --write-json-result")
	}
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	var got installChromeJSONResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json: %v\n%s", err, raw)
	}
	if !got.OK {
		t.Fatalf("ok=false; json=%s", raw)
	}
	if got.Path == "" || !strings.Contains(got.Path, "extensions/browser-agent/") {
		t.Fatalf("path=%q", got.Path)
	}
	if len(got.UI) != 2 {
		t.Fatalf("ui=%+v", got.UI)
	}
	if got.UI[0].Kind != "open-url" || got.UI[0].App != chrome.DefaultAppName || got.UI[0].URL != chrome.ExtensionsURL {
		t.Fatalf("ui[0]=%+v", got.UI[0])
	}
	if got.UI[1].Kind != "reveal" || got.UI[1].Path != got.Path {
		t.Fatalf("ui[1]=%+v path=%q", got.UI[1], got.Path)
	}
}

func TestCliInstallChromeWriteJSONResultOpenError(t *testing.T) {
	outFile := filepath.Join(t.TempDir(), "ba-ext.json")
	var stdout, stderr bytes.Buffer
	err := HandleCLI(
		[]string{"install-chrome-extension", "--open", "--write-json-result", outFile},
		nil,
		&stdout,
		&stderr,
	)
	if err == nil {
		t.Fatal("expected error for --open and --write-json-result")
	}
	if !strings.Contains(err.Error(), "--write-json-result") || !strings.Contains(err.Error(), "--open") {
		t.Fatalf("error = %v", err)
	}
	if _, statErr := os.Stat(outFile); !os.IsNotExist(statErr) {
		t.Fatalf("result file should not be written on conflict; stat=%v", statErr)
	}
}

func TestCliInstallChromeWriteJSONResultRequiresPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := HandleCLI(
		[]string{"install-chrome-extension", "--write-json-result="},
		nil,
		&stdout,
		&stderr,
	)
	if err == nil {
		t.Fatal("expected error for empty --write-json-result")
	}
	if !strings.Contains(err.Error(), "--write-json-result") {
		t.Fatalf("error = %v", err)
	}
}

func TestCliInstallChromeWriteJSONResultExtensionDir(t *testing.T) {
	if !ExtensionEmbedComplete() {
		t.Skip("chrome extension embed is placeholder-only")
	}
	home := t.TempDir()
	ext := t.TempDir()
	outFile := filepath.Join(t.TempDir(), "ba-ext.json")
	err := inject.WithChromeLoadHooks(&inject.ChromeLoadHooks{
		LoadFn: func(ctx context.Context, extensionDir string, dryRun, dumpTree bool) error {
			t.Fatal("UI load should not run with --write-json-result")
			return nil
		},
	}, func() error {
		var stdout, stderr bytes.Buffer
		return HandleCLI(
			[]string{"install-chrome-extension", "--write-json-result", outFile, "--extension-dir", ext},
			map[string]string{"HOME": home},
			&stdout,
			&stderr,
		)
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	var got installChromeJSONResult
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json: %v\n%s", err, raw)
	}
	if got.Path != ext {
		t.Fatalf("path=%q want %q", got.Path, ext)
	}
	if len(got.UI) != 2 || got.UI[1].Path != ext {
		t.Fatalf("ui=%+v", got.UI)
	}
}
