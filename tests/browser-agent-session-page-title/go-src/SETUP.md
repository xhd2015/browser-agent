# Scenario

**Feature**: pure-Go fallback session HTML title format (source contract)

```
Test Client -> read browseragent/server.go (writeFallbackSessionHTML)
  title must use {sessionId} - Browser Agent format
```

## Preconditions

- Mode `ModeGoSrc`.
- ModuleRoot resolved by root Setup.
- Embed is typically always present in-module, so fallback is hard to force via
  HTTP alone — this grouping seals the fallback source contract.

## Steps

1. Set `Mode = ModeGoSrc`.
2. Leaf sets `GoSrcProbe`.

## Context

- Requirement surface: `writeFallbackSessionHTML` same title format as inject.
- Classic TDD: current static `browser-agent session` → RED until implementer.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeGoSrc
	return nil
}
```
