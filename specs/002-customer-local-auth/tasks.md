---

description: "顧客のアカウント作成・ログイン・ログアウト（自前認証）の実装タスク"
---

# Tasks: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Input**: `/specs/002-customer-local-auth/` の設計書

**Prerequisites**: [plan.md](./plan.md)、[spec.md](./spec.md)、[research.md](./research.md)、[data-model.md](./data-model.md)、[contracts/customer-auth-api.md](./contracts/customer-auth-api.md)、[quickstart.md](./quickstart.md)

**Tests**: constitution VI と `.claude/rules/testing.md` により、ドメインとユースケースのテストは必須。受け入れ条件ごとに 1 つ以上のテストケースを対応させ、禁止された遷移が拒否されることも確かめる。テストは実装より先に書き、失敗することを確かめてから実装する。

**Organization**: ユーザーストーリーごとに分け、それぞれ単独で実装・検証できるようにする。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 並行して進められる（別ファイルで、終わっていないタスクに依存しない）
- **[Story]**: 対応するユーザーストーリー（US1・US2・US3）

## 共通の約束

- バックエンドのテストのパッケージ名は `xxx_test`。リポジトリなどは手書きのフェイクに差し替え、DB に繋がない。エラーは `errors.Is` で比べ、`t.Run` のケース名は日本語にする
- テストの実行は `docker compose run --rm --no-deps api go test ./...`
- フロントのコマンドはリポジトリ直下から `pnpm --filter web <script>`。`any` は使わない
- コメントとコミットメッセージは日本語。用語は `docs/design/01-ubiquitous-language.md` に合わせ、「ユーザー」「会員」「`users`」は使わない

---

## Phase 1: Setup（依存・設定・設計書）

**Purpose**: 実装の前に、依存と設定を足し、plan で見つかった設計書のずれを直す

- [x] T001 `backend` で `go get golang.org/x/crypto` を実行し、`backend/go.mod` と `backend/go.sum` に版を固定する（bcrypt のみ使う。research R1）
- [x] T002 [P] `compose.yaml` の `api` の `environment` に `SESSION_COOKIE_SECURE: ${SESSION_COOKIE_SECURE:-true}` を足し、`.env.example` に「`false` にすると Cookie の `Secure` を外す。ローカルで Cookie が保存されないときだけ使う」という説明付きで `SESSION_COOKIE_SECURE=true` を足す（research R3）
- [x] T003 [P] `docs/design/01-ubiquitous-language.md` の認証主体（`auth_subject`）の「MVP は自前で発行」を、`03` に合わせて「MVP では NULL（外部の認証サービスに切り替えたときに入れる）」に直す
- [x] T004 [P] `docs/design/03-logical-data-model.md` の 5 章に、`customers` の `CHECK (email = lower(email))`（小文字にそろえた値だけを保存する。research R5）を書き足す
- [x] T005 [P] `docs/design/04-api-list.md` の共通規約の冪等性に、「`POST /customers` は例外。2 回目の同じ要求は `409 email_already_used` を返す。アカウントが 2 つ作られない点は共通規約と同じ」を書き足す（research R6）

---

## Phase 2: Foundational（全ストーリーの前提）

**Purpose**: テーブル、ドメイン、セッションの発行と確認、認証ミドルウェア、`GET /me`、フロントの API クライアントの土台

**⚠️ CRITICAL**: このフェーズが終わるまでユーザーストーリーに着手しない

### DB とクエリ

