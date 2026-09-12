package browseragent

import (
	"fmt"
	"io"
	"path"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

const howToRootHelp = `Usage: browser-agent how-to [<topic>] [flags]

Print agent playbooks (procedure text on stdout). Does not open a browser or
touch credentials. Audience: automation agents, not humans.

Topics:
  credentials/export   Firefox cookies → dump → adhoc Chrome setCookie (import included)

With no topic, list available topics.
`

const howToCredentialsExportHelp = `Usage: browser-agent how-to credentials/export [flags]

Print an agent playbook: export Firefox cookies to a JSON dump, then import
into an isolated Chrome session via session new --adhoc-browser-profile
(Network.setCookie). Covers dump and import; no separate import topic.

Flags:
  --host <host>   Example host in the playbook (default: example.com)
  -h, --help      Show this help
`

type howToTopic struct {
	ID          string
	Summary     string
	printPlaybook func(stdout io.Writer, host string) error
}

func howToTopics() []howToTopic {
	return []howToTopic{
		{
			ID:            "credentials/export",
			Summary:       "Firefox cookies → dump → adhoc Chrome setCookie (import included)",
			printPlaybook: printCredentialsExportPlaybook,
		},
	}
}

func findHowToTopic(id string) (howToTopic, bool) {
	id = strings.TrimSpace(strings.TrimPrefix(id, "/"))
	for _, t := range howToTopics() {
		if t.ID == id {
			return t, true
		}
	}
	return howToTopic{}, false
}

func cliHowTo(args []string, env map[string]string, stdout, stderr io.Writer) error {
	_ = env
	_ = stderr
	if stdout == nil {
		stdout = io.Discard
	}

	if len(args) == 0 {
		return printHowToTopicList(stdout)
	}

	topicID := strings.TrimSpace(args[0])
	if topicID == "-h" || topicID == "--help" || topicID == "help" {
		_, err := io.WriteString(stdout, howToRootHelp)
		return err
	}

	rest := args[1:]
	topic, ok := findHowToTopic(topicID)
	if !ok {
		_ = printHowToTopicList(stdout)
		return fmt.Errorf("unknown how-to topic %q; see list above", topicID)
	}

	host := "example.com"
	helpFn := func() {
		txt := strings.TrimPrefix(howToCredentialsExportHelp, "\n")
		_, _ = io.WriteString(stdout, txt)
		if !strings.HasSuffix(txt, "\n") {
			_, _ = io.WriteString(stdout, "\n")
		}
	}
	remaining, err := lessflags.String("--host", &host).
		HelpFunc("-h,--help", helpFn).
		HelpNoExit().
		Parse(rest)
	if err == lessflags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("unrecognized extra args: %v", remaining)
	}

	host = strings.TrimSpace(host)
	if host == "" {
		return fmt.Errorf("--host must be non-empty")
	}
	// Strip scheme/path if an agent passes a URL by mistake.
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	host = strings.TrimPrefix(host, ".")
	if host == "" {
		return fmt.Errorf("--host must be a hostname")
	}

	if topic.printPlaybook == nil {
		return fmt.Errorf("how-to topic %q has no playbook", topic.ID)
	}
	return topic.printPlaybook(stdout, host)
}

func printHowToTopicList(stdout io.Writer) error {
	var b strings.Builder
	b.WriteString("browser-agent how-to topics (agent playbooks)\n\n")
	topics := howToTopics()
	width := 0
	for _, t := range topics {
		if n := len(t.ID); n > width {
			width = n
		}
	}
	for _, t := range topics {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, t.ID, t.Summary)
	}
	b.WriteString("\nRun: browser-agent how-to <topic>\n")
	_, err := io.WriteString(stdout, b.String())
	return err
}

func printCredentialsExportPlaybook(stdout io.Writer, host string) error {
	dump := path.Join("/tmp", "ba-ff-cookies-"+sanitizeHostForFilename(host)+".json")
	domainDot := "." + host

	_, err := fmt.Fprintf(stdout, `How to: Firefox cookies → dump → adhoc Chrome import (agent playbook)

Audience: automation agents.
Goal: reuse a Firefox-logged-in identity in Chrome (HAR / CDP). This topic
covers dump and import — there is no separate credentials/import topic.
Never print cookie values (chat, logs, screenshots, tool stdout). Write only
to a mode-0600 dump file; log names + value lengths only.

Why: Firefox CDP cookies incomplete; page eval often CSP-blocked; HAR is
Chrome-only; curl+Cookie often hits Cloudflare.

1. Locate cookies DB (macOS)
     ls ~/Library/Application\ Support/Firefox/Profiles/*/cookies.sqlite
     Active profile is usually *.default-release.

2. Copy then query (never open the live DB while Firefox holds the lock)
     cp "<profile>/cookies.sqlite" /tmp/ff-cookies-copy.sqlite
     sqlite3 /tmp/ff-cookies-copy.sqlite \
       "SELECT name, host, path, isSecure, isHttpOnly, expiry
        FROM moz_cookies
        WHERE host = '%[1]s' OR host = '%[2]s' OR host LIKE '%%%[1]s'
        ORDER BY host, name;"

3. Write CDP dump (mode 0600) → %[3]s
     Fields per Network.setCookie: name, value, domain, path, secure,
     httpOnly, sameSite, expires, url
     expires: if moz_cookies.expiry > 1e12, use expiry/1000 (ms→s); else seconds
     url: https://%[1]s/
     domain: prefer leading-dot (e.g. %[2]s) when present
     sameSite: "None" when secure; else "Lax" unless the row says otherwise
     Do not commit the dump; do not cat values into the transcript.

4. Import into isolated Chrome
     browser-agent session new --adhoc-browser-profile
     browser-agent session create-tab --session-id <id> 'https://%[1]s/'
     browser-agent session info --session-id <id> --json
     For each cookie in the dump (values stay in file/script, not chat):
       browser-agent session cdp --session-id <id> --tab-id <tab> Network.setCookie '<json>'
     Expect {"success":true} per cookie.

5. Verify then HAR
     browser-agent session cdp --session-id <id> --tab-id <tab> Page.navigate '{"url":"https://%[1]s/"}'
     Confirm logged-in UI (screenshot / innerText). Do not eval document.cookie
     into the transcript if it holds session tokens.
     browser-agent session har start <id>
     # … exercise the flow …
     browser-agent session har end <id>

Pitfalls
  | Wrong | Correct |
  | Firefox Network.getCookies / page eval for auth | copied cookies.sqlite |
  | open live cookies.sqlite while Firefox runs | cp then query the copy |
  | pass Firefox ms expiry straight to CDP | /1000 when expiry > 1e12 |
  | curl Cookie header on Cloudflare sites | import into Chrome |
  | session new (default Chrome) for import | session new --adhoc-browser-profile |
  | print sso / cf_clearance / session values | dump file + name/length logs |

Cleanup: rm -f %[3]s /tmp/ff-cookies-copy.sqlite
Host override: browser-agent how-to credentials/export --host %[1]s
`, host, domainDot, dump)
	return err
}

func sanitizeHostForFilename(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	var b strings.Builder
	for _, r := range host {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == '.':
			b.WriteByte('-')
		}
	}
	out := b.String()
	if out == "" {
		return "host"
	}
	return out
}
