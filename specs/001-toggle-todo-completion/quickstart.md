# 動作確認手順

**Feature**: Todo の完了状態の切り替え | **Date**: 2026-09-20

このリポジトリにはテストコードがないため（research.md の決定 6）、以下の手順で手動確認する。

## 準備

```bash
docker compose up -d
```

DB が未作成の場合は先にマイグレーションを適用する（本機能で新しいマイグレーションは増えない）。

```bash
docker compose run --rm migrate up
```

## スキーマ変更なしの確認

本機能ではマイグレーションを追加していないため、version が変わらないことを確認する。

```bash
docker compose run --rm migrate version
```

→ 本機能の実装前後で同じ値（`2`）のままであること。

## API の確認

確認用の Todo を 1 つ作る。

```bash
curl -s -X POST http://localhost:8080/api/todos -H 'Content-Type: application/json' -d '{"title":"動作確認"}'
```

→ `{"id":<N>,"title":"動作確認","completed":false}` が返る。この `<N>` を以降で使う。

完了にする（US1 / FR-002）。

```bash
curl -s -X PATCH http://localhost:8080/api/todos/<N> -H 'Content-Type: application/json' -d '{"completed":true}'
```

→ `{"id":<N>,"title":"動作確認","completed":true}` が返る。

保存されているか、一覧を取り直して確認する（FR-004）。

```bash
curl -s http://localhost:8080/api/todos
```

→ `<N>` の項目が `"completed":true` になっている。

未完了に戻す（US2 / FR-003）。

```bash
curl -s -X PATCH http://localhost:8080/api/todos/<N> -H 'Content-Type: application/json' -d '{"completed":false}'
```

→ `"completed":false` が返る。

存在しない ID を指定する（FR-005）。

```bash
curl -s -o /dev/null -w '%{http_code}\n' -X PATCH http://localhost:8080/api/todos/999999 -H 'Content-Type: application/json' -d '{"completed":true}'
```

→ `404`

表題が変わっていないことを確認する（FR-006）。

```bash
curl -s http://localhost:8080/api/todos
```

→ `<N>` の `title` が `"動作確認"` のままであること。

## 画面の確認

http://localhost:3000 を開く。

| 確認項目 | 期待 | 対応 |
| --- | --- | --- |
| 完了かどうかが見て分かる | 完了と未完了が視覚的に区別されている | FR-001 / SC-003 |
| 1 回の操作で切り替わる | 項目の操作 1 回で状態が変わる | SC-001 |
| 再読み込みで保たれる | ページを再読み込みしても状態が同じ | FR-004 / SC-002 |
| 並び順が変わらない | 切り替えても一覧の順序が変わらない | FR-007 |
| 既存操作が壊れていない | 追加フォームと削除ボタンが従来どおり動く | FR-008 / SC-004 |

## 後片付け

```bash
curl -s -X DELETE http://localhost:8080/api/todos/<N>
```
