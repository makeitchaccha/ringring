# Stage 1: Build the Go app
FROM golang:1.23 AS builder

RUN apt update && apt install -y make && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CGO_ENABLED=1
RUN go build -ldflags="-s -w" -o ringring cmd/ringring/main.go

# Stage 2: Create a small image with the Go binary
FROM debian:stable-slim AS runner
WORKDIR /app
RUN apt update
RUN apt install -y sqlite3 ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/build/ringring ./ringring
COPY --from=builder /build/build/deploy ./deploy
COPY --from=builder /build/locales ./locales

# Command to run the executable
CMD [ "/app/ringring" ]