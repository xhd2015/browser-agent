# Scenario

**Feature**: GET on prepare-upgrade path is rejected (POST-only)

```
GET /v1/admin/prepare-upgrade -> 405 Method Not Allowed
```

## Preconditions

- PrepareUpgradeCase = get-method-not-allowed.
- HTTPMethod = GET.
- No WS required.

## Steps

1. Set case to GET the admin path.

## Context

- Guards accidental browser navigation / caching of a mutation endpoint.

```go
import (
	"net/http"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PrepareUpgradeCase = PrepareUpgradeGetNotAllowed
	req.HTTPMethod = http.MethodGet
	req.SessionIDs = []string{"sess-admin-a"}
	req.ConnectForHTTP = false
	return nil
}
```
