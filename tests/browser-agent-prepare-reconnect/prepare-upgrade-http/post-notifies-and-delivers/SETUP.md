# Scenario

**Feature**: POST prepare-upgrade notifies connected session and delivers WS envelope

```
Create sess-admin-a; hello WS
POST /v1/admin/prepare-upgrade {reason:"daemon-upgrade", delay_ms:1000}
  -> 200 {ok:true, notified:["sess-admin-a"]}
  -> WS type=prepare_reconnect delay_ms=1000
```

## Preconditions

- PrepareUpgradeCase = post-notifies-and-delivers.
- ConnectForHTTP = true; one session.
- Body includes reason + delay_ms.

## Steps

1. Set case, session, POST body fields.

## Context

- End-to-end admin path used by upgrade orchestration.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PrepareUpgradeCase = PrepareUpgradePostNotifies
	req.HTTPMethod = "POST"
	req.SessionIDs = []string{"sess-admin-a"}
	req.ConnectForHTTP = true
	req.HTTPBodyReason = "daemon-upgrade"
	req.HTTPBodyDelayMS = 1000
	req.HTTPBodyIncludeDelay = true
	return nil
}
```
