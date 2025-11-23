FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reviewer-service ./cmd/reviewer-service

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/reviewer-service /usr/local/bin/reviewer-service
COPY --from=builder /app/internal/config /config
COPY --from=builder /app/api/openapi /app/api/openapi

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/reviewer-service", "-config", "/config/config.yaml"]
