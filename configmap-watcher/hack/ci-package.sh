#!/usr/bin/env bash
set -euo pipefail

BASH_ROOT="$(dirname "${BASH_SOURCE[0]}")/.."

cd "$BASH_ROOT"
make package

mkdir -p "/source/out"
cp -a "/source/dist" "/source/out/dist"
cp -a "/source/images" "/source/out/images"
