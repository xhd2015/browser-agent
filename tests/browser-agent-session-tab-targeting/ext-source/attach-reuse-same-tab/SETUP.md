# Scenario

**Feature**: debugger attach reuse + per-session serialize (policy B)

```
Same tab_id between jobs -> reuse attach (no duplicate attach)
Attach work per session -> serialize via attachLock
Different tab_id -> peers stay attached (multi-attach set; not switch-detach)
```

## Preconditions

- ExtSourceTarget = attach-reuse-same-tab.

## Steps

1. Set `ExtSourceTarget = ExtSrcAttachReuseSameTab`.

## Context

- Screenshot fix: avoid double-attach race; clear error if DevTools already attached.
- Multi-attach keep peers is asserted in
  `browser-agent-session-attach-gate/ext-source/multi-attach-keep-peers`.
- Switch-detach (policy A) is obsolete and must not be required here.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcAttachReuseSameTab
	return nil
}
```
