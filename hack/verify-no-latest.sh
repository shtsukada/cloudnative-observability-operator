#!/usr/bin/env bash
set -euo pipefail

# 監査対象:
#   - Helm/manifest YAML: charts/**, config/**
#   - Dockerfile / Dockerfile.*
# 除外:
#   - ドキュメント: **/*.md, **/README*
#   - Makefile
#   - 自分自身
#   - バックアップ/テンプレ類: **/*.bak, **/*.tmpl

# 対象ファイル一覧を取得（git管理下のみ）
files="$(git ls-files \
  'charts/**' \
  'config/**' \
  'Dockerfile' \
  'Dockerfile.*' \
  ':!:**/*.md' \
  ':!:**/README*' \
  ':!:Makefile' \
  ':!:hack/verify-no-latest.sh' \
  ':!:**/*.bak' \
  ':!:**/*.tmpl' \
  || true)"

# 何もなければOK
if [ -z "${files}" ]; then
  echo "OK: no ':latest' usage found."
  exit 0
fi

# 検出パターン:
#   - ":latest"
#   - "tag: latest"
#   - "image: <repo>:latest"
PATTERN='(:latest\b|tag:\s*latest\b|image:\s*[^\s]+:latest\b)'

# grep はヒット無しで終了コード1を返すので || true で握る
matches="$(echo "${files}" | xargs grep -n -E "${PATTERN}" || true)"

if [ -n "${matches}" ]; then
  echo "${matches}"
  echo "::error ::Detected disallowed ':latest' tag usage. Pin image tags explicitly."
  exit 2
fi

echo "OK: no ':latest' usage found."
