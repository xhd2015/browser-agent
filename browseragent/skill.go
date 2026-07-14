package browseragent

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/skills/skillcmd"
)

//go:embed SKILL.md
var skillContent string

const skillName = ProductName

// browserAgentSkill returns the Shape-1 single skill definition.
func browserAgentSkill() *skillcmd.SingleSkill {
	return &skillcmd.SingleSkill{
		Name:        skillName,
		RootContent: skillContent,
		Usage:       "browser-agent skill --install",
	}
}

// cliSkill handles: skill [--list|--show|--install …]
// Writes to the provided stdout/stderr (does not rely on os.Stdout alone).
func cliSkill(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	sk := browserAgentSkill()
	parsed, err := skillcmd.ParseSkillArgs(args)
	if err != nil {
		// skillcmd bare / missing-action error already mentions --show/--list/--install.
		return err
	}

	switch parsed.Action {
	case skillcmd.ActionHelp:
		help := strings.TrimSpace(sk.Help)
		if help == "" {
			help = skillcmd.DefaultSingleSkillHelp(sk.Usage, sk.Name)
		}
		if !strings.HasSuffix(help, "\n") {
			help += "\n"
		}
		_, _ = io.WriteString(stdout, help)
		return nil

	case skillcmd.ActionList:
		// Shape 1: skill name + trailing newline.
		_, err := fmt.Fprintln(stdout, sk.Name)
		return err

	case skillcmd.ActionShow:
		content, err := loadSkillContent(sk, parsed.Header, parsed.Rest)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		_, err = io.WriteString(stdout, content)
		return err

	case skillcmd.ActionInstall:
		// Install uses skillcmd's HandleInstall (may write to process stdout for
		// install progress). Not asserted by vite-skill tests; still support it.
		return sk.Handle(append([]string{"--install"}, parsed.Rest...))

	default:
		return fmt.Errorf("unknown skill action %q", parsed.Action)
	}
}

func loadSkillContent(sk *skillcmd.SingleSkill, header bool, rest []string) (string, error) {
	// Reuse SingleSkill show path logic without printing to os.Stdout.
	// Only root content is required for Shape 1 tests.
	if len(rest) == 0 {
		content := sk.RootContent
		if header {
			out, err := skillcmd.FormatHeaderWithDelimiters(content)
			if err != nil {
				return "", err
			}
			return out, nil
		}
		return content, nil
	}
	// Nested topics: fall back to SingleSkill.Handle via stdout redirect is
	// awkward; return a clear error when TreeFS is unset (default).
	if sk.TreeFS == nil {
		return "", fmt.Errorf("unknown topic path: %s", strings.Trim(rest[0], "/"))
	}
	// For multi-topic skills, delegate content load through Handle show.
	// Not used by browser-agent Shape 1.
	return "", fmt.Errorf("topic paths not supported via package skill load")
}
