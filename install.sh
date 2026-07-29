#!/usr/bin/env bash
# Install browser-agent from GitHub Releases.
# Prefers $GOPATH/bin when that directory exists; otherwise /usr/local/bin.
#
#   curl -fsSL https://raw.githubusercontent.com/xhd2015/browser-agent/master/install.sh | bash
#   INSTALL_TAG=v0.3.1 bash install.sh
set -eo pipefail

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: please install browser-agent using Windows Subsystem for Linux' >&2
    exit 1
fi

error() {
    echo "error: $*" >&2
    exit 1
}

command -v curl >/dev/null || error 'curl is required to install browser-agent'
command -v tar >/dev/null || true # not required for raw binaries; keep soft

case $(uname -ms) in
    'Darwin x86_64')
        target=darwin-amd64
        ;;
    'Darwin arm64')
        target=darwin-arm64
        ;;
    'Linux aarch64' | 'Linux arm64')
        target=linux-arm64
        ;;
    'Linux x86_64' | *)
        target=linux-amd64
        ;;
esac

REPO="https://github.com/xhd2015/browser-agent"
BIN_NAME="browser-agent"

if [[ -n "${INSTALL_TAG:-}" ]]; then
    install_version=${INSTALL_VERSION:-$INSTALL_TAG}
    install_version=${install_version/#v}
    file="${BIN_NAME}-v${install_version}-${target}"
    uri="${REPO}/releases/download/${INSTALL_TAG}/${file}"
else
    latestURL="${REPO}/releases/latest"
    headers=$(curl "$latestURL" -so /dev/null -D -)
    if [[ "$headers" != *302* ]]; then
        error "expect 302 from $latestURL"
    fi

    location=$(echo "$headers" | grep -i "^location: " | tail -1)
    if [[ -z $location ]]; then
        error "expect 302 location from $latestURL"
    fi
    locationURL=${location#*: }
    locationURL=${locationURL//$'\r'/}
    locationURL=${locationURL//$'\n'/}
    locationURL=$(echo "$locationURL" | tr -d '[:space:]')

    versionName=""
    if [[ "$locationURL" = *'/tag/v'* ]]; then
        versionName=${locationURL##*/tag/v}
    elif [[ "$locationURL" = *"/browser-agent-v"* ]]; then
        versionName=${locationURL##*/browser-agent-v}
    fi

    if [[ -z $versionName ]]; then
        error "expect tag format v1.x.x (from $locationURL)"
    fi

    file="${BIN_NAME}-v${versionName}-${target}"
    uri="${latestURL}/download/${file}"
fi

# Resolve install directory: $GOPATH/bin if it exists, else /usr/local/bin.
if [[ -z "${GOPATH:-}" ]] && command -v go >/dev/null 2>&1; then
    GOPATH=$(go env GOPATH 2>/dev/null || true)
fi

dest_dir=""
use_sudo=
if [[ -n "${GOPATH:-}" && -d "${GOPATH}/bin" ]]; then
    dest_dir="${GOPATH}/bin"
    if [[ ! -w "$dest_dir" ]]; then
        error "GOPATH/bin exists but is not writable: $dest_dir"
    fi
else
    dest_dir=/usr/local/bin
    if touch "${dest_dir}/.write_test" 2>/dev/null; then
        rm -f "${dest_dir}/.write_test"
        use_sudo=
    else
        use_sudo=sudo
    fi
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

curl --fail --location --progress-bar --output "${tmp_dir}/${file}" "$uri" \
    || error "failed to download browser-agent from \"$uri\""

chmod +x "${tmp_dir}/${file}"
mv "${tmp_dir}/${file}" "${tmp_dir}/${BIN_NAME}"

echo "installing browser-agent to ${dest_dir}"
if [[ -n "$use_sudo" ]]; then
    if [[ -f "${dest_dir}/${BIN_NAME}" ]]; then
        $use_sudo mv "${dest_dir}/${BIN_NAME}" "${dest_dir}/${BIN_NAME}_backup"
    fi
    $use_sudo install "${tmp_dir}/${BIN_NAME}" "${dest_dir}/${BIN_NAME}"
    $use_sudo rm -f "${dest_dir}/${BIN_NAME}_backup" || true
else
    install "${tmp_dir}/${BIN_NAME}" "${dest_dir}/${BIN_NAME}"
fi

echo "Successfully installed, to get started, run:"
echo "  browser-agent --help"
echo "  browser-agent session new"
