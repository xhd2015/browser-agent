# Scenario

**Feature**: heal reconnects WS only — no debugger re-attach storm

```
healSessionsFromTabs
  -> tabs.query + parse /go?session=
  -> maybeRegisterGoTab / connectSession
  -> NO maybeEagerAttach / attachDebuggerForSession in heal body
```

## Preconditions

- ExtSourceTarget = re-attach-on-heal (leaf name historical; contract is WS-only heal).

## Steps

1. Set `ExtSourceTarget = ExtSrcReAttachOnHeal`.

## Context

- Debugger attach inside heal freezes Chrome and blocks toolbar popup for 10–30s+.
- Content-tab multi-attach remains on create/navigate and jobs after heal reconnects WS.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcReAttachOnHeal
	return nil
}
```
