#!/usr/bin/env bash
set -euo pipefail

BASH_ROOT="$(dirname "${BASH_SOURCE[0]}")/.."

cd "$BASH_ROOT"

apt update
apt install -y gettext-base build-essential jq

mkdir -p hack/tools/bin

PROJ="$1" make restore
