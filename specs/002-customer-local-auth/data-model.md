# Data Model: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Feature**: [spec.md](./spec.md) | **Date**: 2026-10-12

constitution 原則 I に従い、API 契約（`contracts/`）より先にドメインモデルを決める。
テーブルの正は `docs/design/03-logical-data-model.md` の 4.1 で、ここではそれをどう層に分けて持つかを書く。

## 1. 層ごとの置き場所

| 層 | パッケージ | 置くもの |
| --- | --- | --- |
| ドメイン | `backend/domain` | `Customer`、`Email`、`CustomerSession`、`SessionTokenHash`、`CustomerRepository`、`CustomerSessionRepository`、ドメインのエラー |
| アプリケーション | `backend/usecase` | `PasswordAuthUsecase`、`CustomerSessionUsecase`、`CustomerUsecase`、パスワードの方針、ポート（`PasswordHasher`・`PasswordCredentialStore`・`SessionTokenGenerator`・`CustomerNotifier`） |
| インフラ | `backend/repository` | 3 テーブルのリポジトリ実装（sqlc の生成コードはこの中だけで使う） |
| インフラ | `backend/auth` | bcrypt のハッシュ化、セッショントークンの生成とハッシュ化 |
| インフラ | `backend/notification` | 何もしない `NoopCustomerNotifier` |
| インターフェース | `backend/controller` | ハンドラ、認証ミドルウェアと `CustomerAuthenticator`（IF と Cookie セッションの実装）、Cookie、エラー応答 |

パスワードとそのハッシュは、ドメイン層に入れない（`01` の 1.2「ドメイン層は顧客 ID しか知らず、パスワードと認証方式はインフラ層に閉じる」、research R8）。

## 2. ドメイン（`backend/domain`）

### 2.1 `Email`（値オブジェクト）

| 項目 | 内容 |
| --- | --- |
| 生成 | `NewEmail(raw string) (Email, error)` |
| 正規化 | 前後の空白を取り除き、小文字にする（R5） |
| 検証 | `net/mail.ParseAddress` で解釈でき、表示名がなく、解釈結果が入力と一致し、254 文字以下 |
| エラー | `ErrInvalidEmail` |

### 2.2 `Customer`（エンティティ）

| 項目 | 型 | 説明 |
| --- | --- | --- |
| `ID` | `int64` | DB が採番 |
| `Email` | `Email` | 必須。顧客ごとに一意 |
| `FamilyName` / `GivenName` | `*string` | 未登録なら `nil` |
| `FamilyNameKana` / `GivenNameKana` | `*string` | 未登録なら `nil` |
| `PhoneNumber` | `*string` | 未登録なら `nil` |

- 本機能では氏名・フリガナ・電話番号を書き込まない（spec の Assumptions）。読み出して `nil` かどうかで「未登録」を表す（FR-014）
- 生成関数は持たない。作成はパスワードの登録と一体で行うため、`PasswordCredentialStore` が `Customer` を返す（R9）

### 2.3 `SessionTokenHash`（値オブジェクト）

- Cookie の値を SHA-256 にかけた 16 進文字列。生の値はドメインに持ち込まない（FR-018）
- 中身の作り方はインフラ（`backend/auth`）が決める。ドメインは比べるだけ

### 2.4 `CustomerSession`（エンティティ）

| 項目 | 型 | 説明 |
| --- | --- | --- |
| `ID` | `int64` | DB が採番 |
| `CustomerID` | `int64` | 持ち主 |
| `TokenHash` | `SessionTokenHash` | 照合に使う |
| `ExpiresAt` | `time.Time` | 作成時刻 + 30 日。延長しない（FR-008、Clarifications Q5） |
| `RevokedAt` | `*time.Time` | ログアウトした時刻。`nil` なら未ログアウト |
| `CreatedAt` | `time.Time` | 作成時刻（FR-013 の記録） |

| 振る舞い | 内容 |
| --- | --- |
| `NewCustomerSession(customerID, tokenHash, now)` | `ExpiresAt = now + SessionLifetime`（`SessionLifetime = 30 * 24h`） |
| `IsActive(now) bool` | `RevokedAt == nil` かつ `now < ExpiresAt` |
| `Revoke(now) error` | 有効なら `RevokedAt = now`。すでにログアウト済みまたは期限切れなら `ErrSessionNotActive` |

