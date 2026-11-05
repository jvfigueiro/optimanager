#!/usr/bin/env bash
set -euo pipefail

# scripts/init-env.sh
# Initializes a secure .env file with a cryptographically generated JWT_SECRET
# without overwriting an existing .env file.

ENV_FILE=".env"
ENV_EXAMPLE=".env.example"

if [ -f "$ENV_FILE" ]; then
    echo "[init-env] Arquivo $ENV_FILE já existe. Nenhuma alteração foi realizada para preservar as configurações atuais."
    exit 0
fi

echo "[init-env] Criando $ENV_FILE a partir de $ENV_EXAMPLE..."

if command -v openssl >/dev/null 2>&1; then
    NEW_SECRET=$(openssl rand -hex 32)
else
    NEW_SECRET=$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')
fi

if [ -f "$ENV_EXAMPLE" ]; then
    sed "s|^JWT_SECRET=.*|JWT_SECRET=${NEW_SECRET}|" "$ENV_EXAMPLE" > "$ENV_FILE"
else
    cat <<EOF > "$ENV_FILE"
PORT=8080
DB_PATH=/data/optimanager.db
TZ=America/Sao_Paulo
JWT_SECRET=${NEW_SECRET}
EOF
fi

chmod 600 "$ENV_FILE"
echo "[init-env] Sucesso! Arquivo $ENV_FILE gerado com segredo HMAC seguro de 64 caracteres hexadecimais."

