import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Config, AutocompletePresets } from '../types'
import { api } from '../api'
import { useToastStore } from './toast'

export const DEFAULT_AUTOCOMPLETE_PRESETS: Record<string, AutocompletePresets> = {
  informatica: {
    marcas: ['Asus', 'Acer', 'Dell', 'Lenovo', 'HP', 'Apple', 'Samsung', 'LG', 'Positivo'],
    modelos: ['Notebook', 'Computador Desktop', 'All-in-One', 'Servidor', 'Monitor', 'N/A'],
    servicos: [
      'Formatação e Instalação de Sistema Operacional',
      'Limpeza Preventiva e Troca de Pasta Térmica',
      'Troca de SSD / Upgrade de Memória RAM',
      'Reparo de Placa-Mãe / Troca de Conector DC Jack',
      'Troca de Tela / Display',
      'Remoção de Vírus e Otimização de Desempenho',
    ],
  },
  celulares: {
    marcas: ['Apple (iPhone)', 'Samsung', 'Motorola', 'Xiaomi', 'LG', 'Realme'],
    modelos: ['Smartphone', 'Tablet', 'Smartwatch', 'iPad', 'Outro'],
    servicos: [
      'Troca de Tela / Display Frontal',
      'Substituição de Bateria',
      'Troca de Conector de Carga (Sub-placa)',
      'Desoxidação / Limpeza após contato com água',
      'Reparo de Alto-falante / Microfone',
      'Restauração de Software e Backup',
    ],
  },
  mecanica: {
    marcas: ['Chevrolet', 'Fiat', 'Ford', 'Honda', 'Hyundai', 'Renault', 'Toyota', 'Volkswagen'],
    modelos: ['Carro Passeio', 'SUV', 'Caminhonete', 'Utilitário', 'Moto'],
    servicos: [
      'Troca de Óleo e Filtros',
      'Revisão do Sistema de Freios (Pastilhas/Discos)',
      'Alinhamento e Balanceamento',
      'Suspensão e Amortecedores',
      'Diagnóstico Eletrônico via Scanner (Injeção)',
      'Substituição de Correia Dentada e Tensor',
    ],
  },
  confeitaria: {
    marcas: ['Artesanal / Própria', 'Tradicional', 'Diet / Zero Açúcar', 'Vegana'],
    modelos: ['Bolo de Aniversário (2kg)', 'Bolo de Casamento', 'Cento de Salgados', 'Cento de Doces Finos', 'Torta Holandesa', 'Kit Festa'],
    servicos: [
      'Confecção de Bolo Decorado Personalizado',
      'Produção de Doces Gourmet',
      'Kit Festa Completo com Entrega',
      'Degustação de Sabores',
      'Serviço de Buffet e Mesa de Doces',
    ],
  },
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
    custom_presets: JSON.stringify(DEFAULT_AUTOCOMPLETE_PRESETS.informatica),
    logo_base64: '',
    cor_primaria: '#0B5694',
    cor_secundaria: '#083f6e',
  })

  const isDarkMode = ref(false)
  const loading = ref(false)

  const presets = computed<AutocompletePresets>(() => {
    if (!config.value.custom_presets) return DEFAULT_AUTOCOMPLETE_PRESETS.informatica
    try {
      const parsed = JSON.parse(config.value.custom_presets)
      return {
        marcas: Array.isArray(parsed.marcas) ? parsed.marcas : DEFAULT_AUTOCOMPLETE_PRESETS.informatica.marcas,
        modelos: Array.isArray(parsed.modelos) ? parsed.modelos : DEFAULT_AUTOCOMPLETE_PRESETS.informatica.modelos,
        servicos: Array.isArray(parsed.servicos) ? parsed.servicos : DEFAULT_AUTOCOMPLETE_PRESETS.informatica.servicos,
      }
    } catch {
      return DEFAULT_AUTOCOMPLETE_PRESETS.informatica
    }
  })

  async function updatePresets(newPresets: AutocompletePresets) {
    return saveConfig({
      custom_presets: JSON.stringify(newPresets),
    })
  }

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
