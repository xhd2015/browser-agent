package browseragent

import (
	"context"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
	ghrelease "github.com/xhd2015/dot-pkgs/go-pkgs/git/github/release"
)

const (
	// releaseOwner is the GitHub owner for browser-agent releases.
	releaseOwner = "xhd2015"
	// releaseRepo is the GitHub repo name for browser-agent releases.
	releaseRepo = "browser-agent"
)

// FetchLatestReleaseVersion returns the latest published release version from
// GitHub Releases (e.g. "1.0.14"), stripping the "v" prefix. It follows the
// releases/latest 302 redirect — no API token or rate-limit concern.
// Tests override via inject.FetchLatestVersionFn.
func FetchLatestReleaseVersion(ctx context.Context) (string, error) {
	if inj.FetchLatestVersionFn != nil {
		return inj.FetchLatestVersionFn(ctx)
	}
	return ghrelease.FetchLatestReleaseVersion(ctx, releaseOwner, releaseRepo)
}
