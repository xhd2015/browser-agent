package main

import "testing"

func TestParseSemver(t *testing.T) {
	v, err := parseSemver("0.3.1")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "0.3.1" {
		t.Fatalf("got %s", v.String())
	}
	v2, err := parseSemver("v1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	if v2.major != 1 || v2.minor != 2 || v2.patch != 3 {
		t.Fatalf("got %+v", v2)
	}
	if _, err := parseSemver("1.2"); err == nil {
		t.Fatal("want error for incomplete semver")
	}
	if _, err := parseSemver("01.2.3"); err == nil {
		t.Fatal("want error for leading zero")
	}
}

func TestBumpPatch(t *testing.T) {
	v, _ := parseSemver("0.3.1")
	n := v.bumpPatch()
	if n.String() != "0.3.2" {
		t.Fatalf("got %s", n.String())
	}
}

func TestCmp(t *testing.T) {
	a, _ := parseSemver("0.3.1")
	b, _ := parseSemver("0.3.2")
	c, _ := parseSemver("1.0.0")
	if a.cmp(b) >= 0 || b.cmp(a) <= 0 {
		t.Fatal("0.3.1 vs 0.3.2")
	}
	if a.cmp(c) >= 0 {
		t.Fatal("0.3.1 vs 1.0.0")
	}
	if a.cmp(a) != 0 {
		t.Fatal("equal")
	}
}

func TestGitTagFor(t *testing.T) {
	if gitTagFor("0.3.1") != "v0.3.1" {
		t.Fatal(gitTagFor("0.3.1"))
	}
	if gitTagFor("v0.3.1") != "v0.3.1" {
		t.Fatal(gitTagFor("v0.3.1"))
	}
}

func TestTagExistsAndLatest(t *testing.T) {
	tags := []semver{
		{0, 2, 0},
		{0, 2, 4},
		{0, 2, 3},
	}
	if !tagExists(tags, "0.2.4") {
		t.Fatal("expected 0.2.4 exists")
	}
	if tagExists(tags, "0.3.1") {
		t.Fatal("0.3.1 should not exist")
	}
	latest := latestSemverTag(tags)
	if latest == nil || latest.String() != "0.2.4" {
		t.Fatalf("latest=%v", latest)
	}
}
