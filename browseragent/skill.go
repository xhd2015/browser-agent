package browseragent

import (
	_ "embed"
	"fmt"
	"io"
	"strings"

	"github.com/xhd2015/skills/skillcmd"
)

//go:embed SKILL.md
var browserAgentSkillContent string

//go:embed skills/browser-agent-to-api/SKILL.md
var browserAgentToAPISkillContent string

const (
	skillName             = ProductName
	browserAgentToAPIName = "browser-agent-to-api"
)

const browserAgentSkillHelp = `Usage: browser-agent skill --list
       browser-agent skill --show [--header] [<name>]
       browser-agent skill [<name>] --show [--header]
       browser-agent skill --install [<name>] [OPTIONS] [<dir>]
       browser-agent skill [<name>] --install [OPTIONS] [<dir>]

Embedded skills:
  browser-agent         Drive a live browser session
  browser-agent-to-api  Capture browser behavior and derive reusable API calls

Omitting <name> defaults to browser-agent for backward compatibility.
Run 'browser-agent skill --install [<name>] --help' for install target flags.
`

func browserAgentSkill() *skillcmd.SingleSkill {
	return &skillcmd.SingleSkill{
		Name:        skillName,
		RootContent: browserAgentSkillContent,
		Usage:       "browser-agent skill --install",
	}
}

func browserAgentToAPISkill() *skillcmd.SingleSkill {
	return &skillcmd.SingleSkill{
		Name:        browserAgentToAPIName,
		RootContent: browserAgentToAPISkillContent,
		Usage:       "browser-agent skill --install browser-agent-to-api",
	}
}

func browserAgentSkills() []*skillcmd.SingleSkill {
	return []*skillcmd.SingleSkill{browserAgentSkill(), browserAgentToAPISkill()}
}

func findBrowserAgentSkill(name string) (*skillcmd.SingleSkill, bool) {
	for _, skill := range browserAgentSkills() {
		if skill.Name == name {
			return skill, true
		}
	}
	return nil, false
}

// cliSkill handles: skill [--list|--show|--install …]
// Writes list/show/help to the provided writers; install uses skillcmd's installer.
func cliSkill(args []string, env map[string]string, stdout, stderr io.Writer) error {
	_ = env
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	parsed, err := skillcmd.ParseSkillArgs(args)
	if err != nil {
		return err
	}

	switch parsed.Action {
	case skillcmd.ActionHelp:
		_, err := io.WriteString(stdout, browserAgentSkillHelp)
		return err

	case skillcmd.ActionList:
		for _, skill := range browserAgentSkills() {
			if _, err := fmt.Fprintln(stdout, skill.Name); err != nil {
				return err
			}
		}
		return nil

	case skillcmd.ActionShow:
		skill := browserAgentSkill()
		if len(parsed.Rest) > 1 {
			return fmt.Errorf("unexpected arguments: %v", parsed.Rest[1:])
		}
		if len(parsed.Rest) == 1 {
			var ok bool
			skill, ok = findBrowserAgentSkill(parsed.Rest[0])
			if !ok {
				return fmt.Errorf("unknown skill %q", parsed.Rest[0])
			}
		}
		content, err := loadSkillContent(skill, parsed.Header, nil)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		_, err = io.WriteString(stdout, content)
		return err

	case skillcmd.ActionInstall:
		skill := browserAgentSkill()
		rest := make([]string, 0, len(parsed.Rest))
		selectedName := ""
		for _, arg := range parsed.Rest {
			if selected, ok := findBrowserAgentSkill(arg); ok {
				if selectedName != "" && selectedName != selected.Name {
					return fmt.Errorf("multiple skill names provided: %s and %s", selectedName, selected.Name)
				}
				selectedName = selected.Name
				skill = selected
				continue
			}
			rest = append(rest, arg)
		}
		return skill.Handle(append([]string{"--install"}, rest...))

	default:
		return fmt.Errorf("unknown skill action %q", parsed.Action)
	}
}

func loadSkillContent(skill *skillcmd.SingleSkill, header bool, rest []string) (string, error) {
	if len(rest) == 0 {
		content := skill.RootContent
		if header {
			out, err := skillcmd.FormatHeaderWithDelimiters(content)
			if err != nil {
				return "", err
			}
			return out, nil
		}
		return content, nil
	}
	if skill.TreeFS == nil {
		return "", fmt.Errorf("unknown topic path: %s", strings.Trim(rest[0], "/"))
	}
	return "", fmt.Errorf("topic paths not supported via package skill load")
}
