FROM golang:1.27
LABEL MAINTAINER="ryota.kashio"

RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl bash unzip \
    && rm -rf /var/lib/apt/lists/*

# air をインストール
RUN go install github.com/air-verse/air@latest

# bun をインストール（.air.toml の pre_cmd で使用）
RUN curl -fsSL https://bun.sh/install | bash
ENV PATH="/root/.bun/bin:$PATH"

WORKDIR /app/frontend
COPY frontend/package.json frontend/*.lock /app/frontend/
RUN bun install

WORKDIR /app

# Go モジュールをビルド時にキャッシュ（ボリュームマウント対象の /app 外に保存されるため上書きされない）
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

EXPOSE 1323

# ソースはボリュームマウントで提供、air でホットリロード起動
CMD ["air", "-c", ".air.toml"]
