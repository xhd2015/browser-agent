## Expected

Requirement **T1** (inject SPA path via `injectSessionBoot` / GET `/go`):

- HTTP status **200**; body is HTML.
- Page `<title>` text equals **`sess-title - Browser Agent`**
  (i.e. `SessionID + " - Browser Agent"`, spaces around hyphen).
- Case-insensitive `<title>` tag OK; first title wins.
- Body / title must **not** use the old static sole title
  **`Browser Agent Session`** as the page title.

## Side Effects

- None.

## Errors

- Static SPA title, missing session id in title, wrong suffix (`Agent` alone),
  or wrong spacing fails.

## Exit Code

- Not asserted.

```go
import (
	"net/http"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	assertHTTPStatus(t, resp, http.StatusOK)
	assertHTMLContentType(t, resp)
	body := resp.BodyString
	if strings.TrimSpace(body) == "" {
		t.Fatal("HTML body empty")
	}

	sid := req.SessionID
	if sid == "" {
		sid = resp.RealSessionID
	}
	if sid == "" {
		sid = "sess-title"
	}

	want := expectedSessionTitle(sid) // sid + " - Browser Agent"
	title := strings.TrimSpace(resp.PageTitle)
	if title == "" {
		title = extractHTMLTitle(body)
	}

	ok := title == want || titleTagContains(body, want)
	if !ok {
		t.Fatalf("page title = %q, want %q; probe=%s body=%s",
			title, want, resp.ProbeURL, truncate(body, 700))
	}

	// Forbid static sole SPA title.
	assertNotStaticSPATitle(t, title, body)
}
```
