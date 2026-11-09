import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Cliente } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const useClientesStore = defineStore('clientes', () => {
  const toast = useToastStore()

  const clientes = ref<Cliente[]>([])
  const loading = ref(false)
  const searchQuery = ref('')

  async function loadClientes() {
    loading.value = true
    try {
      const res = await api.getClientes(searchQuery.value)
      clientes.value = res
    } catch (err: any) {
      toast.show('Erro ao listar clientes: ' + err.message, 'error')
    } finally {
      loading.value = false
    }
  }

  function setSearch(query: string) {
    searchQuery.value = query
    loadClientes()
  }

  async function createCliente(cli: Omit<Cliente, 'id'>) {
    loading.value = true
    try {
      const novo = await api.createCliente(cli)
      clientes.value.unshift(novo)
      toast.show('Cliente salvo! Já disponível no autocompletar.')
      return novo
    } catch (err: any) {
      toast.show('Erro ao cadastrar cliente: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateCliente(id: number, cli: Partial<Cliente>) {
    loading.value = true
    try {
      const updated = await api.updateCliente(id, cli)
      const idx = clientes.value.findIndex((c) => c.id === id)
      if (idx !== -1) {
        clientes.value[idx] = updated
      }
      toast.show('Cliente atualizado com sucesso!')
      return updated
    } catch (err: any) {
      toast.show('Erro ao atualizar cliente: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteCliente(id: number) {
    try {
      await api.deleteCliente(id)
      clientes.value = clientes.value.filter((c) => c.id !== id)
      toast.show('Cliente removido com sucesso.')
    } catch (err: any) {
      toast.show(err.message, 'error')
    }
  }

  async function batchDelete(ids: number[]) {
    if (!ids.length) return
    loading.value = true
    try {
      const res = await api.batchClientes({ action: 'delete', ids })
      await loadClientes()
      if (res.deleted.length > 0 && res.failed.length === 0) {
        toast.show(`${res.deleted.length} clientes excluídos com sucesso.`)
      } else if (res.deleted.length > 0 && res.failed.length > 0) {
        toast.show(
          `${res.deleted.length} clientes excluídos. ${res.failed.length} não puderam ser excluídos pois possuem ordens vinculadas.`,
          'info'
        )
      } else if (res.failed.length > 0) {
        toast.show(
          `Nenhum cliente pôde ser excluído pois todos possuem ordens de serviço vinculadas no histórico.`,
          'error'
        )
      }
      return res
    } catch (err: any) {
      toast.show('Erro ao excluir clientes em lote: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    clientes,
    loading,
    searchQuery,
    loadClientes,
    setSearch,
    createCliente,
    updateCliente,
    deleteCliente,
    batchDelete,
  }
})

