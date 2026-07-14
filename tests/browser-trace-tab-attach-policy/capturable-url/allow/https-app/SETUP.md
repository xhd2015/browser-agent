# Scenario

**Feature**: representative HTTPS app URL is capturable; attempt when gates open (#4)

```
URL = "https://app.example.com/app/weekly"
IsCapturableTabURL(URL) -> true
ShouldAttemptAttach(true, true, false, URL) -> true
```

## Preconditions

- Representative product app URL on a public HTTPS host.
- Gates fully open.

## Steps

1. Set `URL = "https://app.example.com/app/weekly"`.

## Context

- Mirrors the common case: tab navigates from blank/chrome to a real app URL.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.URL = "https://app.example.com/app/weekly"
	return nil
}
```