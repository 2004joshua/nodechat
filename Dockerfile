# Stage 1: Build Go binary with CGO enabled
FROM golang:1.21-alpine AS builder
# Install build tools and SQLite C headers for go-sqlite3
RUN apk add --no-cache gcc musl-dev sqlite-dev ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /src/cmd/peer
RUN go build -o /app/nodechat

# Stage 2: Build React UI
FROM node:18-alpine AS ui-builder
WORKDIR /ui
COPY ui/package*.json ./
RUN npm ci --prefer-offline --no-audit --progress=false
COPY ui .
RUN npm run build

# Stage 3: Final minimal runtime image
FROM alpine:3.18
# Install SSL certs
RUN apk add --no-cache ca-certificates
WORKDIR /app

# Copy the Go server binary and React UI artifacts
COPY --from=builder /app/nodechat ./nodechat
COPY --from=ui-builder /ui/build ./ui/build

# Create persistent storage directories
RUN mkdir -p uploads databases

# Expose P2P and HTTP ports
EXPOSE 9000 8080

# Command entrypoint: pass flags at runtime
ENTRYPOINT ["/app/nodechat"]
# No default CMD so you must supply --port, --api-port, --username, [--connect]
