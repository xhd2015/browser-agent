package browseragent

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent/inject"
)

// withFetchLatest pins both inject.ClientVersionOverride and inject.FetchLatestVersionFn.
func withFetchLatest(localVer, latestVer string, fn func()) {
	prevV := inject.ClientVersionOverride
	prevF := inject.FetchLatestVersionFn
	inject.ClientVersionOverride = func() string { return localVer }
	inject.FetchLatestVersionFn = func(ctx context.Context) (string, error) {
		return latestVer, nil
	}
	defer func() {
		inject.ClientVersionOverride = prevV
		inject.FetchLatestVersionFn = prevF
	}()
	fn()
}

func TestCLI_VersionFetchLatest(t *testing.T) {
	withFetchLatest("1.0.13", "1.0.14", func() {
		var out, errOut bytes.Buffer
		if e := HandleCLI([]string{"--version", "--fetch-latest"}, nil, &out, &errOut); e != nil {
			t.Fatalf("HandleCLI err = %v", e)
		}
		if got := out.String(); got != "1.0.14\n" {
			t.Fatalf("stdout = %q, want %q", got, "1.0.14\n")
		}
		if n := errOut.Len(); n != 0 {
			t.Fatalf("stderr = %q, want empty", errOut.String())
		}
	})
}

func TestCLI_VersionFetchLatestVerboseUpToDate(t *testing.T) {
	withFetchLatest("1.0.13", "1.0.13", func() {
		var out bytes.Buffer
		if e := HandleCLI([]string{"--version", "--fetch-latest", "-v"}, nil, &out, nil); e != nil {
			t.Fatalf("HandleCLI err = %v", e)
		}
		got := out.String()
		if !strings.Contains(got, "local:") {
			t.Fatalf("stdout missing local: line\n%s", got)
		}
		if !strings.Contains(got, "latest:") {
			t.Fatalf("stdout missing latest: line\n%s", got)
		}
		if !strings.Contains(got, "up-to-date") {
			t.Fatalf("stdout missing up-to-date\n%s", got)
		}
	})
}

func TestCLI_VersionFetchLatestVerboseUpdateAvailable(t *testing.T) {
	withFetchLatest("1.0.12", "1.0.13", func() {
		var out bytes.Buffer
		if e := HandleCLI([]string{"--version", "--fetch-latest", "-v"}, nil, &out, nil); e != nil {
			t.Fatalf("HandleCLI err = %v", e)
		}
		got := out.String()
		if !strings.Contains(got, "local:   1.0.12") {
			t.Fatalf("stdout missing local line\n%s", got)
		}
		if !strings.Contains(got, "latest:  1.0.13") {
			t.Fatalf("stdout missing latest line\n%s", got)
		}
		if !strings.Contains(got, "update available: 1.0.12 → 1.0.13") {
			t.Fatalf("stdout missing update available line\n%s", got)
		}
	})
}

func TestCLI_VersionFetchLatestVerboseLocalNewer(t *testing.T) {
	withFetchLatest("1.0.15", "1.0.13", func() {
		var out bytes.Buffer
		if e := HandleCLI([]string{"--version", "--fetch-latest", "-v"}, nil, &out, nil); e != nil {
			t.Fatalf("HandleCLI err = %v", e)
		}
		got := out.String()
		if !strings.Contains(got, "local is newer") {
			t.Fatalf("stdout missing local is newer\n%s", got)
		}
	})
}

func TestCLI_VersionFetchLatestError(t *testing.T) {
	prevV := inject.ClientVersionOverride
	prevF := inject.FetchLatestVersionFn
	inject.ClientVersionOverride = func() string { return "1.0.13" }
	inject.FetchLatestVersionFn = func(ctx context.Context) (string, error) {
		return "", context.DeadlineExceeded
	}
	defer func() {
		inject.ClientVersionOverride = prevV
		inject.FetchLatestVersionFn = prevF
	}()

	var errOut bytes.Buffer
	err := HandleCLI([]string{"--version", "--fetch-latest"}, nil, nil, &errOut)
	if err == nil {
		t.Fatal("expected error for fetch failure")
	}
	if !strings.Contains(err.Error(), "failed to fetch latest version") {
		t.Fatalf("error = %q, want contains 'failed to fetch latest version'", err.Error())
	}
}

func TestCLI_VersionFetchLatestUnknownFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	err := HandleCLI([]string{"--version", "--bogus"}, nil, &out, &errOut)
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Fatalf("error = %q, want contains 'unknown flag'", err.Error())
	}
}

func TestCLI_VersionFetchLatestOrderInvariant(t *testing.T) {
	withFetchLatest("1.0.12", "1.0.13", func() {
		// -v before --fetch-latest should work the same.
		var out bytes.Buffer
		if e := HandleCLI([]string{"--version", "-v", "--fetch-latest"}, nil, &out, nil); e != nil {
			t.Fatalf("HandleCLI err = %v", e)
		}
		if !strings.Contains(out.String(), "update available") {
			t.Fatalf("stdout missing update available\n%s", out.String())
		}
	})
}
