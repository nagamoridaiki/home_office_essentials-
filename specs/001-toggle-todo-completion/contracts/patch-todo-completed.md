# API 契約: Todo の完了状態を更新する

**Feature**: Todo の完了状態の切り替え | **Date**: 2026-09-20

## 新規に追加する経路

```text
PATCH /api/todos/{id}
```

既存の 3 経路（`GET /api/todos`・`POST /api/todos`・`DELETE /api/todos/{id}`）は
**一切変更しない**（FR-008 / FR-009）。

## リクエスト

| 項目 | 位置 | 型 | 必須 | 説明 |
| --- | --- | --- | --- | --- |
| `id` | パス | 整数 | ✅ | 更新対象の Todo の識別子 |
| `completed` | ボディ | 真偽値 | ✅ | 更新後の完了状態 |

`Content-Type: application/json`

```json
{ "completed": true }
```

表題は受け取らない。完了状態の変更で表題が変わることはない（FR-006）。

## 応答

### 200 OK

更新後の Todo を返す。形は既存の `POST /api/todos` の応答と同じ。

```json
{ "id": 1, "title": "牛乳を買う", "completed": true }
```

### 400 Bad Request

- `id` が整数として解釈できない
- ボディが JSON として解釈できない、または `completed` が真偽値でない

### 404 Not Found

指定された `id` の Todo が存在しない（FR-005）。この場合、一覧の内容は変わらない。

### 500 Internal Server Error

データベースへの接続や更新に失敗した。

## 冪等性

同じ `completed` の値で複数回呼び出しても結果は変わらない。すでに完了の項目に
`{"completed": true}` を送った場合もエラーとせず、`200 OK` でその状態を返す。

## CORS

`PATCH` は単純リクエストではないため、ブラウザは事前に `OPTIONS` による事前確認を送る。

既存の `backend/controller/cors.go` を実測したところ、**CORS の変更は不要**である。

| 設定 | 現在の値 | 本機能への十分性 |
| --- | --- | --- |
| `Access-Control-Allow-Origin` | `http://localhost:3000` | ✅ フロントのオリジンに限定されている |
| `Access-Control-Allow-Methods` | `GET, POST, PATCH, DELETE, OPTIONS` | ✅ `PATCH` を含む |
| `Access-Control-Allow-Headers` | `Content-Type` | ✅ JSON ボディの送信に足りる |
| `OPTIONS` の扱い | `200 OK` を返して終了 | ✅ 事前確認に応答する |

## 既存契約への影響

| 経路 | 変更 |
| --- | --- |
| `GET /api/todos` | なし。応答にはすでに `completed` が含まれている |
| `POST /api/todos` | なし |
| `DELETE /api/todos/{id}` | なし |
