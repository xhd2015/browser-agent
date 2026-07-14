# Scenario

**Feature**: embedded package extension manifest is a production source under contract

```
ModuleRoot/browseragent/embedded/extension/manifest.json
  -> ValidateExtensionManifestJSON
```

## Preconditions

- ManifestSource = embedded.
- File must exist under module root (shipped with package).

## Steps

1. Set ManifestSource to embedded.
2. Leaf asserts Validate OK.

## Context

- Requirement A4. This is what install/extract ships to operators.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.ManifestSource = ManifestSourceEmbedded
	return nil
}
```
