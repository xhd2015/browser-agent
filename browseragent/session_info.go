package browseragent

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// mergeSessionInfoEnrichment copies enriched browser fields to the top level of
// session info --json output (tabs[], job_target, recommended_cli).
func mergeSessionInfoEnrichment(outObj, browser map[string]any) {
	if outObj == nil || browser == nil {
		return
	}
	for _, k := range []string{"tabs", "job_target", "recommended_cli"} {
		if v, ok := browser[k]; ok && v != nil {
			outObj[k] = v
		}
	}
}

// FormatSessionInfoOptions controls human session info rendering.
type FormatSessionInfoOptions struct {
	Color serveColor
}

// FormatSessionInfo writes human-readable session info sections to w.
func FormatSessionInfo(w io.Writer, snap sessionSnapshot, browser map[string]any, opts FormatSessionInfoOptions) error {
	if w == nil {
		w = io.Discard
	}

	if browser != nil {
		if tabs, ok := browser["tabs"].([]any); ok && hasEnrichedTabIndex(tabs) {
			return formatCompactEnrichedSessionInfo(w, snap, browser, tabs, opts)
		}
	}

	if _, err := fmt.Fprintf(w, "Session: %s\n", snap.SessionID); err != nil {
		return err
	}

	created := formatSessionCreated(snap.CreatedAt)
	if _, err := fmt.Fprintf(w, "Created: %s\n", created); err != nil {
		return err
	}

	statusOut := snap.StatusLabel
	if statusOut == "" {
		statusOut = snap.Status
	}
	if _, err := fmt.Fprintf(w, "Status: %s\n", opts.Color.enabledStatus(snap.Status, statusOut)); err != nil {
		return err
	}

	if snap.Phase != "" {
		if _, err := fmt.Fprintf(w, "Phase: %s\n", snap.Phase); err != nil {
			return err
		}
	}

	extLine := "disconnected"
	if snap.Extension.Connected {
		extLine = "connected"
		if snap.Extension.Version != "" {
			extLine += " (v" + snap.Extension.Version + ")"
		}
	}
	if _, err := fmt.Fprintf(w, "Extension: %s\n", extLine); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Pages: %s\n", formatPageCountDisplay(snap.SessionPageCount)); err != nil {
		return err
	}

	browserLine := formatBrowserListDisplay(snap.Browsers)
	if _, err := fmt.Fprintf(w, "Browsers: %s\n", browserLine); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "Jobs: %d inflight\n", snap.InflightJobs); err != nil {
		return err
	}

	if snap.SessionURL != "" {
		if _, err := fmt.Fprintf(w, "Session URL (session_url): %s\n", snap.SessionURL); err != nil {
			return err
		}
	} else if snap.SessionID != "" {
		if _, err := fmt.Fprintf(w, "Session URL (session_url): /go?session=%s\n", snap.SessionID); err != nil {
			return err
		}
	}

	if len(snap.SessionPages) > 0 {
		if _, err := fmt.Fprintln(w, "Session pages:"); err != nil {
			return err
		}
		for _, p := range snap.SessionPages {
			line := p.URL
			if line == "" && p.TabID > 0 {
				line = fmt.Sprintf("tab %d", p.TabID)
			}
			if _, err := fmt.Fprintf(w, "  - %s\n", line); err != nil {
				return err
			}
		}
	}

	if browser != nil {
		if err := formatSessionTabsSection(w, browser); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Next steps:"); err != nil {
		return err
	}

	if snap.Extension.Connected {
		if _, err := fmt.Fprintln(w, "  - Extension connected; browser-agent jobs are ready."); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(w, "  - Run: browser-agent install-chrome-extension"); err != nil {
			return err
		}
		if snap.ExtensionInstallPath != "" {
			if _, err := fmt.Fprintf(w, "  - Load unpacked from: %s\n", snap.ExtensionInstallPath); err != nil {
				return err
			}
		}
	}

	if snap.SessionPageCount != nil && *snap.SessionPageCount == 0 {
		if _, err := fmt.Fprintf(w, "  - No session pages open — run: browser-agent session delete --session-id %s\n", snap.SessionID); err != nil {
			return err
		}
	} else if !snap.Extension.Connected {
		if _, err := fmt.Fprintf(w, "  - To clean up an unused session: browser-agent session delete --session-id %s\n", snap.SessionID); err != nil {
			return err
		}
	}

	return nil
}

func (c serveColor) enabledStatus(status, label string) string {
	if !c.enabled {
		return label
	}
	return statusColor(c, status, label)
}

func formatSessionCreated(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	age := formatSessionAge(t)
	ts := t.Format(time.RFC3339)
	if age != "" {
		return ts + " (" + age + ")"
	}
	return ts
}

func formatSessionAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd ago", days)
	}
	return ""
}

func formatPageCountDisplay(count *int) string {
	if count == nil {
		return "—"
	}
	return fmt.Sprintf("%d", *count)
}

