#!/usr/bin/env bash
set -euo pipefail

TAG="${1:?tag like v1.2.3}"
CHART_DIR="${2:?charts/... path}"
CHART_FILE="${CHART_DIR}/Chart.yaml"

VERSION="${TAG#v}"

tmp="$(mktemp)"
awk -v ver="$VERSION" '
  $1=="version:" { print "version: " ver; next }
  $1=="appVersion:" { print "appVersion: " ver; next }
  { print }
' "$CHART_FILE" > "$tmp"
mv "$tmp" "$CHART_FILE"

echo "Chart.yaml updated: version/appVersion -> ${VERSION}"
