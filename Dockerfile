# ==========================================
# Stage 1: Frontend Build (Vite + Vue 3)
# ==========================================
FROM node:22-alpine AS frontend-builder
WORKDIR /app/ui

COPY ui/package*.json ./
RUN npm ci

COPY ui/ ./
RUN npm run build

# ==========================================
# Stage 2: Static Go Binary Compilation
# ==========================================
FROM golang:alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Copy compiled UI assets for embedding
COPY --from=frontend-builder /app/ui/dist ./ui/dist

# Pure compilation without CGO (CGO_ENABLED=0) with debug symbols stripped
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o optimanager ./cmd/server

# ==========================================
# Stage 3: Ultralight Final Image (~25 MB)
# ==========================================
FROM alpine:3.21

# Install TLS certificates and time zone data (America/Sao_Paulo)
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

WORKDIR /app

# Copy static binary containing embedded backend, database migrations, and frontend
COPY --from=backend-builder /app/optimanager /app/optimanager

# Persistent volume directory for SQLite
RUN mkdir -p /data && chmod 777 /data
VOLUME ["/data"]

ENV DB_PATH=/data/optimanager.db
ENV PORT=8585

EXPOSE 8585

ENTRYPOINT ["/app/optimanager"]
