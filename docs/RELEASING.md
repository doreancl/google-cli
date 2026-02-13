---
summary: "Release checklist for google-cli (GitHub release + Homebrew tap)"
---

# Releasing `google-cli`

This document defines the real release flow for this repo.

Always do **all** steps below (CI + changelog + tag + GitHub release artifacts + tap update + Homebrew sanity install). No partial releases.

Shortcut scripts (preferred, keep notes non-empty):
```sh
scripts/release.sh X.Y.Z
scripts/verify-release.sh X.Y.Z
```

## 0) Prereqs
- Clean working tree on `main`.
- Go toolchain installed (Go version comes from `go.mod`).
- `make` works locally.

## 1) Verify build is green
```sh
make ci
```

Confirm GitHub Actions `ci` is green for the commit you’re tagging:
```sh
gh run list -L 5 --branch main
```

## 2) Update changelog
- Update `CHANGELOG.md` for the version you’re releasing.

Example heading:
- `## 0.1.0 - 2025-12-12`

## 3) Commit, tag & push
```sh
git checkout main
git pull

# commit changelog + any release tweaks
git commit -am "release: vX.Y.Z"

git tag -a vX.Y.Z -m "Release X.Y.Z"
git push origin main --tags
```

## 4) Verify GitHub release artifacts
The tag push triggers `.github/workflows/release.yml` (GoReleaser). Ensure it completes successfully and the release has assets.

```sh
gh run list -L 5 --workflow release.yml
gh release view vX.Y.Z
```

Ensure GitHub release notes are not empty (mirror the changelog section).

If the workflow needs a rerun:
```sh
gh workflow run release.yml -f tag=vX.Y.Z
```

## 6) Sanity-check install from tap
```sh
brew update
brew uninstall gogcli || true
brew untap steipete/tap || true
brew tap steipete/tap
brew install steipete/tap/gogcli
brew test steipete/tap/gogcli

gog --help
```
## Notes
- `gog` currently does not print a version string; use tags + changelog as the source of truth.
- If you later add `gog version`, update this doc to validate `gog version` post-install.
