package browseragent

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// serveColor applies optional ANSI styling to serve operator stderr messages.
type serveColor struct {
	enabled bool
}

func newServeColor(stderr io.Writer, env map[string]string, forceColor, noColor bool) serveColor {
	if noColor {
		return serveColor{enabled: false}
	}
	if forceColor {
		return serveColor{enabled: true}
	}
	if noColorEnvSet(env) {
		return serveColor{enabled: false}
	}
	return serveColor{enabled: stderrIsTTY(stderr)}
}

func noColorEnvSet(env map[string]string) bool {
	if env != nil {
		v, ok := env["NO_COLOR"]
		if !ok {
			return false
		}
		return strings.TrimSpace(v) != ""
	}
	v, ok := os.LookupEnv("NO_COLOR")
	return ok && strings.TrimSpace(v) != ""
}

func stderrIsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func (c serveColor) wrap(code, s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (c serveColor) yellow(s string) string  { return c.wrap("33", s) }
func (c serveColor) red(s string) string     { return c.wrap("31", s) }
func (c serveColor) green(s string) string   { return c.wrap("32", s) }
func (c serveColor) gray(s string) string    { return c.wrap("90", s) }
func (c serveColor) orange(s string) string  { return c.wrap("38;5;208", s) }

func (c serveColor) writeWarning(w io.Writer, msg string) {
	prefix := c.yellow("warning:")
	_, _ = fmt.Fprintf(w, "%s %s\n", prefix, msg)
}

func (c serveColor) writeError(w io.Writer, msg string) {
	prefix := c.red("Error:")
	_, _ = fmt.Fprintf(w, "%s %s\n", prefix, msg)
}

func (c serveColor) writeStopped(w io.Writer, pid int, baseDir string) {
	head := c.green("stopped daemon")
	meta := c.gray(fmt.Sprintf("(pid %d) in %s — shutdown/kill complete", pid, baseDir))
	_, _ = fmt.Fprintf(w, "%s %s\n", head, meta)
}