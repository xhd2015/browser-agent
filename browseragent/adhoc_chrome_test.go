package browseragent

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

func TestAdhocBrowserProfileDataDir(t *testing.T) {
	got := AdhocBrowserProfileDataDir("sess-abc12")
	want := filepath.Join(os.TempDir(), "browser-agent-adhoc-sess-abc12", "data")
	if got != want {
		t.Fatalf("AdhocBrowserProfileDataDir = %q, want %q", got, want)
	}
	root := AdhocBrowserProfileRoot("sess-abc12")
	if root != filepath.Join(os.TempDir(), "browser-agent-adhoc-sess-abc12") {
		t.Fatalf("AdhocBrowserProfileRoot = %q", root)
	}
}

func TestBuildManagedChromeArgs_adhocShape(t *testing.T) {
	dataDir := filepath.Join(os.TempDir(), "browser-agent-adhoc-sess-x", "data")
	ext := "/tmp/ext/browser-agent/1.0.0"
	url := "http://127.0.0.1:43761/go?session=sess-x"
	args := BuildManagedChromeArgs(dataDir, ext, url)
	want := []string{
		"--user-data-dir=" + dataDir,
		"--load-extension=" + ext,
		"--new-window",
		url,
	}
	if len(args) != len(want) {
		t.Fatalf("args=%v want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d]=%q want %q; full=%v", i, args[i], want[i], args)
		}
	}
}

func TestSessionNew_adhocBrowserProfileRequiresChrome(t *testing.T) {
	err := SessionNew(SessionNewConfig{
		BaseDir:             t.TempDir(),
		AdhocBrowserProfile: true,
		Browser:             "firefox",
		Stdout:              io.Discard,
		Stderr:              io.Discard,
	})
	if err == nil || !strings.Contains(err.Error(), "--adhoc-browser-profile requires chrome") {
		t.Fatalf("err=%v", err)
	}
}

func TestFormatSessionNewOutput_adhoc(t *testing.T) {
	var buf bytes.Buffer
	result := &postCreateSessionResult{
		SessionID:  "sess-abc12",
		SessionURL: "http://127.0.0.1:43761/go?session=sess-abc12",
	}
	dataDir := AdhocBrowserProfileDataDir("sess-abc12")
	if err := formatSessionNewOutput(&buf, result, "http://127.0.0.1:43761", "/tmp/ext", "chrome", dataDir); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{
		"session-id: sess-abc12",
		"profile:     adhoc  " + dataDir,
		"loaded  via --load-extension",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Load unpacked once") {
		t.Fatalf("adhoc output should not tell operators to Load unpacked:\n%s", got)
	}
}

func TestFormatSessionNewOutput_defaultChrome(t *testing.T) {
	var buf bytes.Buffer
	result := &postCreateSessionResult{
		SessionID:  "sess-def",
		SessionURL: "http://127.0.0.1:43761/go?session=sess-def",
	}
	if err := formatSessionNewOutput(&buf, result, "http://127.0.0.1:43761", "/tmp/ext", "chrome", ""); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "profile:     adhoc") {
		t.Fatalf("default output should not show adhoc profile:\n%s", got)
	}
	if !strings.Contains(got, "Load unpacked once") {
		t.Fatalf("default output missing Load unpacked note:\n%s", got)
	}
}

func TestCliSessionNewHelpMentionsAdhoc(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"session", "new", "--help"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("help: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "--adhoc-browser-profile") {
		t.Fatalf("session new help missing --adhoc-browser-profile:\n%s", got)
	}
}

func TestOpenAdhocChrome_capturesArgs(t *testing.T) {
	var got []string
	err := inj.WithManagedChromeHooks(&inj.ManagedChromeHooks{
		LaunchFn: func(args []string) error {
			got = append([]string(nil), args...)
			return nil
		},
	}, func() error {
		dataDir := t.TempDir()
		return openAdhocChrome("http://127.0.0.1:43761/go?session=sess-x", "/tmp/ext", dataDir)
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "--user-data-dir=") || !strings.Contains(joined, "--load-extension=/tmp/ext") {
		t.Fatalf("adhoc launch args missing profile/extension: %v", got)
	}
	if !strings.Contains(joined, "--new-window") || !strings.Contains(joined, "http://127.0.0.1:43761/go?session=sess-x") {
		t.Fatalf("adhoc launch args missing window/url: %v", got)
	}
}
