FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/job-api ./cmd/server


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
 && addgroup -S app -g 10001 \
 && adduser  -S app -u 10001 -G app

WORKDIR /app

COPY --from=build /out/job-api ./job-api
COPY migrations ./migrations

RUN mkdir -p /app/logs && chown -R app:app /app

USER app:app

EXPOSE 3000

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- "http://127.0.0.1:${PORT:-3000}/healthz" >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/job-api"]
