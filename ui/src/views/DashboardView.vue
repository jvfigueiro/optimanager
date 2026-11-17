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
  PlusCircle,
  FileText,
  CreditCard,
  Archive,
  Zap,
  MapPin,
  Info,
  X,
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

const DISMISSED_RECS_KEY = 'optimanager_dismissed_recommendations'
const dismissedRecs = ref<string[]>(loadDismissedRecs())

function loadDismissedRecs(): string[] {
  try {
    const raw = localStorage.getItem(DISMISSED_RECS_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) return parsed
    }
  } catch {
    // ignore storage parse errors
  }
  return []
}

function dismissRecommendation(id: string) {
  if (!dismissedRecs.value.includes(id)) {
    dismissedRecs.value.push(id)
    localStorage.setItem(DISMISSED_RECS_KEY, JSON.stringify(dismissedRecs.value))
  }
}

function openViewOS(os: OrdemServico) {
  viewingOS.value = os
  isOSModalOpen.value = true
}

function closeViewOS() {
  isOSModalOpen.value = false
  viewingOS.value = null
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

const recomendacoesVisiveis = computed(() => {
  const list = [
    {
      id: 'session_security',
      titulo: 'Sessão & Segurança de Dados Ativa',
      descricao: `Auto-desconexão configurada para ${autoLogoutTexto.value} de inatividade. Banco SQLite operando em modo WAL de alto rendimento com backups compactados em formato ZIP incluindo logotipo e configurações.`,
      actionLabel: 'Configurações',
      actionView: 'settings' as const,
    },
  ]
  return list.filter((rec) => !dismissedRecs.value.includes(rec.id))
})
</script>

<template>
  <div class="space-y-6 pb-6">
    <!-- Header de Boas-Vindas -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
      <div>
        <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight">
          Olá {{ authStore.user?.nome || 'Usuário' }}!
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Painel de controle definitivo para ordens de serviço, clientes e faturamento
        </p>
      </div>
    </div>

    <!-- Cards de Métricas do Topo -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <!-- Card 1: Faturamento de Hoje -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-blue-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <DollarSign class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Faturamento de Hoje</h3>
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

      <!-- Card 3: Total Acumulado -->
      <div class="bg-white dark:bg-[#111928] rounded-2xl border border-slate-200/90 dark:border-slate-800/80 p-4 sm:p-5 shadow-xs hover:border-slate-300 dark:hover:border-slate-700 transition flex flex-col justify-between">
        <div>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-xl bg-purple-600 flex items-center justify-center text-white shadow-xs shrink-0">
              <TrendingUp class="w-5 h-5 stroke-[2.2]" />
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white leading-tight">Total Acumulado</h3>
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
    </div>

    <!-- Cards Posicionados no Topo: Serviços em Aberto & Pagamentos Pendentes -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-5">
      <!-- Serviços em Aberto -->
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
            Nenhum serviço em aberto no momento.
          </p>
        </div>
      </div>

      <!-- Pagamentos Pendentes -->
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

    <!-- Seção Ações Frequentes Estilo OptiAssist -->
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

      <!-- Grid de 3 colunas -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-2.5 sm:gap-3">
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
        <button
          @click="nav.navigate('finance')"
          class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
        >
          <CreditCard class="w-4 h-4 text-slate-500 dark:text-slate-400" />
          <span>Fluxo de Caixa</span>
        </button>
        <a
          href="https://www.nfse.gov.br/EmissorNacional"
          target="_blank"
          rel="noopener noreferrer"
          class="w-full bg-slate-100/90 dark:bg-[#131d2e] hover:bg-slate-200/90 dark:hover:bg-[#182438] text-slate-800 dark:text-slate-200 border border-slate-200/80 dark:border-slate-700/60 font-semibold text-xs sm:text-sm py-2.5 px-4 rounded-xl transition flex items-center justify-center gap-2"
        >
          <FileText class="w-4 h-4 text-slate-500 dark:text-slate-400" />
          <span>NFS-e Nacional</span>
        </a>
        <button
          @click="nav.navigate('settings')"
          class="w-full bg-sky-500 hover:bg-sky-600 text-white font-bold text-xs sm:text-sm py-2.5 px-4 rounded-xl shadow-xs transition flex items-center justify-center gap-2"
        >
          <Archive class="w-4 h-4 stroke-[2.2]" />
          <span>Gerenciar Backups</span>
        </button>
      </div>

      <!-- Atalhos personalizados -->
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

    <!-- Recomendações -->
    <div
      v-if="recomendacoesVisiveis.length"
      class="rounded-2xl border border-sky-200/90 dark:border-sky-900/40 bg-sky-50/50 dark:bg-sky-950/20 p-4 sm:p-5 space-y-2.5"
    >
      <div class="flex items-center gap-2 mb-1">
        <MapPin class="w-4 h-4 text-sky-600 dark:text-sky-400" />
        <h4 class="font-bold text-sm text-sky-950 dark:text-sky-200">
          Recomendações
        </h4>
      </div>
      <div
        v-for="rec in recomendacoesVisiveis"
        :key="rec.id"
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white/80 dark:bg-[#0e1626]/80 p-3.5 rounded-xl border border-sky-100 dark:border-sky-900/30"
      >
        <div class="flex items-start gap-2.5">
          <Zap class="w-4 h-4 text-sky-500 fill-sky-500 shrink-0 mt-0.5" />
          <div>
            <p class="text-xs font-bold text-slate-900 dark:text-white">
              {{ rec.titulo }}
            </p>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
              {{ rec.descricao }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <button
            @click="nav.navigate(rec.actionView)"
            class="text-xs font-semibold px-3 py-1.5 rounded-lg bg-sky-100/90 hover:bg-sky-200 dark:bg-sky-900/40 dark:hover:bg-sky-900/70 text-sky-900 dark:text-sky-200 flex items-center gap-1 transition"
          >
            <Info class="w-3.5 h-3.5" />
            <span>{{ rec.actionLabel }}</span>
          </button>
          <button
            @click="dismissRecommendation(rec.id)"
            title="Dispensar recomendação"
            class="text-xs font-semibold px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 flex items-center gap-1 transition"
          >
            <X class="w-3.5 h-3.5" />
            <span>Dispensar</span>
          </button>
        </div>
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
