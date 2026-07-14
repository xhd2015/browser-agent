package browseragent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Job is a unit of work for the extension agent.
type Job struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id,omitempty"`
	Type      string         `json:"type"`
	Params    map[string]any `json:"params,omitempty"`
	TimeoutMS int64          `json:"timeout_ms,omitempty"`
	Status    string         `json:"status"`
	Result    *JobResult     `json:"result,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
}

// JobResult is the outcome of a job.
type JobResult struct {
	JobID      string         `json:"job_id"`
	OK         bool           `json:"ok"`
	Error      string         `json:"error,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	DurationMS int64          `json:"duration_ms,omitempty"`
}

// JobQueue is an in-memory FIFO job queue with wait/complete.
type JobQueue struct {
	mu      sync.Mutex
	order   []string
	jobs    map[string]*Job
	waiters map[string][]chan JobResult
	cond    *sync.Cond
}

// NewJobQueue creates an empty job queue.
func NewJobQueue() *JobQueue {
	q := &JobQueue{
		jobs:    make(map[string]*Job),
		waiters: make(map[string][]chan JobResult),
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue assigns an id if missing, status queued, and appends to FIFO.
func (q *JobQueue) Enqueue(j Job) (Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if j.ID == "" {
		j.ID = newID("job")
	}
	if j.Status == "" {
		j.Status = JobStatusQueued
	}
	if j.CreatedAt.IsZero() {
		j.CreatedAt = time.Now()
	}
	// Copy params map so callers can mutate freely.
	if j.Params != nil {
		cp := make(map[string]any, len(j.Params))
		for k, v := range j.Params {
			cp[k] = v
		}
		j.Params = cp
	}
	stored := j
	q.jobs[j.ID] = &stored
	q.order = append(q.order, j.ID)
	q.cond.Broadcast()
	return stored, nil
}

// Dequeue blocks until a queued job is available or ctx is done.
// Dequeued jobs move to running.
func (q *JobQueue) Dequeue(ctx context.Context) (Job, error) {
	// Wake cond on context cancel.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			q.mu.Lock()
			q.cond.Broadcast()
			q.mu.Unlock()
		case <-stop:
		}
	}()

	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		for i, id := range q.order {
			j, ok := q.jobs[id]
			if !ok {
				continue
			}
			if j.Status != JobStatusQueued {
				continue
			}
			// Remove from FIFO head position among queued; keep map entry.
			q.order = append(q.order[:i], q.order[i+1:]...)
			j.Status = JobStatusRunning
			out := *j
			return out, nil
		}
		if err := ctx.Err(); err != nil {
			return Job{}, err
		}
		q.cond.Wait()
		if err := ctx.Err(); err != nil {
			return Job{}, err
		}
	}
}

// TryDequeue returns the next queued job without blocking.
func (q *JobQueue) TryDequeue() (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, id := range q.order {
		j, ok := q.jobs[id]
		if !ok || j.Status != JobStatusQueued {
			continue
		}
		q.order = append(q.order[:i], q.order[i+1:]...)
		j.Status = JobStatusRunning
		return *j, true
	}
	return Job{}, false
}

// Wait blocks until the job reaches a terminal state or ctx is done.
// On context timeout/cancel while still pending, status becomes expired.
func (q *JobQueue) Wait(ctx context.Context, id string) (JobResult, error) {
	q.mu.Lock()
	j, ok := q.jobs[id]
	if !ok {
		q.mu.Unlock()
		return JobResult{}, fmt.Errorf("unknown job id %q", id)
	}
	if isTerminal(j.Status) {
		res := resultOf(j)
		q.mu.Unlock()
		return res, nil
	}
	ch := make(chan JobResult, 1)
	q.waiters[id] = append(q.waiters[id], ch)
	q.mu.Unlock()

	select {
	case res := <-ch:
		return res, nil
	case <-ctx.Done():
		q.mu.Lock()
		defer q.mu.Unlock()
		// Remove this waiter if still present.
		if list, ok := q.waiters[id]; ok {
			out := list[:0]
			for _, w := range list {
				if w != ch {
					out = append(out, w)
				}
			}
			if len(out) == 0 {
				delete(q.waiters, id)
			} else {
				q.waiters[id] = out
			}
		}
		j, ok := q.jobs[id]
		if !ok {
			return JobResult{JobID: id, OK: false, Error: "timeout"}, fmt.Errorf("timeout waiting for job %s", id)
		}
		if isTerminal(j.Status) {
			return resultOf(j), nil
		}
		// Expire pending job.
		j.Status = JobStatusExpired
		res := JobResult{
			JobID: id,
			OK:    false,
			Error: "timeout waiting for job result",
		}
		j.Result = &res
		// Do not notify other waiters with success; notify with expire if any remain.
		q.notifyWaitersLocked(id, res)
		return res, fmt.Errorf("timeout waiting for job %s", id)
	}
}

