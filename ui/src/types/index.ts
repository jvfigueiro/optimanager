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
  vt_valor_padrao_centavos?: number
  termo_garantia: string
  termo_responsabilidade?: string
  obs_padrao?: string
  custom_presets?: string
  logo_base64: string
  cor_primaria: string
  cor_secundaria: string
  modulos_ativos?: string
  proximo_numero_os?: number
  versao?: number
  auto_logout_enabled?: number
  auto_logout_minutos?: number
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

  // Personalização de OS e Documentos
  os_proximo_numero?: number
  os_secao_problema_inf_ativa?: boolean
  os_secao_problema_inf_rotulo?: string
  os_secao_servico_exec_ativa?: boolean
  os_secao_servico_exec_rotulo?: string
}

export interface Cliente {
  id: number
  nome: string
  doc: string
  contato: string
  desde_ano: number
  versao?: number
  created_at?: string
}

export type OSStatus = string
export type PagamentoStatus = 'Pendente' | 'Pago'
export type PagamentoForma = string

export interface OrdemServico {
  id: number
  data_emissao: string
  hora_emissao?: string
  data_entrega?: string
  hora_entrega?: string
  status: OSStatus
  cliente_id: number
  cliente_nome?: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento: string
  servico: string
  obs_tecnicas: string
  valor_total: number
  valor_total_centavos?: number
  pagamento_status: PagamentoStatus
  pagamento_forma: PagamentoForma
  pago_em?: string
  versao?: number
  created_at: string
  whatsapp_url?: string
}

export interface CreateOrdemPayload {
  data_emissao?: string
  hora_emissao?: string
  data_entrega?: string
  hora_entrega?: string
  status?: string
  cliente_id?: number
  cliente_nome: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento?: string
  servico: string
  obs_tecnicas?: string
  valor_total: number
  valor_total_centavos?: number
  pagamento_status?: string
  pagamento_forma?: string
  is_vt?: boolean
}

export interface UpdateOrdemPayload {
  data_emissao?: string
  hora_emissao?: string
  data_entrega?: string
  hora_entrega?: string
  status?: string
  cliente_id?: number
  cliente_nome?: string
  cliente_doc?: string
  cliente_contato?: string
  equipamento?: string
  servico: string
  obs_tecnicas?: string
  valor_total: number
  valor_total_centavos?: number
  pagamento_status?: string
  pagamento_forma?: string
  versao?: number
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
  faturamento_hoje_centavos?: number
  caixa_recebido: number
  caixa_recebido_centavos?: number
  pendente_receber: number
  pendente_receber_centavos?: number
  faturamento_total: number
  faturamento_total_centavos?: number
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
  perfil?: 'administrador' | 'atendente' | 'tecnico'
  ativo?: boolean
  versao?: number
  permissoes?: string[]
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
  logo_base64?: string
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

export interface BackupInfo {
  filename: string
  origem: 'manual' | 'auto' | 'prerestore' | 'premigration' | string
  tamanho_bytes: number
  tamanho_formatado: string
  data_criacao: string
  data_criacao_formatada: string
  copia_externa_ok: boolean
}

export interface BackupConfigPayload {
  backup_auto_enabled: number
  backup_frequencia: 'diario' | 'semanal'
  backup_dia_semana: number
  backup_horario: string
  backup_retencao_auto: number
  backup_retencao_manual: number
  backup_retencao_prerestore: number
  backup_retencao_premigration: number
}

export interface BackupStatusResponse extends BackupConfigPayload {
  backups: BackupInfo[]
  external_dir: string
  backup_ultimo_sucesso: string
  backup_ultimo_erro: string
  backup_ultimo_erro_em: string
  backup_ultimo_aviso_externo: string
  backup_ultimo_aviso_externo_em: string
}
