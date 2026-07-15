package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/browser-agent/browsertrace"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/skills/skillcmd"
)

//go:embed SKILL.md
var skillContent string

const skillName = "browser-trace"

const help = `
Usage: browser-trace [options]
       browser-trace skill --show [--header]
       browser-trace skill --install [OPTIONS] [<dir>]
       browser-trace skill --list

Start a local control server, open Chrome in a new window, and record network
traffic from all tabs in that window via Chrome-Ext-Capture-API.

Commands:
  skill --show [--header]    Print embedded agent skill (SKILL.md)
  skill --install [OPTIONS]  Install skill into agent skill directories
  skill --list               Print skill name
  skill --help               Skill subcommand help
  (see: browser-trace skill --install --help)
  assets ensure|status       Hydrate / report extension assets (see: browser-trace assets --help)

Options:
  --addr <host:port>         Listen address (default: 127.0.0.1:43759)
  --base-dir <path>          Session parent directory (default: ~/.tmp/browser-trace)
  --ready-timeout <duration> Wait for extension recording (default: 30s)
  --complete-timeout <duration> Wait for final HAR after stop (default: 30s)
  --no-open-chrome           Do not launch Chrome (for tests / headless use)
  --install-chrome-extension Extract embedded extension and print Load unpacked help
  -v, --verbose              Extra lifecycle detail (hello, start/stop, complete)
  --quiet                    Suppress info progress on stderr (errors still printed)
  --no-log-file              Do not write {sessionDir}/browser-trace.log
  -h, --help                 Show this help

Stop recording with Ctrl-C (CLI) or Stop on the extension popup.
On success, prints the session directory path and writes recording.har + meta.json.
Progress milestones go to stderr by default (and browser-trace.log unless --quiet
or --no-log-file).

Agent workflow is documented in the skill:
  browser-trace skill --show
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) > 0 && args[0] == "skill" {
		return singleSkill().Handle(args[1:])
	}
	// Optional short alias (skill-cli Shape 1): top-level install → skill --install
	if len(args) > 0 && args[0] == "install" {
		return singleSkill().Handle(append([]string{"--install"}, args[1:]...))
	}
	// Asset hydrate: browser-trace assets ensure|status|--help
	if len(args) > 0 && args[0] == "assets" {
		return browsertrace.HandleCLI(args, nil, os.Stdout, os.Stderr)
	}
	return runCapture(args)
}

func singleSkill() *skillcmd.SingleSkill {
	return &skillcmd.SingleSkill{
		Name:        skillName,
		RootContent: skillContent,
		Usage:       "browser-trace skill --install",
	}
}

func runCapture(args []string) error {
	var (
		addr                   string
		baseDir                string
		readyTimeout           time.Duration
		completeTimeout        time.Duration
		noOpenChrome           bool
		installChromeExtension bool
		verbose                bool
		quiet                  bool
		noLogFile              bool
	)

	rest, err := flags.
		String("--addr", &addr).
		String("--base-dir", &baseDir).
		Duration("--ready-timeout", &readyTimeout).
		Duration("--complete-timeout", &completeTimeout).
		Bool("--no-open-chrome", &noOpenChrome).
		Bool("--install-chrome-extension", &installChromeExtension).
		Bool("-v,--verbose", &verbose).
		Bool("--quiet", &quiet).
		Bool("--no-log-file", &noLogFile).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		if err == flags.ErrHelp {
			return nil
		}
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(rest, " "))
	}

	if installChromeExtension {
		if err := browsertrace.InstallChromeExtension(os.Stdout, baseDir); err != nil {
			return fmt.Errorf("install chrome extension: %w", err)
		}
		return nil
	}

	cfg := browsertrace.Config{
		Addr:            addr,
		BaseDir:         baseDir,
		ReadyTimeout:    readyTimeout,
		CompleteTimeout: completeTimeout,
		NoOpenChrome:    noOpenChrome,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
		Verbose:         verbose,
		Quiet:           quiet,
		NoLogFile:       noLogFile,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	result, runErr := browsertrace.Run(ctx, cfg)
	exitCode := 0
	if result != nil {
		exitCode = result.ExitCode
	} else if runErr != nil {
		exitCode = 1
	}
	if runErr != nil && (result == nil || result.Stderr == "") {
		// Run already prints to stderr for most failures; ensure non-zero exit.
		if exitCode == 0 {
			exitCode = 1
		}
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
	return nil
}
