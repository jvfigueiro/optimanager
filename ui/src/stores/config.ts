import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Config, AutocompletePresets } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const DEFAULT_ESTADOS_SERVICO = ['Avaliação/Orçamento', 'Em Serviço', 'Aguardando', 'Concluído']
export const DEFAULT_FORMAS_PAGAMENTO = ['Pendente', 'Dinheiro', 'Cartão', 'Boleto', 'Pix']

export const DEFAULT_AUTOCOMPLETE_PRESETS: AutocompletePresets = {
  marcas_titulo: 'Fabricante',
  marcas_visivel: true,
  marcas: ['Asus', 'Acer', 'Dell', 'Lenovo', 'HP', 'Apple', 'Samsung', 'LG', 'Positivo'],

  modelos_titulo: 'Modelo',
  modelos_visivel: true,
  modelos: ['Notebook', 'Computador Desktop', 'All-in-One', 'Servidor', 'Monitor', 'N/A'],

  servicos_titulo: 'Serviços Comuns',
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

  os_proximo_numero: 1000,
  os_secao_problema_inf_ativa: true,
  os_secao_problema_inf_rotulo: 'Descrição do Serviço',
  os_secao_servico_exec_ativa: false,
  os_secao_servico_exec_rotulo: 'Caixa personalizada 2',
}

export const useConfigStore = defineStore('config', () => {
  const toast = useToastStore()
  const config = ref<Config>({
    id: 1,
    fantasia: 'Minha Empresa Tech',
    razao_social: 'Soluções em Tecnologia & Serviços Ltda',
    cnpj: '',
    responsavel: 'Administrador',
    fones: '(11) 98765-4321',
    email_contato: 'contato@empresa.com.br',
    endereco: 'Av. Paulista, 1000 - Bela Vista',
    cidade: 'São Paulo - SP',
    vt_valor_padrao: 80,
    vt_valor_padrao_centavos: 8000,
    termo_garantia: 'Garantia legal de 90 dias.',
    termo_responsabilidade: 'Garantia legal de 90 dias conforme Art. 26 do CDC. Equipamentos não retirados em 90 dias estarão sujeitos a custas de guarda.',
    obs_padrao: '',
    custom_presets: JSON.stringify(DEFAULT_AUTOCOMPLETE_PRESETS),
    logo_base64: '',
    cor_primaria: '#0B5694',
    cor_secundaria: '#083f6e',
    auto_logout_enabled: 0,
    auto_logout_minutos: 30,
  })

  const isDarkMode = ref(false)
  const loading = ref(false)

  const presets = computed<AutocompletePresets>(() => {
    if (!config.value.custom_presets) return DEFAULT_AUTOCOMPLETE_PRESETS
    try {
      const parsed = JSON.parse(config.value.custom_presets)
      return {
        marcas_titulo: typeof parsed.marcas_titulo === 'string' && parsed.marcas_titulo.trim() ? parsed.marcas_titulo : 'Fabricante',
        marcas_visivel: parsed.marcas_visivel !== false,
        marcas: Array.isArray(parsed.marcas) ? parsed.marcas : DEFAULT_AUTOCOMPLETE_PRESETS.marcas,

        modelos_titulo: typeof parsed.modelos_titulo === 'string' && parsed.modelos_titulo.trim() ? parsed.modelos_titulo : 'Modelo',
        modelos_visivel: parsed.modelos_visivel !== false,
        modelos: Array.isArray(parsed.modelos) ? parsed.modelos : DEFAULT_AUTOCOMPLETE_PRESETS.modelos,

        servicos_titulo: typeof parsed.servicos_titulo === 'string' && parsed.servicos_titulo.trim() ? parsed.servicos_titulo : 'Serviços Comuns',
        servicos_visivel: parsed.servicos_visivel !== false,
        servicos: Array.isArray(parsed.servicos) ? parsed.servicos : DEFAULT_AUTOCOMPLETE_PRESETS.servicos,

        estados_servico: Array.isArray(parsed.estados_servico) && parsed.estados_servico.length > 0
          ? parsed.estados_servico
          : DEFAULT_ESTADOS_SERVICO,

        formas_pagamento: Array.isArray(parsed.formas_pagamento) && parsed.formas_pagamento.length > 0
          ? parsed.formas_pagamento
          : DEFAULT_FORMAS_PAGAMENTO,

        os_proximo_numero: typeof config.value.proximo_numero_os === 'number' && config.value.proximo_numero_os >= 1
          ? config.value.proximo_numero_os
          : (typeof parsed.os_proximo_numero === 'number' && parsed.os_proximo_numero >= 1
            ? parsed.os_proximo_numero
            : 1000),
        os_secao_problema_inf_ativa: parsed.os_secao_problema_inf_ativa !== false,
        os_secao_problema_inf_rotulo: typeof parsed.os_secao_problema_inf_rotulo === 'string' && parsed.os_secao_problema_inf_rotulo.trim()
          ? parsed.os_secao_problema_inf_rotulo
          : 'Descrição do Serviço',
        os_secao_servico_exec_ativa: parsed.os_secao_servico_exec_ativa === true,
        os_secao_servico_exec_rotulo: typeof parsed.os_secao_servico_exec_rotulo === 'string' && parsed.os_secao_servico_exec_rotulo.trim()
          ? parsed.os_secao_servico_exec_rotulo
          : 'Caixa personalizada 2',
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

  function parseHexRgb(hex: string): [number, number, number] | null {
    const clean = hex.trim().replace('#', '')
    if (clean.length === 3) {
      const r = parseInt(clean[0] + clean[0], 16)
      const g = parseInt(clean[1] + clean[1], 16)
      const b = parseInt(clean[2] + clean[2], 16)
      if (!Number.isNaN(r) && !Number.isNaN(g) && !Number.isNaN(b)) return [r, g, b]
    }
    if (clean.length === 6) {
      const r = parseInt(clean.slice(0, 2), 16)
      const g = parseInt(clean.slice(2, 4), 16)
      const b = parseInt(clean.slice(4, 6), 16)
      if (!Number.isNaN(r) && !Number.isNaN(g) && !Number.isNaN(b)) return [r, g, b]
    }
    return null
  }

  // Aplica variáveis CSS no elemento raiz do documento
  function applyColorTheme(primaryHex: string, secondaryHex: string) {
    if (!primaryHex) return
    const root = document.documentElement
    root.style.setProperty('--color-primary', primaryHex)

    const primaryRgb = parseHexRgb(primaryHex)
    if (primaryRgb) {
      root.style.setProperty('--color-primary-rgb', `${primaryRgb[0]} ${primaryRgb[1]} ${primaryRgb[2]}`)
      // Versão suavizada/clara para boa legibilidade em fundos escuros (dark mode)
      const softR = Math.round(primaryRgb[0] * 0.55 + 255 * 0.45)
      const softG = Math.round(primaryRgb[1] * 0.55 + 255 * 0.45)
      const softB = Math.round(primaryRgb[2] * 0.55 + 255 * 0.45)
      root.style.setProperty('--color-primary-soft', `rgb(${softR}, ${softG}, ${softB})`)
    }

    const darkHex = secondaryHex || primaryHex
    root.style.setProperty('--color-primary-dark', darkHex)
    const darkRgb = parseHexRgb(darkHex)
    if (darkRgb) {
      root.style.setProperty('--color-primary-dark-rgb', `${darkRgb[0]} ${darkRgb[1]} ${darkRgb[2]}`)
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
      toast.show('Configurações salvas com sucesso!')
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
