# Scenario

**Feature**: ValidateExtensionManifestJSON seals MV3 permission + host contract

```
# pure fixtures and production FS manifests
Test Client -> ValidateExtensionManifestJSON(bytes)
  -> nil | error mentioning missing permission/host
```

## Preconditions

- Mode = manifest-validate for this subtree.
- Pure leaves supply ManifestJSON; FS leaves set ManifestSource embedded|shell.

## Steps

1. Set Mode to ModeManifestValidate.
2. Leave ManifestSource / bytes for child Setup.

## Context

- Required permissions: debugger, tabs, alarms, storage.
- Required hosts: 43761 loopback + broad (`<all_urls>` or equivalent).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeManifestValidate
	return nil
}
```
