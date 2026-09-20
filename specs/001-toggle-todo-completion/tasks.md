# Tasks: Todo の完了状態の切り替え

**Feature**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Date**: 2026-09-20

## Format: `[ID] [P?] [Story] Description`

- `[P]` は他のタスクと並行して進められるもの（別ファイル・未完了タスクへの依存なし）
- `[US1]` / `[US2]` は spec.md のユーザーストーリーとの対応
- テストのタスクは生成していない。research.md の決定 6 のとおり、このリポジトリには
  テスト基盤が存在せず、本機能でも新設しないため

## Path Conventions

| 記号 | 実際のパス |
| --- | --- |
| クエリ | `backend/db/queries/todo.sql` |
| 生成物 | `backend/repository/internal/db/`（**手で編集しない**） |
| ドメイン | `backend/domain/todo.go` |
| ユースケース | `backend/usecase/todo.go` |
| リポジトリ実装 | `backend/repository/postgres_todo.go` |
| コントローラ | `backend/controller/todo.go` |
| 画面 | `apps/web/src/app/page.tsx` |

## Phase 1: Setup (Shared Infrastructure)

- [ ] T001 開発環境を起動し、着手前の状態を記録する（`docker compose up -d` の後、既存の一覧・追加・削除が動くことと `docker compose run --rm migrate version` が `2` であることを確認）

> 新しい依存の追加もプロジェクトの初期化も不要。既存の構成に経路を足すだけのため、Setup はこの 1 件のみ。

## Phase 2: Foundational (Blocking Prerequisites)

**このフェーズは US1 と US2 の両方が依存する。** 完了状態の更新は真偽値 1 つを受け取る 1 本の経路で
扱う設計（research.md の決定 3）のため、バックエンドの経路は 2 つのストーリーで共有される。

- [ ] T002 [P] 完了状態を更新するクエリ `UpdateTodoCompleted`（`UPDATE ... RETURNING *`）を `backend/db/queries/todo.sql` に追記する
- [ ] T003 [P] `TodoRepository` に `UpdateCompleted(ctx, id, completed) (Todo, error)` を追加し、対象が見つからない場合の誤り（`ErrTodoNotFound`）を `backend/domain/todo.go` に定義する
- [ ] T004 sqlc を再生成する（`docker compose run --rm sqlc generate`）。`backend/repository/internal/db/` が更新される。生成物は手で編集しない（T002 に依存）
- [ ] T005 `UpdateCompleted` を `backend/repository/postgres_todo.go` に実装する。更新行が 0 件のときは `ErrTodoNotFound` に変換する（T003・T004 に依存）
- [ ] T006 [P] 完了状態を切り替えるユースケースを `backend/usecase/todo.go` に追加する（T003 に依存）
- [ ] T007 `PATCH /api/todos/{id}` の経路とリクエスト・レスポンス変換を `backend/controller/todo.go` に追加する。`ErrTodoNotFound` は 404、ボディや ID の解釈失敗は 400 に対応づける（T005・T006 に依存）

> `backend/db/migrations/` には**何も追加しない**。`completed` 列と `todos_set_updated_at`
> トリガーは既存のため（plan.md / research.md の決定 1）。
> CORS も変更不要（`backend/controller/cors.go` は `PATCH` と `Content-Type` を許可済み）。

**チェックポイント**: `quickstart.md` の「API の確認」の手順が最後まで通ること。この時点で
画面はまだ変わっていない。

## Phase 3: User Story 1 - 終わった作業を完了として記録する (Priority: P1) 🎯 MVP

**ゴール**: 利用者が一覧上で項目を完了にでき、完了していることが見て分かる。

**独立した検証条件**: 未完了の項目を 1 つ完了にし、表示が変わること、および画面を再読み込み
しても完了のままであることを確認できる（spec.md の US1）。

- [ ] T008 [US1] `Todo` 型の `completed` を使い、一覧の各項目に完了・未完了の視覚的な区別を `apps/web/src/app/page.tsx` に追加する（FR-001）
- [ ] T009 [US1] 項目を完了にする操作を `apps/web/src/app/page.tsx` に追加し、成功後は既存の `fetchTodos()` で一覧を再取得する（FR-002 / FR-004、research.md の決定 5）
- [ ] T010 [US1] `quickstart.md` の「画面の確認」のうち、完了にする操作と再読み込み後の保持を確認する（SC-001 / SC-002 / SC-003）

**チェックポイント**: この時点で US1 は単独で成立し、MVP として提供できる。

## Phase 4: User Story 2 - 誤って完了にした項目を戻す (Priority: P2)

**ゴール**: 完了にした項目を未完了に戻せる。

**独立した検証条件**: 完了済みの項目を 1 つ未完了に戻し、表示が変わること、および再読み込み後も
未完了のままであることを確認できる（spec.md の US2）。

