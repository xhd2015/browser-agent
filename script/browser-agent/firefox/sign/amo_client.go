package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const amoAPIBase = "https://addons.mozilla.org/api/v5"

// AMOVersionStatus is a simplified view of one add-on version on AMO.
type AMOVersionStatus struct {
	Version   string // package version string, e.g. "1.0.4"
	Channel   string
	// State is one of: signed, pending, rejected, disabled, unknown, not_found
	State       string
	FileURL     string
	IsSigned    bool
	FileStatus  string
	RawFileJSON json.RawMessage // for debugging
}

// mintAMOJWT builds a short-lived HS256 JWT for AMO (exp-iat ≤ 5 minutes).
// Uses ~4 minutes to leave margin against clock skew.
func mintAMOJWT(issuer, secret string, now time.Time) (string, error) {
	issuer = strings.TrimSpace(issuer)
	secret = strings.TrimSpace(secret)
	if issuer == "" || secret == "" {
		return "", fmt.Errorf("jwt issuer and secret are required")
	}
	iat := now.UTC().Unix()
	exp := iat + 4*60 // 4 minutes (AMO max is 5)
	jti, err := randomJTI()
	if err != nil {
		return "", err
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadObj := map[string]any{
		"iss": issuer,
		"jti": jti,
		"iat": iat,
		"exp": exp,
	}
	payloadJSON, err := json.Marshal(payloadObj)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig, nil
}

func randomJTI() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b[:]), nil
}

func amoGET(path string, issuer, secret string) ([]byte, int, error) {
	token, err := mintAMOJWT(issuer, secret, time.Now())
	if err != nil {
		return nil, 0, err
	}
	u := path
	if !strings.HasPrefix(u, "http") {
		u = strings.TrimRight(amoAPIBase, "/") + "/" + strings.TrimLeft(path, "/")
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "JWT "+token)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// lookupAMOVersion finds version wantVer for guid (e.g. browser-agent@xhd2015).
func lookupAMOVersion(guid, wantVer, issuer, secret string) (*AMOVersionStatus, error) {
	wantVer = strings.TrimSpace(wantVer)
	guid = strings.TrimSpace(guid)
	if wantVer == "" || guid == "" {
		return nil, fmt.Errorf("guid and version are required")
	}

	// Include unlisted (our channel). Paginate a few pages if needed.
	encoded := url.PathEscape(guid)
	path := fmt.Sprintf("addons/addon/%s/versions/?page_size=50&filter=all_with_unlisted", encoded)

	for page := 1; page <= 10; page++ {
		p := path
		if page > 1 {
			p = fmt.Sprintf("addons/addon/%s/versions/?page_size=50&filter=all_with_unlisted&page=%d", encoded, page)
		}
		body, code, err := amoGET(p, issuer, secret)
		if err != nil {
			return nil, fmt.Errorf("AMO versions request: %w", err)
		}
		if code == http.StatusUnauthorized || code == http.StatusForbidden {
			return nil, fmt.Errorf("AMO auth failed (HTTP %d): %s\n  hint: check jwt_issuer/jwt_secret and system clock (JWT exp must be ≤5m); see %s",
				code, truncate(string(body), 300), amoAPIKeyURL)
		}
		if code == http.StatusNotFound {
			return &AMOVersionStatus{Version: wantVer, State: "not_found"}, nil
		}
		if code < 200 || code >= 300 {
			return nil, fmt.Errorf("AMO versions HTTP %d: %s", code, truncate(string(body), 400))
		}

		var pageResp struct {
			Next    *string          `json:"next"`
			Results []amoVersionJSON `json:"results"`
		}
		if err := json.Unmarshal(body, &pageResp); err != nil {
			return nil, fmt.Errorf("parse AMO versions JSON: %w", err)
		}
		for _, r := range pageResp.Results {
			if strings.TrimSpace(r.Version) != wantVer {
				continue
			}
			st := interpretVersion(r)
			st.Version = wantVer
			return st, nil
		}
		if pageResp.Next == nil || *pageResp.Next == "" {
			break
		}
	}
	return &AMOVersionStatus{Version: wantVer, State: "not_found"}, nil
}

type amoVersionJSON struct {
	Version string `json:"version"`
	Channel string `json:"channel"`
	// Newer API: single file object
	File *amoFileJSON `json:"file"`
	// Older / alternate: files array
	Files []amoFileJSON `json:"files"`
}

type amoFileJSON struct {
	ID                       int64  `json:"id"`
	Status                   string `json:"status"`
	URL                      string `json:"url"`
	IsMozillaSignedExtension bool   `json:"is_mozilla_signed_extension"`
	IsSigned                 bool   `json:"is_signed"`
	Hash                     string `json:"hash"`
}

func interpretVersion(r amoVersionJSON) *AMOVersionStatus {
	st := &AMOVersionStatus{
		Version: r.Version,
		Channel: r.Channel,
		State:   "unknown",
	}
	f := pickFile(r)
	if f == nil {
		st.State = "pending"
		return st
	}
	st.FileURL = strings.TrimSpace(f.URL)
	st.FileStatus = strings.TrimSpace(f.Status)
	st.IsSigned = f.IsMozillaSignedExtension || f.IsSigned
	// status values seen: public, unreviewed, disabled, nominated, …
	switch {
	case st.FileStatus == "disabled" || st.FileStatus == "rejected":
		st.State = "rejected"
	case st.FileURL != "" && (st.IsSigned || st.FileStatus == "public"):
		// public + download URL is enough (unlisted signed packages often use status=public).
		st.State = "signed"
		st.IsSigned = true
	case st.IsSigned && st.FileURL != "":
		st.State = "signed"
	default:
		st.State = "pending"
	}
	return st
}

func pickFile(r amoVersionJSON) *amoFileJSON {
	if r.File != nil {
		return r.File
	}
	if len(r.Files) > 0 {
		return &r.Files[0]
	}
	return nil
}

// downloadAMOFile downloads a signed xpi URL (with JWT auth) to destPath.
func downloadAMOFile(fileURL, destPath, issuer, secret string) error {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return fmt.Errorf("empty file URL")
	}
	token, err := mintAMOJWT(issuer, secret, time.Now())
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "JWT "+token)
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("download HTTP %d: %s", resp.StatusCode, truncate(string(b), 300))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return fmt.Errorf("download is not a zip/xpi (len=%d)", len(data))
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0o644)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}