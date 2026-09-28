# Releasing ytui

This document covers:
1. Building and uploading binary assets to a GitHub Release
2. Publishing the Homebrew cask to the `mohsinkaleem/homebrew-tap` tap
3. The install script

---

## 1. Releasing binary assets with GoReleaser

[GoReleaser](https://goreleaser.com) is the standard tool for Go projects. It cross-compiles for every platform, creates a GitHub Release, uploads the archives, and can auto-generate the Homebrew formula.

### Install GoReleaser (once)

```sh
brew install goreleaser
```

### What's already set up

- [.goreleaser.yml](.goreleaser.yml) — builds linux/darwin/windows for amd64 and arm64, archives them with `README.md` and `LICENSE`, writes `checksums.txt`, and pushes the Homebrew cask to the tap
- [.github/workflows/release.yml](.github/workflows/release.yml) — runs the tests and GoReleaser when a `v*` tag is pushed
- [.github/workflows/ci.yml](.github/workflows/ci.yml) — build, vet, test (Linux, macOS, Windows), golangci-lint and `goreleaser check` on every push and PR to `main`

### Before tagging

```sh
make fmt vet test
golangci-lint run ./...
goreleaser check
HOMEBREW_TAP_GITHUB_TOKEN= goreleaser release --snapshot --clean   # builds every archive and the cask into dist/ without publishing
```

`GITHUB_TOKEN` is provided automatically by GitHub Actions. Publishing the cask also needs the `HOMEBREW_TAP_GITHUB_TOKEN` secret (see below).

Commit subjects starting with `feat:` and `fix:` are grouped under Features and Bug fixes in the release notes; `docs:`, `test:`, `ci:` and `chore:` commits are left out.

### Tag and trigger a release

```sh
git tag -a v1.0.0 -m "v1.0.0"
git push origin v1.0.0
```

The workflow runs the tests, cross-compiles for all platforms, and publishes the GitHub Release with the archives and `checksums.txt`. Tags with a suffix such as `v1.1.0-rc.1` are marked as pre-releases.

Once the tag is pushed, `go install github.com/mohsinkaleem/ytui-go/cmd/ytui@v1.0.0` also works, and `ytui --version` reports the tag.

---

### Manual release (no CI)

If you want to release from your local machine instead:

```sh
export GITHUB_TOKEN=<your-token>
export HOMEBREW_TAP_GITHUB_TOKEN=<token-with-access-to-the-tap>
goreleaser release --clean
```

---

## 2. Homebrew tap

GoReleaser generates a Homebrew **cask** (`Casks/ytui.rb`, macOS and Linux) and pushes it to [mohsinkaleem/homebrew-tap](https://github.com/mohsinkaleem/homebrew-tap) on every stable release. The cask depends on the `yt-dlp` formula and clears the macOS quarantine flag after install, since the binaries are not notarized. Pre-releases are skipped.

Users install with:

```sh
brew install mohsinkaleem/tap/ytui
```

### One-time setup

1. Create the public repository `mohsinkaleem/homebrew-tap`.
2. Create a [fine-grained personal access token](https://github.com/settings/personal-access-tokens/new) with access to **only** `homebrew-tap` and the **Contents: Read and write** permission.
3. Add it to this repository's secrets as `HOMEBREW_TAP_GITHUB_TOKEN`:
   ```sh
   gh secret set HOMEBREW_TAP_GITHUB_TOKEN -R mohsinkaleem/ytui-go
   ```

When the token expires, the GitHub Release is still published but the cask step fails; renew the token, update the secret and re-run the release job.

---

## 3. Install script

[scripts/install.sh](scripts/install.sh) is served from `main` and always installs the latest release, so it needs no changes per release. It relies on the archive names (`ytui_<version>_<os>_<arch>.tar.gz`) and `checksums.txt`; keep it in sync if `.goreleaser.yml` changes them.

---

## Quick-reference checklist

| Step | Action |
|------|--------|
| 1 | Make sure CI is green on `main` |
| 2 | Run the "Before tagging" checks locally |
| 3 | Push a `v*` tag to trigger the release |
| 4 | Verify the release appears at `github.com/mohsinkaleem/ytui-go/releases` |
| 5 | Verify the cask was pushed to `mohsinkaleem/homebrew-tap` and `brew upgrade ytui` picks it up |
