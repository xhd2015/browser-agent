## Expected

- HTTP status **200**.
- `jobs` is non-empty.
- First job has non-empty `id`/`job_id`.
- First job `type` is **`eval`**.

## Side Effects

- Job status becomes running (leased) for later result.

## Errors

- Empty jobs, wrong type, or route missing fails.

## Exit Code

- 0.

```go
import (
	"net/http"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d want 200; body=%s (route POST /v1/ext/poll must exist)",
			resp.StatusCode, truncate(resp.BodyString, 500))
	}
	if resp.JobCount < 1 {
		t.Fatalf("JobCount=%d want >=1 after enqueue; body=%s",
			resp.JobCount, truncate(resp.BodyString, 500))
	}
	if resp.FirstJobID == "" {
		t.Fatalf("first job missing id/job_id; body=%s", truncate(resp.BodyString, 500))
	}
	if resp.FirstJobType != "eval" {
		t.Fatalf("first job type=%q want eval; body=%s",
			resp.FirstJobType, truncate(resp.BodyString, 500))
	}
	assertJSONContentType(t, resp.ContentType)
	assertExitZero(t, resp)
}
```
