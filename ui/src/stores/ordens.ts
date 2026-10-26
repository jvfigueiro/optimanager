import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { OrdemServico, CreateOrdemPayload, UpdateOrdemPayload } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const useOrdensStore = defineStore('ordens', () => {
  const toast = useToastStore()

  const ordens = ref<OrdemServico[]>([])
  const loading = ref(false)
  const currentFiltro = ref('TODOS')
  const searchQuery = ref('')
  const selectedOSForDoc = ref<OrdemServico | null>(null)

  async function loadOrdens() {
    loading.value = true
    try {
      const res = await api.getOrdens(currentFiltro.value, searchQuery.value)
      ordens.value = res
    } catch (err: any) {
      toast.show('Erro ao listar ordens: ' + err.message, 'error')
    } finally {
      loading.value = false
    }
  }

  function setFiltro(filtro: string) {
    currentFiltro.value = filtro
    loadOrdens()
  }

  function setSearch(query: string) {
    searchQuery.value = query
    loadOrdens()
  }

  async function createOrdem(payload: CreateOrdemPayload) {
    loading.value = true
    try {
      const nova = await api.createOrdem(payload)
      ordens.value.unshift(nova)
      toast.show(`✅ OS #${nova.id} registrada com sucesso!`)
      return nova
    } catch (err: any) {
      toast.show('Erro ao criar OS: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateOrdem(id: number, payload: UpdateOrdemPayload) {
    loading.value = true
    try {
      const atualizada = await api.updateOrdem(id, payload)
      const idx = ordens.value.findIndex((o) => o.id === id)
      if (idx !== -1) {
        ordens.value[idx] = atualizada
      }
      toast.show(`✅ OS #${id} atualizada com sucesso!`)
      return atualizada
    } catch (err: any) {
      toast.show('Erro ao atualizar OS: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function liquidarOS(id: number, forma = 'Pix') {
    try {
      const res = await api.baixaPagamento(id, forma, 'Pago')
      const idx = ordens.value.findIndex((o) => o.id === id)
      if (idx !== -1) {
        ordens.value[idx] = res
      }
      toast.show(`💰 Pagamento da OS #${id} baixado com sucesso!`)
      return res
    } catch (err: any) {
      toast.show('Erro ao liquidar OS: ' + err.message, 'error')
      throw err
    }
  }

  function openWhatsApp(os: OrdemServico) {
    if (!os.whatsapp_url) {
      toast.show('⚠️ Cliente sem número de telefone/WhatsApp cadastrado.', 'info')
      return
    }
    window.open(os.whatsapp_url, '_blank')
  }

  return {
    ordens,
    loading,
    currentFiltro,
    searchQuery,
    selectedOSForDoc,
    loadOrdens,
    setFiltro,
    setSearch,
    createOrdem,
    updateOrdem,
    liquidarOS,
    openWhatsApp,
  }
})
