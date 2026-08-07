package browseragent

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent/inject"
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
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.noOpen || !opts.dryRun || opts.extensionDir != "/tmp/ext" || !opts.keepOld || !opts.debugScreenshot || opts.screenshotDir != "/tmp/shots" {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestParseInstallChromeExtOptionsHelp(t *testing.T) {
	var buf bytes.Buffer
	_, err := parseInstallChromeExtOptions([]string{"--help"}, &buf)
	if err != lessflags.ErrHelp {
		t.Fatalf("err=%v want ErrHelp", err)
	}
	if !strings.Contains(buf.String(), "install-chrome-extension") || !strings.Contains(buf.String(), "--debug-screenshot") {
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
