#!/usr/bin/env bash
set -euo pipefail
IMAGE_REF="${1:?image ref}"
OUT="${2:?output path .spdx.json}"

syft packages "${IMAGE_REF}" -o spdx-json > "${OUT}"
echo "SBOM written: ${OUT}"
