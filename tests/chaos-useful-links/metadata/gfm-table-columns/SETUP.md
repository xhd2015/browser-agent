# Scenario

**Feature**: useful-links style table fills Seed Env/Kind/Title/Market/URL

```
two GFM rows -> two seeds with table metadata
```

## Preconditions

- Parent fixture useful-links-table.md.

## Steps

1. WantCount=2.
2. Assertions check field mapping per URL.

## Context

- Row1: live, app_logs, ID live app logs, ID, https://logs.example.com/id
- Row2: test, grafana, Stress board, MY, https://grafana.example.com/stress

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.WantCount = 2
	req.WantCountSet = true
	req.WantURLs = []string{
		"logs.example.com/id",
		"grafana.example.com/stress",
	}
	return nil
}
```
