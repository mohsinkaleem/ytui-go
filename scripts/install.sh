#!/bin/sh
# Install ytui from the GitHub releases.
#
#   curl -fsSL https://raw.githubusercontent.com/mohsinkaleem/ytui-go/main/scripts/install.sh | sh
#
# Environment:
#   YTUI_VERSION      release to install, e.g. v1.0.1 (default: latest)
#   YTUI_INSTALL_DIR  where to put the binary (default: /usr/local/bin if
#                     writable, otherwise ~/.local/bin)

set -eu

REPO="mohsinkaleem/ytui-go"

err() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}

fetch() { # fetch <url> <dest>
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$1" -o "$2"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$2" "$1"
	else
		err "curl or wget is required"
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d ' ' -f 1
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$1" | cut -d ' ' -f 1
	else
		err "sha256sum or shasum is required to verify the download"
	fi
}

case "$(uname -s)" in
Linux) os=linux ;;
Darwin) os=darwin ;;
*) err "unsupported OS $(uname -s); on Windows, download the zip from https://github.com/$REPO/releases" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) err "unsupported architecture $(uname -m)" ;;
esac

if [ -n "${YTUI_VERSION:-}" ]; then
	base="https://github.com/$REPO/releases/download/v${YTUI_VERSION#v}"
else
	base="https://github.com/$REPO/releases/latest/download"
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# checksums.txt lists every archive, so it also tells us the versioned file name.
fetch "$base/checksums.txt" "$tmp/checksums.txt" || err "could not download $base/checksums.txt"
line=$(grep "_${os}_${arch}\.tar\.gz\$" "$tmp/checksums.txt" || true)
[ -n "$line" ] || err "no release archive for $os/$arch"
want=${line%% *}
archive=${line##* }

printf 'Downloading %s\n' "$archive"
fetch "$base/$archive" "$tmp/$archive" || err "could not download $base/$archive"
[ "$(sha256 "$tmp/$archive")" = "$want" ] || err "checksum mismatch for $archive"
tar -xzf "$tmp/$archive" -C "$tmp" ytui

dir=${YTUI_INSTALL_DIR:-}
if [ -z "$dir" ]; then
	if [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
		dir=/usr/local/bin
	else
		dir="$HOME/.local/bin"
	fi
fi
mkdir -p "$dir"
cp "$tmp/ytui" "$dir/ytui"
chmod 755 "$dir/ytui"

printf 'Installed %s to %s\n' "$("$dir/ytui" --version)" "$dir/ytui"

case ":$PATH:" in
*":$dir:"*) ;;
*) printf '\n%s is not on your PATH. Add it with:\n  export PATH="%s:$PATH"\n' "$dir" "$dir" ;;
esac

if ! command -v yt-dlp >/dev/null 2>&1; then
	printf '\nytui needs yt-dlp: https://github.com/yt-dlp/yt-dlp#installation\n'
fi
