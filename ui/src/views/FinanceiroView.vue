<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useToastStore } from '../stores/toast'
import { formatBRL, formatDate } from '../utils/formatters'
import {
  DollarSign,
  TrendingUp,
  Calendar,
  Download,
  Printer,
  CreditCard,
  Clock,
  CheckCircle2,
  Award,
  Receipt,
} from 'lucide-vue-next'

const ordensStore = useOrdensStore()
const toast = useToastStore()

type PeriodoPreset = 'mes_atual' | 'mes_anterior' | 'ultimos_30' | 'ultimos_90' | 'ano_atual' | 'personalizado'
const periodoSelecionado = ref<PeriodoPreset>('mes_atual')

function toISODate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const dataInicio = ref('')
const dataFim = ref('')
const filtroStatusPagamento = ref<'TODOS' | 'PAGO' | 'PENDENTE'>('TODOS')

function aplicarPreset(preset: PeriodoPreset) {
  periodoSelecionado.value = preset
  const hoje = new Date()

  if (preset === 'mes_atual') {
    const primeiroDia = new Date(hoje.getFullYear(), hoje.getMonth(), 1)
    const ultimoDia = new Date(hoje.getFullYear(), hoje.getMonth() + 1, 0)
    dataInicio.value = toISODate(primeiroDia)
    dataFim.value = toISODate(ultimoDia)
  } else if (preset === 'mes_anterior') {
    const primeiroDia = new Date(hoje.getFullYear(), hoje.getMonth() - 1, 1)
    const ultimoDia = new Date(hoje.getFullYear(), hoje.getMonth(), 0)
    dataInicio.value = toISODate(primeiroDia)
    dataFim.value = toISODate(ultimoDia)
  } else if (preset === 'ultimos_30') {
    const passado = new Date()
    passado.setDate(hoje.getDate() - 30)
    dataInicio.value = toISODate(passado)
    dataFim.value = toISODate(hoje)
  } else if (preset === 'ultimos_90') {
    const passado = new Date()
    passado.setDate(hoje.getDate() - 90)
    dataInicio.value = toISODate(passado)
    dataFim.value = toISODate(hoje)
  } else if (preset === 'ano_atual') {
    const primeiroDia = new Date(hoje.getFullYear(), 0, 1)
    const ultimoDia = new Date(hoje.getFullYear(), 11, 31)
    dataInicio.value = toISODate(primeiroDia)
    dataFim.value = toISODate(ultimoDia)
  }
}

onMounted(async () => {
  aplicarPreset('mes_atual')
  if (!ordensStore.ordens.length) {
    await ordensStore.loadOrdens()
  }
})

// Ordens filtradas pelo período selecionado e status de pagamento
const ordensFiltradas = computed(() => {
  return ordensStore.ordens.filter((o) => {
    // Extrai data no formato YYYY-MM-DD
    const emissao = o.data_emissao ? o.data_emissao.substring(0, 10) : ''
    if (!emissao) return false

    if (dataInicio.value && emissao < dataInicio.value) return false
    if (dataFim.value && emissao > dataFim.value) return false

    if (filtroStatusPagamento.value === 'PAGO' && o.pagamento_status !== 'Pago') return false
    if (filtroStatusPagamento.value === 'PENDENTE' && o.pagamento_status !== 'Pendente') return false

    return true
  })
})

// Métricas agregadas
const totalFaturado = computed(() => {
  return ordensFiltradas.value.reduce((acc, o) => acc + (Number(o.valor_total) || 0), 0)
})

const totalRecebido = computed(() => {
  return ordensFiltradas.value
    .filter((o) => o.pagamento_status === 'Pago')
    .reduce((acc, o) => acc + (Number(o.valor_total) || 0), 0)
})

const totalPendente = computed(() => {
  return ordensFiltradas.value
    .filter((o) => o.pagamento_status === 'Pendente')
    .reduce((acc, o) => acc + (Number(o.valor_total) || 0), 0)
})

const ticketMedio = computed(() => {
  const count = ordensFiltradas.value.length
  return count > 0 ? totalFaturado.value / count : 0
})

// Agrupamento por dia para identificar o dia de maior faturamento
interface FaturamentoDia {
  data: string
  total: number
  recebido: number
  pendente: number
  quantidade: number
}