- [x] T006 `docker compose run --rm migrate create -ext sql -dir /migrations -seq create_customers` で作り、`backend/db/migrations/000003_create_customers.{up,down}.sql` を書く。列は `id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY`、`email text NOT NULL`、`family_name`・`given_name`・`family_name_kana`・`given_name_kana`・`phone_number`（いずれも `text` で NULL 可）、`created_at`・`updated_at`（`timestamptz NOT NULL DEFAULT now()`）。制約は `UNIQUE (email)` と `CHECK (email = lower(email))`。`set_updated_at()` を呼ぶトリガー `customers_set_updated_at` を作る。up/down は `BEGIN;`〜`COMMIT;` で囲み、`IF NOT EXISTS`／`IF EXISTS`／`OR REPLACE` を使う（`.claude/rules/database.md`）
- [x] T007 `backend/db/migrations/000004_create_customer_credentials.{up,down}.sql` を同じ手順で書く。列は `id`、`customer_id bigint NOT NULL`（FK → `customers.id` `ON DELETE RESTRICT`）、`password_hash text`、`auth_issuer text`、`auth_subject text`、`created_at`・`updated_at`。制約は `UNIQUE (customer_id)`（Clarifications Q1）と `UNIQUE (auth_issuer, auth_subject)`。トリガーも作る（T006 の後）
- [x] T008 `backend/db/migrations/000005_create_customer_sessions.{up,down}.sql` を同じ手順で書く。列は `id`、`customer_id bigint NOT NULL`（FK → `customers.id` `ON DELETE RESTRICT`）、`token_hash text NOT NULL`、`expires_at timestamptz NOT NULL`、`revoked_at timestamptz`（NULL 可）、`created_at`・`updated_at`。制約は `UNIQUE (token_hash)`、インデックスは `customer_id`。トリガーも作る（T007 の後）
- [x] T009 `docker compose run --rm migrate up` で 000003〜000005 を適用し、`down 3` → `up` を流して何度でも壊れないことを確かめる
- [x] T010 [P] `backend/db/queries/customer.sql` に sqlc のクエリを書く：`GetCustomerByID`（:one）、`CreateCustomer`（:one。`email` だけを入れ、全列を返す）、`CreateCustomerCredential`（:exec。`customer_id`・`password_hash`）、`GetPasswordHashByEmail`（:one。`customers` と `customer_credentials` を結合し `customer_id`・`password_hash` を返す）
- [x] T011 [P] `backend/db/queries/customer_session.sql` にクエリを書く：`CreateCustomerSession`（:one）、`GetCustomerSessionByTokenHash`（:one）、`RevokeCustomerSession`（:execrows。`WHERE id = $1 AND revoked_at IS NULL` で `revoked_at` を埋める）
- [x] T012 `docker compose run --rm sqlc generate` で `backend/repository/internal/db/` を再生成する（手で編集しない。T009・T010・T011 の後）

### ドメイン（テストを先に書く）

- [x] T013 [P] `backend/domain/customer_test.go`（`package domain_test`）に `NewEmail` のテーブル駆動テストを書く：前後の空白を除き小文字になる（`" Taro@Example.com "` → `taro@example.com`）／空文字・`x`・`a@`・表示名付き（`"Taro" <a@b.jp>`）・255 文字以上は `ErrInvalidEmail`／254 文字ちょうどは通る
- [x] T014 [P] `backend/domain/customer.go` に書く：`Email` 値オブジェクトと `NewEmail(raw string) (Email, error)`（「前後の空白を取り除き、小文字にする」「`net/mail.ParseAddress` で解釈でき、表示名がなく、解釈結果が入力と一致し、254 文字以下」）、`Customer`（`ID int64`、`Email Email`、`FamilyName`・`GivenName`・`FamilyNameKana`・`GivenNameKana`・`PhoneNumber` は `*string` で「未登録なら `nil`」）、`CustomerRepository`（`FindByID(ctx, id int64) (Customer, error)`。なければ `ErrCustomerNotFound`）、エラー `ErrInvalidEmail`・`ErrEmailAlreadyUsed`・`ErrCustomerNotFound`。生成関数は持たない（data-model 2.2）
- [x] T015 [P] `backend/domain/customer_session_test.go` に `CustomerSession` のテストを書く：`NewCustomerSession` の `ExpiresAt` が `now + 30 日`／`IsActive` は期限の直前で true、ちょうど・過ぎたら false、ログアウト済みなら false／有効なセッションの `Revoke` は `RevokedAt` を埋める／**禁止された遷移**：ログアウト済みの `Revoke` と期限切れの `Revoke` は `ErrSessionNotActive` で、`RevokedAt` が変わらない／ログアウト済みは時刻を戻しても `IsActive` が false のまま
- [x] T016 [P] `backend/domain/customer_session.go` に書く：`SessionTokenHash`（比べるだけの値。作り方は持たない）、`SessionLifetime = 30 * 24 * time.Hour`、`CustomerSession`（`ID`・`CustomerID`・`TokenHash`・`ExpiresAt`・`RevokedAt *time.Time`・`CreatedAt`）、`NewCustomerSession(customerID, tokenHash, now)`、`IsActive(now) bool`（`RevokedAt == nil` かつ `now < ExpiresAt`）、`Revoke(now) error`（有効でなければ `ErrSessionNotActive`。`RevokedAt` を戻す操作は作らない）、`CustomerSessionRepository`（`Create`、`FindByTokenHash`（なければ `ErrSessionNotFound`）、`Revoke(ctx, id int64, at time.Time) error`）、エラー `ErrSessionNotFound`・`ErrSessionNotActive`

