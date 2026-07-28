# Scenario

**Feature**: POST prepare-upgrade with no connected extension returns empty notified

```
Create sess-admin-a; no WS
POST /v1/admin/prepare-upgrade
  -> 200 {ok:true, notified:[]}
```

## Preconditions

- PrepareUpgradeCase = post-empty-notified.
- ConnectForHTTP = false.
- Session still created so registry is non-empty.

## Steps

1. Set case without connecting WS.

## Context

- Upgrade may run when extensions are already offline; still succeeds with empty list.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PrepareUpgradeCase = PrepareUpgradePostEmpty
	req.HTTPMethod = "POST"
	req.SessionIDs = []string{"sess-admin-a"}
	req.ConnectForHTTP = false
	req.HTTPBodyReason = "daemon-upgrade"
	req.HTTPBodyDelayMS = 1000
	req.HTTPBodyIncludeDelay = true
	return nil
}
```
