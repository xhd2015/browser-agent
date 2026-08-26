## Expected

- Exit 0 JSON; authorization redacted; no raw Bearer JWT; response result id 42

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp.CLIErr != "" {
		t.Fatalf("CLIErr=%q", resp.CLIErr)
	}
	assertExitZero(t, resp)
	if strings.Contains(resp.Stdout, "Bearer eyJ") {
		t.Fatal("raw bearer leaked")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(resp.Stdout), &payload); err != nil {
		t.Fatalf("json: %v\n%s", err, truncate(resp.Stdout, 400))
	}
	headers := payload["request"].(map[string]any)["headers"].(map[string]any)
	if headers["authorization"] != "<REDACTED>" {
		t.Fatalf("authorization=%v", headers["authorization"])
	}
	result := payload["response"].(map[string]any)["body"].(map[string]any)["result"].(map[string]any)
	if int(result["id"].(float64)) != 42 {
		t.Fatalf("result=%v", result)
	}
}
```
