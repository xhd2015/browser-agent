package inject

// ClientVersionOverride, when non-nil, replaces embedded ClientVersion() for tests.
var ClientVersionOverride func() string