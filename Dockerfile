FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o inteam-http ./cmd/server

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/inteam-http /app/inteam-http
COPY internal/frontend /app/internal/frontend
COPY migrations /app/migrations

EXPOSE 8080

ENTRYPOINT ["/app/inteam-http"]

