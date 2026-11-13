<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useNavigationStore, type ViewType } from '../stores/navigation'
import { useConfigStore } from '../stores/config'
import { useDashboardStore } from '../stores/dashboard'
import { useClientesStore } from '../stores/clientes'
import { useAuthStore } from '../stores/auth'
import {
  LayoutDashboard,
  Wrench,
  Users,
  DollarSign,
  FileText,
  Settings,
  Plus,
  Copy,
  Moon,
  Sun,
  Menu,
  X,
  PanelLeftClose,
  PanelLeftOpen,
  LogOut,
  Download,
} from 'lucide-vue-next'

import DashboardView from '../views/DashboardView.vue'
import ServicosView from '../views/ServicosView.vue'
import ClientesView from '../views/ClientesView.vue'
import FinanceiroView from '../views/FinanceiroView.vue'
import RelatoriosView from '../views/RelatoriosView.vue'
import ConfiguracoesView from '../views/ConfiguracoesView.vue'
import NovaOSModal from './NovaOSModal.vue'
import ToastContainer from './ToastContainer.vue'

const nav = useNavigationStore()
const configStore = useConfigStore()
const dashboardStore = useDashboardStore()
const clientesStore = useClientesStore()
const authStore = useAuthStore()

const pageMeta: Record<ViewType, { title: string; sub: string }> = {
  dashboard: {
    title: 'Dashboard',
    sub: 'Visão geral do faturamento, pendências e atalhos rápidos',
  },
  orders: {
    title: 'Serviços',
    sub: 'Relação de atendimentos em aberto na bancada e serviços já concluídos',
  },
  clients: {
    title: 'Clientes',
    sub: 'Cadastro rápido para autocompletar em Ordens de Serviço e Relatórios',
  },
  finance: {
    title: 'Financeiro',
    sub: 'Métricas de faturamento, fluxo de recebimentos, pico diário e exportação de dados',
  },
  reports: {
    title: 'Relatórios de Serviço',
    sub: 'Emita Relatórios Técnicos, Orçamentos ou Recibos prontos para impressão',
  },
  settings: {
    title: 'Configurações do Sistema',
    sub: 'Personalização visual, dados da empresa, termos de garantia e gestão de usuários',
  },
}

const currentMeta = computed(() => pageMeta[nav.currentView])

const abertosCount = computed(
  () => dashboardStore.data.metricas.qtd_abertos || 0
)

// PWA Install Prompt
const deferredPrompt = ref<any>(null)
const canInstallPwa = ref(false)

onMounted(() => {
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault()
    deferredPrompt.value = e
    canInstallPwa.value = true
  })
})

async function instalarPwa() {
  if (!deferredPrompt.value) return
  deferredPrompt.value.prompt()
  const { outcome } = await deferredPrompt.value.userChoice
  if (outcome === 'accepted') {
    canInstallPwa.value = false
  }
  deferredPrompt.value = null
}
</script>

