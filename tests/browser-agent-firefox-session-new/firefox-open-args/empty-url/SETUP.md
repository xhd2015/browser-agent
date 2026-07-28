# Scenario

**Feature**: BuildFirefoxOpenArgs with empty URL

```
BuildFirefoxOpenArgs("") -> no --load-extension / --user-data-dir
```

## Preconditions

- FirefoxOpenArgsOp = empty-url.

## Steps

1. Set FirefoxOpenArgsOp = FirefoxOpenArgsOpEmptyURL.
2. Clear URL.

## Context

- Empty URL still forbids managed Firefox profile flags.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.FirefoxOpenArgsOp = FirefoxOpenArgsOpEmptyURL
	req.URL = ""
	return nil
}
```
