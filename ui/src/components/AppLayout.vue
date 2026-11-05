<script setup lang="ts">
import { computed } from 'vue'
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
  UserCheck,
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
</script>

<template>
  <div class="h-screen flex overflow-hidden bg-slate-100 dark:bg-slate-900 text-slate-800 dark:text-slate-100">
    <!-- OVERLAY MOBILE DA SIDEBAR -->
    <div
      v-if="nav.isMobileSidebarOpen"
      @click="nav.toggleMobileSidebar()"
      class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 z-40 lg:hidden no-print transition-opacity"
    ></div>

    <!-- SIDEBAR (MENU LATERAL RETRÁTIL NO DESKTOP E DRAWER NO MOBILE) -->
    <aside
      class="fixed lg:static inset-y-0 left-0 z-50 bg-opti text-white flex flex-col justify-between transform transition-all duration-200 ease-in-out shadow-xl no-print shrink-0 overflow-hidden"
      :class="[
        nav.isMobileSidebarOpen ? 'translate-x-0 w-64' : '-translate-x-full w-64',
        nav.isDesktopSidebarOpen
          ? 'lg:translate-x-0 lg:w-64'
          : 'lg:-translate-x-full lg:w-0'
      ]"
    >
      <!-- Topo: Identificação & Links -->
      <div class="w-64">
        <!-- Logo & Identificação -->
        <div class="p-5 border-b border-white/10 flex items-center justify-between">
          <div class="flex items-center gap-3 min-w-0">
            <img
              v-if="configStore.config.logo_base64"
              :src="configStore.config.logo_base64"
              alt="Logo"
              class="w-10 h-10 object-contain rounded-lg bg-white p-0.5 shrink-0"
            />
            <div
              v-else
              class="bg-white text-opti font-black px-2.5 py-1.5 rounded-lg text-lg tracking-tight shadow shrink-0"
            >
              OPTI
            </div>
            <div class="truncate">
              <h1 class="font-bold text-base leading-tight truncate">
                {{ configStore.config.fantasia }}
              </h1>
              <span class="text-[11px] text-blue-200 block truncate">
                {{ configStore.config.razao_social }}
              </span>
              <span
                v-if="configStore.config.cnpj"
                class="text-[11px] text-blue-200/80 font-mono block truncate"
              >
                {{ configStore.config.cnpj }}
              </span>
            </div>
          </div>

          <!-- Botão Fechar (Mobile) -->
          <button
            @click="nav.toggleMobileSidebar()"
            class="lg:hidden text-blue-200 hover:text-white p-1 rounded-lg transition"
            title="Fechar menu"
          >
            <X class="w-6 h-6" />
          </button>
        </div>

        <!-- Botão Principal de Ação e Copiar CNPJ -->
        <div class="p-4 space-y-2">
          <button
            @click="nav.openNovaOS()"
            class="w-full bg-emerald-500 hover:bg-emerald-600 active:bg-emerald-700 text-white font-bold py-2.5 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
          >
            <Plus class="w-5 h-5" />
            <span>Novo Serviço</span>
          </button>

          <!-- Botão Copiar CNPJ (Movido para a barra lateral) -->
          <button
            @click="dashboardStore.copyCNPJ()"
            class="w-full bg-white/10 hover:bg-white/20 active:bg-white/25 text-blue-100 hover:text-white text-xs font-semibold py-2 px-3 rounded-xl transition flex items-center justify-center gap-2"
          >
            <Copy class="w-3.5 h-3.5" />
            <span>Copiar CNPJ</span>
          </button>
        </div>

        <!-- Links de Navegação -->
        <nav class="px-3 space-y-1 text-sm font-medium">
          <button
            @click="nav.navigate('dashboard')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'dashboard'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <LayoutDashboard class="w-5 h-5 shrink-0" />
            <span>Dashboard</span>
          </button>

          <button
            @click="nav.navigate('orders')"
            class="w-full flex items-center justify-between px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'orders'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <div class="flex items-center gap-3">
              <Wrench class="w-5 h-5 shrink-0" />
              <span>Serviços</span>
            </div>
            <span
              v-if="abertosCount > 0"
              class="text-xs bg-amber-400 text-slate-900 font-bold px-2 py-0.5 rounded-full"
            >
              {{ abertosCount }}
            </span>
          </button>

          <button
            @click="nav.navigate('clients')"
            class="w-full flex items-center justify-between px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'clients'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <div class="flex items-center gap-3">
              <Users class="w-5 h-5 shrink-0" />
              <span>Clientes</span>
            </div>
            <span
              v-if="clientesStore.clientes.length > 0"
              class="text-xs bg-opti-dark text-blue-200 px-2 py-0.5 rounded-full"
            >
              {{ clientesStore.clientes.length }}
            </span>
          </button>

          <button
            @click="nav.navigate('finance')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'finance'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <DollarSign class="w-5 h-5 shrink-0" />
            <span>Financeiro</span>
          </button>

          <button
            @click="nav.navigate('reports')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'reports'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <FileText class="w-5 h-5 shrink-0" />
            <span>Relatórios de Serviço</span>
          </button>

          <button
            @click="nav.navigate('settings')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'settings'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <Settings class="w-5 h-5 shrink-0" />
            <span>Configurações</span>
          </button>
        </nav>
      </div>

      <!-- Rodapé da Sidebar (Usuário Logado & Logout) -->
      <div class="w-64 p-4 border-t border-white/10 bg-opti-dark/60 text-xs">
        <div
          v-if="authStore.user"
          class="flex items-center justify-between gap-2"
        >
          <div class="flex items-center gap-2 min-w-0">
            <UserCheck class="w-4 h-4 text-emerald-300 shrink-0" />
            <div class="truncate">
              <p class="font-bold text-white truncate">{{ authStore.user.nome }}</p>
              <p class="text-[10px] text-blue-200 truncate">{{ authStore.user.email }}</p>
            </div>
          </div>
          <button
            @click="authStore.logout()"
            class="p-1.5 rounded-lg bg-white/10 hover:bg-rose-600 text-blue-100 hover:text-white transition shrink-0"
            title="Sair do sistema"
          >
            <LogOut class="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>

    <!-- ÁREA CENTRAL DE CONTEÚDO -->
    <div class="flex-1 flex flex-col h-screen overflow-hidden min-w-0">
      <!-- HEADER CONTEXTUAL / TOPBAR -->
      <header
        class="bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 px-4 sm:px-6 py-3.5 flex justify-between items-center shrink-0 no-print"
      >
        <div class="flex items-center gap-3">
          <!-- Botão Exibir/Ocultar Barra Lateral (Mobile e Desktop) -->
          <button
            @click="nav.toggleSidebar()"
            class="p-2 rounded-xl bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-600 transition flex items-center gap-1.5"
            :title="nav.isDesktopSidebarOpen ? 'Ocultar barra lateral de menus' : 'Exibir barra lateral de menus'"
          >
            <Menu class="w-5 h-5 lg:hidden" />
            <PanelLeftClose v-if="nav.isDesktopSidebarOpen" class="w-5 h-5 hidden lg:block" />
            <PanelLeftOpen v-else class="w-5 h-5 hidden lg:block text-opti dark:text-blue-400" />
          </button>

          <div>
            <h2 class="font-bold text-lg text-slate-800 dark:text-white leading-tight">
              {{ currentMeta.title }}
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 hidden sm:block">
              {{ currentMeta.sub }}
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <!-- Alternador de Tema Claro/Escuro -->
          <button
            @click="configStore.toggleTheme()"
            class="p-2 rounded-xl text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
            :title="configStore.isDarkMode ? 'Mudar para tema claro' : 'Mudar para tema escuro'"
          >
            <Sun v-if="configStore.isDarkMode" class="w-4 h-4 text-amber-400" />
            <Moon v-else class="w-4 h-4 text-slate-600" />
          </button>

          <!-- Botão Novo Serviço (Visível no mobile ou quando a sidebar desktop está recolhida) -->
          <button
            @click="nav.openNovaOS()"
            :class="nav.isDesktopSidebarOpen ? 'lg:hidden' : 'flex'"
            class="bg-emerald-500 hover:bg-emerald-600 text-white font-bold px-3 py-2 rounded-xl text-xs shadow flex items-center gap-1 transition"
          >
            <Plus class="w-3.5 h-3.5" />
            <span>Novo Serviço</span>
          </button>

          <!-- Botão Sair Rápido -->
          <button
            @click="authStore.logout()"
            class="p-2 rounded-xl text-slate-500 hover:text-rose-600 hover:bg-rose-50 dark:hover:bg-slate-700 transition"
            title="Sair do sistema"
          >
            <LogOut class="w-4 h-4" />
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
        class="lg:hidden fixed bottom-0 left-0 right-0 bg-white dark:bg-slate-800 border-t border-slate-200 dark:border-slate-700 flex justify-around items-center py-2 z-30 text-xs font-medium text-slate-500 dark:text-slate-400 no-print"
      >
        <button
          @click="nav.navigate('dashboard')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'dashboard' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <LayoutDashboard class="w-5 h-5" />
          <span>Dashboard</span>
        </button>

        <button
          @click="nav.navigate('orders')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'orders' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <Wrench class="w-5 h-5" />
          <span>Serviços</span>
        </button>

        <button
          @click="nav.navigate('finance')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'finance' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <DollarSign class="w-5 h-5" />
          <span>Financeiro</span>
        </button>

        <button
          @click="nav.navigate('reports')"
          class="flex flex-col items-center gap-1 px-2.5 py-1 transition"
          :class="nav.currentView === 'reports' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <FileText class="w-5 h-5" />
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
