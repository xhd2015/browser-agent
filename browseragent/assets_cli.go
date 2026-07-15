package browseragent

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"
)

const assetsHelp = `Usage: browser-agent assets <command>

Manage hydrated browser-agent assets (session-page + extension).

Commands:
  ensure    Download missing assets into the local cache (uses BROWSER_AGENT_ASSET_BASE_URL)
  status    Report embed and cache completeness for session-page and extension (no network)

Env:
  XDG_CACHE_HOME                 Cache root isolation (default: ~/.cache)
  BROWSER_AGENT_ASSET_BASE_URL   Download base, e.g. https://…/releases/download

Examples:
  browser-agent assets status
  browser-agent assets ensure
  browser-agent assets --help
`

// cliAssets dispatches: assets ensure|status|--help|help|-h
func cliAssets(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
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
	// Apply env map into process for AssetCacheRoot / EnsureAsset when provided.
	applyAssetsEnv(env)

	cfg := AssetDownloadConfig{
		BaseURL: assetBaseURLFromEnv(env),
	}
	version := ClientVersion()
	product := ProductName
	kinds := []string{AssetKindSessionPage, AssetKindExtension}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, kind := range kinds {
		dir, err := EnsureAsset(ctx, product, version, kind, cfg)
		if err != nil {
			return fmt.Errorf("assets ensure %s: %w", kind, err)
		}
		_, _ = fmt.Fprintf(stdout, "ensured %s -> %s\n", kind, dir)
	}
	return nil
}

func cliAssetsStatus(args []string, env map[string]string, stdout, stderr io.Writer) error {
	if hasHelpFlag(args) {
		return writeAssetsHelp(stdout)
	}
	applyAssetsEnv(env)

	version := normalizeCacheVersion(ClientVersion())
	product := ProductName

	// session-page
	spEmbed := SessionPageEmbedComplete()
	spCache := CacheComplete(product, version, AssetKindSessionPage)
	spPath := AssetCacheDir(product, version, AssetKindSessionPage)
	_, _ = fmt.Fprintf(stdout, "session-page:\n")
	_, _ = fmt.Fprintf(stdout, "  embed:  %s\n", completeLabel(spEmbed))
	_, _ = fmt.Fprintf(stdout, "  cache:  %s\n", completeLabel(spCache))
	_, _ = fmt.Fprintf(stdout, "  path:   %s\n", spPath)

	// extension
	extEmbed := ExtensionEmbedComplete()
	extCache := CacheComplete(product, version, AssetKindExtension)
	extPath := AssetCacheDir(product, version, AssetKindExtension)
	_, _ = fmt.Fprintf(stdout, "extension:\n")
	_, _ = fmt.Fprintf(stdout, "  embed:  %s\n", completeLabel(extEmbed))
	_, _ = fmt.Fprintf(stdout, "  cache:  %s\n", completeLabel(extCache))
	_, _ = fmt.Fprintf(stdout, "  path:   %s\n", extPath)

	// Optional: note live embed roots for operators.
	if sub, err := fs.Sub(embeddedSessionPage, embeddedSessionPageRoot); err == nil && EmbedCompleteFS(sub, AssetKindSessionPage) {
		_, _ = fmt.Fprintf(stdout, "note: live session-page embed is complete\n")
	}
	return nil
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
