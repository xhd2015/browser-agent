package clicolor

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestResolve_modes(t *testing.T) {
	if !Resolve(ColorAlways, false, "1") {
		t.Fatal("Always must ignore TTY and NO_COLOR")
	}
	if Resolve(ColorNever, true, "") {
		t.Fatal("Never must be off")
	}
	if Resolve(ColorAuto, false, "") {
		t.Fatal("Auto + non-TTY → off")
	}
	if Resolve(ColorAuto, true, "1") {
		t.Fatal("Auto + NO_COLOR → off")
	}
	if !Resolve(ColorAuto, true, "") {
		t.Fatal("Auto + TTY + empty NO_COLOR → on")
	}
	if Resolve(ColorAuto, true, "") != true {
		t.Fatal("empty NO_COLOR must not disable")
	}
}

func TestParseFlags_conflict(t *testing.T) {
	_, _, err := ParseFlags([]string{"--color", "--no-color"})
	if err == nil || !strings.Contains(err.Error(), "cannot be specified together") {
		t.Fatalf("want conflict error, got %v", err)
	}
}

func TestParseFlags_modes(t *testing.T) {
	m, rem, err := ParseFlags([]string{"--color", "rest"})
	if err != nil {
		t.Fatal(err)
	}
	if m != ColorAlways || len(rem) != 1 || rem[0] != "rest" {
		t.Fatalf("mode=%v remain=%v", m, rem)
	}
	m, _, err = ParseFlags([]string{"--no-color"})
	if err != nil || m != ColorNever {
		t.Fatalf("no-color: %v %v", m, err)
	}
	m, _, err = ParseFlags(nil)
	if err != nil || m != ColorAuto {
		t.Fatalf("default auto: %v %v", m, err)
	}
}

func TestParseFlags_leavesCommandFlags(t *testing.T) {
	// Regression: strict less-flags rejected --git-add-generated before generate saw it.
	m, rem, err := ParseFlags([]string{"--git-add-generated", "--color"})
	if err != nil {
		t.Fatal(err)
	}
	if m != ColorAlways {
		t.Fatalf("mode=%v want Always", m)
	}
	if len(rem) != 1 || rem[0] != "--git-add-generated" {
		t.Fatalf("remain=%v want [--git-add-generated]", rem)
	}
}

func TestStyle_noEscapeWhenDisabled(t *testing.T) {
	c := Style{enabled: false}
	if strings.Contains(c.Red("Error:"), "\x1b") {
		t.Fatal("disabled must not emit SGR")
	}
	c = Style{enabled: true}
	if !strings.Contains(c.Green("ok"), "\x1b[32m") {
		t.Fatal("enabled green should use 32")
	}
	if !strings.HasSuffix(c.Green("ok"), "\x1b[0m") {
		t.Fatal("must reset")
	}
}

func TestErrorLine_prefix(t *testing.T) {
	c := Style{enabled: true}
	var buf bytes.Buffer
	c.ErrorLine(&buf, "boom")
	s := buf.String()
	if !strings.Contains(s, "Error:") || !strings.Contains(s, "boom") {
		t.Fatalf("got %q", s)
	}
}

func TestNewStyle_respectsNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Force Always still wins when constructed via Resolve path for Always.
	if !Resolve(ColorAlways, false, os.Getenv("NO_COLOR")) {
		t.Fatal("--color / Always must win over NO_COLOR")
	}
	// Auto + NO_COLOR off even if we pretend TTY.
	if Resolve(ColorAuto, true, os.Getenv("NO_COLOR")) {
		t.Fatal("NO_COLOR disables auto")
	}
}
