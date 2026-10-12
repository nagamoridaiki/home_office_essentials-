# Research: 顧客のアカウント作成・ログイン・ログアウト（自前認証）

**Feature**: [spec.md](./spec.md) | **Date**: 2026-10-12

Technical Context に残った未決事項を、ここで 1 つずつ決める。各項目は「決めたこと / 理由 / 検討した別案」で書く。

## R1. パスワードの保存方式

- **決めたこと**: bcrypt（`golang.org/x/crypto/bcrypt`）。コストは 12
- **理由**:
  - 標準ライブラリにはパスワード向けのハッシュ関数がない（`crypto/sha256` は速すぎて総当たりに弱い）。constitution IV の「標準ライブラリで足りるか」を確かめたうえで、Go チームが保守する `x/crypto` を足す
  - 将来 IdP に移るとき、bcrypt は取り込み・移行の対応が広い（#14 の論点）
  - spec の上限 72 バイト（FR-002）は bcrypt の入力上限に合わせた値
- **検討した別案**: argon2id（同じ `x/crypto` にある）。強度は上だが、移行先の IdP が対応しているか未確認のため見送る

## R2. ログイン状態を表す値（セッショントークン）

- **決めたこと**:
  - `crypto/rand` で 32 バイトを作り、base64url（パディングなし）にした文字列を Cookie に入れる
  - DB の `customer_sessions.token_hash` には SHA-256 の 16 進文字列だけを保存する
- **理由**: 32 バイトの乱数は推測できないので、照合用のハッシュは速い SHA-256 で足りる（bcrypt は不要）。DB が漏れても Cookie の値は復元できない（FR-018・SC-004）
- **検討した別案**: JWT を Cookie に入れる方式。ログアウトの即時失効（FR-011）に DB 照合が結局必要になるため採らない

## R3. Cookie の属性

- **決めたこと**:

  | 属性 | 値 |
  | --- | --- |
  | 名前 | `customer_session` |
  | `HttpOnly` | 付ける（FR-019） |
  | `Secure` | 既定で付ける。環境変数 `SESSION_COOKIE_SECURE=false` で外せる |
  | `SameSite` | `Lax` |
  | `Path` | `/` |
  | `Max-Age` | 2592000（30 日。FR-008） |

- **理由**:
  - `04` の共通規約（`HttpOnly` / `Secure` / `SameSite=Lax`）どおり
  - Chrome・Firefox は `http://localhost` を安全な接続とみなし、`Secure` 付きの Cookie を受け取る。ブラウザによっては受け取らないことがあるため、ローカルで困ったときだけ外せる逃げ道を環境変数で用意する
  - Cookie はポートで区別されないため、`localhost:8080` が発行した Cookie は `localhost:3000` からの `fetch`（`credentials: 'include'`）でも送られる
- **検討した別案**: `SameSite=Strict`。外部サイトのリンクから来たときにログアウト扱いに見えるため採らない

## R4. CORS と CSRF

- **決めたこと**:
  - `controller/cors.go` に `Access-Control-Allow-Credentials: true` を足す。許可するオリジンは今のまま `http://localhost:3000` の 1 つだけ（`*` は使わない）
  - 状態を変える `POST` は `Content-Type: application/json` 以外を 415 で拒否する
- **理由**:
  - JSON の `POST` と `DELETE` は、ブラウザが事前確認（プリフライト）を送る。CORS で許可したオリジン以外からは本体が送られない
  - `SameSite=Lax` と合わせて、別サイトからの書き換え要求を防げる
- **検討した別案**: CSRF トークン。ローカル用の MVP では上の 2 つで足りる。stg・本番でドメインが分かれたときに #14 で見直す

## R5. メールアドレスの扱い

- **決めたこと**:
  - 前後の空白を取り除き、すべて小文字にしてから保存・照合する
  - 形式は `net/mail.ParseAddress` で確かめる。表示名付き（`"Taro" <a@b>`）の入力は拒否する。長さは 254 文字まで
- **理由**:
  - 保存する値を小文字にそろえれば、`03` の `customers.email` の UNIQUE がそのまま大文字・小文字を無視した重複判定になる（FR-003・SC-006）
  - 関数インデックス（`lower(email)`）を足す設計変更が要らない
- **検討した別案**: 入力どおりに保存し、`UNIQUE (lower(email))` を張る。`03` の制約を変えることになるので採らない。ローカル部の大文字・小文字の違いが失われる点は、EC では実害がないとして受け入れる

## R6. メールアドレスの重複と二重送信

- **決めたこと**:
  - 重複は DB の UNIQUE 違反（SQLSTATE `23505`）で判定し、リポジトリで `domain.ErrEmailAlreadyUsed` に変換する
  - 事前の存在確認（SELECT）はしない
  - 2 回目の要求も同じ違反になり、409 を返す（Clarifications Q2）
- **理由**: 同時に 2 件届いても、DB の制約だけが確実に 1 件に絞れる。事前確認をすると、確認と登録の間に別の要求が入る隙ができる
- **`04` の共通規約との関係**:
  - `04` の「顧客の POST は二重送信でも同じ結果を返す」に、`POST /customers` は例外として 2 回目を 409 にする、と書き足す（tasks で設計書を直す）
  - アカウントが 2 つ作られない点は共通規約と同じ

## R7. ログイン失敗時の見分けにくさ（FR-007）

- **決めたこと**:
  - メールアドレスが未登録のときも、形式が不正なときも、パスワードが違うときも、同じエラー（`401` と同じメッセージ）を返す
  - 未登録のときも、起動時に作ったダミーのハッシュと照合し、bcrypt の照合を 1 回行う
