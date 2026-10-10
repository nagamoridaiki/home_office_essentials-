#!/usr/bin/env bash
# やりたいこと: Claude が書いたコードの書式を、毎回そろった状態にする
# していること: Claude がファイルを編集した直後に呼ばれ、Go は gofmt、apps/web の TS は eslint --fix で整形する
f=$(jq -r '.tool_response.filePath // .tool_input.file_path // empty')
[ -z "$f" ] || [ ! -f "$f" ] && exit 0
root="${CLAUDE_PROJECT_DIR:-$(pwd)}"
rel="${f#"$root"/}"

case "$rel" in
  backend/repository/internal/db/*) ;;
  backend/*.go) gofmt -w "$f" ;;
  apps/web/*.ts|apps/web/*.tsx)
    (cd "$root/apps/web" && ./node_modules/.bin/eslint --fix "$f") ;;
esac
exit 0
