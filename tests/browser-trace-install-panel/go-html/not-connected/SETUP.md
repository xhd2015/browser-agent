# Scenario

**Feature**: /go install panel present and expanded when not connected (req #1, #5)

```
Control Server Session: no hello yet; extract already done
Test Client -> GET /go?session=<SessionSuffix>
HTML contains always-visible install panel:
  - markers (data-browser-trace-install / id)
  - expanded (open and/or data-default-open=true)
  - absolute path guidance
  - chrome://extensions text
```

## Preconditions

- DoHello false (not connected).
- Expect expanded default because `!(connected && supports)`.

## Steps

1. Ensure `DoHello = false`.

## Context

- Regression for not-connected path/chrome text (requirement #5) lives here
  alongside always-visible + expanded (#1).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DoHello = false
	return nil
}
```