<template>
  <div class="h-screen flex overflow-hidden bg-slate-50 dark:bg-[#0a0f1d] text-slate-800 dark:text-slate-100">
    <!-- OVERLAY MOBILE DA SIDEBAR -->
    <div
      v-if="nav.isMobileSidebarOpen"
      @click="nav.toggleMobileSidebar()"
      class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 z-40 lg:hidden no-print transition-opacity"
    ></div>

    <!-- SIDEBAR (ESTILO OPTIASSIST: TEMA ADAPTATIVO, BORDAS FINAS E PILULAS SUAVES) -->
    <aside
      class="fixed lg:static inset-y-0 left-0 z-50 bg-white dark:bg-[#0c1322] text-slate-800 dark:text-slate-100 flex flex-col justify-between transform transition-all duration-200 ease-in-out border-r border-slate-200/90 dark:border-slate-800/90 shadow-xs no-print shrink-0 overflow-hidden"
      :class="[
        nav.isMobileSidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full w-64',
        nav.isDesktopSidebarOpen
          ? 'lg:translate-x-0 lg:w-64'
          : 'lg:-translate-x-full lg:w-0'
      ]"
    >
      <!-- Topo: Identificação & Links -->
      <div class="w-64 flex flex-col h-full overflow-y-auto no-scrollbar">
        <!-- Logo & Identificação (Estilo OptiAssist) -->
        <div class="p-4 border-b border-slate-200/80 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2.5 min-w-0">
            <img
              v-if="configStore.config.logo_base64"
              :src="configStore.config.logo_base64"
              alt="Logo"
              class="w-8 h-8 object-contain shrink-0"
            />
            <img
              v-else
              src="/icon.png"
              alt="OptiManager"
              class="w-8 h-8 object-contain shrink-0"
            />
            <div class="truncate">
              <span class="font-extrabold text-base tracking-tight text-slate-900 dark:text-white block truncate">
                {{ configStore.config.fantasia || 'OptiManager' }}
              </span>
              <span
                v-if="configStore.config.razao_social"
                class="text-[11px] text-slate-500 dark:text-slate-400 block truncate"
              >
                {{ configStore.config.razao_social }}
              </span>
              <span
                v-if="configStore.config.cnpj"
                class="text-[11px] font-mono text-slate-400 dark:text-slate-500 block truncate mt-0.5"
              >
                {{ configStore.config.cnpj }}
              </span>
            </div>
          </div>

          <!-- Botão Fechar (Mobile) -->
          <button
            @click="nav.toggleMobileSidebar()"
            class="lg:hidden text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-lg transition"
            title="Fechar menu"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Botões de Ação Rápida (Estilo OptiAssist) -->
        <div class="p-3.5 space-y-2">
          <button
            @click="nav.openNovaOS()"
            class="w-full bg-sky-500 hover:bg-sky-600 active:bg-sky-700 text-white font-bold py-2.5 px-4 rounded-xl shadow-xs transition flex items-center justify-center gap-2 text-xs"
          >
            <Plus class="w-4 h-4" />
            <span>Novo Serviço</span>
          </button>

          <button
            @click="dashboardStore.copyCNPJ()"
            class="w-full bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold py-2 px-3 rounded-xl transition flex items-center justify-center gap-1.5 border border-slate-200/80 dark:border-slate-700/80"
          >
            <Copy class="w-3.5 h-3.5" />
            <span>Copiar CNPJ</span>
          </button>
        </div>

        <!-- Links de Navegação (Pílulas arredondadas estilo OptiAssist) -->
        <nav class="px-3 flex-1 space-y-1 text-xs font-medium">
          <button
            @click="nav.navigate('dashboard')"
            class="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'dashboard'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <LayoutDashboard class="w-4 h-4 shrink-0" />
            <span>Dashboard</span>
          </button>

          <button
            @click="nav.navigate('orders')"
            class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'orders'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <div class="flex items-center gap-3">
              <Wrench class="w-4 h-4 shrink-0" />
              <span>Serviços</span>
            </div>
            <span
              v-if="abertosCount > 0"
              class="text-[10px] bg-amber-400 text-slate-900 font-bold px-2 py-0.5 rounded-full"
            >
              {{ abertosCount }}
            </span>
          </button>

          <button
            @click="nav.navigate('clients')"
            class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'clients'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <div class="flex items-center gap-3">
              <Users class="w-4 h-4 shrink-0" />
              <span>Clientes</span>
            </div>
            <span
              v-if="clientesStore.clientes.length > 0"
              class="text-[10px] bg-slate-200 dark:bg-slate-700 text-slate-700 dark:text-slate-300 font-semibold px-2 py-0.5 rounded-full"
            >
              {{ clientesStore.clientes.length }}
            </span>
          </button>

          <button
            @click="nav.navigate('finance')"
            class="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'finance'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <DollarSign class="w-4 h-4 shrink-0" />
            <span>Financeiro</span>
          </button>

          <button
            @click="nav.navigate('reports')"
            class="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'reports'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <FileText class="w-4 h-4 shrink-0" />
            <span>Relatórios</span>
          </button>

          <!-- Divisor de Grupo (Estilo OptiAssist: Settings & About) -->
          <div class="pt-3 pb-1">
            <div class="border-t border-slate-200/80 dark:border-slate-800"></div>
          </div>

          <button
            @click="nav.navigate('settings')"
            class="w-full flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition"
            :class="
              nav.currentView === 'settings'
                ? 'bg-sky-100/90 text-sky-950 dark:bg-blue-600/20 dark:text-sky-300 font-bold border border-sky-200/90 dark:border-blue-500/30 shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100/80 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800/70'
            "
          >
            <Settings class="w-4 h-4 shrink-0" />
            <span>Configurações</span>
          </button>

          <!-- Botão Instalar App PWA -->
          <button
            v-if="canInstallPwa"
            @click="instalarPwa"
            class="w-full flex items-center justify-between px-3.5 py-2 rounded-xl transition font-medium bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/80 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 text-xs shadow-2xs mt-2"
            title="Instalar aplicativo no computador ou smartphone"
          >
            <div class="flex items-center gap-2">
              <Download class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
              <span>Instalar PWA</span>
            </div>
            <span class="text-[9px] bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 px-1 py-0.5 rounded font-bold">App</span>
          </button>
        </nav>
      </div>

      <!-- Rodapé da Sidebar (Usuário Logado & Logout) -->
      <div class="w-64 p-3 border-t border-slate-200/80 dark:border-slate-800 bg-slate-50/60 dark:bg-[#0c1322] text-xs">
        <div
          v-if="authStore.user"
          class="flex items-center justify-between gap-2 p-1.5 rounded-xl bg-white dark:bg-slate-800/60 border border-slate-200/70 dark:border-slate-700/60"
        >
          <div class="flex items-center gap-2 min-w-0">
            <div class="w-7 h-7 rounded-lg bg-sky-500/15 text-sky-600 dark:text-sky-400 flex items-center justify-center font-bold text-xs shrink-0">
              {{ authStore.user.nome.charAt(0).toUpperCase() }}
            </div>
            <div class="truncate">
              <p class="font-bold text-slate-800 dark:text-white truncate leading-tight">{{ authStore.user.nome }}</p>
              <p class="text-[10px] text-slate-400 truncate">{{ authStore.user.role === 'admin' ? 'Administrador' : 'Operador' }}</p>
            </div>
          </div>
          <button
            @click="authStore.logout()"
            class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-rose-950/40 transition shrink-0"
            title="Sair do sistema"
          >
            <LogOut class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </aside>

    <!-- ÁREA CENTRAL DE CONTEÚDO -->
    <div class="flex-1 flex flex-col h-screen overflow-hidden min-w-0">
      <!-- HEADER TOPBAR (ESTILO OPTIASSIST) -->
      <header
        class="bg-white dark:bg-[#0c1322] border-b border-slate-200/80 dark:border-slate-800 px-4 sm:px-6 py-2.5 flex justify-between items-center shrink-0 no-print"
      >
        <div class="flex items-center gap-3">
          <!-- Botão Exibir/Ocultar Barra Lateral -->
          <button
            @click="nav.toggleSidebar()"
            class="p-2 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition flex items-center gap-1"
            :title="nav.isDesktopSidebarOpen ? 'Recolher menu lateral' : 'Expandir menu lateral'"
          >
            <Menu class="w-4 h-4 lg:hidden" />
            <PanelLeftClose v-if="nav.isDesktopSidebarOpen" class="w-4 h-4 hidden lg:block" />
            <PanelLeftOpen v-else class="w-4 h-4 hidden lg:block text-sky-600 dark:text-sky-400" />
          </button>

          <div>
            <h2 class="font-black text-base text-slate-900 dark:text-white leading-tight">
              {{ currentMeta.title }}
            </h2>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 hidden sm:block">
              {{ currentMeta.sub }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <!-- Botão Alternar Tema Claro/Escuro (Estilo OptiAssist: Pílula Redonda no Topo) -->
          <button
            @click="configStore.toggleTheme()"
            class="p-2 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 border border-slate-200/80 dark:border-slate-700 text-slate-700 dark:text-slate-200 transition shadow-2xs"
            :title="configStore.isDarkMode ? 'Mudar para tema claro' : 'Mudar para tema escuro'"
          >
            <Sun v-if="configStore.isDarkMode" class="w-4 h-4 text-amber-400" />
            <Moon v-else class="w-4 h-4 text-slate-600" />
          </button>

          <!-- Botão Novo Serviço (Mobile ou sidebar recolhida) -->
          <button
            @click="nav.openNovaOS()"
            :class="nav.isDesktopSidebarOpen ? 'lg:hidden' : 'flex'"
            class="bg-sky-500 hover:bg-sky-600 text-white font-bold px-3 py-1.5 rounded-xl text-xs shadow-xs flex items-center gap-1 transition"
          >
            <Plus class="w-3.5 h-3.5" />
            <span>Novo</span>
          </button>
        </div>
      </header>

      <!-- CONTEÚDO DA VIEW ATIVA -->
      <main class="flex-1 overflow-y-auto p-4 sm:p-6 pb-24 lg:pb-8">
        <DashboardView v-if="nav.currentView === 'dashboard'" />
        <ServicosView v-else-if="nav.currentView === 'orders'" />
        <ClientesView v-else-if="nav.currentView === 'clients'" />
        <FinanceiroView v-else-if="nav.currentView === 'finance'" />
        <RelatoriosView v-else-if="nav.currentView === 'reports'" />
        <ConfiguracoesView v-else-if="nav.currentView === 'settings'" />
      </main>

      <!-- BOTTOM NAVIGATION BAR (FIXA NO MOBILE) -->
      <nav
        class="lg:hidden fixed bottom-0 left-0 right-0 bg-white/95 dark:bg-[#0c1322]/95 backdrop-blur border-t border-slate-200 dark:border-slate-800 flex justify-around items-center py-2 z-30 text-xs font-medium text-slate-500 dark:text-slate-400 no-print"
      >
        <button
          @click="nav.navigate('dashboard')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'dashboard' ? 'text-sky-600 dark:text-sky-400 font-bold' : ''"
        >
          <LayoutDashboard class="w-4 h-4" />
          <span>Dashboard</span>
        </button>

        <button
          @click="nav.navigate('orders')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'orders' ? 'text-sky-600 dark:text-sky-400 font-bold' : ''"
        >
          <Wrench class="w-4 h-4" />
          <span>Serviços</span>
        </button>

        <button
          @click="nav.navigate('finance')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'finance' ? 'text-sky-600 dark:text-sky-400 font-bold' : ''"
        >
          <DollarSign class="w-4 h-4" />
          <span>Financeiro</span>
        </button>

        <button
          @click="nav.navigate('reports')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'reports' ? 'text-sky-600 dark:text-sky-400 font-bold' : ''"
        >
          <FileText class="w-4 h-4" />
          <span>Relatórios</span>
        </button>
      </nav>
    </div>

    <!-- MODAL DE NOVA OS -->
    <NovaOSModal />

    <!-- CONTAINER DE NOTIFICAÇÕES TOAST -->
    <ToastContainer />
  </div>
</template>
