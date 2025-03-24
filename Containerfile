# ────── Stage 1: Build React frontend ──────
FROM node:18-alpine AS ui-builder
WORKDIR /app/ui

# Copy only package manifests first → speeds rebuilds
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

# Copy the rest of the React source and compile it
COPY ui/ ./
RUN npm run build
# Result: production-ready static files in /app/ui/build

# ────── Stage 2: Build Go backend ──────
FROM golang:1.20-alpine AS go-builder
WORKDIR /app

# Copy Go modules, download deps
COPY go.mod go.sum ./
RUN go mod download

# Copy full source and compile the Go binary
COPY . ./
RUN go build -o nodechat cmd/peer/main.go

# ────── Stage 3: Final runtime image ──────
FROM alpine:latest
WORKDIR /app

# Copy only the built artifacts (no build tools)
COPY --from=go-builder /app/nodechat ./nodechat
COPY --from=ui-builder /app/ui/build ./ui/build

# Expose port for HTTP + WebSocket
EXPOSE 8080

# Default command: run your Go server, which will serve the React build
CMD ["./nodechat"]