- **理由**: 照合を省くと応答時間の差で登録の有無が分かってしまう。メッセージをそろえるだけでは足りない
- **検討した別案**: なし。アカウント作成では登録の有無を伝える（Clarifications Q3）ため、ここで完全に隠しきれるわけではないが、FR-007 のとおりログインでは見分けにくくしておく

## R8. 認証方式を差し替えられる境界（FR-020）

- **決めたこと**:
  - 変わらない境目を「**認証ミドルウェアがリクエストから顧客 ID を取り出し、それより内側は顧客 ID だけを使う**」に置く
  - ミドルウェアは controller 側のインターフェース `CustomerAuthenticator`（`AuthenticateRequest(r *http.Request) (customerID int64, err error)`）にだけ依存する
  - 今回の実装 `CookieSessionAuthenticator` は、Cookie を読んで `CustomerSessionUsecase.Authenticate` を呼ぶ
  - ユースケースは次の 3 つに分ける

  | ユースケース | 役割 | #14 で Cookie セッションを続ける場合 | #14 で JWT を毎回検証する場合 |
  | --- | --- | --- | --- |
  | `PasswordAuthUsecase` | アカウント作成・ログインで、お客様が誰かを確かめて顧客 ID を得る | OIDC の callback 用のユースケースに置き換える | 使わなくなる（登録・ログインは IdP が行う）。`(iss, sub)` で顧客を引き、いなければ作るユースケースを新しく作る |
  | `CustomerSessionUsecase` | セッションの発行（`Issue`）／Cookie の値から顧客 ID を引く（`Authenticate`）／ログアウト（`Revoke`） | そのまま使う | 使わなくなる。`CustomerAuthenticator` を JWT を検証する実装（`Authorization: Bearer` を読む）に差し替える |
  | `CustomerUsecase` | 自分の情報を返す | そのまま使う | そのまま使う |

- 決めたことの続き:
  - パスワードの方針（長さ）とハッシュ化は `PasswordAuthUsecase` とその周りのポート（`PasswordHasher`・`PasswordCredentialStore`）だけが知る
  - ドメイン層は顧客 ID・メールアドレス・セッションだけを知る（`01` の「ドメイン層は顧客 ID しか知らず、パスワードと認証方式はインフラ層に閉じる」）
  - `/me` 配下のハンドラは、ミドルウェアがコンテキストに入れた顧客 ID だけを読む。Cookie もセッションも直接触らない
- **理由**:
  - #14 は Cookie セッションと JWT のどちらにするか未定。どちらに決まっても、差し替えるのは `CustomerAuthenticator` の実装と登録・ログインの部分だけで、`/me` 配下のハンドラと `CustomerUsecase` には手が入らない
  - JWT 方式ではログアウトの即時失効（FR-011）が保てないため、ログアウトは変わらない境目に含めない（spec の FR-020）
- **検討した別案**:
  - 1 つの `AuthUsecase` にまとめる。差し替えるときに、使い続ける部分（自分の情報）まで巻き込むため採らない
  - ミドルウェアが `CustomerSessionUsecase` を直接呼ぶ。JWT 方式に移るときにミドルウェアごと書き直すことになるため採らない

## R9. トランザクション境界

- **決めたこと**:
  - `customers` と `customer_credentials` の作成は、リポジトリの 1 メソッド（`PasswordCredentialStore.CreateCustomerWithPassword`）の中で 1 つのトランザクションにする
  - セッションの作成は、その後の別の処理にする
- **理由**:
  - 顧客と認証情報が片方だけできる状態を作らない
  - セッションの作成に失敗しても、アカウントはできているので、お客様はログインし直せる。これを 1 つにまとめるほどの利点はない
- **検討した別案**: ユースケース層でトランザクションを開く仕組み（Unit of Work）を作る。今は 1 か所でしか使わないので作らない

## R10. 確認メールの口（spec の Assumptions）

- **決めたこと**:
  - ユースケース側にポート `CustomerNotifier`（`NotifyRegistered`）を置く
  - 実装は何もしない `NoopCustomerNotifier` にする
  - 通知に失敗しても登録は失敗させない（ログに残すだけ）
- **理由**: 送信の実装を後から差し込めるようにする。spec で「口だけ用意する」と決めている

## R11. エラー応答の形と API のパス

- **決めたこと**:
  - 新しい API は `04` の共通規約に従う。パスに `/api` を付けない（`/customers`・`/sessions`・`/me`）
  - エラーは `{"error":{"code","message","details"}}` で返す
  - 既存の `/api/todos` は変えない
- **理由**: `04` が契約の正。Todo は EC の機能に置き換わる予定の試作なので、合わせ直さない

## R12. 時刻の扱い

- **決めたこと**: ユースケースに `func() time.Time`（今の時刻を返す関数）を渡す。本番は `time.Now`、テストでは固定の時刻を渡す
- **理由**: 30 日の期限切れ（FR-008）を、テストで実際に待たずに確かめられる

## R13. 画面と移動（FR-009a）

- **決めたこと**:
  - 画面は `/signup`（アカウント作成）・`/login`（ログイン）・`/mypage`（自分の情報）の 3 つ
  - 各画面を開いたときに `GET /me` を呼ぶ
    - `/signup`・`/login` は 200 なら `/mypage` へ移す
    - `/mypage` は 401 なら `/login` へ移す
  - ヘッダーにログアウトのボタンを置く
  - API の呼び出しは `apps/web/src/lib/` の API クライアントに集める
- **理由**: 今の Todo 画面と同じくクライアント側で描く方式にそろえる。サーバー側でのリダイレクト（Next.js の proxy）はまだ導入していないので、今回は持ち込まない
- **検討した別案**: Next.js のサーバー側で Cookie を読んで移す。Cookie が `localhost` で共有されるので可能だが、セッションの照合をフロントのサーバーでも行う二重の仕組みになるため見送る
