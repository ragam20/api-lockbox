FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN \
  CGO_ENABLED=0 GOOS=linux go build -o lockbox_app ./cmd/api/main.go && \
  CGO_ENABLED=0 GOOS=linux go build -o lockbox_migrate ./cmd/migrate/main.go

FROM alpine:3
WORKDIR /app
COPY --from=builder /app/lockbox_app .
COPY --from=builder /app/lockbox_migrate .
COPY migrations ./migrations

EXPOSE 8000

CMD [ "/app/lockbox_app" ]
