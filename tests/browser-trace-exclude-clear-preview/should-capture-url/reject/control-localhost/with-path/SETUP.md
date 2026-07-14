# Scenario

**Feature**: reject localhost control path used by preview (req #1)

```
ShouldCaptureURL("http://localhost:43759/preview?session=x") -> false
```

## Preconditions

- Preview path on localhost control host (popup may open this URL).

## Steps

1. Set `CaptureURL = "http://localhost:43759/preview?session=x"`.

## Context

- Preview page loads must not appear as captured HAR entries.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "http://localhost:43759/preview?session=x"
	return nil
}
```
