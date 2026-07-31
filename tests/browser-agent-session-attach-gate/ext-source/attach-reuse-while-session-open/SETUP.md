# Scenario

**Feature**: same-tab attach reuse + per-session attach serialization (policy B)

```
Same tab_id between jobs while control tab open -> reuse attach (no double attach)
Concurrent attach work for one session -> serialize via attachLock
```

## Preconditions

- ExtSourceTarget = attach-reuse-while-session-open.

## Steps

1. Set `ExtSourceTarget = ExtSrcAttachReuseWhileSessionOpen`.

## Context

- Policy B: reuse + lock only. **Does not** require (or allow as success criterion)
  switch-detach of peer tabs — multi-attach keep peers is
  `ext-source/multi-attach-keep-peers`.
- Overlaps `browser-agent-session-tab-targeting/ext-source/attach-reuse-same-tab`
  (also revised away from switch-detach).

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
