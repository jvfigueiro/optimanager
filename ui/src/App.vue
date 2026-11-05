<script setup lang="ts">
import { onMounted, watch } from 'vue'
import AppLayout from './components/AppLayout.vue'
import AuthView from './views/AuthView.vue'
import ToastContainer from './components/ToastContainer.vue'
import { useConfigStore } from './stores/config'
import { useNavigationStore } from './stores/navigation'
import { useClientesStore } from './stores/clientes'
import { useAuthStore } from './stores/auth'

const configStore = useConfigStore()
const nav = useNavigationStore()
const clientesStore = useClientesStore()
const authStore = useAuthStore()

const viewTitles: Record<string, string> = {
  dashboard: 'Dashboard',
  orders: 'Serviços',
  clients: 'Clientes',
  finance: 'Financeiro',
  reports: 'Relatórios de Serviço',
  settings: 'Configurações',
}

function updateDocumentTitle() {
  const empresa = configStore.config.fantasia || 'Empresa'
  if (!authStore.isAuthenticated || authStore.needsSetup) {
    document.title = `OptiManager | Acesso | ${empresa}`
    return
  }
  const tela = viewTitles[nav.currentView] || 'Painel'
  document.title = `OptiManager | ${tela} | ${empresa}`
}

watch(
  [() => nav.currentView, () => configStore.config.fantasia, () => authStore.isAuthenticated],
  () => {
    updateDocumentTitle()
  },
  { immediate: true }
)

onMounted(async () => {
  configStore.initTheme()
  nav.initFromHash()
  await configStore.loadConfig()
  await authStore.checkStatus()
  updateDocumentTitle()
  if (authStore.isAuthenticated) {
    clientesStore.loadClientes()
  }
})

watch(
  () => authStore.isAuthenticated,
  (authenticated) => {
    if (authenticated) {
      clientesStore.loadClientes()
    }
  }
)
</script>

<template>
  <div v-if="!authStore.initialized" class="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-900">
    <div class="w-8 h-8 border-4 border-opti border-t-transparent rounded-full animate-spin"></div>
  </div>
  <template v-else>
    <AppLayout v-if="authStore.isAuthenticated && !authStore.needsSetup" />
    <div v-else>
      <AuthView />
      <ToastContainer />
    </div>
  </template>
</template>
