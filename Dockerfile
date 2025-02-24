# Stage 1: Build the Go app
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CGO_ENABLED=1
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build -ldflags="-s -w" -o ringring cmd/ringring/main.go

# Stage 2: Create a small image with the Go binary
FROM debian:stable-slim AS runner
WORKDIR /app
RUN apt update
RUN apt install -y sqlite3 ca-certificates
COPY --from=builder /app/ringring ./ringring
COPY --from=builder /app/locales ./locales

# Command to run the executable
ENTRYPOINT [ "/app/ringring" ]
