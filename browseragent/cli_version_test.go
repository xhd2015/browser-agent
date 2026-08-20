package browseragent

import (
	"bytes"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent/inject"
)

// withClientVersionOverride pins inj.ClientVersionOverride for the duration of fn.
func withClientVersionOverride(ver string, fn func()) {
	prev := inject.ClientVersionOverride
	inject.ClientVersionOverride = func() string { return ver }
	defer func() { inject.ClientVersionOverride = prev }()
	fn()
}

// TestCLI_Version prints the pinned version on stdout, nothing on stderr, exit 0.
func TestCLI_Version(t *testing.T) {
	const want = "9.9.9"
	withClientVersionOverride(want, func() {
		var out, errOut bytes.Buffer
		if e := HandleCLI([]string{"--version"}, nil, &out, &errOut); e != nil {
			t.Fatalf("HandleCLI(--version) err = %v", e)
		}
		if got := out.String(); got != want+"\n" {
			t.Fatalf("stdout = %q, want %q", got, want+"\n")
		}
		if n := errOut.Len(); n != 0 {
			t.Fatalf("stderr = %q, want empty", errOut.String())
		}
	})
}

// TestCLI_VersionUnsetUsesEmbedded confirms the real embedded version is used
// when no override is active.
func TestCLI_VersionUnsetUsesEmbedded(t *testing.T) {
	prev := inject.ClientVersionOverride
	inject.ClientVersionOverride = nil
	defer func() { inject.ClientVersionOverride = prev }()

	var out bytes.Buffer
	if e := HandleCLI([]string{"--version"}, nil, &out, nil); e != nil {
		t.Fatalf("HandleCLI(--version) err = %v", e)
	}
	// VERSION.txt at HEAD of branch is 1.0.13; assert non-empty + trailing newline.
	got := out.String()
	if got == "" || got[len(got)-1] != '\n' {
		t.Fatalf("stdout = %q, want non-empty with trailing newline", got)
	}
}
