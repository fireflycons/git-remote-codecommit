#!/usr/bin/env bash
# Increment the semantic version in VERSION: bump-version.sh patch|minor|major
set -euo pipefail

file="$(dirname "$0")/../VERSION"
current="$(tr -d '[:space:]' < "$file")"

if [[ ! "$current" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "VERSION contains '$current', which is not MAJOR.MINOR.PATCH" >&2
  exit 1
fi
major="${BASH_REMATCH[1]}" minor="${BASH_REMATCH[2]}" patch="${BASH_REMATCH[3]}"

case "${1:-}" in
  patch) patch=$((patch + 1)) ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  major) major=$((major + 1)); minor=0; patch=0 ;;
  *) echo "usage: $0 patch|minor|major" >&2; exit 2 ;;
esac

next="$major.$minor.$patch"
printf '%s\n' "$next" > "$file"
echo "$current -> $next"
