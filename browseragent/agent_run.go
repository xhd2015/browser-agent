package browseragent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AgentRunSessionIDPrefix namespaces agent-run chat sessions from control plane ids.
const AgentRunSessionIDPrefix = "browser-agent-sess-"

// AgentRunSessionID maps a control-plane session id to the agent-run session id.
// Idempotent: if control already has the prefix, it is returned as-is.
func AgentRunSessionID(controlID string) string {
	if strings.HasPrefix(controlID, AgentRunSessionIDPrefix) {
		return controlID
	}
	return AgentRunSessionIDPrefix + controlID
}

// BuildAgentRunArgs builds argv for agent-run (without the binary name).
//
// Shape:
//
//	run
//	--session-id=<AgentRunSessionID(control)>
//	--agent-runner=grok-tty
//	--auto-send-or-resume
//	--new-terminal
//	optional --dir=<workspace> when workspace non-empty
//	--env BROWSER_AGENT_SESSION_ID=<control>  // control id for nested CLI resolve
//	--no-submit  // ALWAYS — first prompt stays draft (no auto-submit in TTY)
//	--open
//	--
//	<prompt with absolute SYSTEM.md path when playbook path is provided>
func BuildAgentRunArgs(controlSessionID, promptOrSystemPath, workspaceDir string) []string {
	agentID := AgentRunSessionID(controlSessionID)
	args := []string{
		"run",
		"--session-id=" + agentID,
		"--agent-runner=grok-tty",
		"--auto-send-or-resume",
		"--new-terminal",
	}
	if strings.TrimSpace(workspaceDir) != "" {
		args = append(args, "--dir="+workspaceDir)
	}
	// Pass control id via agent-run --env so nested browser-agent session CLI can resolve it.
	// Do not rely on process env overlay.
	args = append(args,
		"--env", "BROWSER_AGENT_SESSION_ID="+controlSessionID,
		"--no-submit",
		"--open",
	)

	// Prefer absolute path to SYSTEM.md so the agent can open the playbook without
	// guessing cwd. Control id still comes from --env, not from prose.
	prompt := formatOpenPrompt(promptOrSystemPath)
	args = append(args, "--", prompt)
	return args
}

// formatOpenPrompt builds the agent-run open inject text.
// When promptOrSystemPath points at SYSTEM.md (or is empty), returns a recipe with
// an absolute playbook path when possible.
func formatOpenPrompt(promptOrSystemPath string) string {
	p := strings.TrimSpace(promptOrSystemPath)
	if p == "" {
		// No path available — still name the file; serve always passes abs path.
		return "Read the playbook at SYSTEM.md and co-pilot the browser-agent session."
	}
	low := strings.ToLower(p)
	if strings.HasSuffix(low, "system.md") || strings.HasSuffix(low, string(filepath.Separator)+"system.md") {
		abs := p
		if !filepath.IsAbs(p) {
			if a, err := filepath.Abs(p); err == nil {
				abs = a
			}
		} else {
			// Clean absolute path for stable display.
			abs = filepath.Clean(p)
		}
		// Prefer real abs when file exists (resolves symlinks when possible).
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			if r, err := filepath.Abs(abs); err == nil {
				abs = r
			}
		}
		return fmt.Sprintf("Read the playbook at %s and co-pilot the browser-agent session.", abs)
	}
	// Non-SYSTEM free-form prompt: pass through unchanged.
	return p
}

// launchAgentRun starts agent-run as a best-effort production launcher.
// Failures are returned to the caller (which logs and continues serve).
// Session id is carried only via BuildAgentRunArgs argv (--session-id + --env);
// no manual cmd.Env overlay for BROWSER_AGENT_SESSION_ID.
func launchAgentRun(controlSessionID, systemPromptPath, workspaceDir string) error {
	bin, err := exec.LookPath("agent-run")
	if err != nil {
		return fmt.Errorf("agent-run not found on PATH: %w", err)
	}
	args := BuildAgentRunArgs(controlSessionID, systemPromptPath, workspaceDir)
	cmd := exec.Command(bin, args...)
	// Inherit process env only; session env is the --env flag in argv.
	if strings.TrimSpace(workspaceDir) != "" {
		cmd.Dir = workspaceDir
	}
	return cmd.Start()
}
