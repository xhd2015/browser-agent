# Scenario

**Feature**: reject control path+query on 127.0.0.1:43759 (req #1)

```
ShouldCaptureURL("http://127.0.0.1:43759/v1/entries?session=abc") -> false
```

## Preconditions

- Non-root path that the agent itself calls while recording (entries push).

## Steps

1. Set `CaptureURL = "http://127.0.0.1:43759/v1/entries?session=abc"`.

## Context

- Without this exclude, periodic POST /v1/entries would self-capture.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.CaptureURL = "http://127.0.0.1:43759/v1/entries?session=abc"
	return nil
}
```
