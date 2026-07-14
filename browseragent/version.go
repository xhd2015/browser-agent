package browseragent

import (
	_ "embed"
	"strconv"
	"strings"

	inj "github.com/xhd2015/browser-agent/browseragent/inject"
)

//go:embed VERSION.txt
var embeddedVersion string

const (
	// DefaultControlHost is the default control plane bind host.
	DefaultControlHost = "127.0.0.1"
	// DefaultControlPort is the default control plane TCP port.
	DefaultControlPort = 43761
)

// DefaultAddr is the product control listen address (host:port).
const DefaultAddr = DefaultControlHost + ":43761"

// ClientVersion returns the embedded CLI/daemon version from VERSION.txt.
func ClientVersion() string {
	if inj.ClientVersionOverride != nil {
		if v := strings.TrimSpace(inj.ClientVersionOverride()); v != "" {
			return v
		}
	}
	return strings.TrimSpace(embeddedVersion)
}

// EffectiveDaemonVersion normalizes a daemon version string; empty → "0.0.0".
func EffectiveDaemonVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "0.0.0"
	}
	return v
}

// CompareVersion compares loose semver tuples major.minor.patch. Pre-release suffixes
// are ignored for numeric tuple parsing; when numeric tuples tie, a pre-release orders
// below the corresponding release (0.2.0-beta < 0.2.0). Returns -1/0/+1.
func CompareVersion(a, b string) int {
	av := parseVersionTuple(a)
	bv := parseVersionTuple(b)
	for i := 0; i < 3; i++ {
		if av[i] < bv[i] {
			return -1
		}
		if av[i] > bv[i] {
			return 1
		}
	}
	apre := hasPrereleaseSuffix(a)
	bpre := hasPrereleaseSuffix(b)
	if apre && !bpre {
		return -1
	}
	if !apre && bpre {
		return 1
	}
	return 0
}

func hasPrereleaseSuffix(v string) bool {
	v = strings.TrimSpace(v)
	i := strings.IndexAny(v, "-+")
	return i >= 0
}

func parseVersionTuple(v string) [3]int {
	v = strings.TrimSpace(v)
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	var out [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		p := strings.TrimSpace(parts[i])
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		out[i] = n
	}
	return out
}

// DefaultControlPortString returns DefaultControlPort as a decimal string.
func DefaultControlPortString() string {
	return strconv.Itoa(DefaultControlPort)
}