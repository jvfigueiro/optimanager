import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { DashboardData, Atalho } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'
import { useConfigStore } from './config'

export const useDashboardStore = defineStore('dashboard', () => {
  const toast = useToastStore()
  const configStore = useConfigStore()

  const data = ref<DashboardData>({
    metricas: {
      faturamento_hoje: 0,
      caixa_recebido: 0,
      pendente_receber: 0,
      faturamento_total: 0,
      qtd_abertos: 0,
      qtd_pendentes: 0,
    },
    servicos_aberto: [],
    cobrancas_pendentes: [],
    atalhos: [],
  })

  const loading = ref(false)

  async function loadDashboard() {
    loading.value = true
    try {
      const res = await api.getDashboard()
      data.value = res
    } catch (err: any) {
      toast.show('Erro ao carregar dashboard: ' + err.message, 'error')
    } finally {
      loading.value = false
    }
  }

  function copyCNPJ(showNotification = true) {
    const cnpj = configStore.config.cnpj || '38.136.905/0001-66'
    navigator.clipboard.writeText(cnpj)
    if (showNotification) {
      toast.show(`📋 CNPJ ${cnpj} copiado para a área de transferência!`)
    }
  }

  function openShortcut(atalho: Atalho) {
    copyCNPJ(false)
    toast.show(`✅ CNPJ copiado! Abrindo ${atalho.titulo}...`)
    setTimeout(() => {
      window.open(atalho.url, '_blank')
    }, 400)
  }

  async function addShortcut(novo: Omit<Atalho, 'id'>) {
    try {
      const created = await api.createAtalho(novo)
      data.value.atalhos.push(created)
      toast.show('🔗 Novo atalho adicionado ao Dashboard!')
    } catch (err: any) {
      toast.show('Erro ao criar atalho: ' + err.message, 'error')
      throw err
    }
  }

  async function removeShortcut(id: number) {
    try {
      await api.deleteAtalho(id)
      data.value.atalhos = data.value.atalhos.filter((a) => a.id !== id)
      toast.show('Atalho removido.')
    } catch (err: any) {
      toast.show('Erro ao remover atalho: ' + err.message, 'error')
    }
  }

  async function liquidarOS(id: number) {
    try {
      await api.baixaPagamento(id, 'Pix', 'Pago')
      toast.show(`💰 Pagamento da OS #${id} baixado com sucesso!`)
      await loadDashboard()
    } catch (err: any) {
      toast.show('Erro ao baixar pagamento: ' + err.message, 'error')
    }
  }

  return {
    data,
    loading,
    loadDashboard,
    copyCNPJ,
    openShortcut,
    addShortcut,
    removeShortcut,
    liquidarOS,
  }
})

