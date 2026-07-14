# Scenario

**Feature**: missing `debugger` permission → validate error mentions debugger (A2)

```
manifest.json permissions without debugger
  -> ValidateExtensionManifestJSON -> error contains "debugger"
```

## Preconditions

- Sibling fixture omits only `debugger` (other required fields present).

## Steps

1. Load `manifest.json` into ManifestJSON.

## Context

- Requirement A2. Regression that shipped storage-only must not pass.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	data := readLeafManifest(t, "manifest.json")
	mustValidJSONObject(t, data)
	req.ManifestJSON = data
	req.ManifestSource = ManifestSourceBytes
	return nil
}
```
