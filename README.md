# Home Office Essentials

## 技術スタック

- **Frontend**: Next.js 16 / React 19 / TypeScript / Tailwind CSS 4
- **Backend**: Go 1.27（標準ライブラリ `net/http`）
- **パッケージマネージャ**: pnpm 11（ワークスペース）

## 開発用データベース（PostgreSQL）

初回のみ、`.env.example` をコピーして `.env` を用意する。

```bash
cp .env.example .env
```

DB だけ起動する。

```bash
docker compose up -d db
```

プロセス確認。

```bash
docker compose ps
```

設定したユーザー・データベースで接続できるか確認する。

```bash
docker compose exec db psql -U app -d home_office_essentials -c "select current_user, current_database();"
```

## 開発サーバーの起動

```bash
pnpm dev
```

次の 2 つが同時に立ち上がります。

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

```bash
# フロントのみ
pnpm --filter web dev

# バックのみ
cd backend && go run .
```