### ユースケース（セッションの発行と確認、自分の情報）

- [x] T017 [P] `backend/usecase/customer_session_test.go`（`package usecase_test`）に、フェイクのリポジトリとトークン生成器・固定の時刻関数を使ったテストを書く：`Issue` が生のトークンを返し、リポジトリにはハッシュだけが渡る（生の値が保存されない。FR-018）／`Authenticate` は有効なトークンで顧客 ID を返す（US2-4）／存在しないトークン・ログアウト済み・期限切れ（時刻を 30 日進める。FR-008）・1 文字変えたトークン（SC-005）はいずれも `ErrUnauthenticated` で区別しない
- [x] T018 `backend/usecase/customer_session.go` に書く：ポート `SessionTokenGenerator`（`Generate() (raw string, hash domain.SessionTokenHash, err error)`、`Hash(raw string) domain.SessionTokenHash`）、`ErrUnauthenticated`、`CustomerSessionUsecase`（コンストラクタで `domain.CustomerSessionRepository`・`SessionTokenGenerator`・`now func() time.Time` を受け取る。research R12）、`Issue(ctx, customerID) (raw string, domain.CustomerSession, error)`、`Authenticate(ctx, raw) (customerID int64, error)`。`Revoke` は US3 で足す（T017 の後）
- [x] T019 [P] `backend/usecase/customer_test.go` に `GetMe` のテストを書く：自分の顧客を返し、未登録の項目は `nil` のまま（US1-2・FR-014）／いない顧客 ID は `domain.ErrCustomerNotFound`
- [x] T020 `backend/usecase/customer.go` に `CustomerUsecase`（`domain.CustomerRepository` を受け取る）と `GetMe(ctx, customerID int64) (domain.Customer, error)` を書く。ポート `CustomerNotifier`（`NotifyRegistered(ctx, c domain.Customer) error`）もここに置く（T019 の後）

### インフラ

- [x] T021 [P] `backend/auth/session_token.go` に `SessionTokenGenerator` の実装を書く：`crypto/rand` で 32 バイト → base64url（パディングなし。`base64.RawURLEncoding`）を生の値にし、`crypto/sha256` の 16 進文字列（64 文字）を `domain.SessionTokenHash` にする（research R2）
- [x] T022 [P] `backend/repository/postgres_customer.go` に `domain.CustomerRepository` の `FindByID` を実装する。`pgx.ErrNoRows` は `domain.ErrCustomerNotFound` に変える。sqlc の型は外に出さず `domain.Customer` に詰め替える（NULL の列は `nil`）（T012 の後）
- [x] T023 [P] `backend/repository/postgres_customer_session.go` に `domain.CustomerSessionRepository` の `Create`・`FindByTokenHash`（なければ `domain.ErrSessionNotFound`）を実装する。`Revoke` は US3 で足す（T012 の後）

### HTTP（controller）

