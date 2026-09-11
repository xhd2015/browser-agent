package browseragent

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var errPathNotFound = errors.New("not found")

func TestResolveChromeBinary_envWins(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "chrome-bin")
	if err := writeFakeBin(t, bin); err != nil {
		t.Fatal(err)
	}
	got, err := resolveChromeBinary(browserBinaryResolveOpts{
		Getenv: func(k string) string {
			if k == EnvChromePath {
				return bin
			}
			return ""
		},
		LookPath: func(string) (string, error) { return "", errPathNotFound },
		PathOK:   fileExistsRegular,
		GOOS:     "darwin",
		Home:     dir,
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != bin {
		t.Fatalf("got %q want %q", got, bin)
	}
}

func TestResolveChromeBinary_homeApplicationsBeforeSystem(t *testing.T) {
	home := t.TempDir()
	homeBin := filepath.Join(home, "Applications", "Google Chrome.app", "Contents", "MacOS", "Google Chrome")
	if err := writeFakeBin(t, homeBin); err != nil {
		t.Fatal(err)
	}
	systemBin := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	ok := map[string]bool{homeBin: true, systemBin: true}
	got, err := resolveChromeBinary(browserBinaryResolveOpts{
		Getenv:   func(string) string { return "" },
		LookPath: func(string) (string, error) { return "", errPathNotFound },
		Home:     home,
		GOOS:     "darwin",
		PathOK:   func(p string) bool { return ok[p] },
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != homeBin {
		t.Fatalf("got %q want home binary %q", got, homeBin)
	}
}

func TestResolveChromeBinary_systemWhenHomeMissing(t *testing.T) {
	home := t.TempDir()
	systemBin := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	got, err := resolveChromeBinary(browserBinaryResolveOpts{
		Getenv:   func(string) string { return "" },
		LookPath: func(string) (string, error) { return "", errPathNotFound },
		Home:     home,
		GOOS:     "darwin",
		PathOK:   func(p string) bool { return p == systemBin },
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != systemBin {
		t.Fatalf("got %q want %q", got, systemBin)
	}
}

func TestResolveChromeBinary_pathBeforeKnownInstall(t *testing.T) {
	pathBin := "/usr/local/bin/google-chrome"
	home := t.TempDir()
	got, err := resolveChromeBinary(browserBinaryResolveOpts{
		Getenv: func(string) string { return "" },
		LookPath: func(name string) (string, error) {
			if name == "google-chrome" {
				return pathBin, nil
			}
			return "", errPathNotFound
		},
		Home:   home,
		GOOS:   "darwin",
		PathOK: func(p string) bool { return p == pathBin },
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != pathBin {
		t.Fatalf("got %q want PATH binary %q", got, pathBin)
	}
}

func TestResolveFirefoxBinary_homeBeforeSystem(t *testing.T) {
	home := t.TempDir()
	homeBin := filepath.Join(home, "Applications", "Firefox.app", "Contents", "MacOS", "firefox")
	systemBin := "/Applications/Firefox.app/Contents/MacOS/firefox"
	ok := map[string]bool{homeBin: true, systemBin: true}
	got, err := resolveFirefoxBinary(browserBinaryResolveOpts{
		Getenv:   func(string) string { return "" },
		LookPath: func(string) (string, error) { return "", errPathNotFound },
		Home:     home,
		GOOS:     "darwin",
		PathOK:   func(p string) bool { return ok[p] },
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != homeBin {
		t.Fatalf("got %q want %q", got, homeBin)
	}
}

func TestResolveChromeBinary_missing(t *testing.T) {
	_, err := resolveChromeBinary(browserBinaryResolveOpts{
		Getenv:   func(string) string { return "" },
		LookPath: func(string) (string, error) { return "", errPathNotFound },
		Home:     t.TempDir(),
		GOOS:     "darwin",
		PathOK:   func(string) bool { return false },
	})
	if err == nil {
		t.Fatal("expected error when no binary")
	}
}

func TestBuildFirefoxOpenArgs_newWindow(t *testing.T) {
	url := "http://127.0.0.1:43761/go?session=s1"
	args := BuildFirefoxOpenArgs(url)
	if len(args) < 2 || args[0] != "-new-window" {
		t.Fatalf("want -new-window first; args=%v", args)
	}
	found := false
	for _, a := range args {
		if a == url {
			found = true
		}
		if strings.Contains(a, "load-extension") || strings.Contains(a, "user-data-dir") {
			t.Fatalf("managed flag in args: %v", args)
		}
	}
	if !found {
		t.Fatalf("url missing: %v", args)
	}
	empty := BuildFirefoxOpenArgs("")
	if len(empty) != 1 || empty[0] != "-new-window" {
		t.Fatalf("empty url want only -new-window; got %v", empty)
	}
}

func TestBuildChromeArgs_newWindow(t *testing.T) {
	args := BuildChromeArgs("http://example.com/go?session=x", "/tmp/ext")
	if len(args) < 2 || args[0] != "--new-window" {
		t.Fatalf("want --new-window first; args=%v", args)
	}
	if args[1] != "http://example.com/go?session=x" {
		t.Fatalf("want session URL second; args=%v", args)
	}
	for _, a := range args {
		if strings.Contains(a, "load-extension") {
			t.Fatalf("default-profile args must not pass --load-extension; args=%v", args)
		}
	}
}

func writeFakeBin(t *testing.T, path string) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755)
}
