# Scenario

**Feature**: serve embedded identity + hello match statuses via GET /v1/session

```
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun)
Serve Runtime -> extract -> ensure bundle-sum -> log embedded version+md5
Fake Extension ?-> WS hello {version, features, bundle_md5?}
GET /v1/session -> bundled_extension + extension_match
```

## Preconditions

- Mode = session-match.
- Leaves set SessionMatchKind (C1–C5).
- Fake WS hello only when kind is not no-hello.

## Steps

1. Set Mode to session-match.
2. Leave SessionMatchKind / hello overrides for leaves.

## Context

- Requirement G3–G5 / scenarios C1–C5.
- Match enum: not_connected | ok | version_mismatch | md5_mismatch | md5_unknown.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeSessionMatch
	req.NoOpenChrome = true
	req.NoAgentRun = true
	return nil
}
```
