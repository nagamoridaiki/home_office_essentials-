---
paths:
  - "backend/db/migrations/**"
---

# DB マイグレーション（golang-migrate / PostgreSQL）

## すべてのテーブルに付ける列

- `created_at` と `updated_at` を必ず付ける
  - 型は `timestamptz NOT NULL DEFAULT now()`
- `updated_at` は DB のトリガーで更新する。テーブルを作るマイグレーションで、共通関数 `set_updated_at()`（`000001`）を呼ぶトリガーも作る

```sql
CREATE OR REPLACE TRIGGER <テーブル名>_set_updated_at
    BEFORE UPDATE ON <テーブル名>
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

## ファイルの書き方

- `docker compose run --rm migrate create -ext sql -dir /migrations -seq <名前>` で作る（6 桁の連番）
- up / down は `BEGIN;` 〜 `COMMIT;` で囲む
- `IF NOT EXISTS` / `IF EXISTS` / `OR REPLACE` を使い、何度流しても壊れないようにする
- 適用済み・コミット済みのファイルは書き換えない。変更は新しい番号のファイルで行う
