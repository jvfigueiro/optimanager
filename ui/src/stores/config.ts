import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Config, AutocompletePresets } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const DEFAULT_ESTADOS_SERVICO = ['Avaliação/Orçamento', 'Em Serviço', 'Aguardando', 'Concluído']
export const DEFAULT_FORMAS_PAGAMENTO = ['Pendente', 'Dinheiro', 'Cartão', 'Boleto', 'Pix']

export const DEFAULT_AUTOCOMPLETE_PRESETS: AutocompletePresets = {
  marcas_titulo: 'Marcas / Fabricantes',
  marcas_visivel: true,
  marcas: ['Asus', 'Acer', 'Dell', 'Lenovo', 'HP', 'Apple', 'Samsung', 'LG', 'Positivo'],

  modelos_titulo: 'Modelos / Tipos / Categorias de Item',
  modelos_visivel: true,
  modelos: ['Notebook', 'Computador Desktop', 'All-in-One', 'Servidor', 'Monitor', 'N/A'],

  servicos_titulo: 'Serviços Frequentes',
  servicos_visivel: true,
  servicos: [
    'Formatação e Instalação de Sistema Operacional',
    'Limpeza Preventiva e Troca de Pasta Térmica',
    'Troca de SSD / Upgrade de Memória RAM',
    'Reparo de Placa-Mãe / Troca de Conector DC Jack',
    'Troca de Tela / Display',
    'Remoção de Vírus e Otimização de Desempenho',
  ],

  estados_servico: DEFAULT_ESTADOS_SERVICO,
  formas_pagamento: DEFAULT_FORMAS_PAGAMENTO,
}

export const useConfigStore = defineStore('config', () => {
  const toast = useToastStore()
  const config = ref<Config>({
    id: 1,
    fantasia: 'Opti Tecnologia',
    razao_social: 'João Vítor Streb Figueiró',
    cnpj: '38.136.905/0001-66',
    responsavel: 'João Figueiró',
    fones: '(55) 99127-2332 | (55) 99674-5762',
    email_contato: '',
    endereco: '',
    cidade: 'Santa Maria - RS',
    vt_valor_padrao: 80,
    termo_garantia: 'Garantia legal de 90 dias para serviços executados e peças aplicadas.',
    termo_responsabilidade: 'Garantia legal de 90 dias conforme Art. 26 do CDC. Equipamentos não retirados em 90 dias estarão sujeitos a custas de guarda.',
    obs_padrao: 'Aparelho recebido com marcas normais de uso. Backup de dados sob responsabilidade do cliente.',
    custom_presets: JSON.stringify(DEFAULT_AUTOCOMPLETE_PRESETS),
    logo_base64: '',
    cor_primaria: '#0B5694',
    cor_secundaria: '#083f6e',
  })

  const isDarkMode = ref(false)
  const loading = ref(false)

  const presets = computed<AutocompletePresets>(() => {
    if (!config.value.custom_presets) return DEFAULT_AUTOCOMPLETE_PRESETS
    try {
      const parsed = JSON.parse(config.value.custom_presets)
      return {
        marcas_titulo: typeof parsed.marcas_titulo === 'string' && parsed.marcas_titulo.trim() ? parsed.marcas_titulo : 'Marcas / Fabricantes',
        marcas_visivel: parsed.marcas_visivel !== false,
        marcas: Array.isArray(parsed.marcas) ? parsed.marcas : DEFAULT_AUTOCOMPLETE_PRESETS.marcas,

        modelos_titulo: typeof parsed.modelos_titulo === 'string' && parsed.modelos_titulo.trim() ? parsed.modelos_titulo : 'Modelos / Tipos / Categorias de Item',
        modelos_visivel: parsed.modelos_visivel !== false,
        modelos: Array.isArray(parsed.modelos) ? parsed.modelos : DEFAULT_AUTOCOMPLETE_PRESETS.modelos,

        servicos_titulo: typeof parsed.servicos_titulo === 'string' && parsed.servicos_titulo.trim() ? parsed.servicos_titulo : 'Serviços Frequentes',
        servicos_visivel: parsed.servicos_visivel !== false,
        servicos: Array.isArray(parsed.servicos) ? parsed.servicos : DEFAULT_AUTOCOMPLETE_PRESETS.servicos,

        estados_servico: Array.isArray(parsed.estados_servico) && parsed.estados_servico.length > 0
          ? parsed.estados_servico
          : DEFAULT_ESTADOS_SERVICO,

        formas_pagamento: Array.isArray(parsed.formas_pagamento) && parsed.formas_pagamento.length > 0
          ? parsed.formas_pagamento
          : DEFAULT_FORMAS_PAGAMENTO,
      }
    } catch {
      return DEFAULT_AUTOCOMPLETE_PRESETS
    }
  })

  async function updatePresets(newPresets: AutocompletePresets) {
    return saveConfig({
      custom_presets: JSON.stringify(newPresets),
    })
  }

  // Inicializa tema: segue automaticamente o sistema operacional (com modo claro como padrão base)
  // ou respeita a escolha manual salva pelo usuário no localStorage
  function initTheme() {
    const saved = localStorage.getItem('optimanager_theme')
    if (saved === 'dark') {
      setDarkMode(true, false)
    } else if (saved === 'light') {
      setDarkMode(false, false)
    } else if (typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      setDarkMode(true, false)
    } else {
      setDarkMode(false, false)
    }

    if (typeof window !== 'undefined' && window.matchMedia) {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        if (!localStorage.getItem('optimanager_theme')) {
          setDarkMode(e.matches, false)
        }
      })
    }
  }

  function setDarkMode(val: boolean, persist = true) {
    isDarkMode.value = val
    if (val) {
      document.documentElement.classList.add('dark')
      if (persist) localStorage.setItem('optimanager_theme', 'dark')
    } else {
      document.documentElement.classList.remove('dark')
      if (persist) localStorage.setItem('optimanager_theme', 'light')
    }
  }

  function toggleTheme() {
    setDarkMode(!isDarkMode.value, true)
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
    presets,
    isDarkMode,
    loading,
    initTheme,
    toggleTheme,
    applyColorTheme,
    loadConfig,
    saveConfig,
    updatePresets,
  }
})
