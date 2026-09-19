# 開発用（next dev でホットリロード）。
# ビルドコンテキストはリポジトリのルート（compose.yaml で指定）

# Node は LTS の 24 系（ホストの 22 より新しい）。pnpm 11 は Node 22 以上が必要
FROM node:24-trixie-slim

# pnpm の単体版が必要とする共有ライブラリ（slim イメージには入っていない）
RUN apt-get update \
    && apt-get install -y --no-install-recommends libatomic1 \
    && rm -rf /var/lib/apt/lists/*

# pnpm は公式イメージから単体版（dist/ を含むフォルダごと）をコピーし、バージョンを固定する
COPY --from=ghcr.io/pnpm/pnpm:11.23.0 /opt/pnpm /opt/pnpm
ENV PATH=/opt/pnpm:$PATH

WORKDIR /app

# 依存の定義だけ先にコピーする。ソースはまだコピーしないので、
# ソースだけを変えたときは下の install の層がキャッシュから再利用される
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
# ワークスペース特有の点: pnpm は pnpm-workspace.yaml の apps/* から package.json を探して
# メンバーを決める。apps/web/package.json がないと web がメンバーとして見つからず、
# エラーにならないまま web の依存（next など）が入らない（ルートの依存だけが入る）
COPY apps/web/package.json apps/web/

# パッケージ（next など）を node_modules に入れる。
# ホスト（macOS）用の node_modules は Linux では動かないため、コンテナの中で入れ直す。
# --frozen-lockfile: lockfile に記録された版だけを入れる（食い違いがあればエラー）
# --mount / --store-dir: ダウンロード済みのパッケージを次回のビルドで再利用する
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile --store-dir /pnpm/store

COPY apps/web apps/web

WORKDIR /app/apps/web
EXPOSE 3000
CMD ["pnpm", "dev"]
