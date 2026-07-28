## Expected

- PrepareUpgradeFn called with delayMS **1000**.
- Call order includes sequence: `prepare` then `sleep` then `kill` (spawn after kill).
- At least one SleepFn duration near **1200ms** (1000+200) OR sleeps of 1000 and 200.

## Side Effects

- Kill/spawn still run.

## Errors

- Missing prepare or kill-before-prepare fails.

## Exit Code

- 0.

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunErr(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureDaemon err=%v want nil", resp.EnsureErr)
	}
	if !resp.PrepareFnCalled {
		t.Fatal("PrepareUpgradeFn not called")
	}
	if resp.PrepareDelayMSArg != 1000 {
		t.Fatalf("PrepareUpgrade delayMS=%d want 1000", resp.PrepareDelayMSArg)
	}
	if !resp.KillFnCalled {
		t.Fatal("KillFn not called")
	}
	assertCallOrderHasSequence(t, resp.CallOrder, "prepare", "sleep", "kill")
	assertCallOrderHasSequence(t, resp.CallOrder, "kill", "spawn")

	// Accept combined 1200ms sleep or 1000+200.
	wantCombined := time.Duration(req.PrepareDelayMS+req.PrepareSlackMS) * time.Millisecond
	ok := false
	var sum time.Duration
	for _, d := range resp.SleepCalls {
		sum += d
		if d == wantCombined || d == time.Duration(req.PrepareDelayMS)*time.Millisecond {
			ok = true
		}
	}
	if !ok && sum < wantCombined {
		t.Fatalf("SleepCalls %v do not cover prepare delay+slack %s", resp.SleepCalls, wantCombined)
	}
	assertExitZero(t, resp)
}
```
