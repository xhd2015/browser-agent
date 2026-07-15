# Scenario

**Feature**: GFM table columns Env/Kind/Title/Market/Link map into Seed fields

```
testdata/useful-links-table.md
  | Env | Kind | Title | Market | Link |
  -> Seed.{Env,Kind,Title,Market,URL}
```

## Preconditions

- Mode is metadata.
- Fixture is useful-links-table.md.

## Steps

1. Set Mode to metadata.
2. Set Fixture to useful-links-table.md.

## Context

- Two table rows: live/app_logs and test/grafana.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeMetadata
	req.Fixture = "useful-links-table.md"
	req.IncludeArchived = false
	req.MaxSeeds = 0
	return nil
}
```
