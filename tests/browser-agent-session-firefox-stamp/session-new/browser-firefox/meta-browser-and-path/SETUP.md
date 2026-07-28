# Scenario

**Feature**: meta.json stamps browser firefox + browser-agent-firefox path

```
SessionNew(Browser=firefox, Home=TestHome)
  -> sessions/<id>/meta.json
       browser: "firefox"
       extension_install_path: …/browser-agent-firefox/…
```

## Preconditions

- SessionNewFirefoxOp = meta-browser-and-path.
- Assert reads on-disk meta only (no /v1/session required).

## Steps

1. Set SessionNewFirefoxOp = SessionNewFirefoxOpMetaBrowserAndPath.

## Context

- Primary durable-stamp leaf for Phase 2.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.SessionNewFirefoxOp = SessionNewFirefoxOpMetaBrowserAndPath
	return nil
}
```
