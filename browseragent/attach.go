package browseragent

import (
	"strings"
	"time"
)

// Attach stage pipeline for Chrome cold-start diagnosis.
// Monotonic while attaching; reset toward page_open on disconnect.
const (
	AttachStageNoPage       = "no_page"
	AttachStagePageOpen     = "page_open"
	AttachStageRegisterSent = "register_sent"
	AttachStageSWAck        = "sw_ack"
	AttachStageWSConnecting = "ws_connecting"
	AttachStageWSOpen       = "ws_open"
	AttachStageHello        = "hello"
	AttachStageReady        = "ready"
)

// Default adaptive wait for session new (when WaitExtensionTimeout is unset).
const (
	DefaultAttachWaitBase  = 15 * time.Second
	DefaultAttachWaitGrant = 10 * time.Second
	DefaultAttachWaitCap   = 45 * time.Second
)

// attachState is ephemeral per-session cold-start progress (not persisted).
type attachState struct {
	Stage            string
	LastError        string
	RegisterAttempts int
	UpdatedAt        time.Time
}

// sessionAttach is the JSON shape exposed on GET /v1/session.
type sessionAttach struct {
	Stage            string    `json:"stage,omitempty"`
	LastError        string    `json:"last_error,omitempty"`
	RegisterAttempts int       `json:"register_attempts,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

// AttachEvent is POST /v1/ext/attach body (page / content script / tests).
type AttachEvent struct {
	SessionID        string `json:"session_id"`
	Stage            string `json:"stage"`
	LastError        string `json:"last_error"`
	RegisterAttempts int    `json:"register_attempts"`
}

func attachStageRank(stage string) int {
	switch strings.TrimSpace(stage) {
	case AttachStageNoPage:
		return 0
	case AttachStagePageOpen:
		return 1
	case AttachStageRegisterSent:
		return 2
	case AttachStageSWAck:
		return 3
	case AttachStageWSConnecting:
		return 4
	case AttachStageWSOpen:
		return 5
	case AttachStageHello:
		return 6
	case AttachStageReady:
		return 7
	default:
		return -1
	}
}

func normalizeAttachStage(stage string) string {
	s := strings.TrimSpace(stage)
	if attachStageRank(s) < 0 {
		return ""
	}
	return s
}

// applyAttachEvent merges an attach event. Stage only moves forward (or stays).
// last_error is replaced when non-empty; cleared when stage reaches sw_ack+.
// register_attempts uses max(current, incoming) when incoming > 0.
func (s *session) applyAttachEvent(ev AttachEvent) {
	stage := normalizeAttachStage(ev.Stage)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.attach.Stage == "" {
		s.attach.Stage = AttachStageNoPage
	}
	changed := false
	if stage != "" {
		if attachStageRank(stage) >= attachStageRank(s.attach.Stage) {
			if stage != s.attach.Stage {
				changed = true
			}
			s.attach.Stage = stage
		}
	}
	if ev.RegisterAttempts > s.attach.RegisterAttempts {
		s.attach.RegisterAttempts = ev.RegisterAttempts
		changed = true
	}
	if err := strings.TrimSpace(ev.LastError); err != "" {
		if err != s.attach.LastError {
			changed = true
		}
		s.attach.LastError = err
	} else if attachStageRank(s.attach.Stage) >= attachStageRank(AttachStageSWAck) && s.attach.LastError != "" {
		s.attach.LastError = ""
		changed = true
	}
	if changed {
		s.attach.UpdatedAt = time.Now()
	}
}

// setAttachStageLocked promotes stage if higher. Caller holds s.mu.
func (s *session) setAttachStageLocked(stage string) {
	stage = normalizeAttachStage(stage)
	if stage == "" {
		return
	}
	if s.attach.Stage == "" {
		s.attach.Stage = AttachStageNoPage
	}
	if attachStageRank(stage) >= attachStageRank(s.attach.Stage) {
		s.attach.Stage = stage
		if attachStageRank(stage) >= attachStageRank(AttachStageSWAck) {
			s.attach.LastError = ""
		}
		s.attach.UpdatedAt = time.Now()
	}
}

func (s *session) setAttachStage(stage string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setAttachStageLocked(stage)
}

func (s *session) resetAttachOnDisconnect() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Keep page_open if we already saw a page; otherwise no_page.
	prev := s.attach.Stage
	attempts := s.attach.RegisterAttempts
	s.attach = attachState{
		RegisterAttempts: attempts,
		UpdatedAt:        time.Now(),
	}
	if attachStageRank(prev) >= attachStageRank(AttachStagePageOpen) ||
		(s.sessionPageCount != nil && *s.sessionPageCount > 0) {
		s.attach.Stage = AttachStagePageOpen
	} else {
		s.attach.Stage = AttachStageNoPage
	}
}

func (s *session) snapshotAttach() *sessionAttach {
	if s.attach.Stage == "" && s.attach.RegisterAttempts == 0 && s.attach.LastError == "" && s.attach.UpdatedAt.IsZero() {
		return nil
	}
	out := &sessionAttach{
		Stage:            s.attach.Stage,
		LastError:        s.attach.LastError,
		RegisterAttempts: s.attach.RegisterAttempts,
		UpdatedAt:        s.attach.UpdatedAt,
	}
	if out.Stage == "" {
		out.Stage = AttachStageNoPage
	}
	return out
}

// attachDeadline computes adaptive wait: base with no progress, +grant per
// stage advance, hard-capped at cap. Explicit WaitExtensionTimeout becomes cap
// (and shrinks base when smaller — doctests use 3s).
func attachDeadline(explicitTimeout time.Duration) (base, grant, cap time.Duration) {
	cap = DefaultAttachWaitCap
	base = DefaultAttachWaitBase
	grant = DefaultAttachWaitGrant
	if explicitTimeout > 0 {
		cap = explicitTimeout
	}
	if base > cap {
		base = cap
	}
	return base, grant, cap
}
