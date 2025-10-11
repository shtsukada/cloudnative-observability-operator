#!/usr/bin/env bash
set -euo pipefail

DIR="${1:-dist}"
( cd "${DIR}" && for f in *; do [ -f "$f" ] &&  sha256sum "$f" > "$f.sha256"; done )
echo "Checksums generated under ${DIR}"
