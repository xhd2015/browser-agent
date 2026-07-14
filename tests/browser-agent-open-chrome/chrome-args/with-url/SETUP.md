# Scenario

**Feature**: explicit URL appended to Chrome argv

```
BuildManagedChromeArgs(dataDir, extPath, url) -> argv includes url
```

## Preconditions

- ChromeArgsOp with-url.

## Steps

1. Set ChromeArgsOp = ChromeArgsOpWithURL.
2. Default URL to https://example.com/session when empty.

## Context

- Mirrors `open-chrome <url>` positional arg.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ChromeArgsOp = ChromeArgsOpWithURL
	if req.URL == "" {
		req.URL = "https://example.com/session"
	}
	return nil
}
```
