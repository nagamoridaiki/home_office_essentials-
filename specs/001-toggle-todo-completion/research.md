# Phase 0: 調査と決定

**Feature**: Todo の完了状態の切り替え | **Date**: 2026-09-20

spec.md に未確定事項（NEEDS CLARIFICATION）はなかったため、本フェーズでは技術選択の決定と
既存コードの実測結果を記録する。

## 実測: 既存コードの状態

計画の前提として、各層に何があるかを実際に確認した。

| 層 | 実測結果 | 出典 |
| --- | --- | --- |
| DB スキーマ | `completed boolean NOT NULL DEFAULT false` が存在。`todos_set_updated_at` トリガーも存在 | `backend/db/migrations/000002_create_todos.up.sql` |
| sqlc クエリ | `ListTodos` / `CreateTodo` / `DeleteTodo` のみ。更新系なし | `backend/db/queries/todo.sql` |
| ドメイン | `Todo.Completed` は存在。`TodoRepository` は `FindAll` / `Create` / `Delete` のみ | `backend/domain/todo.go` |
| ユースケース | `List` / `Create` / `Delete` のみ | `backend/usecase/todo.go` |
| HTTP | `GET /api/todos` / `POST /api/todos` / `DELETE /api/todos/{id}` のみ | `backend/controller/todo.go` |
| 画面 | `Todo` 型に `completed` はあるが、**画面に描画されていない** | `apps/web/src/app/page.tsx` |

**結論**: 完了状態は「データとしては通っているが、読み書きの経路がない」状態。
必要なのは経路の追加のみで、スキーマ変更は不要。

## 決定 1: マイグレーションを作らない

**決定**: `backend/db/migrations/` に新規ファイルを追加しない。

**根拠**: 本機能に必要な `completed` 列と、`updated_at` を更新するトリガーがすでに存在する
（上記の実測による）。追加すべきスキーマ変更がない。

**却下した代替案**: 「完了日時（`completed_at`）列を追加する」。spec.md の Assumptions で
完了日時の記録を範囲外と明記しているため、本機能では不要。将来必要になった時点で
新しい連番のマイグレーションを追加する（constitution 原則 III）。

## 決定 2: リポジトリのインターフェースの形

**決定**: `UpdateCompleted(ctx context.Context, id int64, completed bool) (Todo, error)` を
`TodoRepository` に追加し、SQL 1 本（`UPDATE ... RETURNING *`）で完結させる。

**根拠**:

- `completed` は 2 値であり、spec.md の Assumptions のとおり中間状態がない。
  ドメインが守るべき不変条件が実質存在しないため、「取得してから変更して保存する」形にしても
  現時点では守られるルールが増えない
- 既存の `Delete(ctx, id)` が同じく ID を受け取る形であり、インターフェースの対称性が保たれる
- クエリが 1 回で済み、取得と保存の間に状態が変わる余地がない

**却下した代替案**: `FindByID(ctx, id)` と `Update(ctx, todo)` を追加し、ユースケースで
「取得 → ドメインのメソッドで状態遷移 → 保存」を行う形。

- DDD としてはこちらが素直で、状態遷移をドメインが所有する
- ただし現時点で守るべき不変条件がないため、クエリが 2 回に増え、構造だけが増える
- **Todo に不変条件が増えた時点（例: 完了済みの項目は表題を変更できない）で、この形に
  切り替えることを推奨する。** その際は constitution 原則 I に沿った設計変更として扱う

## 決定 3: HTTP メソッドと経路

**決定**: `PATCH /api/todos/{id}`、リクエストボディは `{"completed": <boolean>}`。

**根拠**:

- 既存の 3 経路がリソース指向（`/api/todos`、`/api/todos/{id}`）で統一されており、その語彙に沿う
- 変更するのは資源の一部（完了状態のみ）であり、部分更新を表す PATCH が適切
- ルーティングは Go 1.22+ のメソッド付きパターンで書ける（`.claude/rules/backend.md`）

**却下した代替案**:

- `PUT /api/todos/{id}`: 資源全体の置換を意味するため、表題も送る必要が生じる。
  FR-006「完了状態の変更は表題を変えてはならない」に対して、余計な事故の余地を作る
- `POST /api/todos/{id}/complete` と `/uncomplete` の 2 経路: 経路が 2 本に増え、
  「切り替え」という 1 つの操作が分裂する。既存のリソース指向からも外れる

## 決定 4: 応答の形

**決定**: 更新後の Todo を JSON で返す（`200 OK`）。

**根拠**: 既存の `POST /api/todos` が作成後の Todo を返しており、その作法に合わせる。
画面側が応答から最新状態を得られるため、再取得の前でも表示を更新できる。

**却下した代替案**: `204 No Content`。本文がないぶん軽いが、既存の作法と異なり、
画面側が必ず再取得を要することになる。

## 決定 5: 画面の更新方式

**決定**: 切り替え操作の後に既存の `fetchTodos()` で一覧を再取得する。

**根拠**: 既存の追加・削除がどちらも「操作 → `fetchTodos()`」の形で統一されている。
同じ形にすることで、画面の状態管理の考え方を増やさない。

**却下した代替案**: 楽観的更新（応答を待たずに画面を先に変える）。操作の反応は速くなるが、
失敗時の巻き戻し処理が必要になり、spec.md の Edge Cases「画面の表示と実際に保存されている
状態が食い違ったままにならないか」に対する実装が複雑になる。単一利用者・少数件の
現状では、再取得で十分。

## 決定 6: テストの扱い

**決定**: 本機能ではテストコードを追加しない。検証は `quickstart.md` の手順による手動確認とする。

**根拠**: このリポジトリには Go・フロントエンドとも既存のテストコードが一切なく、
テスト基盤（テストランナー、DB のテスト用セットアップ）も存在しない。constitution v1.0.0 にも
テストに関する原則はない。本機能の一部としてテスト基盤を新設すると、変更範囲が
spec.md の範囲を大きく超える。

**却下した代替案**: 本機能に合わせてテスト基盤を導入する。価値はあるが、独立した課題として
扱うべきであり、この 1 機能の計画に混ぜると spec.md との対応が取れなくなる。