func formatBrowserListDisplay(browsers []string) string {
	if len(browsers) == 0 {
		return "—"
	}
	return strings.Join(browsers, ", ")
}

func formatSessionTabsSection(w io.Writer, browser map[string]any) error {
	tabs, ok := browser["tabs"].([]any)
	if !ok || len(tabs) == 0 {
		return nil
	}
	if hasEnrichedTabIndex(tabs) {
		return formatEnrichedTabsTable(w, browser, tabs)
	}
	if _, err := fmt.Fprintln(w, "Browser tabs:"); err != nil {
		return err
	}
	for _, item := range tabs {
		tab, ok := item.(map[string]any)
		if !ok {
			continue
		}
		title, _ := tab["title"].(string)
		url, _ := tab["url"].(string)
		if title == "" {
			title = url
		}
		if title == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "  - %s\n", title); err != nil {
			return err
		}
	}
	return nil
}

func formatCompactEnrichedSessionInfo(w io.Writer, snap sessionSnapshot, browser map[string]any, tabs []any, opts FormatSessionInfoOptions) error {
	statusOut := snap.StatusLabel
	if statusOut == "" {
		statusOut = snap.Status
	}
	if _, err := fmt.Fprintf(w, "Session %s                                    %s\n",
		snap.SessionID, opts.Color.enabledStatus(snap.Status, statusOut)); err != nil {
		return err
	}

	extLine := "Extension disconnected"
	if snap.Extension.Connected {
		extLine = "Extension connected"
		if snap.Extension.Version != "" {
			extLine += " (v" + snap.Extension.Version + ")"
		}
	}
	if _, err := fmt.Fprintf(w, "%s\n", extLine); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "  Idx  ID        Active  Role          Title\n"); err != nil {
		return err
	}

	for _, item := range tabs {
		tab, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if err := writeEnrichedTabRow(w, tab); err != nil {
			return err
		}
	}

	return writeEnrichedTabsFooter(w, browser)
}

func writeEnrichedTabRow(w io.Writer, tab map[string]any) error {
	idx := tabIntField(tab, "index")
	id := tabIntField(tab, "id")
	if id == 0 {
		id = tabIntField(tab, "tab_id")
	}
	activeMark := " "
	if active, _ := tab["active"].(bool); active {
		activeMark = "*"
	}
	role := tabRoleDisplay(tab)
	title := tabStringField(tab, "title")
	if title == "" {
		title = tabStringField(tab, "url")
	}
	if len(title) > 40 {
		title = title[:37] + "..."
	}
	_, err := fmt.Fprintf(w, "  %-4d %-9d %-7s %-13s %s\n", idx, id, activeMark, role, title)
	return err
}

func hasEnrichedTabIndex(tabs []any) bool {
	for _, item := range tabs {
		if tab, ok := item.(map[string]any); ok {
			if _, ok := tab["index"]; ok {
				return true
			}
		}
	}
	return false
}

func formatEnrichedTabsTable(w io.Writer, browser map[string]any, tabs []any) error {
	windowID := "?"
	if v, ok := browser["window_id"]; ok {
		windowID = fmt.Sprintf("%v", v)
	} else if v, ok := browser["windowId"]; ok {
		windowID = fmt.Sprintf("%v", v)
	}
	if _, err := fmt.Fprintf(w, "Tabs (window %s, %d capturable)\n", windowID, len(tabs)); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "  Idx  ID        Active  Role          Title"); err != nil {
		return err
	}
	for _, item := range tabs {
		tab, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if err := writeEnrichedTabRow(w, tab); err != nil {
			return err
		}
	}
	return writeEnrichedTabsFooter(w, browser)
}

func writeEnrichedTabsFooter(w io.Writer, browser map[string]any) error {
	if jt, ok := browser["job_target"].(map[string]any); ok && jt != nil {
		tabIdx := tabIntField(jt, "tab_index")
		tabID := tabIntField(jt, "tab_id")
		reason, _ := jt["reason"].(string)
		if reason == "" {
			reason = "active in session window"
		} else {
			reason = strings.ReplaceAll(reason, "_", " ")
		}
		if _, err := fmt.Fprintf(w, "Job target  idx %d / tab %d  (%s)\n", tabIdx, tabID, reason); err != nil {
			return err
		}
	}
	if hint, ok := browser["recommended_cli"].(string); ok && hint != "" {
		if _, err := fmt.Fprintf(w, "Recommended: %s\n", hint); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "Keep the session page (/go?session=…) open — navigating it away disconnects the extension."); err != nil {
		return err
	}
	return nil
}

func tabIntField(tab map[string]any, key string) int {
	if tab == nil {
		return 0
	}
	switch v := tab[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func tabStringField(tab map[string]any, key string) string {
	if tab == nil {
		return ""
	}
	if s, ok := tab[key].(string); ok {
		return s
	}
	return ""
}

func tabRoleDisplay(tab map[string]any) string {
	role := tabStringField(tab, "role")
	if role == "" {
		return "user"
	}
	return strings.ReplaceAll(role, "_", "-")
}