# Scenario

**Feature**: Firefox public manifest is MV3 with stable gecko.id

```
public/manifest.json -> manifest_version 3 + browser_specific_settings.gecko.id
```

## Preconditions

- PackageLayoutOp = manifest-gecko-id.

## Steps

1. Set PackageLayoutOp = PackageLayoutOpManifestGeckoID.

## Context

- Stable id example: browser-agent@xhd2015 (any non-empty gecko id OK).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PackageLayoutOp = PackageLayoutOpManifestGeckoID
	return nil
}
```