> **実装の大半は US1 と共有される。** 切り替えを 1 つの操作で双方向に扱う設計のため、
> T009 の操作がそのまま「戻す」側にも働く。したがって本フェーズは、双方向に動くことの確認と、
> 契約で定めた冪等性の担保が中心になる。水増しのタスクは作らない。

- [ ] T011 [US2] 完了済みの項目を未完了に戻す経路が `apps/web/src/app/page.tsx` で動くことを確認し、片方向にしか働かない実装になっていれば双方向に直す（FR-003）
- [ ] T012 [US2] 同じ `completed` の値を続けて送っても `200 OK` で同じ状態が返ること（冪等性）を確認する。`contracts/patch-todo-completed.md` の「冪等性」に対応
- [ ] T013 [US2] `quickstart.md` の「未完了に戻す」手順を実行して確認する

**チェックポイント**: 完了と未完了を何度往復しても、最後の操作どおりの状態になる。

## Phase 5: Polish & Cross-Cutting Concerns

- [ ] T014 [P] 存在しない ID への `PATCH` が `404` を返し、一覧の内容が変わらないことを確認する（FR-005）
- [ ] T015 [P] 完了状態の変更で表題が変わらないこと、一覧の並び順が変わらないことを確認する（FR-006 / FR-007）
- [ ] T016 [P] 既存の一覧・追加・削除が着手前と同じ手順で動くことを確認する（FR-008 / FR-009 / SC-004）
- [ ] T017 [P] `docker compose run --rm migrate version` が T001 で記録した値（`2`）のままであることを確認する（マイグレーションを追加していないことの裏取り）
- [ ] T018 [P] `README.md` の API の記述に `PATCH /api/todos/{id}` を追記する
- [ ] T019 `quickstart.md` の手順を最初から通して実行し、すべて期待どおりであることを確認する

## Dependencies & Execution Order

### Phase Dependencies

```text
Phase 1 (Setup)
   ↓
Phase 2 (Foundational) ← US1 と US2 の両方がここに依存する
   ↓
Phase 3 (US1 / P1) ── MVP はここまで
   ↓
Phase 4 (US2 / P2)
   ↓
Phase 5 (Polish)
```

### User Story Dependencies

- **US1**: Phase 2 の完了に依存。US2 には依存しない
- **US2**: Phase 2 に依存。加えて、画面の操作を US1（T009）と共有するため、実務上は US1 の後に行う

通常は各ストーリーが独立するが、本機能は「1 つの真偽値を双方向に切り替える」という性質上、
US2 の実装が US1 に含まれる。これは設計上の選択の結果であり、spec.md の分割が誤っていたわけではない
（利用者から見れば「完了にする」と「戻す」は別の価値である）。

### Within Each User Story

- Phase 2 の T002 と T003 は別ファイルのため並行可能。T004 は T002 の後
- Phase 3 の T008 と T009 は同一ファイル（`page.tsx`）のため直列
- Phase 5 はすべて確認作業で、互いに独立

### Parallel Opportunities

- **Phase 2 の開始時**: T002（SQL）と T003（ドメイン）を同時に着手できる
- **Phase 2 の中盤**: T005（リポジトリ実装）と T006（ユースケース）は別ファイルで、
  どちらも T003 の後に着手できる
- **Phase 5**: T014〜T018 の 5 件は互いに独立

## Parallel Example: Phase 2

```text
# 別ファイルのため同時に着手できる:
T002  backend/db/queries/todo.sql
T003  backend/domain/todo.go

# T003 の完了後、こちらも同時に着手できる:
T005  backend/repository/postgres_todo.go   （T004 の生成物も必要）
T006  backend/usecase/todo.go
```

## Implementation Strategy

### MVP First (User Story 1 Only)

Phase 1 → Phase 2 → Phase 3 まで（T001〜T010）。この時点で「終わった作業を完了として記録する」
という中心的な価値が提供できる。US2 以降を行わなくても、利用者にとって成立した機能になる。

### Incremental Delivery

1. Phase 2 の完了時点で、API としては完成している（`curl` で双方向に切り替えられる）
2. Phase 3 の完了時点で、画面から使える（MVP）
3. Phase 4 で戻す操作の確実性を担保
4. Phase 5 で既存機能への影響がないことを確認

### Parallel Team Strategy

本機能は変更量が小さく（手で触るのは 6 ファイル）、分担の利点より調整の手間が上回る。
1 人が Phase 順に進めることを推奨する。

## Notes

- テストのタスクは意図的に生成していない。理由は research.md の決定 6 を参照
- `backend/repository/internal/db/` は sqlc の生成物であり、T004 の再生成以外で触らない
  （constitution 原則 III）
- 各フェーズのチェックポイントを通過してから次へ進むこと。特に Phase 2 の
  チェックポイント（`quickstart.md` の API 確認）は、画面の作業に入る前の分岐点になる
