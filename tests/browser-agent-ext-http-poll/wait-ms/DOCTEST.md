# NormalizeExtPollWaitMS — pure wait_ms helper (nested root)

Nested Classic TDD tree for the pure poll wait normalizer used by
`POST /v1/ext/poll`. Kept separate so the parent HTTP route tree still **compiles**
while this symbol is unimplemented.

## Version

0.0.2

# DSN (Domain Specific Notion)

**NormalizeExtPollWaitMS** is a pure function on the control plane:

- Input `waitMS <= 0` (omitted / zero) → **DefaultExtPollWaitMS (25000)**
- Input within `(0, MaxExtPollWaitMS]` → unchanged
- Input above **MaxExtPollWaitMS (30000)** → clamped to 30000

**Test Client** calls the package helper with fixed integers; no HTTP, no registry.

```text
NormalizeExtPollWaitMS(0)     -> 25000
NormalizeExtPollWaitMS(5000)  -> 5000
NormalizeExtPollWaitMS(99999) -> 30000
```

## Decision Tree

```
wait-ms/
├── default/       <=0 → 25000
├── within-cap/    5000 → 5000
└── over-cap/      99999 → 30000
```

## Test Index

| Leaf | Scenario |
|------|----------|
| `default` | input 0 → 25000 |
| `within-cap` | input 5000 → 5000 |
| `over-cap` | input 99999 → 30000 |

**Leaf count: 3**

## How to Run

```sh
doctest vet ./tests/browser-agent-ext-http-poll/wait-ms
doctest test ./tests/browser-agent-ext-http-poll/wait-ms   # RED until helper exists
```

### Implementer contract

```text
const DefaultExtPollWaitMS = 25000
const MaxExtPollWaitMS     = 30000
func NormalizeExtPollWaitMS(waitMS int) int
```

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/browser-agent/browseragent"
)

// WaitMSCase — pure helper scenarios.
const (
	WaitMSDefault   = "default"
	WaitMSWithinCap = "within-cap"
	WaitMSOverCap   = "over-cap"
)

// Request is narrowed root→leaf.
type Request struct {
	WaitMSCase  string
	WaitMSInput int
	ModuleRoot  string
}

// Response holds the normalized wait.
type Response struct {
	NormalizedWaitMS int
	ExitCode         int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	if req.WaitMSCase == "" {
		t.Fatal("WaitMSCase must be set by leaf Setup")
	}
	n := browseragent.NormalizeExtPollWaitMS(req.WaitMSInput)
	return &Response{ExitCode: 0, NormalizedWaitMS: n}, nil
}

// silence unused
var _ = fmt.Sprintf
```
