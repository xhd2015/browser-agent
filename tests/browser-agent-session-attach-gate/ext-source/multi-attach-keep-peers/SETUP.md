# Scenario

**Feature**: multi-tab attach set keeps peer tabs attached (policy B)

```
attachDebuggerForSession(S, tabA) -> set = {A}
attachDebuggerForSession(S, tabB) -> set = {A,B}
  (A stays attached; no switch-detach of A)
```

## Preconditions

- ExtSourceTarget = multi-attach-keep-peers.

## Steps

1. Set `ExtSourceTarget = ExtSrcMultiAttachKeepPeers`.

## Context

- Current sticky code detaches previous `attachedTabId` when attaching a different tab
  — **RED** until multi-set policy B lands.
- Same-tab reuse and attachLock are covered by `attach-reuse-while-session-open`.
- Global `attachedTabs` Map alone is not per-session set state.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcMultiAttachKeepPeers
	return nil
}
```
