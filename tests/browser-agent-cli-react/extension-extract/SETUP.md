# Scenario

**Feature**: extract embedded Chrome-Ext-Browser-Agent (or mini fixture)

```
Test Client -> ExtractEmbeddedExtension(BaseDir)
  -> {BaseDir}/extension/{version}/manifest.json
Test Client -> InstallChromeExtension(stdout, BaseDir)
  -> path + Load unpacked + chrome://extensions + \n
```

## Preconditions

- ModeExtensionExtract.
- BaseDir empty temp (root Setup).
- Embedded mini MV3 includes version + 43761 hosts (see testdata/mini-extension).
- No webpack in CI.

## Steps

1. Set Mode = ModeExtensionExtract.
2. Default ExtractPasses = 1.
3. ExtractOp set by leaf.

## Context

- Asserts absolute path layout and trailing newline on install stdout.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeExtensionExtract
	if req.ExtractPasses == 0 {
		req.ExtractPasses = 1
	}
	return nil
}
```
