
#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

docker build -f docker/Dockerfile -t zouppp-test .
