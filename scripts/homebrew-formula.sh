#!/usr/bin/env bash
# Print the Homebrew formula for a release: homebrew-formula.sh <version> <SHA256SUMS>
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "usage: $0 <version> <SHA256SUMS>" >&2
  exit 2
fi
version="$1" sums="$2"
base="https://github.com/fireflycons/git-remote-codecommit/releases/download/v$version"

sha() {
  local file="git-remote-codecommit_${version}_$1.tar.gz" hash
  hash="$(awk -v f="$file" '$2 == f || $2 == "*" f { print $1 }' "$sums")"
  if [ -z "$hash" ]; then
    echo "no SHA256 for $file in $sums" >&2
    exit 1
  fi
  echo "$hash"
}

darwin_arm64="$(sha darwin_arm64)"
darwin_amd64="$(sha darwin_amd64)"
linux_arm64="$(sha linux_arm64)"
linux_amd64="$(sha linux_amd64)"

cat <<EOF
class GitRemoteCodecommit < Formula
  desc "Git remote helper for AWS CodeCommit using AWS credentials"
  homepage "https://github.com/fireflycons/git-remote-codecommit"
  version "$version"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "$base/git-remote-codecommit_${version}_darwin_arm64.tar.gz"
      sha256 "$darwin_arm64"
    end
    on_intel do
      url "$base/git-remote-codecommit_${version}_darwin_amd64.tar.gz"
      sha256 "$darwin_amd64"
    end
  end

  on_linux do
    on_arm do
      url "$base/git-remote-codecommit_${version}_linux_arm64.tar.gz"
      sha256 "$linux_arm64"
    end
    on_intel do
      url "$base/git-remote-codecommit_${version}_linux_amd64.tar.gz"
      sha256 "$linux_amd64"
    end
  end

  def install
    bin.install "git-remote-codecommit"
  end

  test do
    assert_match "malformed", shell_output("#{bin}/git-remote-codecommit origin not-a-url 2>&1", 1)
  end
end
EOF
