# Scenario

**Feature**: control-server live entries API + HTML preview (mock extension push)

```
# Start Control Server without Chrome
Test Client -> browsertrace.Run(NoOpenChrome) -> Control Server @ free port

# Mock Extension push (product: ~1s while recording; clear = empty POST)
Mock Extension -> POST /v1/entries {session_id, entries, count}
Mock Extension -> POST /v1/entries {session_id, entries: [], count: 0}  # clear

# Surfaces under test
Test Client -> GET /v1/entries?session=… -> JSON snapshot
Test Client -> GET /preview?session=… -> HTML live viewer
```

## Preconditions

- Mode is HTTP (`ModeHTTP`).
- Live session id is `SessionSuffix` from root Setup.
- Probe and stage flags set by descendants.
- Routes under test do not yet exist on green product → TDD red until implementer adds them.

## Steps

1. Set `Mode = ModeHTTP` (`"http"`).
2. Ensure `NoOpenChrome = true`.
3. Descendants set `Probe`, staging POSTs, and known vs unknown session.

## Context

- No real extension process: harness is the mock pusher.
- Popup Clear confirm is product UI; this tree only checks the empty POST contract.
- Fallback extension `preview.html` is out of scope.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeHTTP
	req.NoOpenChrome = true
	return nil
}
```
