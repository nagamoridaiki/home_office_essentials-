---
paths:
  - "apps/web/**"
---

# Frontend（Next.js）

## 技術前提

- Next.js 16（App Router、`apps/web/src/app`）/ React 19 / TypeScript / Tailwind CSS 4
- コマンドはリポジトリ直下から `pnpm --filter web <script>` で実行する
- `any` 型は使わない。API のレスポンスにも型を付ける

## 責務の置き場所

- 画面（`app/`）は表示と入力の受け取りに集中させる。ビジネスルールを書かない（判断はバックエンドのドメイン層が持つ）
- バックエンドとの通信は API クライアントに集約し、コンポーネントから直接 `fetch` を散らさない
- 型・変数名・UI 文言はユビキタス言語（`docs/design/01-ubiquitous-language.md`）に揃える

## 変更時

- API の契約（`docs/design/04-api-list.md`）を変えるときは、バックエンドのハンドラと API クライアント／型をセットで直す
- `package.json` や設定ファイルを変えたら `docker compose up -d --build web` で作り直す