const faturamentoPorDia = computed(() => {
  const mapa = new Map<string, FaturamentoDia>()

  for (const o of ordensFiltradas.value) {
    const d = o.data_emissao ? o.data_emissao.substring(0, 10) : 'Sem data'
    const valor = Number(o.valor_total) || 0
    const isPago = o.pagamento_status === 'Pago'

    if (!mapa.has(d)) {
      mapa.set(d, {
        data: d,
        total: 0,
        recebido: 0,
        pendente: 0,
        quantidade: 0,
      })
    }

    const diaObj = mapa.get(d)!
    diaObj.total += valor
    diaObj.quantidade += 1
    if (isPago) {
      diaObj.recebido += valor
    } else {
      diaObj.pendente += valor
    }
  }

  // Ordena por data decrescente
  return Array.from(mapa.values()).sort((a, b) => b.data.localeCompare(a.data))
})

// Identificação do dia que mais faturou
const diaMaiorFaturamento = computed(() => {
  if (!faturamentoPorDia.value.length) return null

  let topDia = faturamentoPorDia.value[0]
  for (const d of faturamentoPorDia.value) {
    if (d.total > topDia.total) {
      topDia = d
    }
  }
  return topDia.total > 0 ? topDia : null
})

// Distribuição por Forma de Pagamento
const distribuicaoFormas = computed(() => {
  const mapa = new Map<string, { forma: string; total: number; count: number }>()

  for (const o of ordensFiltradas.value) {
    const f = o.pagamento_forma || 'Pendente'
    const v = Number(o.valor_total) || 0

    if (!mapa.has(f)) {
      mapa.set(f, { forma: f, total: 0, count: 0 })
    }
    const item = mapa.get(f)!
    item.total += v
    item.count += 1
  }

  return Array.from(mapa.values()).sort((a, b) => b.total - a.total)
})