// Complete finishes a job with a result. Unknown id → error.
// Late complete after expired/failed/done is safe (ignored or soft error; status stays non-success if already expired).
func (q *JobQueue) Complete(id string, res JobResult) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	j, ok := q.jobs[id]
	if !ok {
		return fmt.Errorf("unknown job id %q: not found", id)
	}
	if res.JobID == "" {
		res.JobID = id
	}
	switch j.Status {
	case JobStatusDone, JobStatusFailed, JobStatusExpired:
		// Late / double complete: do not flip expired → done.
		return nil
	}
	if res.OK {
		j.Status = JobStatusDone
	} else {
		j.Status = JobStatusFailed
	}
	cp := res
	j.Result = &cp
	// Remove from FIFO if still queued.
	q.removeFromOrderLocked(id)
	q.notifyWaitersLocked(id, res)
	return nil
}

// Fail marks a non-terminal job as failed with an error message.
func (q *JobQueue) Fail(id string, errMsg string) error {
	return q.Complete(id, JobResult{JobID: id, OK: false, Error: errMsg})
}

// FailAllInflight fails every queued/running job (WS disconnect policy v1).
func (q *JobQueue) FailAllInflight(errMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for id, j := range q.jobs {
		if j.Status == JobStatusQueued || j.Status == JobStatusRunning {
			j.Status = JobStatusFailed
			res := JobResult{JobID: id, OK: false, Error: errMsg}
			j.Result = &res
			q.removeFromOrderLocked(id)
			q.notifyWaitersLocked(id, res)
		}
	}
}

// Get returns a job snapshot by id.
func (q *JobQueue) Get(id string) (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	j, ok := q.jobs[id]
	if !ok {
		return Job{}, false
	}
	return *j, true
}

// InflightCount returns queued plus running jobs.
func (q *JobQueue) InflightCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := 0
	for _, id := range q.order {
		j, ok := q.jobs[id]
		if !ok {
			continue
		}
		if j.Status == JobStatusQueued || j.Status == JobStatusRunning {
			n++
		}
	}
	return n
}

// SnapshotQueued returns a copy of currently queued (not yet dequeued/running) jobs in FIFO order.
func (q *JobQueue) SnapshotQueued() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	var out []Job
	for _, id := range q.order {
		if j, ok := q.jobs[id]; ok && j.Status == JobStatusQueued {
			out = append(out, *j)
		}
	}
	return out
}

// MarkRunning sets status running if still queued.
func (q *JobQueue) MarkRunning(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if j, ok := q.jobs[id]; ok && j.Status == JobStatusQueued {
		j.Status = JobStatusRunning
		q.removeFromOrderLocked(id)
	}
}

func (q *JobQueue) notifyWaitersLocked(id string, res JobResult) {
	list := q.waiters[id]
	delete(q.waiters, id)
	for _, ch := range list {
		select {
		case ch <- res:
		default:
		}
	}
}

func (q *JobQueue) removeFromOrderLocked(id string) {
	for i, oid := range q.order {
		if oid == id {
			q.order = append(q.order[:i], q.order[i+1:]...)
			return
		}
	}
}

func isTerminal(st string) bool {
	switch st {
	case JobStatusDone, JobStatusFailed, JobStatusExpired:
		return true
	default:
		return false
	}
}

func resultOf(j *Job) JobResult {
	if j.Result != nil {
		return *j.Result
	}
	switch j.Status {
	case JobStatusExpired:
		return JobResult{JobID: j.ID, OK: false, Error: "timeout"}
	case JobStatusFailed:
		return JobResult{JobID: j.ID, OK: false, Error: "failed"}
	case JobStatusDone:
		return JobResult{JobID: j.ID, OK: true}
	default:
		return JobResult{JobID: j.ID, OK: false}
	}
}

func newID(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}
