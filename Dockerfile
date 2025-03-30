# Build stage
FROM debian:11 AS builder
WORKDIR /app
# 依存関係をダウンロード（go.mod, go.sumがある場合）
RUN apt-get update && apt-get upgrade -y && apt-get install -y curl gcc
COPY go.mod go.sum ./
RUN curl -OL https://go.dev/dl/go1.23.7.linux-amd64.tar.gz &&\
    tar -C /usr/local -xzf go1.23.7.linux-amd64.tar.gz &&\
    rm -rf go1.23.7.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
RUN go mod download
# ソースコードをコピー
COPY . .
# アプリケーションのビルド
RUN go build -v -o main

# Final stage: 軽量な実行環境
# FROM gcr.io/distroless/base-debian10
FROM debian:11 AS deploy
WORKDIR /
RUN apt-get update && apt-get upgrade -y
RUN apt-get install -y openssl ca-certificates
# ビルドしたバイナリをコピー
COPY --from=builder /app/main .
# Cloud Runのためにポート8080を公開
EXPOSE 8080
# コンテナ起動時に実行されるコマンド
# CMD ["./main"]
ENTRYPOINT ["./main"]