// Exportação dos dados para arquivo CSV (compatível com Excel / LibreOffice)
function exportarCSV() {
  if (!ordensFiltradas.value.length) {
    toast.show('Nenhum registro encontrado no período para exportação.', 'info')
    return
  }

  const cabecalho = [
    'ID',
    'Data de Emissão',
    'Hora',
    'Cliente',
    'CPF/CNPJ',
    'Telefone/Contato',
    'Equipamento',
    'Serviço',
    'Valor Total (R$)',
    'Status OS',
    'Status Pagamento',
    'Forma de Pagamento',
    'Data Entrega',
  ]

  const linhas = ordensFiltradas.value.map((o) => {
    return [
      o.id,
      formatDate(o.data_emissao),
      o.hora_emissao || '',
      `"${(o.cliente_nome || '').replace(/"/g, '""')}"`,
      `"${o.cliente_doc || ''}"`,
      `"${o.cliente_contato || ''}"`,
      `"${(o.equipamento || '').replace(/"/g, '""')}"`,
      `"${(o.servico || '').replace(/"/g, '""')}"`,
      (Number(o.valor_total) || 0).toFixed(2).replace('.', ','),
      `"${o.status}"`,
      `"${o.pagamento_status}"`,
      `"${o.pagamento_forma || ''}"`,
      o.data_entrega ? formatDate(o.data_entrega) : '',
    ].join(';')
  })

  // Linhas de resumo no rodapé
  const linhaTotais = [
    'TOTAIS',
    '',
    '',
    `Total de Registros: ${ordensFiltradas.value.length}`,
    '',
    '',
    '',
    `Faturamento Total:`,
    totalFaturado.value.toFixed(2).replace('.', ','),
    `Recebido: ${totalRecebido.value.toFixed(2).replace('.', ',')}`,
    `Pendente: ${totalPendente.value.toFixed(2).replace('.', ',')}`,
    '',
    '',
  ].join(';')

  // BOM UTF-8 (\uFEFF) para garantir acentuação correta no Excel
  const conteudoCSV = '\uFEFF' + cabecalho.join(';') + '\n' + linhas.join('\n') + '\n\n' + linhaTotais

  const blob = new Blob([conteudoCSV], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `relatorio_financeiro_${dataInicio.value}_a_${dataFim.value}.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)

  toast.show('Relatório financeiro exportado com sucesso em formato CSV!')
}

function imprimirRelatorio() {
  window.print()
}
</script>

<template>
  <div class="space-y-6">
    <!-- CABEÇALHO & FILTROS DE PERÍODO -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 space-y-4 no-print">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-100 dark:border-slate-700/60 pb-4">
        <div>
          <h2 class="font-black text-xl text-slate-800 dark:text-white flex items-center gap-2">
            <DollarSign class="w-6 h-6 text-opti" />
            <span>Painel Financeiro</span>
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Acompanhe o faturamento, fluxo de recebimentos, pico diário de vendas e exporte relatórios consolidados.
          </p>
        </div>

        <div class="flex items-center gap-2 flex-wrap">
          <button
            @click="exportarCSV"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-xs"
            title="Exportar dados para arquivo CSV/Excel"
          >
            <Download class="w-4 h-4" />
            <span>Exportar CSV / Excel</span>
          </button>

          <button
            @click="imprimirRelatorio"
            class="px-3.5 py-2 border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 hover:bg-slate-50 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-2xs"
            title="Imprimir ou gerar PDF financeiro"
          >
            <Printer class="w-4 h-4" />
            <span>Imprimir / PDF</span>
          </button>
        </div>
      </div>

      <!-- Barra de Filtros Rápidos -->
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-1.5 text-xs font-semibold">
          <span class="text-slate-400 dark:text-slate-500 text-[11px] uppercase font-bold mr-1 flex items-center gap-1">
            <Calendar class="w-3.5 h-3.5" />
            <span>Período:</span>
          </span>

          <button
            @click="aplicarPreset('mes_atual')"
            class="px-3 py-1.5 rounded-xl transition border text-xs"
            :class="
              periodoSelecionado === 'mes_atual'
                ? 'bg-opti text-white border-opti font-bold shadow-xs'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-600 hover:bg-slate-200'
            "
          >
            Este Mês
          </button>

          <button
            @click="aplicarPreset('mes_anterior')"
            class="px-3 py-1.5 rounded-xl transition border text-xs"
            :class="
              periodoSelecionado === 'mes_anterior'
                ? 'bg-opti text-white border-opti font-bold shadow-xs'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-600 hover:bg-slate-200'
            "
          >
            Mês Anterior
          </button>

          <button
            @click="aplicarPreset('ultimos_30')"
            class="px-3 py-1.5 rounded-xl transition border text-xs"
            :class="
              periodoSelecionado === 'ultimos_30'
                ? 'bg-opti text-white border-opti font-bold shadow-xs'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-600 hover:bg-slate-200'
            "
          >
            Últimos 30 Dias
          </button>

          <button
            @click="aplicarPreset('ano_atual')"
            class="px-3 py-1.5 rounded-xl transition border text-xs"
            :class="
              periodoSelecionado === 'ano_atual'
                ? 'bg-opti text-white border-opti font-bold shadow-xs'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-600 hover:bg-slate-200'
            "
          >
            Este Ano
          </button>

          <button
            @click="periodoSelecionado = 'personalizado'"
            class="px-3 py-1.5 rounded-xl transition border text-xs"
            :class="
              periodoSelecionado === 'personalizado'
                ? 'bg-opti text-white border-opti font-bold shadow-xs'
                : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-600 hover:bg-slate-200'
            "
          >
            Personalizado
          </button>
        </div>

        <!-- Seletor de Datas Personalizadas -->
        <div class="flex items-center gap-2 text-xs">
          <div class="flex items-center gap-1.5 bg-slate-50 dark:bg-slate-700 p-1 rounded-xl border border-slate-200 dark:border-slate-600">
            <input
              type="date"
              v-model="dataInicio"
              @change="periodoSelecionado = 'personalizado'"
              class="bg-transparent px-2 py-1 text-slate-800 dark:text-white font-semibold text-xs focus:outline-none"
              title="Data inicial"
            />
            <span class="text-slate-400 font-bold">até</span>
            <input
              type="date"
              v-model="dataFim"
              @change="periodoSelecionado = 'personalizado'"
              class="bg-transparent px-2 py-1 text-slate-800 dark:text-white font-semibold text-xs focus:outline-none"
              title="Data final"
            />
          </div>

          <!-- Filtro de Pagamento -->
          <select
            v-model="filtroStatusPagamento"
            class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-2.5 py-2 text-xs font-semibold focus:ring-2 focus:ring-opti"
          >
            <option value="TODOS">Todos os Pagamentos</option>
            <option value="PAGO">Apenas Pagos</option>
            <option value="PENDENTE">Apenas Pendentes</option>
          </select>
        </div>
      </div>
    </div>

    <!-- CARDS DE MÉTRICAS PRINCIPAIS -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- 1. Faturamento Total do Período -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 relative overflow-hidden">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-bold uppercase text-slate-500 dark:text-slate-400">
            Faturamento do Período
          </span>
          <div class="w-8 h-8 rounded-xl bg-opti/10 dark:bg-opti/20 flex items-center justify-center text-opti dark:text-blue-400">
            <TrendingUp class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-slate-900 dark:text-white">
          {{ formatBRL(totalFaturado) }}
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ ordensFiltradas.length }} serviço(s) emitido(s) no período
        </p>
      </div>

      <!-- 2. Total Recebido (Caixa) -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 relative overflow-hidden">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-bold uppercase text-emerald-700 dark:text-emerald-400">
            Recebido (Caixa)
          </span>
          <div class="w-8 h-8 rounded-xl bg-emerald-100 dark:bg-emerald-950/60 flex items-center justify-center text-emerald-600 dark:text-emerald-400">
            <CheckCircle2 class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400">
          {{ formatBRL(totalRecebido) }}
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          {{ totalFaturado > 0 ? ((totalRecebido / totalFaturado) * 100).toFixed(0) : 0 }}% do faturamento liquidado
        </p>
      </div>

      <!-- 3. Cobranças Pendentes -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 relative overflow-hidden">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs font-bold uppercase text-amber-700 dark:text-amber-400">
            Pendente a Receber
          </span>
          <div class="w-8 h-8 rounded-xl bg-amber-100 dark:bg-amber-950/60 flex items-center justify-center text-amber-600 dark:text-amber-400">
            <Clock class="w-4 h-4" />
          </div>
        </div>
        <div class="text-2xl font-black text-amber-600 dark:text-amber-400">
          {{ formatBRL(totalPendente) }}
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Valores aguardando liquidação
        </p>
      </div>

      <!-- 4. Dia em que mais faturou (Destaque Solicitado) -->
      <div class="bg-gradient-to-br from-opti to-opti-dark text-white rounded-2xl shadow-xs p-5 relative overflow-hidden flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs font-bold uppercase text-blue-200 flex items-center gap-1">
              <Award class="w-3.5 h-3.5 text-amber-300" />
              <span>Dia de Maior Faturamento</span>
            </span>
          </div>
          <div v-if="diaMaiorFaturamento" class="space-y-0.5">
            <div class="text-2xl font-black text-white">
              {{ formatBRL(diaMaiorFaturamento.total) }}
            </div>
            <p class="text-xs text-blue-100 font-semibold">
              {{ formatDate(diaMaiorFaturamento.data) }} • {{ diaMaiorFaturamento.quantidade }} serviço(s)
            </p>
          </div>
          <div v-else class="text-sm text-blue-200 py-2">
            Nenhum faturamento registrado no período selecionado.
          </div>
        </div>
        <p class="text-[11px] text-blue-200/80 pt-2 border-t border-white/10 mt-2">
          Ticket Médio do Período: <strong>{{ formatBRL(ticketMedio) }}</strong>
        </p>
      </div>
    </div>

    <!-- DETALHAMENTO DIÁRIO & FORMAS DE PAGAMENTO -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Tabela Agrupada por Dia (com destaque de pico) -->
      <div class="lg:col-span-2 bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 space-y-3">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-700/60 pb-3">
          <div class="flex items-center gap-2">
            <Calendar class="w-4 h-4 text-opti" />
            <h3 class="font-bold text-sm text-slate-800 dark:text-white">
              Demonstrativo por Dia
            </h3>
          </div>
          <span class="text-xs text-slate-400">
            {{ faturamentoPorDia.length }} dia(s) com movimento
          </span>
        </div>

        <div v-if="faturamentoPorDia.length" class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse">
            <thead>
              <tr class="text-slate-400 uppercase text-[10px] border-b border-slate-200 dark:border-slate-700">
                <th class="py-2.5 px-3">Data</th>
                <th class="py-2.5 px-3 text-center">Qtd</th>
                <th class="py-2.5 px-3 text-right">Recebido</th>
                <th class="py-2.5 px-3 text-right">Pendente</th>
                <th class="py-2.5 px-3 text-right">Total Faturado</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700/60">
              <tr
                v-for="dia in faturamentoPorDia"
                :key="dia.data"
                class="hover:bg-slate-50 dark:hover:bg-slate-750 transition"
                :class="diaMaiorFaturamento && dia.data === diaMaiorFaturamento.data ? 'bg-amber-50/60 dark:bg-amber-950/20' : ''"
              >
                <td class="py-3 px-3 whitespace-nowrap font-semibold text-slate-800 dark:text-white flex items-center gap-1.5">
                  <span>{{ formatDate(dia.data) }}</span>
                  <span
                    v-if="diaMaiorFaturamento && dia.data === diaMaiorFaturamento.data"
                    class="px-2 py-0.5 rounded-full text-[9px] font-black uppercase bg-amber-200 text-amber-900 dark:bg-amber-900/60 dark:text-amber-200 flex items-center gap-0.5"
                    title="Pico de faturamento do período"
                  >
                    <Award class="w-2.5 h-2.5" />
                    <span>Pico</span>
                  </span>
                </td>
                <td class="py-3 px-3 text-center font-bold text-slate-600 dark:text-slate-300">
                  {{ dia.quantidade }}
                </td>
                <td class="py-3 px-3 text-right text-emerald-600 dark:text-emerald-400 font-semibold">
                  {{ formatBRL(dia.recebido) }}
                </td>
                <td class="py-3 px-3 text-right text-amber-600 dark:text-amber-400 font-semibold">
                  {{ formatBRL(dia.pendente) }}
                </td>
                <td class="py-3 px-3 text-right font-black text-slate-900 dark:text-white">
                  {{ formatBRL(dia.total) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="text-xs text-slate-400 py-8 text-center">
          Nenhuma movimentação registrada no período selecionado.
        </p>
      </div>

      <!-- Resumo por Formas de Pagamento -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 space-y-4">
        <div class="flex items-center gap-2 border-b border-slate-100 dark:border-slate-700/60 pb-3">
          <CreditCard class="w-4 h-4 text-opti" />
          <h3 class="font-bold text-sm text-slate-800 dark:text-white">
            Formas de Pagamento
          </h3>
        </div>

        <div v-if="distribuicaoFormas.length" class="space-y-3">
          <div
            v-for="df in distribuicaoFormas"
            :key="df.forma"
            class="space-y-1 text-xs"
          >
            <div class="flex justify-between items-center">
              <span class="font-bold text-slate-700 dark:text-slate-200">
                {{ df.forma }} ({{ df.count }})
              </span>
              <span class="font-black text-slate-900 dark:text-white">
                {{ formatBRL(df.total) }}
              </span>
            </div>
            <div class="w-full h-2 rounded-full bg-slate-100 dark:bg-slate-700 overflow-hidden">
              <div
                class="h-full bg-opti rounded-full"
                :style="{ width: `${totalFaturado > 0 ? (df.total / totalFaturado) * 100 : 0}%` }"
              ></div>
            </div>
            <p class="text-[10px] text-slate-400 text-right">
              {{ totalFaturado > 0 ? ((df.total / totalFaturado) * 100).toFixed(1) : 0 }}% do total
            </p>
          </div>
        </div>
        <p v-else class="text-xs text-slate-400 py-6 text-center">
          Sem movimentações no período.
        </p>
      </div>
    </div>

    <!-- LISTAGEM DETALHADA DAS ORDENS DO PERÍODO -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 space-y-3">
      <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-700/60 pb-3">
        <div class="flex items-center gap-2">
          <Receipt class="w-4 h-4 text-opti" />
          <h3 class="font-bold text-sm text-slate-800 dark:text-white">
            Detalhamento de Ordens & Serviços do Período
          </h3>
        </div>
        <span class="text-xs text-slate-400 font-semibold">
          {{ ordensFiltradas.length }} registro(s)
        </span>
      </div>

      <div v-if="ordensFiltradas.length" class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr class="text-slate-400 uppercase text-[10px] border-b border-slate-200 dark:border-slate-700">
              <th class="py-2.5 px-3">OS</th>
              <th class="py-2.5 px-3">Data</th>
              <th class="py-2.5 px-3">Cliente</th>
              <th class="py-2.5 px-3">Serviço / Descrição</th>
              <th class="py-2.5 px-3">Forma Pgto</th>
              <th class="py-2.5 px-3 text-center">Status Pgto</th>
              <th class="py-2.5 px-3 text-right">Valor</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700/60">
            <tr
              v-for="o in ordensFiltradas"
              :key="o.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-750 transition"
            >
              <td class="py-3 px-3 font-bold text-opti dark:text-blue-400">
                {{ o.id }}
              </td>
              <td class="py-3 px-3 text-slate-600 dark:text-slate-300 whitespace-nowrap">
                {{ formatDate(o.data_emissao) }}
              </td>
              <td class="py-3 px-3 font-semibold text-slate-800 dark:text-white whitespace-nowrap">
                {{ o.cliente_nome }}
              </td>
              <td class="py-3 px-3 text-slate-600 dark:text-slate-300 max-w-xs truncate">
                {{ o.servico }}
              </td>
              <td class="py-3 px-3 text-slate-600 dark:text-slate-300">
                {{ o.pagamento_forma || 'Pendente' }}
              </td>
              <td class="py-3 px-3 text-center">
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-bold"
                  :class="
                    o.pagamento_status === 'Pago'
                      ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300'
                      : 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300'
                  "
                >
                  {{ o.pagamento_status }}
                </span>
              </td>
              <td class="py-3 px-3 text-right font-black text-slate-900 dark:text-white whitespace-nowrap">
                {{ formatBRL(o.valor_total) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="text-xs text-slate-400 py-8 text-center">
        Nenhuma ordem de serviço localizada nos filtros aplicados.
      </p>
    </div>
  </div>
</template>
