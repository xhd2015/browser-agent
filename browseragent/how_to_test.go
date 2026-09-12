package browseragent

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestHowToList(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleCLI([]string{"how-to"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("how-to: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{
		"credentials/export",
		"Firefox cookies",
		"browser-agent how-to <topic>",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("list missing %q\n%s", want, got)
		}
	}
	if strings.Contains(got, "sso=") || strings.Contains(got, "cf_clearance=") {
		t.Fatalf("list leaked cookie-looking values:\n%s", got)
	}
}

func TestHowToRootHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"how-to", "--help"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("how-to --help: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "Usage: browser-agent how-to") || !strings.Contains(got, "credentials/export") {
		t.Fatalf("root help = %q", got)
	}
}

func TestHowToCredentialsExport(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"how-to", "credentials/export", "--host", "grok.com"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("credentials/export: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{
		"Audience: automation agents",
		"Never print cookie values",
		"no separate credentials/import topic",
		"cookies.sqlite",
		"Network.setCookie",
		"expiry > 1e12",
		"grok.com",
		".grok.com",
		"/tmp/ba-ff-cookies-grok-com.json",
		"session har start",
		"--adhoc-browser-profile",
		"Import into isolated Chrome",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("playbook missing %q\n%s", want, got)
		}
	}
	for _, banned := range []string{
		"sso=",
		"cf_clearance=",
		"restok_",
		"DevTools",
		"Application →",
		"Future topic credentials/import",
	} {
		if strings.Contains(got, banned) {
			t.Fatalf("playbook contains banned %q", banned)
		}
	}
}

func TestHowToCredentialsExportHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"how-to", "credentials/export", "-h"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("credentials/export -h: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{
		"Usage: browser-agent how-to credentials/export",
		"--host",
		"--adhoc-browser-profile",
		"no separate import topic",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("topic help missing %q:\n%s", want, got)
		}
	}
}

func TestHowToUnknownTopic(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := HandleCLI([]string{"how-to", "credentials/import"}, nil, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), `unknown how-to topic "credentials/import"`) {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(stdout.String(), "credentials/export") {
		t.Fatalf("unknown topic should still list topics:\n%s", stdout.String())
	}
}

func TestHowToListedInRootHelp(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"--help"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("--help: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "how-to") || !strings.Contains(got, "credentials/export") {
		t.Fatalf("full help missing how-to:\n%s", got)
	}
}

func TestSanitizeHostForFilename(t *testing.T) {
	if got, want := sanitizeHostForFilename("Grok.COM"), "grok-com"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := sanitizeHostForFilename("@@@"); got != "host" {
		t.Fatalf("empty sanitize = %q", got)
	}
}

func TestHowToCredentialsExportStripsURLHost(t *testing.T) {
	var stdout bytes.Buffer
	if err := HandleCLI([]string{"how-to", "credentials/export", "--host", "https://grok.com/path"}, nil, &stdout, io.Discard); err != nil {
		t.Fatalf("export: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "https://grok.com/") || !strings.Contains(got, "/tmp/ba-ff-cookies-grok-com.json") {
		t.Fatalf("expected cleaned host in playbook:\n%s", got)
	}
	if strings.Contains(got, "https://https://") {
		t.Fatalf("double scheme:\n%s", got)
	}
}
