package main

import (
	"strings"
	"testing"
)

func TestBumpVersion_patch(t *testing.T) {
	got, err := bumpVersion("1.0.6", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.0.7" {
		t.Fatalf("got %q", got)
	}
}

func TestBumpVersion_minor(t *testing.T) {
	got, err := bumpVersion("1.0.6", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.1.0" {
		t.Fatalf("got %q", got)
	}
}

func TestBumpVersion_major(t *testing.T) {
	got, err := bumpVersion("1.0.6", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "2.0.0" {
		t.Fatalf("got %q", got)
	}
}

func TestBumpVersion_invalid(t *testing.T) {
	cases := []string{"", "1.0", "1.0.0-beta", "a.b.c", "1.0.0.1"}
	for _, c := range cases {
		if _, err := bumpVersion(c, false, false); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}

func TestBumpVersion_majorMinorConflict(t *testing.T) {
	_, err := bumpVersion("1.0.0", true, true)
	if err == nil || !strings.Contains(err.Error(), "cannot combine") {
		t.Fatalf("got %v", err)
	}
}

func TestHelpText(t *testing.T) {
	h := helpText()
	for _, s := range []string{"--minor", "--major", "--dry-run", "--skip-bundle", "--color"} {
		if !strings.Contains(h, s) {
			t.Fatalf("help missing %s", s)
		}
	}
}
