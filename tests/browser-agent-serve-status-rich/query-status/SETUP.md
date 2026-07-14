# Scenario

**Feature**: `QueryDaemonStatus` rich field population

```
QueryDaemonStatus(baseDir) -> DaemonStatus
  running:     DaemonVersion + extension fields + sessions
  not running: extension fields from canonical extract
```

## Preconditions

- Mode `ModeQueryStatus`.
- Leaf sets `QueryStatusOp`.
- Isolated `HOME` from root Setup.

## Steps

1. Set `Mode = ModeQueryStatus`.

## Context

- Must not mutate `server.json` after query.

```go
import (
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeQueryStatus
	return nil
}

// richStatusString reads a string field from DaemonStatus via reflection so
// query leaves compile before implementer adds struct fields (runtime RED).
func richStatusString(st browseragent.DaemonStatus, field string) string {
	v := reflect.ValueOf(st)
	f := v.FieldByName(field)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return strings.TrimSpace(f.String())
}

func assertRichStatusFieldsPresent(t *testing.T, st browseragent.DaemonStatus) {
	t.Helper()
	if richStatusString(st, "DaemonVersion") == "" && st.Running {
		t.Fatalf("DaemonVersion empty while Running=true; status=%+v", st)
	}
	extPath := richStatusString(st, "ExtensionPath")
	if extPath == "" {
		t.Fatalf("ExtensionPath empty; status=%+v", st)
	}
	assertCanonicalPathSegment(t, extPath)
	if richStatusString(st, "ExtensionVersion") == "" {
		t.Fatalf("ExtensionVersion empty; status=%+v", st)
	}
	if richStatusString(st, "ExtensionMD5") == "" {
		t.Fatalf("ExtensionMD5 empty; status=%+v", st)
	}
}
```