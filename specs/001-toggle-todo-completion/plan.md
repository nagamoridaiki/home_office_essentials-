# Implementation Plan: Todo の完了状態の切り替え

**Branch**: `feature/issue-9-spec-kit` | **Date**: 2026-09-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-toggle-todo-completion/spec.md`

## Summary

一覧の各 Todo に完了状態の表示と切り替え手段を追加する。完了状態を表す列とその更新用トリガーは
すでに DB にあり、ドメインの型にも `Completed` が存在するため、**スキーマ変更は行わない**。
不足しているのは「更新する経路」だけなので、クエリ・リポジトリ・ユースケース・HTTP・画面の
各層に 1 本ずつ経路を通す。

既存の一覧・追加・削除は変更しない。新しい外部ライブラリも追加しない。

## Technical Context

**Language/Version**: Go 1.27.0（`backend`）／ TypeScript 5 + React 19.2.8 + Next.js 16.3.2（`apps/web`）

**Primary Dependencies**: 標準ライブラリ `net/http`、`jackc/pgx/v5` 5.11.0、Tailwind CSS 4。
コード生成は sqlc 1.31.1（コンテナ経由）。**本機能で新規の依存は追加しない**

**Storage**: PostgreSQL 18。既存の `todos` テーブルを使う（`completed boolean NOT NULL DEFAULT false`
と `todos_set_updated_at` トリガーは既存。**マイグレーションは不要**）

**Testing**: このリポジトリには現時点でテストコードが存在しない（Go・フロントとも）。
constitution にもテストに関する原則はない。本機能でもテスト基盤の新設は行わず、
検証は quickstart.md の手順による手動確認とする

**Target Platform**: Docker Compose（`web` / `api` / `db`）によるローカル開発環境

**Project Type**: Web application（フロントエンドとバックエンドを分離したモノレポ）

**Performance Goals**: 既存の一覧・追加・削除と同等。本機能固有の性能目標は設けない

**Constraints**: 既存 API の応答形を壊さない（FR-008 / FR-009）。外部ライブラリを増やさない

**Scale/Scope**: 単一利用者。手で変更するのは 6 ファイル（SQL 1・ドメイン 1・ユースケース 1・
リポジトリ実装 1・コントローラ 1・画面 1）。これに加えて sqlc の生成物が再生成で更新される

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` v1.0.0 に照らして評価する。

| 原則 | 評価 | 根拠 |
|---|---|---|
| I. 依存方向とドメインの独立 | ✅ 通過 | リポジトリ IF はドメイン側（`backend/domain/todo.go`）に追加し、実装はインフラ側（`backend/repository/`）に置く。ドメインは HTTP と DB を知らないまま。ハンドラにはビジネスルールを置かず、変換のみ |
| II. ユビキタス言語と境界 | ✅ 通過 | 「完了 / completed」を DB 列・ドメインの項目・API の項目・画面表示で同一の語に統一する。契約を増やすため、フロントエンドの型と呼び出しを同じ変更でそろえる |
| III. 生成物と手書きコードの分離 | ✅ 通過 | SQL を `backend/db/queries/todo.sql` に追記し、`backend/repository/internal/db/` は sqlc の再生成で更新する（手で編集しない）。既存マイグレーションは書き換えず、**新規マイグレーションも作らない**（必要な列とトリガーが既存のため） |
| IV. 再現可能な開発環境 | ✅ 通過 | sqlc はコンテナ経由（`docker compose run --rm sqlc generate`）で実行する。新規の外部ライブラリを追加せず、HTTP は標準ライブラリのまま |
| V. 秘密情報の取り扱い | ✅ 通過 | 本機能は秘密情報を扱わない。設定ファイルの追加も行わない |

**判定**: 違反なし。Phase 0 に進む。

### Post-Design Re-check

Phase 1（data-model / contracts / quickstart）の設計後に再評価した。

- 新たに `PATCH /api/todos/{id}` を追加するが、既存 3 経路の形と応答は変更していない（原則 II）
- `data-model.md` で定めた状態遷移はドメイン層に閉じ、HTTP の語彙を含まない（原則 I）
- 追加する SQL は 1 本で、生成物は再生成のみ（原則 III）

**判定**: 設計後も違反なし。

## Project Structure

### Documentation (this feature)

```text
specs/001-toggle-todo-completion/
├── plan.md              # このファイル
├── research.md          # Phase 0 の調査と決定
├── data-model.md        # Phase 1 のデータモデルと状態遷移
├── quickstart.md        # Phase 1 の動作確認手順
├── contracts/           # Phase 1 の API 契約
│   └── patch-todo-completed.md
├── checklists/
│   └── requirements.md  # /speckit-specify が生成した品質チェックリスト
└── spec.md              # 仕様
```

### Source Code (repository root)

```text
backend/
├── db/
│   └── queries/
│       └── todo.sql                  # 変更: 完了状態を更新するクエリを追記
├── domain/
│   └── todo.go                       # 変更: 状態遷移とリポジトリ IF を追加
├── usecase/
│   └── todo.go                       # 変更: 完了状態を切り替えるユースケースを追加
├── repository/
│   ├── postgres_todo.go              # 変更: 追加した IF を実装
│   └── internal/db/                  # 変更: sqlc の再生成（手で編集しない）
└── controller/
    └── todo.go                       # 変更: PATCH の経路とリクエスト/レスポンス変換を追加

apps/web/src/app/
└── page.tsx                          # 変更: 完了状態の表示と切り替え操作を追加
```

**Structure Decision**: 既存のディレクトリ構成をそのまま使う。新しいパッケージもディレクトリも
作らない。既存の 4 層（domain / usecase / repository / controller）に 1 本ずつ経路を足すだけで、
本機能に必要な範囲は満たせる。`backend/db/migrations/` には**何も追加しない**。

## Complexity Tracking

> Constitution Check に違反がないため、記入なし。

本機能で既存の構成を逸脱する点はない。新規パッケージ・新規ライブラリ・新規マイグレーションの
いずれも発生しない。
