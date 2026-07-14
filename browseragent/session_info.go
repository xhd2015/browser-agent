package browseragent

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// FormatSessionInfoOptions controls human session info rendering.
type FormatSessionInfoOptions struct {
	Color serveColor
}

// FormatSessionInfo writes human-readable session info sections to w.
func FormatSessionInfo(w io.Writer, snap sessionSnapshot, browser map[string]any, opts FormatSessionInfoOptions) error {
	if w == nil {
		w = io.Discard
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
		if tabs, ok := browser["tabs"].([]any); ok && len(tabs) > 0 {
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