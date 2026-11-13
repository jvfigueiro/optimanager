<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useDashboardStore } from '../stores/dashboard'
import { useNavigationStore } from '../stores/navigation'
import { useOrdensStore } from '../stores/ordens'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import { useAuthStore } from '../stores/auth'
import { formatBRL, formatDate } from '../utils/formatters'
import GerenciarAtalhosModal from '../components/GerenciarAtalhosModal.vue'
import EditarOSModal from '../components/EditarOSModal.vue'
import type { OrdemServico } from '../types'
import {
  DollarSign,
  Clock,
  CheckCircle2,
  TrendingUp,
  SlidersHorizontal,
  ExternalLink,
  Wrench,
  ChevronRight,
  Eye,
  Trash2,
  Users,
  Copy,
  PlusCircle,
  FileText,
  CreditCard,
  Building2,
  Archive,
  Database,
  ShieldCheck,
  Zap,
  MapPin,
  Info,
} from 'lucide-vue-next'

const dashboardStore = useDashboardStore()
const nav = useNavigationStore()
const ordensStore = useOrdensStore()
const clientesStore = useClientesStore()
const configStore = useConfigStore()
const authStore = useAuthStore()

const isShortcutModalOpen = ref(false)
const viewingOS = ref<OrdemServico | null>(null)
const isOSModalOpen = ref(false)
const cnpjCopied = ref(false)

function openViewOS(os: OrdemServico) {
  viewingOS.value = os
  isOSModalOpen.value = true
}

function closeViewOS() {
  isOSModalOpen.value = false
  viewingOS.value = null
}

async function handleCopyCNPJ() {
  await dashboardStore.copyCNPJ()
  cnpjCopied.value = true
  setTimeout(() => {
    cnpjCopied.value = false
  }, 2000)
}

onMounted(() => {
  dashboardStore.loadDashboard()
  if (!clientesStore.clientes.length) {
    clientesStore.loadClientes()
  }
})

function irParaServicos(filtro = 'TODOS') {
  ordensStore.setFiltro(filtro)
  nav.navigate('orders')
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'Concluído':
      return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
    case 'Em Serviço':
      return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30'
    case 'Avaliação':
      return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/30'
    case 'Orçamento':
      return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30'
    default:
      return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/30'
  }
}

const autoLogoutTexto = computed(() => {
  if (!configStore.config.auto_logout_enabled) {
    return 'Desativado'
  }
  return `${configStore.config.auto_logout_minutos} min`
})
</script>

