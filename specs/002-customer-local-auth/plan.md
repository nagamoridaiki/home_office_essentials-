# Implementation Plan: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Branch**: `feat/issue-13-customer-local-auth` | **Date**: 2026-10-12 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-customer-local-auth/spec.md`

## Summary

メールアドレスとパスワードで顧客がアカウントを作り、ログインし、ログアウトできるようにする。設計書どおりの 3 テーブル（`customers` / `customer_credentials` / `customer_sessions`）を新しく作り、外部の認証サービスは使わない。

技術的な方針：

- パスワードは bcrypt で保存する
- ログイン状態は、乱数のトークンを `HttpOnly` Cookie に入れ、DB にはそのハッシュだけを置くサーバー側セッションで持つ
- 将来の OIDC への切り替え（#14）に備え、認証ミドルウェアが依存するインターフェース `CustomerAuthenticator`（リクエスト → 顧客 ID）を境目にする。#14 が Cookie セッションと JWT のどちらになっても、差し替えるのはこの実装と登録・ログインの部分だけで、`/me` 配下は変えない（FR-020・research R8）
- 画面は `/signup`・`/login`・`/mypage` とヘッダーのログアウトを足す

## Technical Context

**Language/Version**: Go 1.27.0（`backend`）／ TypeScript 5 + React 19.2.8 + Next.js 16.3.2（`apps/web`）

**Primary Dependencies**:
- 既存：標準ライブラリ `net/http`・`net/mail`・`crypto/rand`・`crypto/sha256`、`jackc/pgx/v5` 5.11.0、sqlc 1.31.1（コンテナ経由）、Tailwind CSS 4
- **新規：`golang.org/x/crypto`（bcrypt のみ使う）**。標準ライブラリにパスワード向けのハッシュがないため（research R1）。版は `go get` 時点の最新を `go.mod` に固定する

**Storage**: PostgreSQL 18。新しいマイグレーション 3 本（`000003_create_customers`・`000004_create_customer_credentials`・`000005_create_customer_sessions`）

**Testing**: 標準の `testing`。ドメインとユースケースはテスト必須（constitution VI・`.claude/rules/testing.md`）。ユースケースは手書きのフェイクで DB に繋がない。HTTP ハンドラは Cookie の属性と認証ミドルウェアだけをテストする。不変条件に関わる SQL（メールアドレスの一意性・認証情報の 1 人 1 件）は `quickstart.md` の手動手順で確かめる

**Target Platform**: Docker Compose（`web` / `api` / `db`）によるローカル開発環境

**Project Type**: Web application（フロントエンドとバックエンドを分離したモノレポ）

**Performance Goals**: 画面操作の目標は SC-001（登録 1 分以内）・SC-002（ログイン 30 秒以内）。bcrypt コスト 12 の照合は 1 回あたり数百ミリ秒程度で、この目標に収まる

**Constraints**:
- 秘密の値（パスワード・トークン）を平文で保存しない（FR-017・FR-018）
- Cookie を JS から読めない（FR-019）
- 既存の `/api/todos` を変えない
- `04` の共通規約に従う（パスにプレフィックスなし、エラー形式）