- [x] T024 [P] `backend/controller/error_response.go` に `04` のエラー形式 `{"error":{"code","message","details"}}` を書き出す関数を書く。`details` は検証エラー以外は空配列。`code` は `invalid_request`(400)・`unsupported_media_type`(415)・`validation_failed`(422)・`email_already_used`(409)・`invalid_credentials`(401)・`unauthenticated`(401)・`internal_error`(500)。JSON 本文の読み取りで `Content-Type` が `application/json` 以外なら 415、読めなければ 400 にする共通関数もここに置く（research R4）
- [x] T025 [P] `backend/controller/session_cookie.go` に Cookie `customer_session` の発行と削除を書く。発行は `HttpOnly; SameSite=Lax; Path=/; Max-Age=2592000`、`Secure` は設定値で切り替える（既定 true）。削除は同じ名前・`Path`・`Max-Age=0`（contracts「セッション Cookie」）
- [x] T026 [P] `backend/controller/cors.go` の `EnableCORS` に `Access-Control-Allow-Credentials: true` を足す。許可するオリジンは `http://localhost:3000` の 1 つのまま（`*` にしない）
- [x] T027 [P] `backend/controller/auth_middleware_test.go`（`package controller_test`）に `httptest` で書く：`RequireCustomer` は認証に成功すると次のハンドラでコンテキストから顧客 ID が読める／失敗すると 401 `unauthenticated` で次のハンドラを呼ばない／`CookieSessionAuthenticator` は Cookie がないとき `ErrUnauthenticated`
- [x] T028 `backend/controller/auth_middleware.go` に書く：`CustomerAuthenticator`（`AuthenticateRequest(r *http.Request) (customerID int64, err error)`）、`CookieSessionAuthenticator`（Cookie `customer_session` を読み `CustomerSessionUsecase.Authenticate` を呼ぶ）、`RequireCustomer(a CustomerAuthenticator) func(http.Handler) http.Handler`、`CustomerIDFromContext(ctx) (int64, bool)`（research R8・data-model 4 章。T027 の後）
- [x] T029 `backend/controller/me.go` に `GET /me` を書く。顧客 ID は `CustomerIDFromContext` だけから読み、リクエストに顧客の識別子を受け取らない（FR-016）。応答は contracts の「自分の情報」の形（`email`・`family_name`・`given_name`・`family_name_kana`・`given_name_kana`・`phone_number`、未登録は `null`、`id` は返さない）。自分の情報を JSON にする関数は `POST /customers`・`POST /sessions` でも使うので、このファイルに置く
- [x] T030 `backend/main.go` に依存の組み立てを足す：環境変数 `SESSION_COOKIE_SECURE`（`false` のときだけ `Secure` を外す）、`CustomerSessionUsecase`（`time.Now`）、`CustomerUsecase`、`CookieSessionAuthenticator`、`GET /me` を `RequireCustomer` で包んで登録する。既存の `/api/todos` は変えない

### フロント（API クライアントの土台）

- [x] T031 [P] `apps/web/src/types/customer.ts` に `Customer`（`email: string`、`family_name`・`given_name`・`family_name_kana`・`given_name_kana`・`phone_number` は `string | null`）と `ApiError`（`code`・`message`・`details: { field: string; code: string; message: string }[]`）の型を書く
- [x] T032 [P] `apps/web/src/lib/api-client.ts` に fetch の共通処理を書く：ベース URL は `http://localhost:8080`、`credentials: "include"`、本文があれば `Content-Type: application/json`、失敗時は `ApiError` とステータスを持つエラーを投げる
- [x] T033 `apps/web/src/lib/customer-api.ts` に `getMe(): Promise<Customer>`（`GET /me`）を書く。401 を呼び出し側が見分けられるようにする（T031・T032 の後）

**Checkpoint**: `go test ./...` が通り、DB に直接作った顧客とセッションで `GET /me` が 200、Cookie がなければ 401 になる

---

## Phase 3: User Story 1 - アカウントを作って使い始める (Priority: P1) 🎯 MVP

**Goal**: メールアドレスとパスワードだけでアカウントを作り、そのままログインした状態で自分の情報を見られる

**Independent Test**: `/signup` で新しいメールアドレスを登録し、`/mypage` にメールアドレスと「未登録」が表示される（quickstart 3 章の 1〜5、6 章の 1〜2）

### Tests for User Story 1 ⚠️

> 先に書き、失敗することを確かめる

- [x] T034 [US1] `backend/usecase/password_auth_test.go` に `SignUp` のテストを書く（フェイクの `PasswordHasher`・`PasswordCredentialStore`・`CustomerNotifier`）：正しい入力で顧客が作られ、メールアドレスが小文字にそろう（US1-1）／ストアに渡るのは生のパスワードではなくハッシュ（FR-017）／メールアドレスの形式不正とパスワード 7 文字を同時に送ると、両方の項目の理由を持つ `ValidationError` が返り、ストアを呼ばない（US1-4・FR-002）／空のメールアドレスとパスワードは `required`／パスワードは「8 文字（rune）以上、72 バイト以下」で、8 文字ちょうど・72 バイトちょうどは通り、73 バイトは `too_long`／ストアが `domain.ErrEmailAlreadyUsed` を返すとそのまま伝わる（US1-3・FR-003・FR-004）／通知が失敗しても `SignUp` は成功する（R10）

### Implementation for User Story 1

