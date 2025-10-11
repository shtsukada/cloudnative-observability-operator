#!/usr/bin/env bash
set -euo pipefail

CHART_DIR="${1:?charts/... path}"
CHART_REPO="${2:?ghcr.io/<owner>/charts}"
TAG="${3:?vX.Y.Z}"

VERSION="${TAG#v}"
mkdir -p dist

helm package "${CHART_DIR}" --version "${VERSION}" --app-version "${VERSION}" --destination dist

for tgz in dist/*.tgz; do
  helm push "$tgz" "oci://${CHART_REPO}"
done

echo "Chart pushed to oci://${CHART_REPO} with tag ${TAG}"
