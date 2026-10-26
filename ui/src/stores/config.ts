import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Config } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const useConfigStore = defineStore('config', () => {
  const toast = useToastStore()
  const config = ref<Config>({
    id: 1,
    fantasia: 'Opti Tecnologia',
    razao_social: 'João Vítor Streb Figueiró',
    cnpj: '38.136.905/0001-66',
    responsavel: 'João Figueiró',
    fones: '(55) 99127-2332 | (55) 99674-5762',
    cidade: 'Santa Maria - RS',
    vt_valor_padrao: 80,
    termo_garantia: 'Garantia legal de 90 dias para serviços executados e peças aplicadas.',
    logo_base64: '',
    cor_primaria: '#0B5694',
    cor_secundaria: '#083f6e',
  })

  const isDarkMode = ref(false)
  const loading = ref(false)

  // Inicializa tema a partir do localStorage ou preferência do sistema
  function initTheme() {
    const saved = localStorage.getItem('optimanager_theme')
    if (saved === 'dark' || (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
      setDarkMode(true)
    } else {
      setDarkMode(false)
    }
  }

  function setDarkMode(val: boolean) {
    isDarkMode.value = val
    if (val) {
      document.documentElement.classList.add('dark')
      localStorage.setItem('optimanager_theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      localStorage.setItem('optimanager_theme', 'light')
    }
  }

  function toggleTheme() {
    setDarkMode(!isDarkMode.value)
  }

  // Aplica variáveis CSS no elemento raiz do documento
  function applyColorTheme(primaryHex: string, secondaryHex: string) {
    if (!primaryHex) return
    const root = document.documentElement
    root.style.setProperty('--color-primary', primaryHex)

    if (secondaryHex) {
      root.style.setProperty('--color-primary-dark', secondaryHex)
    } else {
      root.style.setProperty('--color-primary-dark', primaryHex)
    }

    // Calcula uma versão suave para fundos (light)
    root.style.setProperty('--color-primary-light', primaryHex + '18') // ~10% de opacidade
  }

  async function loadConfig() {
    loading.value = true
    try {
      const data = await api.getConfig()
      config.value = data
      applyColorTheme(data.cor_primaria, data.cor_secundaria)
    } catch (err: any) {
      console.error('Erro ao carregar configurações:', err)
    } finally {
      loading.value = false
    }
  }

  async function saveConfig(updated: Partial<Config>) {
    loading.value = true
    try {
      const data = await api.updateConfig(updated)
      config.value = data
      applyColorTheme(data.cor_primaria, data.cor_secundaria)
      toast.show('✅ Configurações salvas com sucesso!')
      return data
    } catch (err: any) {
      toast.show('Erro ao salvar configurações: ' + err.message, 'error')
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    config,
    isDarkMode,
    loading,
    initTheme,
    toggleTheme,
    applyColorTheme,
    loadConfig,
    saveConfig,
  }
})

