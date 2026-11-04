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
  servicos: {
    title: 'Ordens de Serviço',
    sub: 'Relação de atendimentos em aberto na bancada e serviços já concluídos',
  },
  clientes: {
    title: 'Base de Clientes',
    sub: 'Cadastro rápido para autocompletar em Ordens de Serviço e Relatórios',
  },
  relatorios: {
    title: 'Central de Relatórios & PDFs',
    sub: 'Emita Relatórios Técnicos, Orçamentos ou Recibos prontos para impressão',
  },
  configuracoes: {
    title: 'Configurações do Sistema',
    sub: 'Personalização visual, dados do MEI, termos de garantia e gestão de usuários',
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

          <!-- Botão Ocultar Barra Lateral (Desktop) -->
          <button
            @click="nav.toggleDesktopSidebar()"
            class="hidden lg:flex text-blue-200 hover:text-white hover:bg-white/10 p-1.5 rounded-lg transition"
            title="Ocultar barra lateral"
          >
            <PanelLeftClose class="w-5 h-5" />
          </button>
        </div>

        <!-- Botão Principal de Ação -->
        <div class="p-4">
          <button
            @click="nav.openNovaOS()"
            class="w-full bg-emerald-500 hover:bg-emerald-600 active:bg-emerald-700 text-white font-bold py-2.5 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
          >
            <Plus class="w-5 h-5" />
            <span>Novo Serviço / OS</span>
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
            @click="nav.navigate('servicos')"
            class="w-full flex items-center justify-between px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'servicos'
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
            @click="nav.navigate('clientes')"
            class="w-full flex items-center justify-between px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'clientes'
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
            @click="nav.navigate('relatorios')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'relatorios'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <FileText class="w-5 h-5 shrink-0" />
            <span>Relatórios & PDFs</span>
          </button>

          <button
            @click="nav.navigate('configuracoes')"
            class="w-full flex items-center gap-3 px-3.5 py-3 rounded-xl transition font-medium"
            :class="
              nav.currentView === 'configuracoes'
                ? 'bg-white text-opti font-bold shadow-xs'
                : 'text-blue-100 hover:bg-opti-dark'
            "
          >
            <Settings class="w-5 h-5 shrink-0" />
            <span>Configurações</span>
          </button>
        </nav>
      </div>

      <!-- Rodapé da Sidebar (Dados do MEI & Usuário Logado) -->
      <div class="w-64 p-4 border-t border-white/10 bg-opti-dark/60 text-xs space-y-2.5">
        <div>
          <div class="flex justify-between items-center text-blue-200">
            <span>CNPJ MEI:</span>
            <button
              @click="dashboardStore.copyCNPJ()"
              class="underline hover:text-white font-semibold flex items-center gap-1"
              title="Copiar CNPJ"
            >
              <Copy class="w-3 h-3" />
              <span>Copiar</span>
            </button>
          </div>
          <p class="font-mono font-bold text-white text-sm mt-0.5">
            {{ configStore.config.cnpj }}
          </p>
        </div>

        <!-- Usuário Logado & Logout -->
        <div
          v-if="authStore.user"
          class="pt-2.5 border-t border-white/10 flex items-center justify-between gap-2"
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

          <!-- Botão Copiar CNPJ -->
          <button
            @click="dashboardStore.copyCNPJ()"
            class="text-xs bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 font-medium px-3 py-2 rounded-xl transition flex items-center gap-1.5"
          >
            <Copy class="w-3.5 h-3.5" />
            <span class="hidden sm:inline">Copiar CNPJ</span>
          </button>

          <!-- Botão Nova OS (Visível no mobile ou quando a sidebar desktop está recolhida) -->
          <button
            @click="nav.openNovaOS()"
            :class="nav.isDesktopSidebarOpen ? 'lg:hidden' : 'flex'"
            class="bg-emerald-500 hover:bg-emerald-600 text-white font-bold px-3 py-2 rounded-xl text-xs shadow flex items-center gap-1 transition"
          >
            <Plus class="w-3.5 h-3.5" />
            <span>Nova OS</span>
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
        <ServicosView v-else-if="nav.currentView === 'servicos'" />
        <ClientesView v-else-if="nav.currentView === 'clientes'" />
        <RelatoriosView v-else-if="nav.currentView === 'relatorios'" />
        <ConfiguracoesView v-else-if="nav.currentView === 'configuracoes'" />
      </main>

      <!-- BOTTOM NAVIGATION BAR (FIXA NO MOBILE) -->
      <nav
        class="lg:hidden fixed bottom-0 left-0 right-0 bg-white dark:bg-slate-800 border-t border-slate-200 dark:border-slate-700 flex justify-around items-center py-2 z-30 text-xs font-medium text-slate-500 dark:text-slate-400 no-print"
      >
        <button
          @click="nav.navigate('dashboard')"
          class="flex flex-col items-center gap-1 px-3 py-1 transition"
          :class="nav.currentView === 'dashboard' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <LayoutDashboard class="w-5 h-5" />
          <span>Dashboard</span>
        </button>

        <button
          @click="nav.navigate('servicos')"
          class="flex flex-col items-center gap-1 px-3 py-1 transition"
          :class="nav.currentView === 'servicos' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <Wrench class="w-5 h-5" />
          <span>Serviços</span>
        </button>

        <button
          @click="nav.navigate('clientes')"
          class="flex flex-col items-center gap-1 px-3 py-1 transition"
          :class="nav.currentView === 'clientes' ? 'text-opti dark:text-blue-400 font-bold' : ''"
        >
          <Users class="w-5 h-5" />
          <span>Clientes</span>
        </button>

        <button
          @click="nav.navigate('relatorios')"
          class="flex flex-col items-center gap-1 px-3 py-1 transition"
          :class="nav.currentView === 'relatorios' ? 'text-opti dark:text-blue-400 font-bold' : ''"
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
