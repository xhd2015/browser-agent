package browsertrace

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/xhd2015/browser-agent/browseragent"
)

// ProductName is the asset-cache product key for browser-trace.
const ProductName = "browser-trace"

const assetsHelp = `Usage: browser-trace assets <command>

Manage hydrated browser-trace assets (extension only).

Commands:
  ensure    Download missing extension into the local cache (uses BROWSER_AGENT_ASSET_BASE_URL)
  status    Report embed and cache completeness for extension (no network)

Env:
  XDG_CACHE_HOME                 Cache root isolation (default: ~/.cache)
  BROWSER_AGENT_ASSET_BASE_URL   Download base, e.g. https://…/releases/download

Examples:
  browser-trace assets status
  browser-trace assets ensure
  browser-trace assets --help
`

// HandleCLI dispatches browser-trace CLI args (after the binary name).
// When env != nil, relevant keys are applied for cache/download isolation
// (XDG_CACHE_HOME, BROWSER_AGENT_ASSET_BASE_URL, HOME, USERPROFILE).
func HandleCLI(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if args == nil {
		args = []string{}
	}
	if len(args) == 0 {
		return fmt.Errorf("missing command; try assets, skill, or capture flags")
	}

	cmd := args[0]
	rest := args[1:]
	switch cmd {
	case "assets":
		return cliAssets(rest, env, stdout, stderr)
	case "-h", "--help", "help":
		// Minimal top-level help mention assets for discoverability.
		_, err := io.WriteString(stdout, "Usage: browser-trace [options] | browser-trace assets ensure|status|--help\n")
		return err
	default:
		return fmt.Errorf("unknown command %q; try assets", cmd)
	}
}

func cliAssets(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if len(args) == 0 || hasHelpFlag(args) {
		return writeAssetsHelp(stdout)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "ensure":
		return cliAssetsEnsure(rest, env, stdout, stderr)
	case "status":
		return cliAssetsStatus(rest, env, stdout, stderr)
	case "help", "-h", "--help":
		return writeAssetsHelp(stdout)
	default:
		_, _ = io.WriteString(stderr, assetsHelp)
		if !strings.HasSuffix(assetsHelp, "\n") {
			_, _ = io.WriteString(stderr, "\n")
		}
		return fmt.Errorf("unknown assets subcommand %q; try ensure, status, or --help", sub)
	}
}

func writeAssetsHelp(stdout io.Writer) error {
	_, err := io.WriteString(stdout, assetsHelp)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(assetsHelp, "\n") {
		_, err = io.WriteString(stdout, "\n")
	}
	return err
}

func cliAssetsEnsure(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		return writeAssetsHelp(stdout)
	}
	applyAssetsEnv(env)

	cfg := browseragent.AssetDownloadConfig{
		BaseURL: assetBaseURLFromEnv(env),
	}
	version := browseragent.ClientVersion()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dir, err := browseragent.EnsureAsset(ctx, ProductName, version, browseragent.AssetKindExtension, cfg)
	if err != nil {
		return fmt.Errorf("assets ensure extension: %w", err)
	}
	_, _ = fmt.Fprintf(stdout, "ensured extension -> %s\n", dir)
	return nil
}

func cliAssetsStatus(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		return writeAssetsHelp(stdout)
	}
	applyAssetsEnv(env)

	version := browseragent.ClientVersion()
	if v := strings.TrimSpace(version); v != "" && !strings.HasPrefix(v, "v") {
		version = "v" + v
	}

	embedOK := extensionEmbedComplete()
	cacheOK := browseragent.CacheComplete(ProductName, version, browseragent.AssetKindExtension)
	path := browseragent.AssetCacheDir(ProductName, version, browseragent.AssetKindExtension)

	_, _ = fmt.Fprintf(stdout, "extension:\n")
	_, _ = fmt.Fprintf(stdout, "  embed:  %s\n", completeLabel(embedOK))
	_, _ = fmt.Fprintf(stdout, "  cache:  %s\n", completeLabel(cacheOK))
	_, _ = fmt.Fprintf(stdout, "  path:   %s\n", path)
	return nil
}

func extensionEmbedComplete() bool {
	sub, err := fs.Sub(embeddedExtension, embeddedExtensionRoot)
	if err != nil {
		return false
	}
	return browseragent.EmbedCompleteFS(sub, browseragent.AssetKindExtension)
}

func completeLabel(ok bool) string {
	if ok {
		return "complete (true)"
	}
	return "incomplete (false)"
}

func applyAssetsEnv(env map[string]string) {
	if env == nil {
		return
	}
	for _, key := range []string{"XDG_CACHE_HOME", "HOME", "USERPROFILE", "BROWSER_AGENT_ASSET_BASE_URL"} {
		if v, ok := env[key]; ok {
			_ = os.Setenv(key, v)
		}
	}
}

func assetBaseURLFromEnv(env map[string]string) string {
	if env != nil {
		if v := strings.TrimSpace(env["BROWSER_AGENT_ASSET_BASE_URL"]); v != "" {
			return v
		}
	}
	return strings.TrimSpace(os.Getenv("BROWSER_AGENT_ASSET_BASE_URL"))
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}