- [x] T035 [US1] `backend/usecase/password_auth.go` に書く：`ValidatePassword(raw string) error`（「8 文字（rune）以上、72 バイト以下」。`ErrPasswordTooShort`・`ErrPasswordTooLong`）、`ValidationError`（項目名と理由コード `required`・`invalid_email`・`too_short`・`too_long` の組の一覧）、ポート `PasswordHasher`（`Hash`・`Compare`（違えば `ErrPasswordMismatch`）・`DummyCompare`）と `PasswordCredentialStore`（`CreateCustomerWithPassword`・`FindPasswordHashByEmail`）、`PasswordAuthUsecase` と `SignUp(ctx, email, password string) (domain.Customer, error)`（data-model 3.3 の手順 1〜4。T034 の後）
- [x] T036 [P] [US1] `backend/auth/bcrypt_hasher.go` に `PasswordHasher` の実装を書く：コスト 12 の `bcrypt.GenerateFromPassword`、`bcrypt.CompareHashAndPassword` の不一致を `usecase.ErrPasswordMismatch` に変える、`DummyCompare` は起動時に 1 度だけ作ったダミーのハッシュと照合して結果を捨てる（research R1・R7）
- [x] T037 [P] [US1] `backend/notification/noop.go` に `NoopCustomerNotifier`（`NotifyRegistered` は何もせず nil を返す）を書く
- [x] T038 [US1] `backend/repository/postgres_customer.go` に `CreateCustomerWithPassword(ctx, email domain.Email, passwordHash string) (domain.Customer, error)` を足す：`CreateCustomer` と `CreateCustomerCredential` を 1 つのトランザクションで行い、UNIQUE 違反（SQLSTATE `23505`、`pgconn.PgError` で判定）を `domain.ErrEmailAlreadyUsed` に変える。事前の存在確認（SELECT）はしない（research R6・R9）
- [x] T039 [US1] `backend/controller/customer_auth.go` に `POST /customers` を書く：本文 `{"email","password"}` を T024 の共通関数で読む（415・400）→ `SignUp` → `ValidationError` は 422 `validation_failed` で `details` に contracts の `message` を入れる／`domain.ErrEmailAlreadyUsed` は 409 `email_already_used` → 成功したら `CustomerSessionUsecase.Issue` を呼び、Cookie を発行して 201 と自分の情報を返す
- [x] T040 [US1] `backend/main.go` に `PasswordAuthUsecase`（bcrypt・`PostgresCustomer` のストア・`NoopCustomerNotifier`）を組み立て、`POST /customers` を登録する
- [x] T041 [US1] `apps/web/src/lib/customer-api.ts` に `signUp(email: string, password: string): Promise<Customer>` を足す
- [x] T042 [P] [US1] `apps/web/src/app/mypage/page.tsx` を作る（クライアントコンポーネント）：開いたときに `getMe` を呼び、401 なら `/login` へ移す。メールアドレスを表示し、氏名・フリガナ・電話番号は `null` なら「未登録」と表示する（FR-014）
- [x] T043 [P] [US1] `apps/web/src/app/signup/page.tsx` を作る（クライアントコンポーネント）：開いたときに `getMe` が 200 なら `/mypage` へ移す（FR-009a）。入力はメールアドレスとパスワードだけ（FR-001）。送信中はボタンを押せなくする（FR-004）。422 は `details` の内容を項目ごとに、409 は「このメールアドレスはすでに使われています」を表示する。201 なら `/mypage` へ移す
- [x] T044 [US1] quickstart 3 章の 1〜5 と 4 章（メールアドレスの重複・大文字の直接 INSERT・パスワードとトークンが平文でない）を手で確かめる

**Checkpoint**: アカウント作成から自分の情報の表示まで単独で動く

---

## Phase 4: User Story 2 - ログインして続きから使う (Priority: P2)

**Goal**: アカウントを持つお客様がメールアドレスとパスワードでログインし、自分の情報を見られる

**Independent Test**: 既存のアカウントで `/login` からログインし、`/mypage` に本人の情報が表示される。パスワード違いと未登録は同じ案内になる（quickstart 3 章の 8〜11、6 章の 3・5・6）

### Tests for User Story 2 ⚠️

- [x] T045 [US2] `backend/usecase/password_auth_test.go` に `Login` のテストを足す：正しいメールアドレス（大文字・前後の空白があっても）とパスワードで顧客 ID が返る（US2-1）／パスワード違いは `ErrInvalidCredentials`（US2-2）／未登録のメールアドレスは `ErrInvalidCredentials` で、フェイクの `DummyCompare` が 1 回呼ばれる（US2-3・FR-007・R7）／形式不正のメールアドレスも `ErrInvalidCredentials` で `DummyCompare` が呼ばれ、`ValidationError` は返さない