<template>
  <div class="space-y-6 pb-6">
    <!-- Header de Boas-Vindas Estilo OptiAssist -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">
          Bem-vindo ao OptiManager!
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Painel de controle definitivo para ordens de serviço, clientes e faturamento
        </p>
      </div>
    </div>

    <!-- 6 Cards de Métricas Estilo OptiAssist -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-4.5">
      <!-- Card 1: Faturamento Hoje -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-blue-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <DollarSign class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Faturamento Hoje</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Entradas liquidadas no dia</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-slate-900 dark:text-white">
              {{ formatBRL(dashboardStore.data.metricas.faturamento_hoje) }}
            </p>
          </div>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 font-medium">
          Serviços quitados e recebidos hoje
        </p>
      </div>

      <!-- Card 2: Caixa Recebido -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-emerald-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <CheckCircle2 class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Caixa Recebido</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Total liquidado acumulado</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-emerald-600 dark:text-emerald-400">
              {{ formatBRL(dashboardStore.data.metricas.caixa_recebido) }}
            </p>
          </div>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 font-medium">
          Ordens finalizadas com pagamento confirmado
        </p>
      </div>

      <!-- Card 3: A Receber / Pendente -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-amber-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <Clock class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">A Receber</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Cobranças em aberto</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-amber-600 dark:text-amber-400">
              {{ formatBRL(dashboardStore.data.metricas.pendente_receber) }}
            </p>
          </div>
        </div>
        <p class="text-xs text-amber-700 dark:text-amber-400 mt-2 font-medium">
          {{ dashboardStore.data.metricas.qtd_pendentes }} ordens com pagamento pendente
        </p>
      </div>

      <!-- Card 4: Bancada / Em Aberto -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-rose-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <Wrench class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Ordens na Bancada</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Serviços em atendimento</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-slate-900 dark:text-white">
              {{ dashboardStore.data.servicos_aberto.length }} em aberto
            </p>
          </div>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 font-medium">
          {{ dashboardStore.data.metricas.qtd_abertos }} serviços ativos na oficina
        </p>
      </div>

      <!-- Card 5: Volume Total Acumulado -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-purple-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <TrendingUp class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Total Acumulado</h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Volume total movimentado</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-purple-600 dark:text-purple-400">
              {{ formatBRL(dashboardStore.data.metricas.faturamento_total) }}
            </p>
          </div>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 font-medium">
          Soma de recebidos e pendentes registrados
        </p>
      </div>

      <!-- Card 6: Base de Clientes & CNPJ (com botão Copiar CNPJ) -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-orange-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <Users class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight truncate">
                {{ configStore.config.fantasia || 'Empresa' }}
              </h3>
              <p class="text-xs text-slate-500 dark:text-slate-400">Cadastro & Documento</p>
            </div>
          </div>
          <div class="mt-4">
            <p class="text-2xl font-black tracking-tight text-slate-900 dark:text-white">
              {{ clientesStore.clientes.length }} clientes
            </p>
          </div>
        </div>
        <div class="mt-2 flex items-center justify-between gap-2 pt-1 border-t border-slate-100 dark:border-slate-800/80">
          <span class="text-xs text-slate-500 dark:text-slate-400 truncate">
            CNPJ: <span class="font-medium text-slate-700 dark:text-slate-300">{{ configStore.config.cnpj || 'Não definido' }}</span>
          </span>
          <button
            @click="handleCopyCNPJ"
            title="Copiar CNPJ da empresa"
            class="text-[11px] font-semibold text-sky-600 dark:text-sky-400 hover:text-sky-700 dark:hover:text-sky-300 flex items-center gap-1 shrink-0 px-2 py-0.5 rounded-md hover:bg-sky-50 dark:hover:bg-sky-950/40 transition"
          >
            <Copy class="w-3 h-3" />
            <span>{{ cnpjCopied ? 'Copiado!' : 'Copiar' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Seção Ações Frequentes Estilo OptiAssist (Common Tasks) -->
    <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-5 shadow-xs">
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-2">
          <Zap class="w-4 h-4 text-sky-500 fill-sky-500" />
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">
            Ações Frequentes
          </h3>
        </div>
        <button
          v-if="authStore.user?.role === 'admin'"
          @click="isShortcutModalOpen = true"
          class="text-xs text-slate-600 dark:text-slate-300 hover:text-sky-600 dark:hover:text-sky-400 bg-slate-100/90 dark:bg-slate-800/90 hover:bg-slate-200 dark:hover:bg-slate-750 font-bold px-3 py-1.5 rounded-xl transition flex items-center gap-1.5 border border-slate-200 dark:border-slate-700"
        >
          <SlidersHorizontal class="w-3.5 h-3.5" />
          <span>Personalizar Atalhos</span>
        </button>
      </div>

      <!-- Grid de 3 colunas idêntico ao OptiAssist -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-2.5 sm:gap-3">
        <!-- Coluna 1 -->
        <div class="space-y-2.5">
          <button
            @click="nav.openNovaOS()"
            class="w-full bg-sky-500 hover:bg-sky-600 text-white font-bold text-xs sm:text-sm py-2.5 px-4 rounded-xl shadow-xs transition flex items-center justify-center gap-2"
          >
            <PlusCircle class="w-4 h-4 stroke-[2.2]" />
            <span>Novo Serviço</span>
          </button>
          <button
            @click="nav.navigate('clients')"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <Users class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>Gerenciar Clientes</span>
          </button>
          <button
            @click="nav.navigate('reports')"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <FileText class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>Relatórios & OS</span>
          </button>
        </div>

        <!-- Coluna 2 -->
        <div class="space-y-2.5">
          <button
            @click="nav.navigate('finance')"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <CreditCard class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>Fluxo de Caixa</span>
          </button>
          <a
            href="https://www.gov.br/empresas-e-negocios/pt-br/empreendedor"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <Building2 class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>Portal do Empreendedor</span>
          </a>
          <a
            href="https://www.nfse.gov.br/EmissorNacional"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <FileText class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>NFS-e Nacional</span>
          </a>
        </div>

        <!-- Coluna 3 -->
        <div class="space-y-2.5">
          <a
            href="http://www8.receita.fazenda.gov.br/SimplesNacional/Aplicacoes/ATSPO/pgmei.app/Identificacao"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <DollarSign class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>PGMEI - Guia DAS</span>
          </a>
          <a
            href="https://solucoes.receita.fazenda.gov.br/servicos/cnpjreva/cnpjreva_solicitacao.asp"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
          >
            <ShieldCheck class="w-4 h-4 text-slate-500 dark:text-slate-400" />
            <span>Consulta Cartão CNPJ</span>
          </a>
          <button
            @click="nav.navigate('settings')"
            class="w-full bg-sky-500 hover:bg-sky-600 text-white font-bold text-xs sm:text-sm py-2.5 px-4 rounded-xl shadow-xs transition flex items-center justify-center gap-2"
          >
            <Archive class="w-4 h-4 stroke-[2.2]" />
            <span>Gerenciar Backups</span>
          </button>
        </div>
      </div>

      <!-- Atalhos adicionais customizados pelo usuário -->
      <div v-if="dashboardStore.data.atalhos.length" class="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800">
        <p class="text-xs font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2">
          Atalhos Personalizados
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
          <div
            v-for="a in dashboardStore.data.atalhos"
            :key="a.id"
            class="group relative bg-slate-50 dark:bg-[#131d2e] hover:bg-sky-50/60 dark:hover:bg-[#182438] border border-slate-200 dark:border-slate-700/60 rounded-xl p-3 transition flex flex-col justify-between cursor-pointer"
            @click="dashboardStore.openShortcut(a)"
          >
            <button
              v-if="authStore.user?.role === 'admin'"
              @click.stop="dashboardStore.removeShortcut(a.id)"
              title="Remover atalho"
              class="opacity-0 group-hover:opacity-100 absolute top-2 right-2 text-slate-400 hover:text-red-500 p-1 transition"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
            <div>
              <div class="flex items-center gap-2 font-bold text-slate-800 dark:text-slate-200 text-xs">
                <ExternalLink class="w-3.5 h-3.5 text-sky-500 shrink-0" />
                <span class="truncate">{{ a.titulo }}</span>
              </div>
              <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5 line-clamp-1">
                {{ a.descricao || a.url }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Listas Operacionais: Serviços na Bancada & Pagamentos Pendentes -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-5">
      <!-- Na Bancada / Em Aberto -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-5 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex justify-between items-center mb-3">
            <h3 class="font-bold text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
              <Wrench class="w-4 h-4 text-sky-500" />
              <span>Serviços em Aberto ({{ dashboardStore.data.servicos_aberto.length }})</span>
            </h3>
            <button
              @click="irParaServicos('ABERTOS')"
              class="text-xs text-sky-600 dark:text-sky-400 font-semibold hover:underline flex items-center gap-0.5"
            >
              <span>Ver todos</span>
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>

          <div
            v-if="dashboardStore.data.servicos_aberto.length"
            class="divide-y divide-slate-100 dark:divide-slate-800/80 text-sm"
          >
            <div
              v-for="o in dashboardStore.data.servicos_aberto"
              :key="o.id"
              class="py-3 flex justify-between items-center gap-2"
            >
              <div class="truncate">
                <div class="flex items-center gap-1.5">
                  <span class="text-xs font-bold text-sky-600 dark:text-sky-400">#{{ o.id }}</span>
                  <span class="font-bold text-slate-800 dark:text-slate-200 truncate">
                    {{ o.cliente_nome }}
                  </span>
                </div>
                <span class="text-xs text-slate-500 dark:text-slate-400 block truncate">
                  {{ o.equipamento }} • {{ o.servico }}
                </span>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <span
                  class="px-2.5 py-0.5 text-[11px] font-bold rounded-full border"
                  :class="getStatusBadgeClass(o.status)"
                >
                  {{ o.status }}
                </span>
                <button
                  @click="openViewOS(o)"
                  title="Visualizar detalhes da OS"
                  class="px-2.5 py-1 text-xs bg-slate-100 hover:bg-sky-50 text-slate-700 dark:text-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 rounded-lg font-bold transition flex items-center gap-1 border border-slate-200 dark:border-slate-700"
                >
                  <Eye class="w-3.5 h-3.5 text-sky-500" />
                  <span>Ver</span>
                </button>
              </div>
            </div>
          </div>
          <p v-else class="text-xs text-slate-400 dark:text-slate-500 py-6 text-center">
            Nenhum serviço em aberto no momento. Bancada livre!
          </p>
        </div>
      </div>

      <!-- Cobranças Pendentes -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-5 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex justify-between items-center mb-3">
            <h3 class="font-bold text-xs uppercase tracking-wider text-amber-700 dark:text-amber-400 flex items-center gap-1.5">
              <Clock class="w-4 h-4" />
              <span>Pagamentos Pendentes ({{ dashboardStore.data.cobrancas_pendentes.length }})</span>
            </h3>
            <button
              @click="irParaServicos('PENDENTE')"
              class="text-xs text-sky-600 dark:text-sky-400 font-semibold hover:underline flex items-center gap-0.5"
            >
              <span>Gerenciar</span>
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>

          <div
            v-if="dashboardStore.data.cobrancas_pendentes.length"
            class="divide-y divide-slate-100 dark:divide-slate-800/80 text-sm"
          >
            <div
              v-for="o in dashboardStore.data.cobrancas_pendentes"
              :key="o.id"
              class="py-3 flex justify-between items-center gap-2"
            >
              <div class="truncate">
                <span class="font-bold text-slate-800 dark:text-slate-200">
                  {{ o.cliente_nome }}
                </span>
                <span class="text-xs text-slate-500 dark:text-slate-400 block">
                  {{ formatDate(o.data_emissao) }} • {{ formatBRL(o.valor_total) }}
                </span>
              </div>
              <button
                @click="dashboardStore.liquidarOS(o.id)"
                class="text-xs bg-emerald-50 dark:bg-emerald-950/60 hover:bg-emerald-600 dark:hover:bg-emerald-600 text-emerald-700 dark:text-emerald-300 hover:text-white font-bold px-3 py-1.5 rounded-lg border border-emerald-200 dark:border-emerald-700 transition shrink-0"
              >
                Receber
              </button>
            </div>
          </div>
          <p v-else class="text-xs text-slate-400 dark:text-slate-500 py-6 text-center">
            Tudo em dia! Nenhum pagamento pendente.
          </p>
        </div>
      </div>
    </div>

    <!-- Recomendações e Status do Sistema Estilo OptiAssist (System & Hardware Recommendations) -->
    <div class="rounded-2xl border border-sky-200/90 dark:border-sky-900/40 bg-sky-50/50 dark:bg-sky-950/20 p-4 sm:p-5">
      <div class="flex items-center gap-2 mb-3">
        <MapPin class="w-4 h-4 text-sky-600 dark:text-sky-400" />
        <h4 class="font-bold text-sm text-sky-950 dark:text-sky-200">
          Status Operacional & Recomendações
        </h4>
      </div>
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white/80 dark:bg-[#0e1626]/80 p-3.5 rounded-xl border border-sky-100 dark:border-sky-900/30">
        <div class="flex items-start gap-2.5">
          <Zap class="w-4 h-4 text-sky-500 fill-sky-500 shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-bold text-slate-900 dark:text-white">
              Sessão & Segurança de Dados Ativa
            </p>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
              Auto-desconexão configurada para {{ autoLogoutTexto }} de inatividade. Banco SQLite operando em modo WAL de alto rendimento com backups compactados em formato ZIP incluindo logotipo e configurações.
            </p>
          </div>
        </div>
        <button
          @click="nav.navigate('settings')"
          class="shrink-0 text-xs font-semibold px-3 py-1.5 rounded-lg bg-sky-100/90 hover:bg-sky-200 dark:bg-sky-900/40 dark:hover:bg-sky-900/70 text-sky-900 dark:text-sky-200 flex items-center gap-1 transition"
        >
          <Info class="w-3.5 h-3.5" />
          <span>Configurações</span>
        </button>
      </div>
    </div>

    <!-- Barra de Pílulas de Status do Rodapé Estilo OptiAssist -->
    <div class="flex flex-wrap items-center justify-center gap-2 pt-2 text-xs">
      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-white dark:bg-[#111928] border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 shadow-2xs">
        <ShieldCheck class="w-3.5 h-3.5 text-sky-500" />
        <span>Sessão: <strong class="text-slate-900 dark:text-slate-200">{{ autoLogoutTexto }}</strong></span>
      </div>

      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-white dark:bg-[#111928] border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 shadow-2xs">
        <Database class="w-3.5 h-3.5 text-emerald-500" />
        <span>Banco SQLite: <strong class="text-emerald-600 dark:text-emerald-400">WAL Ativo</strong></span>
      </div>

      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-white dark:bg-[#111928] border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 shadow-2xs">
        <Archive class="w-3.5 h-3.5 text-sky-500" />
        <span>Backup ZIP: <strong class="text-sky-600 dark:text-sky-400">Disponível</strong></span>
      </div>

      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-white dark:bg-[#111928] border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 shadow-2xs">
        <CheckCircle2 class="w-3.5 h-3.5 text-emerald-500" />
        <span>Servidor: <strong class="text-emerald-600 dark:text-emerald-400">Online (:8585)</strong></span>
      </div>

      <div class="flex items-center gap-1.5 px-3 py-1 rounded-full bg-white dark:bg-[#111928] border border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400 shadow-2xs">
        <FileText class="w-3.5 h-3.5 text-slate-400" />
        <span>Licença: <strong class="text-slate-700 dark:text-slate-300">GNU GPLv3</strong></span>
      </div>
    </div>

    <!-- Modal Gerenciar Atalhos -->
    <GerenciarAtalhosModal
      v-if="isShortcutModalOpen"
      @close="isShortcutModalOpen = false"
    />

    <!-- Modal Visualizar / Editar Ordem de Serviço -->
    <EditarOSModal
      :ordem="viewingOS"
      :is-open="isOSModalOpen"
      @close="closeViewOS"
      @updated="() => dashboardStore.loadDashboard()"
      @deleted="() => dashboardStore.loadDashboard()"
      @duplicated="() => dashboardStore.loadDashboard()"
    />
  </div>
</template>
