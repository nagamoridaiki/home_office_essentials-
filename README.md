# Home Office Essentials

## 技術スタック

- **Frontend**: Next.js 16 / React 19 / TypeScript / Tailwind CSS 4
- **Backend**: Go 1.27（標準ライブラリ `net/http`）
- **Database**: PostgreSQL 18
- **パッケージマネージャ**: pnpm 11（ワークスペース）

## Docker での起動（推奨）

`web`（Next.js）・`api`（Go + air）・`db`（PostgreSQL）の 3 つをコンテナで起動する。

### 初回のみ

`.env.example` をコピーして `.env` を用意する（`.env` はコミットしない）。

```bash
cp .env.example .env
```

DB にテーブルを作る（`db` が止まっていても自動で起動する）。

```bash
docker compose run --rm migrate up
```

### 起動

```bash
docker compose up
```

ログを見ながら起動する。`Ctrl+C` で停止。裏で動かす場合は `-d` を付ける。

```bash
docker compose up -d
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080/api/todos（`/` はルート未定義のため 404）

`api` は `db` が healthy になってから起動する。

### ソースの変更の反映

| 変更したもの | 反映のされ方 |
| --- | --- |
| `backend/` 以下の `.go` | air が自動で再ビルド・再起動 |
| `apps/web/src/`・`apps/web/public/` | Next.js が自動で反映（HMR） |
| `apps/web` の設定ファイル・`package.json` | `docker compose up -d --build web` で作り直す |

### 状態・ログの確認

```bash
docker compose ps
```

```bash
docker compose logs api
```

### 動作確認

DB に設定したユーザー・データベースで接続できるか。

```bash
docker compose exec db psql -U app -d home_office_essentials -c "select current_user, current_database();"
```

`api` のコンテナから `db` に届くか（ネットワークの疎通のみ。アプリからの接続は別 Issue で実装する）。

```bash
docker compose exec api bash -c 'echo > /dev/tcp/db/5432 && echo "db:5432 に接続できた"'
```

### DB マイグレーション

[golang-migrate](https://github.com/golang-migrate/migrate) をコンテナで実行する（ローカルへのインストールは不要）。SQL は `backend/db/migrations/` に置く。書き方のルールは [.claude/rules/database.md](.claude/rules/database.md)。

未適用のものをすべて適用する。

```bash
docker compose run --rm migrate up
```

1 つ前に戻す。

```bash
docker compose run --rm migrate down 1
```

> [!CAUTION]
> 数字を付けない `down` はすべてを戻す（`[y/N]` で確認される）。

いまの version を見る（`2` のように表示される）。

```bash
docker compose run --rm migrate version
```

新しいマイグレーションを作る。`backend/db/migrations/` に `00000N_<名前>.up.sql` と `.down.sql` ができるので、中身を書く。

```bash
docker compose run --rm migrate create -ext sql -dir /migrations -seq <名前>
```

#### 失敗したとき（dirty）

SQL が失敗すると version に `(dirty)` が付き、`up` / `down` が `Dirty database version N. Fix and force version.` で止まる。各ファイルは `BEGIN;` 〜 `COMMIT;` で囲んでいるため、失敗した N の変更は DB に残っていない。

1. N の SQL を直す
2. migrate が覚えている「いま何番まで成功したか」のメモを、1 つ前（N − 1）に書き換える。`force` はテーブルなどには一切触れず、このメモ（DB の `schema_migrations` テーブル）だけを書き換える

   ```bash
   docker compose run --rm migrate force <N − 1>
   ```

3. もう一度 `up` する

失敗時に続けて出る `pg_advisory_unlock` のエラーは、コンテナの終了とともに解消されるので対応不要。

### 停止

```bash
# コンテナを止める（再開は docker compose up -d）
docker compose stop

# コンテナを削除する。DB のデータ（ボリューム）は残る
docker compose down
```

> [!WARNING]
> `docker compose down -v` は DB のデータも削除する。データを消したいときだけ使う。
> 消したあとは `docker compose run --rm migrate up` でテーブルを作り直す。

## Docker を使わない場合

ホストで直接起動する。Docker と同じポート（3000 / 8080）を使うため、同時には起動できない。

```bash
pnpm dev
```

次の 2 つが同時に立ち上がる。

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

```bash
# フロントのみ
pnpm --filter web dev

# バックのみ
cd backend && go run .
```
