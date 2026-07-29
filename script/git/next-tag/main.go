// next-tag keeps root VERSION.txt aligned with git release tags.
//
// VERSION.txt is the source of truth for the next release tag: tag = "v" + version
// (e.g. VERSION 0.3.1 → git tag v0.3.1).
//
// Usage (module root):
//
//	go run ./script/git/next-tag              # status: need bump? next tag?
//	go run ./script/git/next-tag --apply      # write patch+1 to VERSION.txt when needed
//	go run ./script/git/next-tag --check      # exit 1 if VERSION is already a git tag
//	go run ./script/git/next-tag --print-tag   # print only vX.Y.Z for current VERSION
//
// Bump rule:
//   - If VERSION.txt already exists as a git tag (released), next version is patch+1
//     (MAJOR.MINOR.PATCH → MAJOR.MINOR.PATCH+1).
//   - Minor / major bumps are manual: edit VERSION.txt yourself, then release.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	var apply, check, printTag bool
	for _, a := range args {
		switch a {
		case "-h", "--help":
			printHelp()
			return nil
		case "--apply":
			apply = true
		case "--check":
			check = true
		case "--print-tag":
			printTag = true
		default:
			if strings.HasPrefix(a, "-") {
				return fmt.Errorf("unrecognized flag %s (try --help)", a)
			}
			return fmt.Errorf("unexpected argument %q (try --help)", a)
		}
	}
	if apply && check {
		return fmt.Errorf("cannot combine --apply and --check")
	}
	if printTag && (apply || check) {
		return fmt.Errorf("cannot combine --print-tag with --apply or --check")
	}

	root, err := findModuleRoot()
	if err != nil {
		return err
	}

	ver, err := readRootVersion(root)
	if err != nil {
		return err
	}
	cur, err := parseSemver(ver)
	if err != nil {
		return fmt.Errorf("VERSION.txt: %w", err)
	}

	if printTag {
		fmt.Println(gitTagFor(ver))
		return nil
	}

	tags, err := listVersionTags(root)
	if err != nil {
		return err
	}
	latest := latestSemverTag(tags)
	tagged := tagExists(tags, ver)
	needsBump := tagged

	next := cur
	if needsBump {
		next = cur.bumpPatch()
	}

	// VERSION behind latest tag is a problem (SoT drifted).
	if latest != nil && cur.cmp(*latest) < 0 {
		return fmt.Errorf(
			"VERSION.txt %s is behind latest git tag %s — edit VERSION.txt (manual minor/major or set to %s) then re-run",
			ver, gitTagFor(latest.String()), latest.bumpPatch().String(),
		)
	}

	if check {
		if needsBump {
			return fmt.Errorf(
				"VERSION.txt %s is already a git tag (%s); bump needed before release (run: go run ./script/git/next-tag --apply for patch+1 → %s)",
				ver, gitTagFor(ver), next.String(),
			)
		}
		fmt.Printf("ok: VERSION.txt %s is not tagged yet; release tag would be %s\n", ver, gitTagFor(ver))
		return nil
	}

	// Status (default and after --apply planning).
	fmt.Printf("VERSION.txt:     %s\n", ver)
	fmt.Printf("git tag (SoT):   %s\n", gitTagFor(ver))
	if latest != nil {
		fmt.Printf("latest git tag:  %s\n", gitTagFor(latest.String()))
	} else {
		fmt.Printf("latest git tag:  (none)\n")
	}

	if !needsBump {
		fmt.Printf("status:          ready to release (VERSION not yet a git tag)\n")
		fmt.Printf("next action:     git tag %s && go run ./script/github/release\n", gitTagFor(ver))
		fmt.Println("note:            minor/major bumps: edit VERSION.txt manually, then tag/release")
		return nil
	}

	fmt.Printf("status:          VERSION already tagged — needs bump\n")
	fmt.Printf("auto bump rule:  patch+1 only → %s (tag %s)\n", next.String(), gitTagFor(next.String()))
	fmt.Println("note:            for minor/major, edit VERSION.txt manually (do not use --apply)")

	if !apply {
		fmt.Println("hint:            run with --apply to write patch+1 into VERSION.txt")
		return nil
	}

	if err := writeRootVersion(root, next.String()); err != nil {
		return err
	}
	fmt.Printf("wrote VERSION.txt → %s\n", next.String())
	fmt.Printf("next action:     go run ./script/generate && git add VERSION.txt && git commit …\n")
	fmt.Printf("                 git tag %s && go run ./script/github/release\n", gitTagFor(next.String()))
	return nil
}

