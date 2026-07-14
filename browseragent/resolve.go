package browseragent

import (
	"fmt"
	"strings"
)

// ResolveSessionID picks a session id for CLI/side-commands.
// Priority: flag (when flagSet) → non-empty env (when envSet) → error mentioning both.
func ResolveSessionID(flagValue string, flagSet bool, envValue string, envSet bool) (string, error) {
	if flagSet {
		return flagValue, nil
	}
	if envSet {
		if v := strings.TrimSpace(envValue); v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("session id required: pass --session-id or set BROWSER_AGENT_SESSION_ID")
}
