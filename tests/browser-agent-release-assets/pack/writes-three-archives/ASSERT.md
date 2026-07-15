## Expected

- Process exit code **0**.
- Under `--out`, each of the three `AssetReleaseNames(version)` basenames exists.
- Each archive file has **size > 0**.
- Exactly those three release archives are present (no extra non-junk files required
  to be absent of every possible name, but count of listed archives must be **3**
  and set must match the three expected basenames).

## Side Effects

- Writes only under the leaf temp `--out` directory.
- Does not call GitHub / `gh`.

## Errors

- Non-zero exit, missing archive, empty archive, or wrong count fails.

## Exit Code

- 0.

```go
import (
	"sort"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertExitZero(t, resp)

	want := expectedReleaseNames(req.Version)
	if len(want) != 3 {
		t.Fatalf("AssetReleaseNames(%q) len=%d want 3: %v", req.Version, len(want), want)
	}

	got := append([]string(nil), resp.OutBasenames...)
	sort.Strings(got)
	wantSorted := append([]string(nil), want...)
	sort.Strings(wantSorted)

	if len(got) != 3 {
		t.Fatalf("out archive count=%d want 3; got=%v stdout=%s stderr=%s",
			len(got), got, truncate(resp.Stdout, 300), truncate(resp.Stderr, 300))
	}
	for i := range wantSorted {
		if got[i] != wantSorted[i] {
			t.Fatalf("out basenames mismatch:\n  got  %v\n  want %v", got, wantSorted)
		}
	}
	for _, name := range want {
		sz, ok := resp.OutSizes[name]
		if !ok {
			t.Fatalf("missing size for %q; sizes=%v", name, resp.OutSizes)
		}
		if sz <= 0 {
			t.Fatalf("archive %q size=%d want > 0", name, sz)
		}
	}
}
```
