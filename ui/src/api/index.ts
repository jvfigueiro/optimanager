import type {
  Config,
  Cliente,
  OrdemServico,
  CreateOrdemPayload,
  UpdateOrdemPayload,
  Atalho,
  DashboardData,
} from '../types'

const API_BASE = '/api'

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE}${endpoint}`
  const headers = new Headers(options.headers || {})

  if (!(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  if (!response.ok) {
    let errorMsg = `Erro ${response.status}: ${response.statusText}`
    try {
      const errJson = await response.json()
      if (errJson && errJson.error) {
        errorMsg = errJson.error
      }
    } catch {
      // ignora se não for JSON
    }
    throw new Error(errorMsg)
  }

  // Se resposta for 204 No Content
  if (response.status === 204) {
    return {} as T
  }

  return response.json()
}

export const api = {
  // Health
  getHealth: () => request<{ status: string }>('/health'),

  // Config
  getConfig: () => request<Config>('/config'),
  updateConfig: (cfg: Partial<Config>) =>
    request<Config>('/config', {
      method: 'PUT',
      body: JSON.stringify(cfg),
    }),

  // Dashboard
  getDashboard: () => request<DashboardData>('/dashboard'),

  // Clientes
  getClientes: (search = '') =>
    request<Cliente[]>(`/clientes?search=${encodeURIComponent(search)}`),
  createCliente: (cli: Omit<Cliente, 'id'>) =>
    request<Cliente>('/clientes', {
      method: 'POST',
      body: JSON.stringify(cli),
    }),
  updateCliente: (id: number, cli: Partial<Cliente>) =>
    request<Cliente>(`/clientes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(cli),
    }),
  deleteCliente: (id: number) =>
    request<{ message: string }>(`/clientes/${id}`, {
      method: 'DELETE',
    }),

  // Ordens de Serviço
  getOrdens: (filtro = 'TODOS', busca = '') => {
    const params = new URLSearchParams()
    if (filtro && filtro !== 'TODOS') params.set('filtro', filtro)
    if (busca) params.set('busca', busca)
    const qs = params.toString() ? `?${params.toString()}` : ''
    return request<OrdemServico[]>(`/ordens${qs}`)
  },
  getOrdemById: (id: number) => request<OrdemServico>(`/ordens/${id}`),
  createOrdem: (payload: CreateOrdemPayload) =>
    request<OrdemServico>('/ordens', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  updateOrdem: (id: number, payload: UpdateOrdemPayload) =>
    request<OrdemServico>(`/ordens/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),
  baixaPagamento: (id: number, forma = 'Pix', status = 'Pago') =>
    request<OrdemServico>(`/ordens/${id}/baixa`, {
      method: 'PATCH',
      body: JSON.stringify({ forma, status }),
    }),

  // Atalhos
  getAtalhos: () => request<Atalho[]>('/atalhos'),
  createAtalho: (a: Omit<Atalho, 'id'>) =>
    request<Atalho>('/atalhos', {
      method: 'POST',
      body: JSON.stringify(a),
    }),
  deleteAtalho: (id: number) =>
    request<{ message: string }>(`/atalhos/${id}`, {
      method: 'DELETE',
    }),
}
