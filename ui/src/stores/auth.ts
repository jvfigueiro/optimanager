import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Usuario, SetupAdminPayload, CreateUsuarioPayload, UpdateUsuarioPayload } from '../types'
import { api, TOKEN_STORAGE_KEY } from '../api'
import { useToastStore } from './toast'
import { useConfigStore } from './config'

export const useAuthStore = defineStore('auth', () => {
  const toast = useToastStore()
  const configStore = useConfigStore()

  const user = ref<Usuario | null>(null)
  const needsSetup = ref(false)
  const isAuthenticated = ref(false)
  const initialized = ref(false)
  const loading = ref(false)
  const usuarios = ref<Usuario[]>([])

  // Escuta evento de 401 global disparado pelo cliente HTTP
  if (typeof window !== 'undefined') {
    window.addEventListener('optimanager:unauthorized', () => {
      user.value = null
      isAuthenticated.value = false
    })
  }

  async function checkStatus() {
    loading.value = true
    try {
      const status = await api.getAuthStatus()
      needsSetup.value = status.needs_setup
      isAuthenticated.value = status.authenticated
      user.value = status.usuario || null
    } catch (err) {
      console.error('Erro ao verificar status de autenticação:', err)
      isAuthenticated.value = false
    } finally {
      initialized.value = true
      loading.value = false
    }
  }

  async function setupAdmin(payload: SetupAdminPayload) {
    loading.value = true
    try {
      const res = await api.setupAdmin(payload)
      localStorage.setItem(TOKEN_STORAGE_KEY, res.token)
      user.value = res.usuario
      needsSetup.value = false
      isAuthenticated.value = true
      await configStore.loadConfig()
      toast.show(`🎉 Bem-vindo(a), ${res.usuario.nome}! Empresa e administrador configurados.`)
      return res.usuario
    } catch (err: any) {
      toast.show(err.message || 'Erro ao configurar administrador', 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function login(email: string, senha: string) {
    loading.value = true
    try {
      const res = await api.login({ email, senha })
      localStorage.setItem(TOKEN_STORAGE_KEY, res.token)
      user.value = res.usuario
      needsSetup.value = false
      isAuthenticated.value = true
      toast.show(`👋 Olá, ${res.usuario.nome}!`)
      return res.usuario
    } catch (err: any) {
      toast.show(err.message || 'Credenciais inválidas', 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await api.logout()
    } catch {
      // ignora erro de rede no logout
    }
    localStorage.removeItem(TOKEN_STORAGE_KEY)
    user.value = null
    isAuthenticated.value = false
    toast.show('Sessão encerrada.')
  }

  async function loadUsuarios() {
    try {
      usuarios.value = await api.getUsuarios()
    } catch (err: any) {
      toast.show('Erro ao carregar usuários: ' + err.message, 'error')
    }
  }

  async function addUsuario(payload: CreateUsuarioPayload) {
    try {
      const novo = await api.createUsuario(payload)
      usuarios.value.push(novo)
      toast.show(`✅ Usuário ${novo.nome} adicionado com sucesso!`)
      return novo
    } catch (err: any) {
      toast.show(err.message || 'Erro ao criar usuário', 'error')
      throw err
    }
  }

  async function editUsuario(id: number, payload: UpdateUsuarioPayload) {
    try {
      const atualizado = await api.updateUsuario(id, payload)
      const idx = usuarios.value.findIndex((u) => u.id === id)
      if (idx !== -1) {
        usuarios.value[idx] = atualizado
      }
      if (user.value && user.value.id === id) {
        user.value = atualizado
      }
      toast.show(`✅ Usuário ${atualizado.nome} atualizado!`)
      return atualizado
    } catch (err: any) {
      toast.show(err.message || 'Erro ao atualizar usuário', 'error')
      throw err
    }
  }

  async function removeUsuario(id: number) {
    try {
      await api.deleteUsuario(id)
      usuarios.value = usuarios.value.filter((u) => u.id !== id)
      toast.show('🗑️ Usuário removido com sucesso.')
    } catch (err: any) {
      toast.show(err.message || 'Erro ao remover usuário', 'error')
      throw err
    }
  }

  return {
    user,
    needsSetup,
    isAuthenticated,
    initialized,
    loading,
    usuarios,
    checkStatus,
    setupAdmin,
    login,
    logout,
    loadUsuarios,
    addUsuario,
    editUsuario,
    removeUsuario,
  }
})
