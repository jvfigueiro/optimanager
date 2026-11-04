<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useDashboardStore } from '../stores/dashboard'
import { useNavigationStore } from '../stores/navigation'
import { useOrdensStore } from '../stores/ordens'
import { formatBRL, formatDate } from '../utils/formatters'
import NovoAtalhoModal from '../components/NovoAtalhoModal.vue'
import {
  DollarSign,
  Clock,
  CheckCircle2,
  TrendingUp,
  Plus,
  Trash2,
  ExternalLink,
  Wrench,
  ChevronRight,
} from 'lucide-vue-next'

const dashboardStore = useDashboardStore()
const nav = useNavigationStore()
const ordensStore = useOrdensStore()

const isShortcutModalOpen = ref(false)

onMounted(() => {
  dashboardStore.loadDashboard()
})

function irParaServicos(filtro = 'TODOS') {
  ordensStore.setFiltro(filtro)
  nav.navigate('servicos')
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'Concluído':
      return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
    case 'Em Serviço':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-950/60 dark:text-blue-300 border-blue-200 dark:border-blue-800'
    case 'Avaliação':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-950/60 dark:text-purple-300 border-purple-200 dark:border-purple-800'
    case 'Orçamento':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 border-amber-200 dark:border-amber-800'
    default:
      return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300 border-slate-200'
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Cards de Faturamento -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
      <!-- Faturamento Hoje -->
      <div class="bg-white dark:bg-slate-800 p-4 sm:p-5 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between text-slate-400 dark:text-slate-500 mb-1">
            <span class="text-xs font-bold uppercase tracking-wider">Hoje</span>
            <DollarSign class="w-4 h-4 text-slate-400" />
          </div>
          <p class="text-xl sm:text-2xl font-black text-slate-800 dark:text-white mt-1">
            {{ formatBRL(dashboardStore.data.metricas.faturamento_hoje) }}
          </p>
        </div>
        <span class="text-[11px] text-emerald-600 dark:text-emerald-400 font-medium mt-2">
          Serviços finalizados hoje
        </span>
      </div>

      <!-- Caixa Real (Recebido) -->
      <div class="bg-white dark:bg-slate-800 p-4 sm:p-5 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between text-slate-400 dark:text-slate-500 mb-1">
            <span class="text-xs font-bold uppercase tracking-wider">Recebido</span>
            <CheckCircle2 class="w-4 h-4 text-emerald-500" />
          </div>
          <p class="text-xl sm:text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">
            {{ formatBRL(dashboardStore.data.metricas.caixa_recebido) }}
          </p>
        </div>
        <span class="text-[11px] text-slate-500 dark:text-slate-400 mt-2">
          Pagamentos quitados
        </span>
      </div>

      <!-- Pendente (A Receber) -->
      <div class="bg-amber-50/70 dark:bg-amber-950/30 p-4 sm:p-5 rounded-2xl border border-amber-200 dark:border-amber-800/60 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between text-amber-700 dark:text-amber-400 mb-1">
            <span class="text-xs font-bold uppercase tracking-wider">A Receber</span>
            <Clock class="w-4 h-4 text-amber-600 dark:text-amber-400" />
          </div>
          <p class="text-xl sm:text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">
            {{ formatBRL(dashboardStore.data.metricas.pendente_receber) }}
          </p>
        </div>
        <span class="text-[11px] text-amber-700 dark:text-amber-400 font-medium mt-2">
          {{ dashboardStore.data.metricas.qtd_pendentes }} OS aguardando pgto
        </span>
      </div>

      <!-- Faturamento Total -->
      <div class="bg-white dark:bg-slate-800 p-4 sm:p-5 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between text-slate-400 dark:text-slate-500 mb-1">
            <span class="text-xs font-bold uppercase tracking-wider">Total Acumulado</span>
            <TrendingUp class="w-4 h-4 text-opti" />
          </div>
          <p class="text-xl sm:text-2xl font-black text-opti dark:text-blue-400 mt-1">
            {{ formatBRL(dashboardStore.data.metricas.faturamento_total) }}
          </p>
        </div>
        <span class="text-[11px] text-slate-500 dark:text-slate-400 mt-2">
          Total de serviços concluídos
        </span>
      </div>
    </div>

    <!-- Atalhos Rápidos & Portais -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5">
      <div class="flex flex-wrap justify-between items-center gap-2 mb-4">
        <div>
          <h3 class="font-bold text-base text-slate-800 dark:text-white">
            Atalhos Rápidos & Portais do MEI
          </h3>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Ao clicar, o seu CNPJ é copiado automaticamente para agilizar o login
          </p>
        </div>
        <button
          @click="isShortcutModalOpen = true"
          class="text-xs bg-opti-light dark:bg-slate-700 text-opti dark:text-sky-300 hover:bg-opti dark:hover:bg-opti hover:text-white dark:hover:text-white border border-transparent dark:border-slate-600 font-bold px-3 py-2 rounded-xl transition flex items-center gap-1.5 shadow-xs"
        >
          <Plus class="w-3.5 h-3.5" />
          <span>Novo Atalho</span>
        </button>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div
          v-for="a in dashboardStore.data.atalhos"
          :key="a.id"
          class="group relative bg-slate-50 dark:bg-slate-750 hover:bg-blue-50/70 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 rounded-xl p-3.5 transition flex flex-col justify-between cursor-pointer"
          @click="dashboardStore.openShortcut(a)"
        >
          <button
            @click.stop="dashboardStore.removeShortcut(a.id)"
            title="Remover atalho"
            class="opacity-0 group-hover:opacity-100 absolute top-2 right-2 text-slate-400 hover:text-red-500 p-1 transition"
          >
            <Trash2 class="w-3.5 h-3.5" />
          </button>
          <div>
            <div class="flex items-center gap-2 font-bold text-slate-800 dark:text-slate-100 text-sm">
              <ExternalLink class="w-4 h-4 text-opti dark:text-blue-400 shrink-0" />
              <span class="truncate">{{ a.titulo }}</span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-1">
              {{ a.descricao || a.url }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Resumo em Duas Colunas: Bancada & Cobranças Pendentes -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Na Bancada / Em Aberto -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 flex flex-col justify-between">
        <div>
          <div class="flex justify-between items-center mb-3">
            <h3 class="font-bold text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center gap-1.5">
              <Wrench class="w-4 h-4 text-opti" />
              <span>Serviços em Aberto / Bancada ({{ dashboardStore.data.servicos_aberto.length }})</span>
            </h3>
            <button
              @click="irParaServicos('ABERTOS')"
              class="text-xs text-opti dark:text-blue-400 font-semibold hover:underline flex items-center gap-0.5"
            >
              <span>Ver todos</span>
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>

          <div
            v-if="dashboardStore.data.servicos_aberto.length"
            class="divide-y divide-slate-100 dark:divide-slate-700/60 text-sm"
          >
            <div
              v-for="o in dashboardStore.data.servicos_aberto"
              :key="o.id"
              class="py-3 flex justify-between items-center gap-2"
            >
              <div class="truncate">
                <span class="font-bold text-slate-800 dark:text-slate-200">
                  {{ o.cliente_nome }}
                </span>
                <span class="text-xs text-slate-500 dark:text-slate-400 block truncate">
                  {{ o.equipamento }} • {{ o.servico }}
                </span>
              </div>
              <span
                class="px-2.5 py-0.5 text-[11px] font-bold rounded-full border shrink-0"
                :class="getStatusBadgeClass(o.status)"
              >
                {{ o.status }}
              </span>
            </div>
          </div>
          <p v-else class="text-xs text-slate-400 dark:text-slate-500 py-6 text-center">
            Nenhum serviço em aberto no momento. Bancada livre!
          </p>
        </div>
      </div>

      <!-- Cobranças Pendentes -->
      <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 flex flex-col justify-between">
        <div>
          <div class="flex justify-between items-center mb-3">
            <h3 class="font-bold text-xs uppercase tracking-wider text-amber-700 dark:text-amber-400 flex items-center gap-1.5">
              <Clock class="w-4 h-4" />
              <span>Concluídos com Pagamento Pendente ({{ dashboardStore.data.cobrancas_pendentes.length }})</span>
            </h3>
            <button
              @click="irParaServicos('PENDENTE')"
              class="text-xs text-opti dark:text-blue-400 font-semibold hover:underline flex items-center gap-0.5"
            >
              <span>Gerenciar</span>
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>

          <div
            v-if="dashboardStore.data.cobrancas_pendentes.length"
            class="divide-y divide-slate-100 dark:divide-slate-700/60 text-sm"
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
                ✔ Receber
              </button>
            </div>
          </div>
          <p v-else class="text-xs text-slate-400 dark:text-slate-500 py-6 text-center">
            Tudo em dia! Nenhum pagamento pendente.
          </p>
        </div>
      </div>
    </div>

    <!-- Modal Novo Atalho -->
    <NovoAtalhoModal
      v-if="isShortcutModalOpen"
      @close="isShortcutModalOpen = false"
    />
  </div>
</template>
