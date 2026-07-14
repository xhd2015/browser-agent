# Scenario

**Feature**: extract embedded MV3 extension to stable BaseDir path

```
# Extractor reads //go:embed tree and writes versioned directory
Test Client -> ExtractEmbeddedExtension(BaseDir)
Extractor -> {BaseDir}/extension/{version}/manifest.json
Extractor -> (installPath, version)
```

## Preconditions

- Mode is filesystem extract (no HTTP server, no CLI binary).
- BaseDir is empty temp directory (no prior extension/ tree).
- Embedded fixture/production tree includes readable manifest version.

## Steps

1. Set `Mode = ModeExtract` (`"extract"`).
2. Default `ExtractPasses = 1`; re-extract leaf overrides to 2.
3. Do not start control server.

## Context

- Asserts path layout, absolute install path, and manifest version equality.
- No Chrome launch.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtract
	if req.ExtractPasses == 0 {
		req.ExtractPasses = 1
	}
	return nil
}
```
