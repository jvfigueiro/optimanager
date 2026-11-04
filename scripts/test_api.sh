#!/usr/bin/env bash
set -e

PORT=8099
export PORT
export DB_PATH="./data/test_optimanager.db"

# Limpa banco de teste anterior se existir
rm -f "$DB_PATH" "${DB_PATH}-wal" "${DB_PATH}-shm"

echo "=== 1. Iniciando servidor OptiManager na porta $PORT ==="
./server &
SERVER_PID=$!

cleanup() {
    echo ""
    echo "=== Encerrando servidor (PID $SERVER_PID) ==="
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    rm -f "$DB_PATH" "${DB_PATH}-wal" "${DB_PATH}-shm"
}
trap cleanup EXIT

sleep 1

BASE_URL="http://127.0.0.1:$PORT/api"

echo "=== 2. Testando Health Check ==="
HEALTH=$(curl -s "$BASE_URL/health")
echo "Health: $HEALTH"
echo "$HEALTH" | grep -q '"status":"ok"' || { echo "Falha no health check"; exit 1; }

echo ""
echo "=== 3. Testando Status de Autenticação Inicial (Primeiro Acesso) ==="
AUTH_STATUS=$(curl -s "$BASE_URL/auth/status")
echo "Auth Status: $AUTH_STATUS"
echo "$AUTH_STATUS" | grep -q '"needs_setup":true' || { echo "ERRO: needs_setup deveria ser true"; exit 1; }

echo ""
echo "=== 3.1. Configurando Administrador Inicial (/api/auth/setup) ==="
SETUP_RESP=$(curl -s -X POST "$BASE_URL/auth/setup" \
  -H "Content-Type: application/json" \
  -d '{"nome":"João Administrador","email":"admin@optimanager.com.br","senha":"password123"}')
