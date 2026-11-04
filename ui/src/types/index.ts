export interface Config {
  id: number
  fantasia: string
  razao_social: string
  cnpj: string
  responsavel: string
  fones: string
  email_contato: string
  endereco: string
  cidade: string
  vt_valor_padrao: number
  termo_garantia: string
  termo_responsabilidade?: string
  obs_padrao?: string
  custom_presets?: string
  logo_base64: string
  cor_primaria: string
  cor_secundaria: string
}

export interface AutocompletePresets {
  marcas_titulo?: string
  marcas_visivel?: boolean
  marcas: string[]

  modelos_titulo?: string
  modelos_visivel?: boolean
  modelos: string[]

  servicos_titulo?: string
  servicos_visivel?: boolean
  servicos: string[]

  estados_servico?: string[]
  formas_pagamento?: string[]
}

export interface Cliente {
  id: number
  nome: string
  doc: string
  contato: string
  desde_ano: number
  created_at?: string
}

export type OSStatus = 'Avaliação' | 'Orçamento' | 'Em Serviço' | 'Concluído'
export type PagamentoStatus = 'Pendente' | 'Pago'
export type PagamentoForma = 'Pix' | 'Dinheiro' | 'Cartão' | 'Boleto'

export interface OrdemServico {
  id: number
  data_emissao: string
  status: OSStatus
  cliente_id: number
  cliente_nome?: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento: string
  servico: string
  obs_tecnicas: string
  valor_total: number
  pagamento_status: PagamentoStatus
  pagamento_forma: PagamentoForma
  pago_em?: string
  created_at: string
  whatsapp_url?: string
}

export interface CreateOrdemPayload {
  data_emissao?: string
  status?: string
  cliente_id?: number
  cliente_nome: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento?: string
  servico: string
  obs_tecnicas?: string
  valor_total: number
  pagamento_status?: string
  pagamento_forma?: string
  is_vt?: boolean
}

export interface UpdateOrdemPayload {
  data_emissao?: string
  status?: string
  cliente_id?: number
  cliente_nome?: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento?: string
  servico: string
  obs_tecnicas?: string
  valor_total: number
  pagamento_status?: string
  pagamento_forma?: string
}

export interface Atalho {
  id: number
  titulo: string
  descricao: string
  url: string
  ordem: number
}

export interface DashboardMetricas {
  faturamento_hoje: number
  caixa_recebido: number
  pendente_receber: number
  faturamento_total: number
  qtd_abertos: number
  qtd_pendentes: number
}

export interface DashboardData {
  metricas: DashboardMetricas
  servicos_aberto: OrdemServico[]
  cobrancas_pendentes: OrdemServico[]
  atalhos: Atalho[]
}

export interface Usuario {
  id: number
  nome: string
  email: string
  role: 'admin' | 'operador'
  created_at: string
}

export interface SetupAdminPayload {
  nome: string
  email: string
  senha: string
  fantasia?: string
  razao_social?: string
  cnpj?: string
  fones?: string
  email_contato?: string
  endereco?: string
  cidade?: string
}

export interface AuthStatusResponse {
  needs_setup: boolean
  authenticated: boolean
  usuario?: Usuario
}

export interface AuthResponse {
  token: string
  usuario: Usuario
}

export interface CreateUsuarioPayload {
  nome: string
  email: string
  senha: string
  role: 'admin' | 'operador'
}

export interface UpdateUsuarioPayload {
  nome: string
  email: string
  senha?: string
  role: 'admin' | 'operador'
}
