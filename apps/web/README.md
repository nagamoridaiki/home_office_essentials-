# web（Next.js）

フロントエンド。起動・環境構築の手順はリポジトリ直下の `README.md` を正とする。

## Docker を使わずに起動する場合

パッケージマネージャは **pnpm** のみ使う（npm / yarn / bun は使わない）。リポジトリ直下で実行する。

```bash
pnpm install
pnpm --filter web dev
```

http://localhost:3000 で確認できる。
