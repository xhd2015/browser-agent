package main

import (
	"strings"
	"testing"
	"time"
)

func TestMintAMOJWT_shape(t *testing.T) {
	now := time.Unix(1447273096, 0).UTC()
	tok, err := mintAMOJWT("user:1:2", "secret", now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("want 3 JWT parts, got %d", len(parts))
	}
	// exp-iat should be 4 minutes
	// Re-mint and ensure different jti
	tok2, err := mintAMOJWT("user:1:2", "secret", now)
	if err != nil {
		t.Fatal(err)
	}
	if tok == tok2 {
		// jti is random so tokens should differ
		t.Fatal("expected different jti each mint")
	}
}

func TestInterpretVersion_signed(t *testing.T) {
	r := amoVersionJSON{
		Version: "1.0.4",
		Channel: "unlisted",
		File: &amoFileJSON{
			Status:                   "public",
			URL:                      "https://example.com/a.xpi",
			IsMozillaSignedExtension: true,
		},
	}
	st := interpretVersion(r)
	if st.State != "signed" || !st.IsSigned || st.FileURL == "" {
		t.Fatalf("%+v", st)
	}
}

func TestInterpretVersion_pending(t *testing.T) {
	r := amoVersionJSON{
		Version: "1.0.4",
		File: &amoFileJSON{
			Status: "unreviewed",
			URL:    "",
		},
	}
	st := interpretVersion(r)
	if st.State != "pending" {
		t.Fatalf("state=%s", st.State)
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hi", 10) != "hi" {
		t.Fatal()
	}
	if !strings.HasSuffix(truncate("abcdefghij", 5), "…") {
		t.Fatal(truncate("abcdefghij", 5))
	}
}
