package browseragent

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmbeddedSkillListAndShow(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := HandleCLI([]string{"skill", "--list"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "browser-agent\nbrowser-agent-to-api\n"; got != want {
		t.Fatalf("list = %q, want %q", got, want)
	}

	stdout.Reset()
	if err := HandleCLI([]string{"skill", "--show"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "name: browser-agent\n") {
		t.Fatalf("default show = %q", stdout.String())
	}

	stdout.Reset()
	if err := HandleCLI([]string{"skill", "--show", "browser-agent-to-api"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "name: browser-agent-to-api\n") ||
		!strings.Contains(stdout.String(), "session har start") ||
		!strings.Contains(stdout.String(), "har inspect") {
		t.Fatalf("named show = %q", stdout.String())
	}

	stdout.Reset()
	if err := HandleCLI([]string{"skill", "browser-agent-to-api", "--show", "--header"}, nil, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout.String(), "---\n") || !strings.Contains(stdout.String(), "browser-agent-to-api") {
		t.Fatalf("header show = %q", stdout.String())
	}
}

func TestEmbeddedSkillUnknownName(t *testing.T) {
	err := HandleCLI([]string{"skill", "--show", "unknown"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unknown skill") {
		t.Fatalf("error = %v", err)
	}
}

func TestEmbeddedSkillVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	for _, args := range [][]string{
		{"skill", "--version"},
		{"skill", "browser-agent", "--version"},
		{"skill", "--version", "browser-agent"},
	} {
		stdout.Reset()
		if err := HandleCLI(args, nil, &stdout, &stderr); err != nil {
			t.Fatalf("HandleCLI(%v): %v", args, err)
		}
		if got, want := stdout.String(), "1.0.15\n"; got != want {
			t.Fatalf("HandleCLI(%v) = %q, want %q", args, got, want)
		}
	}

	stdout.Reset()
	if err := HandleCLI([]string{"skill", "browser-agent-to-api", "--version"}, nil, &stdout, &stderr); err != nil {
		t.Fatalf("to-api version: %v", err)
	}
	if got, want := stdout.String(), "1.0.15\n"; got != want {
		t.Fatalf("to-api version = %q, want %q", got, want)
	}

	err := HandleCLI([]string{"skill", "unknown", "--version"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "unknown skill") {
		t.Fatalf("unknown version error = %v", err)
	}
}
