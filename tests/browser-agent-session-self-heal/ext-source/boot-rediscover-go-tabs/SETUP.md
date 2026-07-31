# Scenario

**Feature**: boot path rediscovers open `/go?session=` tabs via tabs.query

```
onInstalled / onStartup / healSessions()
  -> chrome.tabs.query({})   // not sessions.keys() alone
  -> parseGoSessionFromURL(tab.url) / isSessionGoPageURL
  -> discover { sessionId, tabId, windowId } for each open control tab
```

## Preconditions

- ExtSourceTarget = boot-rediscover-go-tabs.

## Steps

1. Set `ExtSourceTarget = ExtSrcBootRediscoverGoTabs`.

## Context

- Current boot only loops `sessions.keys()` — empty after cold SW restart — **RED**
  until tabs.query rediscover lands.
- Prefer reusing `parseGoSessionFromURL` / `isSessionGoPageURL` /
  `querySessionControlTabs` (or a dedicated heal helper that queries all tabs and
  filters `/go?session=`).
- Telemetry `collectSessionPageTelemetry` tabs.query is **not** boot heal unless
  wired from onInstalled/onStartup/heal entrypoint.

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ExtSourceTarget = ExtSrcBootRediscoverGoTabs
	return nil
}
```
