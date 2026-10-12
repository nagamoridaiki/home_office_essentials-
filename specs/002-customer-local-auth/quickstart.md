# Quickstart: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Feature**: [spec.md](./spec.md) | **Date**: 2026-10-12

実装後に、受け入れ条件を手元で確かめる手順。API の形は [contracts/customer-auth-api.md](./contracts/customer-auth-api.md)、テーブルは [data-model.md](./data-model.md) を参照する。

## 1. 準備

```bash
cp -n .env.example .env
docker compose up -d --build
docker compose run --rm migrate up
docker compose run --rm sqlc generate
```

`git diff backend/repository/internal/db` に差分が出なければ、生成コードはコミット済みのものと一致している。

## 2. 自動テスト

```bash
docker compose run --rm --no-deps api go test ./...
```

ドメイン（`Email`・`CustomerSession`）とユースケース（`SignUp`・`Login`・`Issue`・`Authenticate`・`Revoke`・`GetMe`）のテストが通ること。

## 3. API で確かめる（curl）

Cookie を `jar.txt` に保存しながら進める。ホスト名は `localhost` を使う（curl もブラウザも、`localhost` なら http でも `Secure` 付きの Cookie を扱う）。保存されない場合は `.env` に `SESSION_COOKIE_SECURE=false` を書いて `api` を作り直す。

| # | 操作 | 期待する結果 | 対応 |
| --- | --- | --- | --- |
| 1 | `POST /customers`（`{"email":" Taro@Example.com ","password":"correct-horse"}`） | 201。`email` が `taro@example.com`、他の項目は `null`。`Set-Cookie` に `HttpOnly`・`SameSite=Lax`・`Max-Age=2592000` | US1-1・US1-2・FR-019 |
| 2 | 同じ要求をもう一度 | 409 `email_already_used` | US1-3・FR-004 |
| 3 | `taro@EXAMPLE.com` で `POST /customers` | 409 | FR-003・SC-006 |
| 4 | `{"email":"x","password":"short"}` で `POST /customers` | 422。`details` に `email` と `password` の両方 | US1-4・FR-002 |
| 5 | 1 の Cookie で `GET /me` | 200 | US1-2 |
| 6 | `DELETE /sessions` | 204。Cookie が消える | US3-1 |
| 7 | 6 の前に控えた Cookie の値を手で付けて `GET /me` | 401 | US3-2・FR-011・SC-003 |
| 8 | 正しいパスワードで `POST /sessions` | 200 と新しい Cookie | US2-1 |
| 9 | パスワードを変えて `POST /sessions` | 401 `invalid_credentials` | US2-2 |
| 10 | 未登録のメールアドレスで `POST /sessions` | 9 と同じ本文の 401 | US2-3・FR-007 |
| 11 | Cookie の値を 1 文字変えて `GET /me` | 401 | SC-005 |
| 12 | 別の jar で 2 回目のログインをし、片方だけ `DELETE /sessions` | もう片方の `GET /me` は 200 のまま | US3-3・FR-012 |

コマンドの例（1 と 5）:

```bash
curl -i -c jar.txt -H 'Content-Type: application/json' \
  -d '{"email":" Taro@Example.com ","password":"correct-horse"}' \
  http://localhost:8080/customers
```

```bash
curl -i -b jar.txt http://localhost:8080/me
```

## 4. DB で確かめる（不変条件・秘密情報）

```bash
docker compose exec db psql -U app -d home_office_essentials
```

| 確かめること | SQL | 期待する結果 | 対応 |
| --- | --- | --- | --- |
| メールアドレスは 1 件だけ | `SELECT email, count(*) FROM customers GROUP BY email HAVING count(*) > 1;` | 0 行 | SC-006 |
| 大文字の直接 INSERT を拒む | `INSERT INTO customers (email) VALUES ('A@b.jp');` | `CHECK` 違反で失敗 | R5 |
| パスワードが平文でない | `SELECT password_hash FROM customer_credentials;` | `$2a$12$` で始まる値だけ | FR-017・SC-004 |
| トークンが平文でない | `SELECT token_hash FROM customer_sessions;` | 64 文字の 16 進。`jar.txt` の値と一致しない | FR-018・SC-004 |
| ログアウトの記録が残る | `SELECT created_at, revoked_at FROM customer_sessions ORDER BY id;` | 6 で使った行の `revoked_at` が埋まっていて、行は消えていない | FR-013 |
| 認証情報は 1 人 1 件 | 同じ `customer_id` で `customer_credentials` に 2 行目を INSERT | UNIQUE 違反で失敗 | Clarifications Q1 |

## 5. 期限切れ（FR-008）

ユースケースのテストで、固定した時刻を 30 日進めて確かめる。手で確かめるときは、`customer_sessions.expires_at` を過去の時刻に更新し、その Cookie で `GET /me` が 401 になることを見る。

## 6. 画面で確かめる（http://localhost:3000）

| # | 操作 | 期待する結果 | 対応 |
| --- | --- | --- | --- |
| 1 | `/signup` で登録 | `/mypage` に移り、メールアドレスと「未登録」が表示される | US1・SC-001（1 分以内） |
| 2 | 登録ボタンを素早く 2 回押す | ボタンが送信中は押せない。アカウントは 1 件 | FR-004 |
| 3 | ログインしたまま `/signup`・`/login` を開く | `/mypage` に移る | FR-009a |
| 4 | ヘッダーのログアウトを押す | ログアウトした状態になり、`/mypage` を開くと `/login` に移る | US3-1 |
| 5 | `/login` でログイン | `/mypage` に移る | US2・SC-002（30 秒以内） |
| 6 | ブラウザを閉じて開き直し、`/mypage` を開く | ログインしたまま | US2-4 |
| 7 | 開発者ツールのコンソールで `document.cookie` | `customer_session` が出てこない | FR-019 |
