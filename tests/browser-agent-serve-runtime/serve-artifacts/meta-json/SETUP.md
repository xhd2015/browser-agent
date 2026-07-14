# Scenario

**Feature**: meta.json written at serve start with session discovery fields (A2)

```
# after health
sessionDir/meta.json → session_id, base_url|session_url, product browser-agent
```

## Preconditions

- ServeArtifactProbe = meta-json.
- NoOpenChrome + NoAgentRun.

## Steps

1. Set ServeArtifactProbe to meta-json.

## Context

- meta.json enables later CLI addr discovery; extras OK.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ServeArtifactProbe = ServeProbeMetaJSON
	return nil
}
```