**Scale/Scope**: ローカルの単一環境。エンドポイント 4 本、画面 3 つとヘッダー 1 つ、テーブル 3 つ

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` v1.1.0 に照らして評価する。

| 原則 | 評価 | 根拠 |
|---|---|---|
| I. 依存方向とドメインの独立 | ✅ 通過 | ドメインは顧客・メールアドレス・セッションだけを持ち、HTTP・DB・bcrypt を知らない。パスワードの扱いはアプリケーション層のポートとインフラ層に閉じる（`01` の 1.2 と一致）。リポジトリ IF はドメイン（顧客・セッション）とアプリケーション（`PasswordCredentialStore`）に置き、実装は `backend/repository` |
| I（ドメインモデルが契約より先） | ✅ 通過 | `data-model.md` を先に作り、`contracts/` はそれに合わせた |
| II. ユビキタス言語 | ✅ 通過 | 顧客 `Customer`、顧客セッション `CustomerSession`、認証主体 `auth_subject` を `01` のとおりに使う。「ユーザー」「会員」「`users`」は使わない。契約を足すので、フロントの API クライアントと型を同じ変更でそろえる |
| III. 生成物と手書きの分離 | ✅ 通過 | クエリは `backend/db/queries/` に書き、`repository/internal/db/` は sqlc で再生成する。既存のマイグレーション（`000001`・`000002`）は書き換えず、新しい番号で足す |
| IV. 再現可能な開発環境 | ✅ 通過（依存を 1 つ追加） | `golang.org/x/crypto` を足す。標準ライブラリにないことを確かめた（research R1）。migrate と sqlc はコンテナ経由のまま。フロントに新しいパッケージは足さない |
| V. 秘密情報 | ✅ 通過 | 新しい設定は `SESSION_COOKIE_SECURE` だけで、秘密ではない。`compose.yaml` で既定値を渡し、`.env.example` に説明を足す。パスワードとトークンは平文で保存しない |
| VI. テストと受け入れ条件 | ✅ 通過 | 受け入れ条件ごとにユースケースのテストを対応させる。セッションの禁止遷移（ログアウト済み・期限切れから有効に戻らない、二重のログアウト）もテストする。不変条件に関わる SQL は `quickstart.md` の 4 章に手順を書いた |

**Plan Gate の追加項目**

| 項目 | 結果 |
| --- | --- |
| 不変条件（`03` の 6 章）への影響 | 「他人の個人データに URL から到達できない」に影響あり。`/me` は URL に識別子を持たず、認証ミドルウェアでセッションの顧客に一括で絞る（data-model 6 章） |
| 状態遷移（`02`） | `02` のどの遷移にも当たらない。セッションの有効・ログアウト済み・期限切れは本機能の中で定義した（data-model 2.4） |

**判定**: 違反なし。Phase 0 に進む。

### Post-Design Re-check

Phase 1（data-model / contracts / quickstart）の設計後に評価し直した。

- 原則 I：`PasswordCredentialStore` が `customers` と `customer_credentials` をまたぐ。パスワードのハッシュを扱うため、ドメインではなくアプリケーション側に置いた。ドメインはパスワードを知らないままで、違反はない
- 原則 I（追加の見直し）：#14 で JWT 方式を選んだ場合に `CustomerSessionUsecase` が使われなくなることを踏まえ、変わらない境目を controller 層の `CustomerAuthenticator` に置き直した（data-model 4 章）。ドメインとユースケースは HTTP を知らないままで、違反はない
- 原則 II：契約の名前（`customer_session` Cookie、`email_already_used` など）は `01` に反しない
- 設計書の変更が 3 か所必要になった。constitution の「設計を変えたら関係する設計書をすべて直す」に従い、tasks に入れる
  - `04` 共通規約の冪等性に、`POST /customers` の例外を書き足す（research R6）
  - `03` の 5 章に `customers` の `CHECK (email = lower(email))` を書き足す（data-model 5 章）
  - `01` の認証主体の「MVP は自前で発行」を、`03` の「MVP では NULL」に合わせる（spec の Assumptions）

**判定**: 違反なし。

## Project Structure

### Documentation (this feature)

```text
specs/002-customer-local-auth/
├── plan.md              # このファイル
├── research.md          # Phase 0
├── data-model.md        # Phase 1
├── quickstart.md        # Phase 1
├── contracts/
│   └── customer-auth-api.md
├── checklists/
│   └── requirements.md
└── tasks.md             # Phase 2（/speckit-tasks で作る）
```

### Source Code (repository root)

```text
backend/
├── main.go                                   # DI に顧客・セッション・認証を追加
├── domain/
│   ├── customer.go                           # Customer, Email, CustomerRepository, エラー
│   ├── customer_test.go
│   ├── customer_session.go                   # CustomerSession, SessionTokenHash, CustomerSessionRepository
│   └── customer_session_test.go
├── usecase/
│   ├── password_auth.go                      # PasswordAuthUsecase, パスワードの方針, ポート
│   ├── password_auth_test.go
│   ├── customer_session.go                   # CustomerSessionUsecase, SessionTokenGenerator
│   ├── customer_session_test.go
│   ├── customer.go                           # CustomerUsecase, CustomerNotifier
│   └── customer_test.go
├── repository/
│   ├── postgres_customer.go                  # CustomerRepository + PasswordCredentialStore の実装
│   └── postgres_customer_session.go
├── auth/
│   ├── bcrypt_hasher.go                      # PasswordHasher の実装
│   └── session_token.go                      # SessionTokenGenerator の実装
├── notification/
│   └── noop.go                               # NoopCustomerNotifier
├── controller/
│   ├── cors.go                               # Allow-Credentials を追加
│   ├── error_response.go                     # 04 のエラー形式
│   ├── session_cookie.go
│   ├── customer_auth.go                      # POST /customers, POST /sessions, DELETE /sessions
│   ├── me.go                                 # GET /me
│   ├── auth_middleware.go                    # RequireCustomer, CustomerAuthenticator, CookieSessionAuthenticator
│   └── auth_middleware_test.go
└── db/
    ├── migrations/000003〜000005_*.{up,down}.sql
    └── queries/customer.sql, customer_session.sql

apps/web/src/
├── lib/
│   ├── api-client.ts                         # fetch の共通処理（credentials: include、エラー形式の型）
│   └── customer-api.ts                       # signUp / login / logout / getMe
├── types/
│   └── customer.ts                           # Customer, ApiError
├── components/
│   └── site-header.tsx                       # ログアウトボタン
└── app/
    ├── layout.tsx                            # ヘッダーを追加
    ├── signup/page.tsx
    ├── login/page.tsx
    └── mypage/page.tsx

docs/design/
├── 01-ubiquitous-language.md                 # 認証主体の記述を 03 に合わせる
├── 03-logical-data-model.md                  # 5 章に email の CHECK を追加
└── 04-api-list.md                            # 共通規約の冪等性に例外を追記

compose.yaml / .env.example                   # SESSION_COOKIE_SECURE
```

**Structure Decision**:
- 既存の `domain` / `usecase` / `repository` / `controller` の並びを踏襲する
- インフラのうち DB 以外（bcrypt・トークン・通知）は `repository` に混ぜず、`auth` と `notification` を新しく作る。どちらもユースケースのポートを満たすだけの薄い実装にする
- フロントは今の Todo 画面と同じくクライアントコンポーネントで書き、API の呼び出しを `lib/` に集める（`.claude/rules/frontend.md`）

## Complexity Tracking

Constitution Check に違反はない。依存の追加（`golang.org/x/crypto`）は原則 IV の手順（標準ライブラリで足りるかを先に確かめる）に沿っており、違反ではないため、この表には載せない。
