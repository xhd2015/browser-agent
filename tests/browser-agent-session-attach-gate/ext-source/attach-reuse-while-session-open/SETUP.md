# Scenario

**Feature**: sticky attach reuse while session open + detach on tab switch (regression)

```
Same tab_id between jobs while control tab open -> reuse attach
Different tab_id -> detach previous; serialize attach per session
```

## Preconditions

- ExtSourceTarget = attach-reuse-while-session-open.

## Steps

1. Set `ExtSourceTarget = ExtSrcAttachReuseWhileSessionOpen`.

## Context

- Regression-friendly: policy keeps sticky attach **while** session control tab is open.
- Overlaps `browser-agent-session-tab-targeting/ext-source/attach-reuse-same-tab`;
  included so attach-gate work does not regress reuse/switch.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAttachReuseWhileSessionOpen
	return nil
}
```