**状態**（`02-state-transitions.md` にはない。セッションの寿命なので本機能の中だけで定義する）

| 遷移元 | 遷移先 | きっかけ | 許可 |
| --- | --- | --- | --- |
| 有効 | ログアウト済み | ログアウト（`Revoke`） | 許可 |
| 有効 | 期限切れ | 時間の経過（`now >= ExpiresAt`） | 許可（保存はしない。`IsActive` で判定） |
| ログアウト済み | 有効 | — | **禁止**（FR-011。`RevokedAt` を戻す操作を持たない） |
| 期限切れ | 有効 | — | **禁止**（延長しない） |
| ログアウト済み・期限切れ | ログアウト済み | 再度のログアウト | **禁止**（`ErrSessionNotActive`） |

### 2.5 リポジトリのインターフェース

```text
CustomerRepository
  FindByID(ctx, id int64) (Customer, error)            // なければ ErrCustomerNotFound

CustomerSessionRepository
  Create(ctx, s CustomerSession) (CustomerSession, error)
  FindByTokenHash(ctx, h SessionTokenHash) (CustomerSession, error)  // なければ ErrSessionNotFound
  Revoke(ctx, id int64, at time.Time) error            // revoked_at IS NULL の行だけ更新する
```

### 2.6 ドメインのエラー

| エラー | 意味 |
| --- | --- |
| `ErrInvalidEmail` | メールアドレスの形式が不正 |
| `ErrEmailAlreadyUsed` | メールアドレスがすでに使われている |
| `ErrCustomerNotFound` | 顧客がいない |
| `ErrSessionNotFound` | セッションがない |
| `ErrSessionNotActive` | セッションがログアウト済みか期限切れ |

## 3. アプリケーション（`backend/usecase`）

### 3.1 パスワードの方針

| 項目 | 内容 |
| --- | --- |
| 検証 | `ValidatePassword(raw string) error`。8 文字（rune）以上、72 バイト以下 |
| エラー | `ErrPasswordTooShort`、`ErrPasswordTooLong` |

### 3.2 ポート

```text
PasswordHasher
  Hash(password string) (string, error)
  Compare(hash, password string) error    // 違えば ErrPasswordMismatch
  DummyCompare(password string)           // 未登録時に時間をそろえるため（R7）

PasswordCredentialStore                   // customers と customer_credentials をまたぐ
  CreateCustomerWithPassword(ctx, email domain.Email, passwordHash string) (domain.Customer, error)
      // 1 トランザクション。UNIQUE 違反は domain.ErrEmailAlreadyUsed
  FindPasswordHashByEmail(ctx, email domain.Email) (customerID int64, passwordHash string, err error)
      // なければ domain.ErrCustomerNotFound

SessionTokenGenerator
  Generate() (raw string, hash domain.SessionTokenHash, err error)
  Hash(raw string) domain.SessionTokenHash

CustomerNotifier
  NotifyRegistered(ctx, c domain.Customer) error
```

`PasswordCredentialStore` はパスワードのハッシュを扱うため、ドメインではなくアプリケーション側に置く（`.claude/rules/backend.md`「リポジトリ IF はドメイン（またはアプリケーション）側」）。

### 3.3 ユースケース

| ユースケース | メソッド | 入力 | 出力 | エラー |
| --- | --- | --- | --- | --- |
| `PasswordAuthUsecase` | `SignUp` | メールアドレス、パスワード | `Customer` | `ValidationError`（項目ごとの理由）、`domain.ErrEmailAlreadyUsed` |
| 〃 | `Login` | メールアドレス、パスワード | 顧客 ID | `ErrInvalidCredentials`（理由を区別しない） |
| `CustomerSessionUsecase` | `Issue` | 顧客 ID | 生のトークン、`CustomerSession` | — |
| 〃 | `Authenticate` | 生のトークン | 顧客 ID | `ErrUnauthenticated` |
| 〃 | `Revoke` | 生のトークン | — | `ErrUnauthenticated` |
| `CustomerUsecase` | `GetMe` | 顧客 ID | `Customer` | `domain.ErrCustomerNotFound` |

