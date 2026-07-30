// Package clicolor implements go-best-practice cli/color for repo scripts:
// --color / --no-color, TTY auto, and NO_COLOR (auto only).
package clicolor

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ColorMode is the three-mode color policy.
type ColorMode int

const (
	// ColorAuto is default: TTY → on unless NO_COLOR non-empty; non-TTY → off.
	ColorAuto ColorMode = iota
	// ColorAlways is --color (force on; ignores TTY and NO_COLOR).
	ColorAlways
	// ColorNever is --no-color (force off).
	ColorNever
)

// Resolve returns whether ANSI escapes should be emitted.
// stdoutIsTTY should come from term.IsTerminal on the writer you print to
// (usually stdout for status; for error-only tools, pass stderr TTY if that
// is the primary human stream — scripts typically use stdout for progress).
func Resolve(mode ColorMode, stdoutIsTTY bool, noColorEnv string) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		if noColorEnv != "" {
			return false
		}
		return stdoutIsTTY
	}
}

// ParseFlags peels only --color / --no-color from args and leaves all other
// tokens in remain (including other flags like --git-add-generated).
//
// Must not use a strict flag parser that rejects unknown flags — color is an
// optional orthogonal layer on top of each command's own flags.
//
// Conflict: both set → error "--color and --no-color cannot be specified together".
// Accepts: --color, --no-color, --color=true|false, --no-color=true|false.
func ParseFlags(args []string) (mode ColorMode, remain []string, err error) {
	var colorFlag, noColorFlag bool
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			remain = append(remain, args[i:]...)
			break
		}
		name, val, hasEq := splitFlag(a)
		switch name {
		case "--color":
			on, ok, perr := parseBoolFlag("color", val, hasEq, args, &i)
			if perr != nil {
				return 0, nil, perr
			}
			if ok {
				colorFlag = on
			} else {
				colorFlag = true // bare --color
			}
		case "--no-color":
			on, ok, perr := parseBoolFlag("no-color", val, hasEq, args, &i)
			if perr != nil {
				return 0, nil, perr
			}
			if ok {
				// --no-color=false means color not forced off via this flag.
				noColorFlag = on
			} else {
				noColorFlag = true // bare --no-color
			}
		default:
			remain = append(remain, a)
		}
	}
	if colorFlag && noColorFlag {
		return 0, nil, fmt.Errorf("--color and --no-color cannot be specified together")
	}
	mode = ColorAuto
	if colorFlag {
		mode = ColorAlways
	}
	if noColorFlag {
		mode = ColorNever
	}
	return mode, remain, nil
}

func splitFlag(a string) (name, val string, hasEq bool) {
	if !strings.HasPrefix(a, "-") {
		return "", "", false
	}
	if i := strings.IndexByte(a, '='); i >= 0 {
		return a[:i], a[i+1:], true
	}
	return a, "", false
}

// parseBoolFlag handles bare --flag (ok=false) or --flag=value / --flag value.
// on is the boolean meaning of the flag when ok is true.
func parseBoolFlag(short string, val string, hasEq bool, args []string, i *int) (on bool, ok bool, err error) {
	if hasEq {
		b, e := parseBoolToken(val)
		if e != nil {
			return false, false, fmt.Errorf("--%s: %w", short, e)
		}
		return b, true, nil
	}
	// Bare --color / --no-color: presence means true (caller sets flag).
	return false, false, nil
}

func parseBoolToken(v string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "yes", "y", "on":
		return true, nil
	case "0", "f", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", v)
	}
}

// FlagHelp is a two-line help snippet for embedding in Usage text.
const FlagHelp = `  --color                 force ANSI color on (even when not a TTY)
  --no-color              force ANSI color off
`

// Style wraps SGR tokens only when enabled.
type Style struct {
	enabled bool
}

// NewStyle builds a Style from mode + real stdout fd + process NO_COLOR.
func NewStyle(mode ColorMode) Style {
	return Style{enabled: Resolve(mode, term.IsTerminal(int(os.Stdout.Fd())), os.Getenv("NO_COLOR"))}
}

// NewStyleForWriter uses writer TTY detection (e.g. os.Stderr for hooks).
func NewStyleForWriter(mode ColorMode, w io.Writer) Style {
	return Style{enabled: Resolve(mode, WriterIsTTY(w), os.Getenv("NO_COLOR"))}
}

// WriterIsTTY reports whether w is a terminal *os.File.
func WriterIsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func (c Style) Enabled() bool { return c.enabled }

func (c Style) wrap(code, s string) string {
	if !c.enabled || s == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// Red colors Error: prefixes and failure tokens.
func (c Style) Red(s string) string { return c.wrap("31", s) }

// Green colors success / pass tokens.
func (c Style) Green(s string) string { return c.wrap("32", s) }

// Yellow colors warning: prefixes.
func (c Style) Yellow(s string) string { return c.wrap("33", s) }

// Gray colors meta labels (paths, counts, dim hints).
func (c Style) Gray(s string) string { return c.wrap("90", s) }

// ErrorLine writes "Error: …\n" with red Error: prefix to w (usually stderr).
func (c Style) ErrorLine(w io.Writer, msg string) {
	msg = strings.TrimRight(msg, "\n")
	_, _ = fmt.Fprintf(w, "%s %s\n", c.Red("Error:"), msg)
}

// WarningLine writes "warning: …\n" with yellow prefix to w.
func (c Style) WarningLine(w io.Writer, msg string) {
	msg = strings.TrimRight(msg, "\n")
	_, _ = fmt.Fprintf(w, "%s %s\n", c.Yellow("warning:"), msg)
}
