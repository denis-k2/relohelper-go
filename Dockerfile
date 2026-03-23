FROM golang:1.26-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api

FROM alpine:3.22
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S app && adduser -S app -G app

COPY --from=builder /out/api /app/api

USER app
EXPOSE 4000 4001

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:4000/healthcheck >/dev/null || exit 1

CMD ["/app/api"]
