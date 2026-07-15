package browseragent

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// debugJobLog writes a single-line job lifecycle message to stderr.
// Detached daemons redirect stderr to {base-dir}/serve.log, so operators can
// tail that file when diagnosing create_tab / Page.navigate failures.
//
// Enable extra verbosity with BROWSER_AGENT_DEBUG=1 (same lines, still always
// on for enqueue/result of create_tab and cdp navigate).
func debugJobLog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	// Always use standard log so timestamps appear when process owns stderr.
	log.Printf("browser-agent: %s", msg)
}

func jobParamsSummary(jobType string, params map[string]any) string {
	if params == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	switch jobType {
	case "create_tab":
		if u, ok := stringParam(params, "url", "URL", "href"); ok {
			parts = append(parts, "url="+truncateForLog(u, 200))
		}
		if a, ok := params["active"]; ok {
			parts = append(parts, fmt.Sprintf("active=%v", a))
		}
	case "cdp":
		if m, ok := stringParam(params, "method", "cdp_method", "cdpMethod"); ok {
			parts = append(parts, "method="+m)
		}
		if nested, ok := params["params"].(map[string]any); ok {
			if u, ok := stringParam(nested, "url"); ok {
				parts = append(parts, "nav_url="+truncateForLog(u, 200))
			}
		}
	case "eval", "run":
		if e, ok := stringParam(params, "expression", "expr", "code", "source", "script"); ok {
			parts = append(parts, "expr="+truncateForLog(e, 80))
		}
	default:
		// Generic: first few string-ish keys.
		n := 0
		for k, v := range params {
			if n >= 3 {
				break
			}
			if s, ok := v.(string); ok && s != "" {
				parts = append(parts, k+"="+truncateForLog(s, 60))
				n++
			}
		}
	}
	return strings.Join(parts, " ")
}

func stringParam(m map[string]any, keys ...string) (string, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t, true
				}
			}
		}
	}
	return "", false
}

func truncateForLog(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func resultDataSummary(data map[string]any) string {
	if data == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if t, ok := data["type"].(string); ok && t != "" {
		parts = append(parts, "type="+t)
	}
	if id, ok := data["tab_id"]; ok {
		parts = append(parts, fmt.Sprintf("tab_id=%v", id))
	}
	if u, ok := data["url"].(string); ok && u != "" {
		parts = append(parts, "url="+truncateForLog(u, 160))
	}
	if m, ok := data["method"].(string); ok && m != "" {
		parts = append(parts, "method="+m)
	}
	if _, ok := data["polyfilled"]; ok {
		parts = append(parts, "polyfilled=true")
	}
	return strings.Join(parts, " ")
}

// debugEnabled returns true when BROWSER_AGENT_DEBUG is set to a truthy value.
func debugEnabled() bool {
	v := strings.TrimSpace(os.Getenv("BROWSER_AGENT_DEBUG"))
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on", "debug":
		return true
	default:
		return false
	}
}

// shouldAlwaysLogJob is true for high-signal navigation/tab jobs even without
// BROWSER_AGENT_DEBUG.
func shouldAlwaysLogJob(jobType string, params map[string]any) bool {
	switch jobType {
	case "create_tab":
		return true
	case "cdp":
		m, _ := stringParam(params, "method", "cdp_method", "cdpMethod")
		return m == "Page.navigate" || strings.HasPrefix(m, "Target.")
	default:
		return debugEnabled()
	}
}

func elapsedMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