### Implementation for User Story 2

- [x] T046 [US2] `backend/usecase/password_auth.go` に `ErrInvalidCredentials` と `Login(ctx, email, password string) (customerID int64, error)` を足す（data-model 3.3 の `Login`。T045 の後）
- [x] T047 [P] [US2] `backend/repository/postgres_customer.go` に `FindPasswordHashByEmail(ctx, email domain.Email) (customerID int64, passwordHash string, err error)` を足す。なければ `domain.ErrCustomerNotFound`
- [x] T048 [US2] `backend/controller/customer_auth.go` に `POST /sessions` を足す：本文を読み（415・400）→ `Login` → 失敗は 401 `invalid_credentials`（`message` は「メールアドレスまたはパスワードが正しくありません」。422 は返さない）→ 成功したら `Issue` で Cookie を発行し、`GetMe` の結果を 200 で返す。他の端末のセッションには触れない（FR-009）
- [x] T049 [US2] `backend/main.go` に `POST /sessions` を登録する
- [x] T050 [US2] `apps/web/src/lib/customer-api.ts` に `login(email: string, password: string): Promise<Customer>` を足す
- [x] T051 [US2] `apps/web/src/app/login/page.tsx` を作る（クライアントコンポーネント）：開いたときに `getMe` が 200 なら `/mypage` へ移す（FR-009a）。送信中はボタンを押せなくする。401 は API の `message` をそのまま表示する。成功したら `/mypage` へ移す。`/signup` へのリンクを置き、`/signup` からも `/login` へのリンクを置く（`apps/web/src/app/signup/page.tsx`）
- [x] T052 [US2] quickstart 3 章の 8〜11 と 6 章の 3・5・6 を手で確かめる

**Checkpoint**: US1 と US2 がそれぞれ単独で動く

---

## Phase 5: User Story 3 - ログアウトする (Priority: P3)

**Goal**: その端末のログイン状態だけを終わらせ、以後そのログイン状態を使えなくする

**Independent Test**: ログインした状態でヘッダーのログアウトを押し、`/mypage` を開くと `/login` に移る。ログアウト前の Cookie で `GET /me` が 401 になり、別の端末は使い続けられる（quickstart 3 章の 6・7・12、6 章の 4）

### Tests for User Story 3 ⚠️

- [ ] T053 [US3] `backend/usecase/customer_session_test.go` に `Revoke` のテストを足す：有効なトークンで `Revoke` すると、その後 `Authenticate` が `ErrUnauthenticated`（US3-1・US3-2・FR-011）／同じ顧客の別のセッションは `Authenticate` が成功したまま（US3-3・FR-012）／**禁止された遷移**：ログアウト済みのトークンでもう一度 `Revoke` すると `ErrUnauthenticated`、期限切れのトークンの `Revoke` も `ErrUnauthenticated`／存在しないトークンは `ErrUnauthenticated`／リポジトリの `Revoke` には現在時刻が渡り、行は消さない（FR-013）

### Implementation for User Story 3

- [ ] T054 [US3] `backend/usecase/customer_session.go` に `Revoke(ctx, raw string) error` を足す：ハッシュで引き、`domain.CustomerSession.Revoke(now)` が `ErrSessionNotActive` なら、またはセッションがなければ `ErrUnauthenticated`。成功したらリポジトリの `Revoke` を呼ぶ（T053 の後）
- [x] T055 [P] [US3] `backend/repository/postgres_customer_session.go` に `Revoke(ctx, id int64, at time.Time) error` を足す。`revoked_at IS NULL` の行だけ更新し、更新が 0 行なら `domain.ErrSessionNotActive` を返す
- [ ] T056 [US3] `backend/controller/customer_auth.go` に `DELETE /sessions` を足す：Cookie を読み `Revoke` を呼ぶ。成功なら Cookie を消して 204、Cookie がない・無効なら Cookie を消して 401 `unauthenticated`。ログアウトは FR-020 の差し替えの境目に含めないので、`RequireCustomer` を通さず Cookie を直接読んでよい（research R8）
- [ ] T057 [US3] `backend/main.go` に `DELETE /sessions` を登録する
- [ ] T058 [US3] `apps/web/src/lib/customer-api.ts` に `logout(): Promise<void>` を足す。204 でも 401 でも成功として扱う（spec の Edge Cases）
- [ ] T059 [US3] `apps/web/src/components/site-header.tsx` を作る（クライアントコンポーネント）：`getMe` が 200 ならログアウトボタン、401 なら「ログイン」「アカウント作成」へのリンクを出す。ログアウトを押したら `logout` の後に `/login` へ移す
- [ ] T060 [US3] `apps/web/src/app/layout.tsx` に `SiteHeader` を足す
- [ ] T061 [US3] quickstart 3 章の 6・7・12、4 章の「ログアウトの記録が残る」、6 章の 4・7 を手で確かめる

