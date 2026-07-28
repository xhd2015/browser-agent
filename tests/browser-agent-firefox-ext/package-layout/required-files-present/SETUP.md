# Scenario

**Feature**: Firefox public package has required MV3 shell files

```
public/{manifest.json,background.js,contentScript.js,popup.html,popup.js}
```

## Preconditions

- PackageLayoutOp = required-files-present.

## Steps

1. Set PackageLayoutOp = PackageLayoutOpRequiredFiles.

## Context

- Thin popup shell OK (React integration later).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PackageLayoutOp = PackageLayoutOpRequiredFiles
	return nil
}
```
