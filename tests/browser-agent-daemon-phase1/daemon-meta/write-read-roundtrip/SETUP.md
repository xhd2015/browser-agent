# Scenario

**Feature**: WriteDaemonMeta then ReadDaemonMeta preserves all fields

```
meta -> WriteDaemonMeta(path) -> ReadDaemonMeta(path) -> same fields
```

## Preconditions

- DaemonMetaOp is write-read-roundtrip.
- Sample meta with known pid, addr, base_url, base_dir, started_at.

## Steps

1. Set DaemonMetaOp to write-read-roundtrip.
2. Set Meta to sampleDaemonMeta(BaseDir).

## Context

- Roundtrip must preserve pid, addr, base_url, base_dir, started_at.
- On-disk JSON should end with trailing newline.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DaemonMetaOp = DaemonMetaWriteReadRoundtrip
	req.Meta = expectedSampleMeta(req.BaseDir)
	return nil
}
```