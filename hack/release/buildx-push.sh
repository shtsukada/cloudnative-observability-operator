#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${1:?ghcr.io}"
OWNER="${2:?owner}"
IMAGE="${3:?image name}"
TAG="${4:?vX.Y.Z}"

REF="${REGISTRY}/${OWNER}/${IMAGE}:${TAG}"

docker buildx build \
  --platform linux/amd64 \
  --tag "${REF}" \
  --push \
  .

echo "Pushed image: ${REF}"