- **`SignUp`**
  1. メールアドレスとパスワードを両方検証し、問題を `ValidationError` にまとめて返す（FR-002）
  2. ハッシュ化する
  3. `CreateCustomerWithPassword` で保存する
  4. `NotifyRegistered` を呼ぶ。失敗してもエラーにしない（R10）
  5. セッションの発行はハンドラが `CustomerSessionUsecase.Issue` を続けて呼ぶ（R8・R9）
- **`Login`**
  - メールアドレスの形式が不正、または未登録なら、`DummyCompare` を呼んだうえで `ErrInvalidCredentials` を返す
  - ハッシュが一致しなければ `ErrInvalidCredentials` を返す（FR-007・R7）
- **`Authenticate`**
  - トークンをハッシュにして引き、`IsActive(now)` でなければ `ErrUnauthenticated` を返す（FR-011・FR-015）
  - ないとき・ログアウト済み・期限切れを区別しない
- **`Revoke`**
  - トークンから自分のセッションだけを `Revoke` する。他の端末のセッションには触れない（FR-012）
  - 有効なセッションがなければ `ErrUnauthenticated` を返す

## 4. 認証ミドルウェアの境目（`backend/controller`）

認証方式を差し替えても変わらない境目（research R8、spec FR-020）。

```text
CustomerAuthenticator
  AuthenticateRequest(r *http.Request) (customerID int64, err error)   // 特定できなければ usecase.ErrUnauthenticated

CookieSessionAuthenticator（今回の実装）
  Cookie customer_session を読む → CustomerSessionUsecase.Authenticate(raw) → 顧客 ID

RequireCustomer(a CustomerAuthenticator) func(http.Handler) http.Handler
  成功：顧客 ID をリクエストのコンテキストに入れて次へ
  失敗：401 unauthenticated
```

- `CustomerAuthenticator` は HTTP のリクエストを受け取るため、ユースケース層ではなく controller 層に置く
- `/me` 配下のハンドラは、コンテキストの顧客 ID だけを読む（`CustomerIDFromContext`）
- #14 で JWT を毎回検証する方式にしたときは、`Authorization: Bearer` を読んで JWT を検証する実装に差し替える。`RequireCustomer` と `/me` 配下のハンドラは変えない

## 5. テーブル（`backend/db/migrations`）

`03` の 4.1 どおり。新しいマイグレーションを 3 本作る（`000003`〜`000005`）。共通の列とトリガーは `.claude/rules/database.md` に従う。

| テーブル | 列 | 制約・インデックス |
| --- | --- | --- |
| `customers` | `email text NOT NULL`、氏名・フリガナ・`phone_number`（`text` NULL 可） | `UNIQUE (email)` |
| `customer_credentials` | `customer_id bigint NOT NULL`、`password_hash text`、`auth_issuer text`、`auth_subject text` | `UNIQUE (customer_id)`（Clarifications Q1）、`UNIQUE (auth_issuer, auth_subject)`、FK → `customers.id` `ON DELETE RESTRICT` |
| `customer_sessions` | `customer_id bigint NOT NULL`、`token_hash text NOT NULL`、`expires_at timestamptz NOT NULL`、`revoked_at timestamptz` | `UNIQUE (token_hash)`、FK → `customers.id` `ON DELETE RESTRICT`、`customer_id` にインデックス |

- `email` には小文字にそろえた値だけを入れる（R5）。DB 側で小文字を強制する `CHECK (email = lower(email))` も付ける
  - `03` にない制約なので、`03` の 5 章に書き足す（tasks で設計書を直す）
- `auth_issuer` / `auth_subject` は本機能では常に NULL（spec の Assumptions）
- セッションの行は消さない。ログアウトは `revoked_at` を埋める（FR-013）

## 6. 不変条件への影響（`03` の 6 章）

| 不変条件 | 影響 | 守り方 |
| --- | --- | --- |
| 1 点ものが二重に売れない | なし | — |
| 在庫ステータスが不正に遷移しない | なし | — |
| 他人の個人データに URL を書き換えても到達できない | **あり** | `/me` は URL に顧客の識別子を持たず、認証ミドルウェアがセッションから引いた顧客 ID だけを使う（FR-016）。ミドルウェアで一括で強制し、ハンドラごとに判断しない。顧客のセッションは管理者のものと別テーブル |
