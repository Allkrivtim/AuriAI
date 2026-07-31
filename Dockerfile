# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

# слой зависимостей — инвалидируется только при изменении go.mod/go.sum
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# исходники
COPY . .

# сборка с кэшем компилятора Go и модулей
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -o /app/bin ./cmd/server

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /server
COPY --from=builder /app/bin /usr/local/bin/server
ENTRYPOINT ["/usr/local/bin/server"]