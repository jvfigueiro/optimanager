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

  async function copyCNPJ(showNotification = true) {
    const cnpj = configStore.config.cnpj?.trim() || ''
    if (!cnpj) {
      if (showNotification) {
        toast.show('Nenhum CNPJ/CPF cadastrado nas configurações da empresa.', 'info')
      }
      return
    }

    let copied = false
    if (navigator?.clipboard?.writeText) {
      try {
        await navigator.clipboard.writeText(cnpj)
        copied = true
      } catch {
        copied = false
      }
    }

    if (!copied) {
      try {
        const textarea = document.createElement('textarea')
        textarea.value = cnpj
        textarea.setAttribute('readonly', '')
        textarea.style.position = 'fixed'
        textarea.style.top = '0'
        textarea.style.left = '0'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        textarea.focus()
        textarea.select()
        copied = document.execCommand('copy')
        document.body.removeChild(textarea)
      } catch {
        copied = false
      }
    }

    if (showNotification) {
      toast.show(`📋 CNPJ ${cnpj} copiado para a área de transferência!`)
    }
  }

  function openShortcut(atalho: Atalho) {
    if (!atalho?.url) return
    let dest = atalho.url.trim()
    if (!/^https?:\/\//i.test(dest)) {
      dest = 'https://' + dest
    }
    // Abre a aba imediatamente no contexto do clique síncrono para evitar bloqueio por popup blocker
    const win = window.open(dest, '_blank', 'noopener,noreferrer')
    if (win) {
      win.focus()
    } else {
      // Fallback via âncora caso window.open seja restrito
      const a = document.createElement('a')
      a.href = dest
      a.target = '_blank'
      a.rel = 'noopener noreferrer'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    }
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

  async function updateShortcut(id: number, atalho: Partial<Atalho>) {
    try {
      const updated = await api.updateAtalho(id, atalho)
      const idx = data.value.atalhos.findIndex((a) => a.id === id)
      if (idx !== -1) {
        data.value.atalhos[idx] = updated
      }
      toast.show('Atalho atualizado com sucesso!')
    } catch (err: any) {
      toast.show('Erro ao atualizar atalho: ' + err.message, 'error')
      throw err
    }
  }

  async function liquidarOS(id: number) {
    try {
      await api.baixaPagamento(id, 'Pix', 'Pago')
      toast.show(`💰 Pagamento da OS ${id} baixado com sucesso!`)
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
    updateShortcut,
    removeShortcut,
    liquidarOS,
  }
})

