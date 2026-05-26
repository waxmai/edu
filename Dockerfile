# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o edu-schedule-system main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/edu-schedule-system .
COPY --from=builder /app/migrate .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations
RUN chown -R app:app /app
USER app

EXPOSE 9999
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:9999/system/ready || exit 1
CMD ["./edu-schedule-system", "-env", "pro", "-config", "/app/configs/pro_configs.toml"]
