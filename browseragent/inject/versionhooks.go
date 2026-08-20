package inject

import "context"

// ClientVersionOverride, when non-nil, replaces embedded ClientVersion() for tests.
var ClientVersionOverride func() string

// FetchLatestVersionFn, when non-nil, replaces FetchLatestReleaseVersion() for tests.
// Returns the latest version string (e.g. "1.0.14") without the "v" prefix.
var FetchLatestVersionFn func(ctx context.Context) (string, error)