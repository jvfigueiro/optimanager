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
      toast.show('👤 Cliente salvo! Já disponível no autocompletar.')
      return novo
    } catch (err: any) {
      toast.show('Erro ao cadastrar cliente: ' + err.message, 'error')
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

  return {
    clientes,
    loading,
    searchQuery,
    loadClientes,
    setSearch,
    createCliente,
    deleteCliente,
  }
})

