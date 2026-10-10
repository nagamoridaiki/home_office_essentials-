#!/usr/bin/env bash
# やりたいこと: Claude が触ってはいけないファイルを、うっかり書き換えないようにする
# していること: Claude がファイルを編集する直前に呼ばれ、次のファイルなら「編集してよいか」を人に確認する
#   - .env 系（秘密情報）
#   - sqlc の生成コード（手で直すと再生成で消える）
#   - コミット済みのマイグレーション（適用済みの環境とずれる）
f=$(jq -r '.tool_input.file_path // empty')
[ -z "$f" ] && exit 0
root="${CLAUDE_PROJECT_DIR:-$(pwd)}"
rel="${f#"$root"/}"

reason=""
case "$rel" in
  .env|.env.*|*/.env|*/.env.*)
    reason="秘密情報のファイル（$rel）です。" ;;
  backend/repository/internal/db/*)
    reason="sqlc の生成コード（$rel）です。SQL を直して再生成してください。" ;;
  backend/db/migrations/*)
    # 新規ファイルは対象外。コミット済みのものだけ止める
    if git -C "$root" ls-files --error-unmatch "$rel" >/dev/null 2>&1; then
      reason="コミット済みのマイグレーション（$rel）です。変更は新しい連番のファイルで行ってください。"
    fi ;;
esac

[ -z "$reason" ] && exit 0
jq -n --arg r "$reason" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"ask",permissionDecisionReason:$r}}'
