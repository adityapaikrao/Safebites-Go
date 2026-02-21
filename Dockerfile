FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20

WORKDIR /app
RUN adduser -D -g '' appuser

COPY --from=builder /out/server /app/server
COPY migrations /app/migrations

EXPOSE 8080
USER appuser

ENTRYPOINT ["/app/server"]
