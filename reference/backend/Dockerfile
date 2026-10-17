# ---- ビルド ----
# go.mod の go 行以上のバージョンを指定する
FROM golang:1.27-alpine AS build
WORKDIR /src

# 依存だけ先に取得（ソースが変わってもキャッシュが効く）
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO_ENABLED=0: Cに依存しない単独の実行ファイルにする（modernc.org/sqlite は Pure Go）
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- 実行 ----
FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir /data && chown app /data
USER app
WORKDIR /app
COPY --from=build /out/server /app/server

ENV PORT=8080 \
    DB_PATH=/data/todo.db
EXPOSE 8080
CMD ["/app/server"]
