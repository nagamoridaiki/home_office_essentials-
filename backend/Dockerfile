# 開発用（air でホットリロード）。本番用のマルチステージビルドは別 Issue
FROM golang:1.27-trixie

RUN go install github.com/air-verse/air@v1.67.4

WORKDIR /app

# 依存の取得を先に行い、ソースだけの変更ではキャッシュを使えるようにする
COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

# .air.toml があれば自動で読み込む（なければ既定値）
CMD ["air"]
