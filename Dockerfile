FROM golang:1.22-bookworm as builder

COPY . /app

WORKDIR /app

RUN go mod download

RUN go build -o main .

FROM debian:bookworm

COPY --from=builder /app/main /app/main

WORKDIR /app

CMD ["./main"]