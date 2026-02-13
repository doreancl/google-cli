#!/usr/bin/env bash
set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
  echo "usage: scripts/verify-release.sh X.Y.Z" >&2
  exit 2
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

changelog="CHANGELOG.md"
if ! rg -q "^## (\\[)?${version}(\\])?( - )" "$changelog"; then
  echo "missing changelog section for $version" >&2
  exit 2
fi
if rg -q "^## (\\[)?${version}(\\])? - [Uu]nreleased" "$changelog"; then
  echo "changelog section still Unreleased for $version" >&2
  exit 2
fi

notes_file="$(mktemp -t google-cli-release-notes)"
awk -v ver="$version" '
  $0 ~ "^## \\[" ver "\\] - " || $0 ~ "^## " ver " - " {print "## " ver; in_section=1; next}
  in_section && /^## / {exit}
  in_section {print}
' "$changelog" | sed '/^$/d' > "$notes_file"

if [[ ! -s "$notes_file" ]]; then
  echo "release notes empty for $version" >&2
  exit 2
fi

release_body="$(gh release view "v$version" --json body -q .body)"
if [[ -z "$release_body" ]]; then
  echo "GitHub release notes empty for v$version" >&2
  exit 2
fi

assets_count="$(gh release view "v$version" --json assets -q '.assets | length')"
if [[ "$assets_count" -eq 0 ]]; then
  echo "no GitHub release assets for v$version" >&2
  exit 2
fi

release_run_id="$(gh run list -L 30 --workflow release.yml --json databaseId,conclusion,headBranch -q ".[] | select(.headBranch==\"v$version\") | select(.conclusion==\"success\") | .databaseId" | head -n1)"
if [[ -z "$release_run_id" ]]; then
  echo "release workflow not green for v$version" >&2
  exit 2
fi

ci_ok="$(gh run list -L 1 --workflow ci.yml --branch main --json conclusion -q '.[0].conclusion')"
if [[ "$ci_ok" != "success" ]]; then
  echo "CI not green for main" >&2
  exit 2
fi

make ci

tmp_assets_dir="$(mktemp -d -t google-cli-release-assets)"
if ! gh release download "v$version" -p checksums.txt -D "$tmp_assets_dir" >/dev/null 2>&1; then
  echo "missing checksums.txt asset in release v$version" >&2
  rm -rf "$tmp_assets_dir"
  rm -f "$notes_file"
  exit 2
fi

rm -rf "$tmp_assets_dir"
rm -f "$notes_file"

echo "Release v$version verified (CI, GitHub release notes/assets, checksums)."
