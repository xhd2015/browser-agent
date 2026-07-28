# Scenario

**Feature**: BuildFirefoxOpenArgs with non-empty session URL

```
BuildFirefoxOpenArgs("http://…/go?session=…") -> argv contains URL; no managed flags
```

## Preconditions

- FirefoxOpenArgsOp = with-url.
- URL non-empty (default set if blank).

## Steps

1. Set FirefoxOpenArgsOp = FirefoxOpenArgsOpWithURL.
2. Set URL to a session-like URL.

## Context

- Argv must include the URL and must not include --load-extension / --user-data-dir.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.FirefoxOpenArgsOp = FirefoxOpenArgsOpWithURL
	req.URL = "http://127.0.0.1:43761/go?session=sess-ff-args-1"
	return nil
}
```
