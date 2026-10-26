import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ViewType = 'dashboard' | 'servicos' | 'clientes' | 'relatorios' | 'configuracoes'

export const useNavigationStore = defineStore('navigation', () => {
  const currentView = ref<ViewType>('dashboard')
  const isNovaOSModalOpen = ref(false)
  const isMobileSidebarOpen = ref(false)

  function navigate(view: ViewType) {
    currentView.value = view
    isMobileSidebarOpen.value = false
    window.location.hash = view
  }

  function initFromHash() {
    const hash = window.location.hash.replace('#', '') as ViewType
    if (['dashboard', 'servicos', 'clientes', 'relatorios', 'configuracoes'].includes(hash)) {
      currentView.value = hash
    }
  }

  function openNovaOS() {
    isNovaOSModalOpen.value = true
    isMobileSidebarOpen.value = false
  }

  function closeNovaOS() {
    isNovaOSModalOpen.value = false
  }

  function toggleMobileSidebar() {
    isMobileSidebarOpen.value = !isMobileSidebarOpen.value
  }

  return {
    currentView,
    isNovaOSModalOpen,
    isMobileSidebarOpen,
    navigate,
    initFromHash,
    openNovaOS,
    closeNovaOS,
    toggleMobileSidebar,
  }
})
