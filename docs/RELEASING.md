---
summary: "Release checklist for google-cli (GitHub release + GoReleaser artifacts)"
---

# Releasing `google-cli`

This document defines the release flow for this repo.

Always do all steps below (CI + changelog + tag + GitHub release artifacts + verification). No partial releases.

Shortcut scripts (preferred):
```sh
scripts/release.sh X.Y.Z
scripts/verify-release.sh X.Y.Z
```

## 0) Prereqs
- Clean working tree on `main`.
- Go toolchain installed (Go version comes from `go.mod`).
- `make` works locally.
- `gh` CLI authenticated with repo write access.

## 1) Verify build is green
```sh
make ci
```

Confirm GitHub Actions `ci` is green for the commit you are tagging:
```sh
gh run list -L 5 --workflow ci.yml --branch main
```

## 2) Update changelog
- Update `CHANGELOG.md` with a version heading for the release.
- Supported headings in automation scripts:
  - `## [X.Y.Z] - YYYY-MM-DD`
  - `## X.Y.Z - YYYY-MM-DD`
- Release section must not be `Unreleased`.

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
The tag push triggers `.github/workflows/release.yml` (GoReleaser). Ensure it completes and the release has assets.

```sh
gh run list -L 5 --workflow release.yml
gh release view vX.Y.Z
```

## 5) Run automated verification
```sh
scripts/verify-release.sh X.Y.Z
```

This verifies:
- release notes are present,
- release assets exist,
- `release.yml` is green for the tag,
- latest `ci.yml` on `main` is green,
- `checksums.txt` is attached.

## Notes
- Artifacts are built from `.goreleaser.yaml`.
- Current binary name is `dorean_g`.
