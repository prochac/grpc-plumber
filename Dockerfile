FROM golang:1.25-bookworm as builder

COPY . /app

WORKDIR /app

RUN go mod download

RUN go build -o build/ ./cmd/...

FROM debian:bookworm

RUN apt update && apt install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && update-ca-certificates

COPY --from=builder /app/build/server /app/build/server

WORKDIR /app

CMD ["./server"]