# Scenario

**Feature**: HandleCLI create-tab posts job type create_tab; fake WS records first job

```
Test Client -> browseragent.Run(NoOpenChrome, NoAgentRun) @ free port
Fake Extension -> WS hello; on first job: record type+params; auto result {tab_id}
Operator -> HandleCLI([session, create-tab, --session-id, --addr, …])
  -> POST /v1/jobs with type=create_tab + params
  -> Fake Extension observes type/params
  -> stdout result + trailing \n; nil error
```

## Preconditions

- Mode is CLI job types.
- Server uses temp BaseDir + known SessionID + free Addr.
- No real Chrome / agent-run.
- Fake WS always enabled for observation (not real CDP).

## Steps

1. Set `Mode = ModeCLIJobTypes`.
2. Enable FakeExtension.
3. Increase MaxDispatchWait for job wait.
4. Leave JobCLI / CreateTabURL to leaf Setup.

## Context

- Harness constructs --addr and --session-id when CLIArgs empty.
- Asserts require ObservedJobType match; CLI success is secondary contract.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = ModeCLIJobTypes
	req.FakeExtension = true
	req.NoOpenChrome = true
	req.NoAgentRun = true
	if req.MaxDispatchWait == 0 {
		req.MaxDispatchWait = 10 * time.Second
	}
	if req.CLIEnv == nil {
		req.CLIEnv = map[string]string{}
	}
	return nil
}
```
