# All-binaries image used on EC2 (api, worker, migrate share the same artifact).
FROM golang:1.25-alpine AS deps
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

FROM deps AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api      ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/worker   ./cmd/worker && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/migrate  ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /out/api     .
COPY --from=builder /out/worker  .
COPY --from=builder /out/migrate .
COPY --from=builder /app/migrations ./migrations
USER appuser
EXPOSE 8080
# Default command runs the API; docker-compose.ec2.yaml overrides `command` per service.
CMD ["./api"]
