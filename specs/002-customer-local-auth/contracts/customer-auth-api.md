# API 契約: 顧客のアカウント作成・ログイン・ログアウト

**Feature**: [spec.md](../spec.md) | **Date**: 2026-10-12

`docs/design/04-api-list.md` の `POST /customers`・`POST /sessions`・`DELETE /sessions`・`GET /me` を、本機能で実装する形まで細かくしたもの。共通規約（パスにプレフィックスなし、`snake_case`、エラー形式）は `04` の 1 章に従う。

## 共通

### セッション Cookie

| 属性 | 値 |
| --- | --- |
| 名前 | `customer_session` |
| 値 | 32 バイトの乱数を base64url にした文字列 |
| 属性 | `HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=2592000` |

- `Secure` は環境変数 `SESSION_COOKIE_SECURE=false` のときだけ外す（research R3）
- 消すときは同じ名前・`Path`・`Max-Age=0` で上書きする

### CORS

- `Access-Control-Allow-Origin: http://localhost:3000`
- `Access-Control-Allow-Credentials: true`

フロントは `fetch(..., { credentials: "include" })` で呼ぶ。

### エラー形式

```json
{
  "error": {
    "code": "validation_failed",
    "message": "入力内容を確認してください",
    "details": [
      { "field": "email", "code": "invalid_email", "message": "メールアドレスの形式が正しくありません" }
    ]
  }
}
```

`details` は検証エラーのときだけ入れ、それ以外は空配列にする。

| `code` | ステータス | 使う場面 |
| --- | --- | --- |
| `invalid_request` | 400 | JSON として読めない |
| `unsupported_media_type` | 415 | `POST` の `Content-Type` が `application/json` でない |
| `validation_failed` | 422 | 入力の検証に失敗（`details` に項目ごとの理由） |
| `email_already_used` | 409 | メールアドレスがすでに使われている |
| `invalid_credentials` | 401 | ログインに失敗（理由を区別しない） |
| `unauthenticated` | 401 | ログインしていない、またはセッションが無効 |
| `internal_error` | 500 | 想定外のエラー |

`details[].code` の値：

| `field` | `code` | `message` |
| --- | --- | --- |
| `email` | `required` | メールアドレスを入力してください |
| `email` | `invalid_email` | メールアドレスの形式が正しくありません |
| `password` | `required` | パスワードを入力してください |
| `password` | `too_short` | パスワードは 8 文字以上で入力してください |
| `password` | `too_long` | パスワードが長すぎます |

### 自分の情報（`Customer`）の形

```json
{
  "email": "taro@example.com",
  "family_name": null,
  "given_name": null,
  "family_name_kana": null,
  "given_name_kana": null,
  "phone_number": null
}
```

- 未登録の項目は `null` を返す（FR-014）
- 顧客の `id` は返さない（`03` の 1 章「顧客向けに連番を出さない」）

---

## `POST /customers` — アカウントを作る

**認証**: 不要

### リクエスト

```json
{ "email": "Taro@example.com", "password": "correct-horse" }
```

| 項目 | 型 | 必須 | 検証 |
| --- | --- | --- | --- |
| `email` | string | ✅ | 前後の空白を除き小文字にしてから、形式を確かめる（254 文字以下） |
| `password` | string | ✅ | 8 文字以上、72 バイト以下 |

### 応答

| ステータス | 本文 | 副作用 |
| --- | --- | --- |
| **201** | 自分の情報（`email` は小文字にそろえた値） | `customers` と `customer_credentials` を作る。`customer_sessions` を作り Cookie を発行する。確認メールの口を呼ぶ（送信はしない） |
| 409 `email_already_used` | エラー | なし。**2 回目の同じ要求もこれになる**（Clarifications Q2・research R6） |
| 422 `validation_failed` | エラー（両方の項目の問題をまとめて返す） | なし |
| 400 / 415 | エラー | なし |

---

## `POST /sessions` — ログインする

**認証**: 不要

### リクエスト

```json
{ "email": "taro@example.com", "password": "correct-horse" }
```

### 応答

| ステータス | 本文 | 副作用 |
| --- | --- | --- |
| **200** | 自分の情報 | `customer_sessions` を作り Cookie を発行する。他の端末のセッションはそのまま（FR-009） |
| 401 `invalid_credentials` | エラー。`message` は「メールアドレスまたはパスワードが正しくありません」 | なし |
| 400 / 415 | エラー | なし |

- メールアドレスの形式が不正なときも、未登録のときも、パスワードが違うときも、401 `invalid_credentials` を返す（FR-007）。422 は返さない

---

## `DELETE /sessions` — ログアウトする

**認証**: 顧客（Cookie）

### 応答

| ステータス | 本文 | 副作用 |
| --- | --- | --- |
| **204** | なし | 今の Cookie のセッションの `revoked_at` を埋める。Cookie を消す |
| 401 `unauthenticated` | エラー | Cookie を消す（期限切れ・ログアウト済みの Cookie を残さない） |

- 他の端末のセッションには触れない（FR-012）
- フロントは 204 でも 401 でも「ログアウトした状態」として画面を切り替える（spec の Edge Cases）

---

## `GET /me` — 自分の情報を返す

**認証**: 顧客（Cookie）

### 応答

| ステータス | 本文 |
| --- | --- |
| **200** | 自分の情報 |
| 401 `unauthenticated` | エラー |

- どの顧客を返すかは、認証ミドルウェアがセッションから引いた顧客 ID だけで決める。リクエストに顧客の識別子を受け取らない（FR-016）

---

## 本機能で変えないもの

- `GET /api/todos`・`POST /api/todos`・`DELETE /api/todos/{id}`
- `PATCH /me`（プロフィール編集。別の機能）
