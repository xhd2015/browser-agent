package inject

// SessionNewHooks records injectable session-new hooks for CLI doctests.
type SessionNewHooks struct {
	OpenChromeFn    func(sessionURL, extensionInstallPath string) error
	AgentRunProbeFn func(sessionID, systemPromptPath, workspaceDir string, env map[string]string) error
}

// SessionNewTestHooks is the doctest assignment target for cliSessionNew hooks.
var SessionNewTestHooks *SessionNewHooks