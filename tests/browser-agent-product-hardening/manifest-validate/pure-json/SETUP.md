# Scenario

**Feature**: pure JSON fixtures exercise ValidateExtensionManifestJSON without FS

```
leaf manifest.json (valid | missing debugger | missing tabs)
  -> ValidateExtensionManifestJSON
```

## Preconditions

- ManifestSource = bytes.
- Each leaf provides sibling `manifest.json` fixture.

## Steps

1. Set ManifestSource to bytes.
2. Leaf Setup loads `manifest.json` into ManifestJSON.

## Context

- Requirement A1–A3.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestSource = ManifestSourceBytes
	return nil
}
```
