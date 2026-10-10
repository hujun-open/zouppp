#!/usr/bin/env bash
set -euo pipefail

# --- mirror the CI setup steps ---
go mod download                       # CI: "download go dep"

rm -f /etc/ppp/options                # CI: remove stock options
mkdir -p /etc/ppp/ipv6-up.d /var/log /var/lib/kea /run/kea
cp -f /src/testdata/pppsvrconf/* /etc/ppp/   # CI: copy test pppd files

# --- run tests exactly like CI (root == the CI "sudo") ---
exec env "PATH=$PATH" go test -failfast -p 1 -v "${@:-./...}"
