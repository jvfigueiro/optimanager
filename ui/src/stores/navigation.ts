import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ViewType = 'dashboard' | 'orders' | 'clients' | 'reports' | 'settings'

export const useNavigationStore = defineStore('navigation', () => {
  const currentView = ref<ViewType>('dashboard')
  const isNovaOSModalOpen = ref(false)
  const isMobileSidebarOpen = ref(false)
  const isDesktopSidebarOpen = ref(localStorage.getItem('optimanager_sidebar_open') !== 'false')

  function navigate(view: ViewType) {
    currentView.value = view
    isMobileSidebarOpen.value = false
    window.location.hash = view
  }

  function initFromHash() {
    let hash = window.location.hash.replace('#', '').toLowerCase()
    // Backwards compatibility / alias mapping
    if (hash === 'servicos' || hash === 'serviços') hash = 'orders'
    if (hash === 'clientes') hash = 'clients'
    if (hash === 'relatorios' || hash === 'relatórios') hash = 'reports'
    if (hash === 'configuracoes' || hash === 'configurações') hash = 'settings'

    if (['dashboard', 'orders', 'clients', 'reports', 'settings'].includes(hash)) {
      currentView.value = hash as ViewType
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

  function toggleDesktopSidebar() {
    isDesktopSidebarOpen.value = !isDesktopSidebarOpen.value
    localStorage.setItem('optimanager_sidebar_open', String(isDesktopSidebarOpen.value))
  }

  function toggleSidebar() {
    if (window.innerWidth >= 1024) {
      toggleDesktopSidebar()
    } else {
      toggleMobileSidebar()
    }
  }

  return {
    currentView,
    isNovaOSModalOpen,
    isMobileSidebarOpen,
    isDesktopSidebarOpen,
    navigate,
    initFromHash,
    openNovaOS,
    closeNovaOS,
    toggleMobileSidebar,
    toggleDesktopSidebar,
    toggleSidebar,
  }
})
