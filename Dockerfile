# Go 1.21を使用した開発環境
FROM golang:1.21-alpine AS development

# 必要なパッケージをインストール
RUN apk add --no-cache git make

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesをキャッシュするため、先にgo.modとgo.sumをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# 本番用ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係をコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 本番用の軽量イメージ
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# タイムゾーンを日本に設定
ENV TZ=Asia/Tokyo

CMD ["./main"]