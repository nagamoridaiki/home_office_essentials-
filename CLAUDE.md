# Home Office Essentials

フロント（Next.js）とバック（Go）を分離したモノレポ。  

## 技術スタック

- Frontend: Next.js 16 / React 19 / TypeScript / Tailwind CSS 4（`apps/web`）
- Backend: Go 1.27（`backend`）— 詳細は `.claude/rules/backend.md`
- パッケージマネージャ: pnpm 11（ワークスペース）

## アーキテクチャ方針（必須）

Spec Kit 向けに同じ方針を `.specify/memory/constitution.md` にも書いている。原則を変えるときは両方を同じコミットで直す（食い違ったらこのファイルが優先）。

クリーンアーキテクチャと DDD を前提にする。特定の画面・エンティティ・API 形にルールを固定しない。

- **依存は外 → 内**。内側（ドメイン）は外側（UI・配送手段・永続化など）を知らない
- 用語はユビキタス言語に揃える。コード名・API・UI 文言を勝手に別名にしない（実際の対応表は `docs/design/01-ubiquitous-language.md`）
- 境界づけられたコンテキストを意識する。またぐときは明示的な連携を検討する
- 新機能は「いまの構成の踏襲」より「依存方向と境界が保たれるか」を優先する
- 試作が方針に合わないときは、無理に温存せず境界に沿って切り直してよい
- ドメイン固有の詳細は実装と会話で決め、このファイルに製品仕様を固定しない

バックエンドの層分け・HTTP・Go 慣例は `.claude/rules/backend.md`、フロントエンドは `.claude/rules/frontend.md` に書く。

## 仕様駆動開発（Spec Kit）

機能の設計〜実装は `speckit-*` スキル（specify → plan → tasks → implement）で進める。成果物は `specs/` 配下に置く。

## 設計書（必須）

EC ドメインを実装する前に `docs/design/README.md` を読む。索引なので、全部読まず必要な章だけ開く。

- 用語（名前の正）: `docs/design/01-ubiquitous-language.md`
- ステータスと許可・禁止の遷移: `docs/design/02-state-transitions.md`
- テーブル・カラム・制約: `docs/design/03-logical-data-model.md`
- エンドポイントの契約: `docs/design/04-api-list.md`

不変条件（1 点ものの二重販売を防ぐ・在庫ステータスを不正に遷移させない・他人のデータに URL から到達させない）を壊す実装をしない。一覧と守り方は `docs/design/03-logical-data-model.md` の 6 章。

## 共通ルール

- パッケージマネージャは **pnpm** を使う（npm / yarn は使わない）
- シークレットや `.env*` をコミットしない
- ドメイン層・ユースケース層を変えたらテストを付け、`go test ./...` を通してから完了とする（詳細は `.claude/rules/testing.md`）
- ユーザーが明示しない限り git commit / push しない

## 関連

- パス別ルール: `.claude/rules/`