func printHelp() {
	fmt.Print(`Usage: go run ./script/git/next-tag [options]

Keep root VERSION.txt as the source of truth for the next git release tag
(tag name = "v" + VERSION, e.g. 0.3.1 → v0.3.1).

If VERSION.txt already exists as a git tag, the automatic next version is
patch+1 (X.Y.Z → X.Y.Z+1). Bumping minor or major is manual: edit VERSION.txt.

Options:
  --apply       When a bump is needed, write patch+1 to VERSION.txt
  --check       Exit 1 if VERSION is already tagged (CI / pre-release gate)
  --print-tag   Print only the git tag for current VERSION (vX.Y.Z) and exit
  -h, --help    Show this help

Examples:
  go run ./script/git/next-tag
  go run ./script/git/next-tag --apply
  go run ./script/git/next-tag --check
  go run ./script/git/next-tag --print-tag
`)
}

type semver struct {
	major, minor, patch int
}

func parseSemver(s string) (semver, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("want MAJOR.MINOR.PATCH, got %q", s)
	}
	var nums [3]int
	for i, p := range parts {
		if p == "" || strings.TrimSpace(p) != p {
			return semver{}, fmt.Errorf("invalid semver segment %q in %q", p, s)
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return semver{}, fmt.Errorf("invalid semver segment %q in %q", p, s)
		}
		// Disallow leading zeros like 01 (except single 0).
		if len(p) > 1 && p[0] == '0' {
			return semver{}, fmt.Errorf("invalid leading zero in %q", s)
		}
		nums[i] = n
	}
	return semver{major: nums[0], minor: nums[1], patch: nums[2]}, nil
}

func (v semver) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v semver) bumpPatch() semver {
	return semver{major: v.major, minor: v.minor, patch: v.patch + 1}
}

// cmp returns -1 if v < o, 0 if equal, 1 if v > o.
func (v semver) cmp(o semver) int {
	if v.major != o.major {
		if v.major < o.major {
			return -1
		}
		return 1
	}
	if v.minor != o.minor {
		if v.minor < o.minor {
			return -1
		}
		return 1
	}
	if v.patch != o.patch {
		if v.patch < o.patch {
			return -1
		}
		return 1
	}
	return 0
}

func gitTagFor(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	return "v" + version
}

func readRootVersion(root string) (string, error) {
	path := filepath.Join(root, "VERSION.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read VERSION.txt: %w", err)
	}
	ver := strings.TrimSpace(string(b))
	ver = strings.TrimPrefix(ver, "v")
	ver = strings.TrimPrefix(ver, "V")
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return "", fmt.Errorf("VERSION.txt is empty")
	}
	if strings.ContainsAny(ver, "/\\ \t\n\r") {
		return "", fmt.Errorf("VERSION.txt has invalid version %q", ver)
	}
	return ver, nil
}

func writeRootVersion(root, ver string) error {
	path := filepath.Join(root, "VERSION.txt")
	return os.WriteFile(path, []byte(ver+"\n"), 0o644)
}

func listVersionTags(root string) ([]semver, error) {
	cmd := exec.Command("git", "tag", "-l", "v*")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git tag -l: %w", err)
	}
	var tags []semver
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Only strict vX.Y.Z (ignore pre-release / weird tags).
		sv, err := parseSemver(line)
		if err != nil {
			continue
		}
		tags = append(tags, sv)
	}
	return tags, nil
}

func tagExists(tags []semver, version string) bool {
	want, err := parseSemver(version)
	if err != nil {
		return false
	}
	for _, t := range tags {
		if t.cmp(want) == 0 {
			return true
		}
	}
	return false
}

func latestSemverTag(tags []semver) *semver {
	if len(tags) == 0 {
		return nil
	}
	best := tags[0]
	for _, t := range tags[1:] {
		if t.cmp(best) > 0 {
			best = t
		}
	}
	return &best
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "VERSION.txt")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("VERSION.txt + go.mod not found from %s", wd)
		}
		dir = parent
	}
}
