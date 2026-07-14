# Scenario

**Feature**: product markers show browser-agent port 43761 (E4)

```
GET /go|/; body or boot config mentions 43761 and/or browser-agent product name
  (install guideline parameterization)
```

## Preconditions

- Same SPA shell as sibling; assert product strings.

## Steps

1. Probe go-html without hello.

## Context

- Product default control port is **43761** (not browser-trace 43759).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	// Product string leaf: same HTML surface; assert 43761 + browser-agent labels.
	req.DoWSHello = false
	req.Probe = ProbeGoHTML
	return nil
}
```

