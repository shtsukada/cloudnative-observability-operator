#!/usr/bin/env bash
set -euo pipefail

CHART_DIR="${1:-charts/cloudnative-observability-operator}"
CHART_YAML="${CHART_DIR}/Chart.yaml"

if [[ ! -f "$CHART_YAML" ]]; then
  echo "::error ::Chart.yaml not found at ${CHART_YAML}"
  exit 1
fi

# yq v4 をGo経由で導入（CIでは Go が入っている前提）
if ! command -v yq >/dev/null 2>&1; then
  echo "Installing yq..."
  go install github.com/mikefarah/yq/v4@latest
  export PATH="${PATH}:$(go env GOPATH)/bin"
fi

version="$(yq '.version' "${CHART_YAML}" | tr -d '"')"
appVersion="$(yq '.appVersion' "${CHART_YAML}" | tr -d '"')"

echo "Chart version=${version} / appVersion=${appVersion}"

if [[ -z "${version}" || -z "${appVersion}" ]]; then
  echo "::error ::version or appVersion is empty."
  exit 1
fi

if [[ "${version}" != "${appVersion}" ]]; then
  echo "::error ::Chart.yaml 'version' must equal 'appVersion' (contract)."
  exit 1
fi

echo "OK: Chart version == appVersion"