echo "Setup Admin: $SETUP_RESP"
TOKEN=$(echo "$SETUP_RESP" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "ERRO: Token não retornado no setup"
    exit 1
fi
AUTH_HEADER="Authorization: Bearer $TOKEN"

echo ""
echo "=== 3.2. Testando Login (/api/auth/login) e Gestão de Usuários (/api/usuarios) ==="
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@optimanager.com.br","senha":"password123"}')
echo "$LOGIN_RESP" | grep -q '"email":"admin@optimanager.com.br"' || { echo "ERRO: Falha no login"; exit 1; }

NOVO_USER=$(curl -s -X POST "$BASE_URL/usuarios" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"nome":"Técnico Operador","email":"tecnico@optimanager.com.br","senha":"senha123","role":"operador"}')
echo "Novo Usuário: $NOVO_USER"
echo "$NOVO_USER" | grep -q "tecnico@optimanager.com.br" || { echo "ERRO ao cadastrar novo usuário"; exit 1; }

echo ""
echo "=== 4. Consultando Configuração Padrão ==="
CONFIG=$(curl -s "$BASE_URL/config")
echo "Config: $CONFIG"
echo "$CONFIG" | grep -q "38.136.905/0001-66" || { echo "Falha na config inicial"; exit 1; }

echo ""
echo "=== 5. Métricas Iniciais do Dashboard ==="
DASH_INIT=$(curl -s -H "$AUTH_HEADER" "$BASE_URL/dashboard")
echo "Dashboard inicial: $DASH_INIT"

echo ""
echo "=== 6. Criando OS com cliente NOVO (Auto-cadastro + Atalho VT) ==="
NOVA_OS=$(curl -s -X POST "$BASE_URL/ordens" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{
    "cliente_nome": "Dra. Carolina Mendes",
    "cliente_contato": "(55) 99887-6655",
    "cliente_doc": "012.345.678-99",
    "servico": "Manutenção preventiva em notebook",
    "is_vt": true,
    "valor_total": 0,
    "status": "Concluído",
    "pagamento_status": "Pendente"
  }')

echo "OS Criada: $NOVA_OS"
OS_ID=$(echo "$NOVA_OS" | grep -o '"id":[0-9]*' | head -n1 | cut -d: -f2)
echo "ID da OS: $OS_ID"

# Valida se o ID gerado é >= 1000
if [ "$OS_ID" -lt 1000 ]; then
    echo "ERRO: O ID da primeira OS deveria iniciar em 1000 ou superior. Recebido: $OS_ID"
    exit 1
fi

# Valida regras de Visita Técnica (VT): prefixo 'VT - ', valor default R$ 80, equipamento 'N/A'
echo "$NOVA_OS" | grep -q '"equipamento":"N/A"' || { echo "ERRO: Equipamento não foi definido como N/A"; exit 1; }
echo "$NOVA_OS" | grep -q 'VT - ' || { echo "ERRO: Serviço não recebeu o prefixo VT -"; exit 1; }
echo "$NOVA_OS" | grep -q '"valor_total":80' || { echo "ERRO: Valor total não adotou o padrão VT de 80.00"; exit 1; }
echo "$NOVA_OS" | grep -q 'https://wa.me/5555998876655' || { echo "ERRO: WhatsApp URL não gerada corretamente"; exit 1; }

echo ""
echo "=== 7. Validando Auto-cadastro do Cliente na base ==="
CLIENTES=$(curl -s -H "$AUTH_HEADER" "$BASE_URL/clientes?search=Carolina")
echo "Busca de clientes: $CLIENTES"
echo "$CLIENTES" | grep -q "Dra. Carolina Mendes" || { echo "ERRO: Cliente Dra. Carolina Mendes não foi cadastrado automaticamente"; exit 1; }

echo ""
echo "=== 8. Verificando reflexo no Dashboard (Pendente a Receber) ==="
DASH_PENDENTE=$(curl -s -H "$AUTH_HEADER" "$BASE_URL/dashboard")
echo "Dashboard com pendência: $DASH_PENDENTE"
echo "$DASH_PENDENTE" | grep -q '"pendente_receber":80' || { echo "ERRO: pendente_receber deveria ser 80"; exit 1; }
echo "$DASH_PENDENTE" | grep -q '"qtd_pendentes":1' || { echo "ERRO: qtd_pendentes deveria ser 1"; exit 1; }

echo ""
echo "=== 9. Baixando pagamento da OS via PATCH /api/ordens/$OS_ID/baixa ==="
BAIXA_RESP=$(curl -s -X PATCH "$BASE_URL/ordens/$OS_ID/baixa" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"forma":"Pix","status":"Pago"}')
echo "Resposta Baixa: $BAIXA_RESP"
echo "$BAIXA_RESP" | grep -q '"pagamento_status":"Pago"' || { echo "ERRO: status do pagamento não mudou para Pago"; exit 1; }

echo ""
echo "=== 10. Conferindo reflexo da baixa no Caixa do Dashboard ==="
DASH_QUITADO=$(curl -s -H "$AUTH_HEADER" "$BASE_URL/dashboard")
echo "Dashboard após quitação: $DASH_QUITADO"
echo "$DASH_QUITADO" | grep -q '"caixa_recebido":80' || { echo "ERRO: caixa_recebido deveria ser 80 após quitação"; exit 1; }
echo "$DASH_QUITADO" | grep -q '"pendente_receber":0' || { echo "ERRO: pendente_receber deveria ser 0 após quitação"; exit 1; }
echo "$DASH_QUITADO" | grep -q '"qtd_pendentes":0' || { echo "ERRO: qtd_pendentes deveria ser 0"; exit 1; }

echo ""
echo "=== 11. Testando Atalhos Rápidos ==="
NOVO_ATALHO=$(curl -s -X POST "$BASE_URL/atalhos" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d '{"titulo":"Banco Inter PJ","descricao":"Internet Banking","url":"https://inter.co","ordem":5}')
echo "Novo Atalho: $NOVO_ATALHO"
echo "$NOVO_ATALHO" | grep -q "Banco Inter PJ" || { echo "ERRO ao criar atalho"; exit 1; }

echo ""
echo "=== 12. Testando Geração de PDFs Nativos via Maroto (/api/ordens/$OS_ID/pdf) ==="
for TIPO in "relatorio" "orcamento" "recibo"; do
    PDF_HEADER=$(curl -s -i -H "$AUTH_HEADER" "$BASE_URL/ordens/$OS_ID/pdf?tipo=$TIPO" | head -n 8)
    echo "Cabeçalho PDF ($TIPO):"
    echo "$PDF_HEADER" | grep -i "Content-Type: application/pdf" || { echo "ERRO: Tipo de conteúdo não é application/pdf para $TIPO"; exit 1; }
    echo "$PDF_HEADER" | grep -i "HTTP/1.1 200 OK" || { echo "ERRO: Status não é 200 para $TIPO"; exit 1; }
done

echo ""
echo "=== 13. Testando Entrega do Frontend SPA Embutido (//go:embed) ==="
ROOT_HTML=$(curl -s "http://127.0.0.1:$PORT/")
echo "$ROOT_HTML" | grep -q "OptiManager 3.0" || { echo "ERRO: HTML raiz do SPA não encontrado"; exit 1; }
echo "HTML do SPA entregue com sucesso!"

echo ""
echo "=== 14. Medindo Uso de Memória RAM (RSS) ==="
ps -o pid,user,%cpu,%mem,vsz,rss,comm -p $SERVER_PID
RSS_KB=$(ps -o rss= -p $SERVER_PID | tr -d ' ')
RSS_MB=$((RSS_KB / 1024))
echo "Consumo total de RAM: ~${RSS_MB} MB (Meta: < 40 MB)"
if [ "$RSS_MB" -gt 40 ]; then
    echo "AVISO: Consumo ultrapassou a meta de 40 MB."
else
    echo "CONFIRMADO: Consumo de memória dentro da meta (< 40 MB)!"
fi

echo ""
echo "=== TODOS OS TESTES PASSARAM COM SUCESSO! ==="
