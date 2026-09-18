# Home Office Essentials

フロント（Next.js）とバック（Go）を分離したモノレポ。  

## 技術スタック

- Frontend: Next.js 16 / React 19 / TypeScript / Tailwind CSS 4（`apps/web`）
- Backend: Go 1.27（`backend`）— 詳細は `.claude/rules/backend.md`
- パッケージマネージャ: pnpm 11（ワークスペース）

## コマンド

```bash
# フロント + バックを同時起動
pnpm dev
# Frontend: http://localhost:3000
# Backend:  http://localhost:8080

# フロントのみ
pnpm --filter web dev

# バックのみ
cd backend && go run .
```

## アーキテクチャ方針（必須）

クリーンアーキテクチャと DDD を前提にする。特定の画面・エンティティ・API 形にルールを固定しない。

- **依存は外 → 内**。内側（ドメイン）は外側（UI・配送手段・永続化など）を知らない
- 用語はユビキタス言語に揃える。コード名・API・UI 文言を勝手に別名にしない
- 境界づけられたコンテキストを意識する。またぐときは明示的な連携を検討する
- 新機能は「いまの構成の踏襲」より「依存方向と境界が保たれるか」を優先する
- 試作が方針に合わないときは、無理に温存せず境界に沿って切り直してよい
- ドメイン固有の詳細は実装と会話で決め、このファイルに製品仕様を固定しない

バックエンドの層分け・HTTP・Go 慣例は `.claude/rules/backend.md` に書く。

## 共通ルール

- パッケージマネージャは **pnpm** を使う（npm / yarn は使わない）
- シークレットや `.env*` をコミットしない
- ユーザーが明示しない限り git commit / push しない

## 関連

- パス別ルール: `.claude/rules/`
