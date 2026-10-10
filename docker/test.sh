#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

docker run --rm -it \
  --privileged \
  -v "$PWD":/src \
  -v zouppp-gocache:/root/.cache/go-build \
  -v zouppp-gomod:/root/go/pkg/mod \
  zouppp-test "$@"
