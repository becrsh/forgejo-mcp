#!/usr/bin/env bash
# Builds our forgejo-mcp image without Docker, as command-center and becrsh/agents build theirs: a static Go binary in
# one layer on a digest-pinned distroless base, put together with crane. Not upstream's Containerfile: its builder
# installs unpinned packages and its runtime image has a shell and a package manager (bcc/README.md). Works on the
# Forgejo runner and on a Mac (`brew install gnu-tar crane`). Needs Go, crane and GNU tar.
#
#   bcc/build-image.sh <image-ref>        push to a registry
#   bcc/build-image.sh -o <file.tar>      write the image to a tarball instead
set -euo pipefail

RUNTIME_USER=65532
# gcr.io/distroless/static-debian12:nonroot, linux/amd64: CA certificates, user 65532, no shell (as command-center)
BASE=gcr.io/distroless/static-debian12@sha256:52dcfbabb7457ea47c82f6e13af8c8a4a1d9f7b0145142b3ecab20f2b888411d

usage() { echo "usage: $0 <image-ref> | $0 -o <file.tar>" >&2; exit 2; }
case "${1:-}" in
  -o) [ $# -eq 2 ] || usage; output=(-o "$2" -t localhost/forgejo-mcp:dev) ;;
  "") usage ;;
  *) [ $# -eq 1 ] || usage; output=(-t "$1") ;;
esac

repo=$(cd "$(dirname "$0")/.." && pwd)
tar=$(command -v gtar || command -v tar)
"$tar" --version 2>/dev/null | grep -q 'GNU tar' || { echo "needs GNU tar (gtar on macOS)" >&2; exit 1; }
revision=$(git -C "$repo" rev-parse --verify -q HEAD || echo unknown)
version=${VERSION:-dev}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/root/opt/app"

echo "Binary"
(cd "$repo" && go mod verify && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -buildvcs=false -ldflags="-s -w -buildid= -X main.Version=$version" \
  -o "$work/root/opt/app/forgejo-mcp" .)

echo "Layer"
chmod -R u=rwX,go=rX "$work/root"
"$tar" -C "$work/root" -cf "$work/layer.tar" --sort=name --mtime=@0 --numeric-owner --format=gnu \
  --owner=0 --group=0 opt

echo "Image"
crane mutate "$BASE" --append "$work/layer.tar" \
  --entrypoint /opt/app/forgejo-mcp \
  --workdir /opt/app \
  --user "$RUNTIME_USER:$RUNTIME_USER" \
  -l org.opencontainers.image.source=https://github.com/becrsh/forgejo-mcp \
  -l "org.opencontainers.image.revision=$revision" \
  -l "org.opencontainers.image.version=$version" \
  "${output[@]}"
