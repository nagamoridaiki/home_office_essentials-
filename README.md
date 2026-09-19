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

### 停止

```bash
# コンテナを止める（再開は docker compose up -d）
docker compose stop

# コンテナを削除する。DB のデータ（ボリューム）は残る
docker compose down
```

> [!WARNING]
> `docker compose down -v` は DB のデータも削除する。データを消したいときだけ使う。

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
