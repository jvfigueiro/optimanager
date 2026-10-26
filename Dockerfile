# ==========================================
# Estágio 1: Build do Frontend (Vite + Vue 3)
# ==========================================
FROM node:22-alpine AS frontend-builder
WORKDIR /app/ui

COPY ui/package*.json ./
RUN npm ci

COPY ui/ ./
RUN npm run build

# ==========================================
# Estágio 2: Compilação do Binário Go Estático
# ==========================================
FROM golang:alpine AS backend-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Substitui/copia os artefatos compilados da UI para o embedding
COPY --from=frontend-builder /app/ui/dist ./ui/dist

# Compilação pura sem CGO (CGO_ENABLED=0) com stripping de debug symbols
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o optimanager ./cmd/server

# ==========================================
# Estágio 3: Imagem Final Ultraleve (~25 MB)
# ==========================================
FROM alpine:3.21

# Instala certificados TLS e fusos horários (America/Sao_Paulo)
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

WORKDIR /app

# Copia apenas o binário estático contendo o backend, banco e frontend embutidos
COPY --from=backend-builder /app/optimanager /app/optimanager

# Diretório para volume persistente do SQLite
RUN mkdir -p /data && chmod 777 /data
VOLUME ["/data"]

ENV DB_PATH=/data/optimanager.db
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/optimanager"]
