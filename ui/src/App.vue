<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import AppLayout from './components/AppLayout.vue'
import AuthView from './views/AuthView.vue'
import ToastContainer from './components/ToastContainer.vue'
import { useConfigStore } from './stores/config'
import { useNavigationStore } from './stores/navigation'
import { useClientesStore } from './stores/clientes'
import { useAuthStore } from './stores/auth'
import { useToastStore } from './stores/toast'

const configStore = useConfigStore()
const nav = useNavigationStore()
const clientesStore = useClientesStore()
const authStore = useAuthStore()
const toast = useToastStore()

const LAST_ACTIVITY_KEY = 'optimanager_last_activity'
let lastActivityWrite = 0
let inactivityTimer: ReturnType<typeof setInterval> | null = null
const versionConflictMessage = ref('')

function handleVersionConflict(ev: Event) {
  const detail = (ev as CustomEvent<{ message?: string }>).detail
  versionConflictMessage.value =
    detail?.message || 'Este registro foi alterado por outra pessoa. Recarregue os dados antes de salvar novamente.'
}

function reloadAfterConflict() {
  versionConflictMessage.value = ''
  window.location.reload()
}

function recordUserActivity() {
  if (!authStore.isAuthenticated || authStore.needsSetup) return
  const now = Date.now()
  if (now - lastActivityWrite > 1000) {
    lastActivityWrite = now
    localStorage.setItem(LAST_ACTIVITY_KEY, String(now))
  }
}

async function checkInactivityTimeout() {
  if (!authStore.isAuthenticated || authStore.needsSetup) return
  if (configStore.config.auto_logout_enabled !== 1) return

  const timeoutMinutos = Number(configStore.config.auto_logout_minutos || 0)
  if (timeoutMinutos <= 0) return

  const rawLast = localStorage.getItem(LAST_ACTIVITY_KEY)
  const lastActivity = rawLast ? Number(rawLast) : Date.now()
  if (!rawLast) {
    localStorage.setItem(LAST_ACTIVITY_KEY, String(lastActivity))
    return
  }

  const elapsedMs = Date.now() - lastActivity
  if (elapsedMs >= timeoutMinutos * 60 * 1000) {
    localStorage.removeItem(LAST_ACTIVITY_KEY)
    await authStore.logout()
    toast.show('Sessão encerrada automaticamente por inatividade.', 'error')
  }
}

const activityEvents = ['mousemove', 'mousedown', 'keydown', 'touchstart', 'scroll', 'click'] as const

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
    document.title = `${empresa} | Acesso | OptiManager`
    return
  }
  const tela = viewTitles[nav.currentView] || 'Painel'
  document.title = `${empresa} | ${tela} | OptiManager`
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
    localStorage.setItem(LAST_ACTIVITY_KEY, String(Date.now()))
    clientesStore.loadClientes()
  }

  for (const ev of activityEvents) {
    window.addEventListener(ev, recordUserActivity, { passive: true })
  }
  window.addEventListener('focus', checkInactivityTimeout)
  window.addEventListener('optimanager:version-conflict', handleVersionConflict)
  inactivityTimer = setInterval(checkInactivityTimeout, 5000)
})

onUnmounted(() => {
  for (const ev of activityEvents) {
    window.removeEventListener(ev, recordUserActivity)
  }
  window.removeEventListener('focus', checkInactivityTimeout)
  window.removeEventListener('optimanager:version-conflict', handleVersionConflict)
  if (inactivityTimer) {
    clearInterval(inactivityTimer)
  }
})

watch(
  () => authStore.isAuthenticated,
  (authenticated) => {
    if (authenticated) {
      localStorage.setItem(LAST_ACTIVITY_KEY, String(Date.now()))
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
    <div v-else class="min-h-screen w-full flex flex-col">
      <AuthView />
      <ToastContainer />
    </div>

    <!-- Diálogo Global de Conflito de Concorrência Otimista (409) -->
    <div
      v-if="versionConflictMessage"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs z-[100] flex items-center justify-center p-4"
    >
      <div class="bg-white dark:bg-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl border border-amber-300 dark:border-amber-700 space-y-4">
        <h3 class="text-base font-bold text-slate-900 dark:text-white">
          Conflito de Edição Detectado
        </h3>
        <p class="text-sm text-slate-600 dark:text-slate-300 leading-relaxed">
          {{ versionConflictMessage }}
        </p>
        <div class="flex justify-end gap-2 pt-2">
          <button
            type="button"
            @click="versionConflictMessage = ''"
            class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
          >
            Fechar
          </button>
          <button
            type="button"
            @click="reloadAfterConflict"
            class="px-4 py-2 rounded-xl text-xs font-bold bg-opti hover:bg-opti-dark text-white shadow-xs transition"
          >
            Recarregar
          </button>
        </div>
      </div>
    </div>
  </template>
</template>