**Checkpoint**: 3 つのストーリーがすべて単独で動く

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T062 `docker compose run --rm --no-deps api go test ./...` がすべて通ることを確かめる（constitution VI。通るまで完了にしない）
- [ ] T063 [P] `backend` で `gofmt -l .` が何も出さず、`go vet ./...` が通ることを確かめる
- [ ] T064 [P] `pnpm --filter web lint` と `pnpm --filter web build` が通ることを確かめる
- [ ] T065 `docker compose run --rm sqlc generate` を流し直し、`git diff backend/repository/internal/db` に差分が出ないことを確かめる
- [ ] T066 quickstart 5 章（期限切れ）を含め、quickstart.md の手順を最初から通しで確かめる。`.env` や `jar.txt` をコミットしない
- [ ] T067 実装と設計書（`docs/design/01`・`03`・`04`、`data-model.md`、`contracts/customer-auth-api.md`）の名前と形が食い違っていないかを見直す。ずれていたら同じ変更で直す

---

## Dependencies & Execution Order

### Phase の依存

- **Setup（Phase 1）**: すぐ始められる。T001 は T036 より前に必要。T003〜T005 は実装と独立
- **Foundational（Phase 2）**: Setup の T001・T002 の後。すべてのストーリーを止める
- **US1（Phase 3）**: Foundational の後
- **US2（Phase 4）**: Foundational の後。T036（bcrypt）と T035（ポートの定義）を US1 と共有するので、US1 の後に進めるのが簡単。並行するなら T035・T036 を先に終える
- **US3（Phase 5）**: Foundational の後。ヘッダー（T059）は US1・US2 の画面があると確かめやすいが、なくても動く
- **Polish（Phase 6）**: 必要なストーリーが終わった後

### Phase 2 の中の順序

- T006 → T007 → T008 → T009 → T012（T010・T011 は T009 と並行で書ける）
- T013・T015 → T014・T016（テストを先に）
- T017 → T018、T019 → T020
- T012 → T022・T023
- T027 → T028 → T029 → T030
- T031・T032 → T033

### 同じファイルを触るタスク（並行させない）

- `backend/usecase/password_auth.go`: T035 → T046
- `backend/usecase/customer_session.go`: T018 → T054
- `backend/repository/postgres_customer.go`: T022 → T038 → T047
- `backend/repository/postgres_customer_session.go`: T023 → T055
- `backend/controller/customer_auth.go`: T039 → T048 → T056
- `backend/main.go`: T030 → T040 → T049 → T057
- `apps/web/src/lib/customer-api.ts`: T033 → T041 → T050 → T058

## Parallel Example

### Phase 2

```text
T013 domain/customer_test.go        | T015 domain/customer_session_test.go
T010 queries/customer.sql           | T011 queries/customer_session.sql
T021 auth/session_token.go          | T024 controller/error_response.go | T025 controller/session_cookie.go | T026 controller/cors.go
T031 types/customer.ts              | T032 lib/api-client.ts
```

### User Story 1

```text
T036 auth/bcrypt_hasher.go | T037 notification/noop.go     （T035 と並行）
T042 app/mypage/page.tsx   | T043 app/signup/page.tsx      （T041 の後）
```

### User Story 3

```text
T055 repository の Revoke   （T053・T054 と並行）
```

## Implementation Strategy

### MVP（US1 だけ）

1. Phase 1 → Phase 2
2. Phase 3（US1）
3. **止めて確かめる**：quickstart 3 章の 1〜5 と 6 章の 1〜2
4. この時点で「登録して自分の情報を見られる」が成り立つ

### 少しずつ足す

1. US1 → 確かめる
2. US2 → 確かめる（ログアウトがなくても、Cookie を消せばログインの確認ができる）
3. US3 → 確かめる
4. Polish でテスト・lint・quickstart を通しで確かめる
