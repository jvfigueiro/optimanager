<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useNavigationStore } from '../stores/navigation'
import { useOrdensStore } from '../stores/ordens'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import { useDashboardStore } from '../stores/dashboard'
import { mergeObsAndProblema } from '../utils/formatters'
import { X, Wrench, CheckCircle, Printer } from 'lucide-vue-next'

const nav = useNavigationStore()
const ordensStore = useOrdensStore()
const clientesStore = useClientesStore()
const configStore = useConfigStore()
const dashboardStore = useDashboardStore()

function getTodayISO() {
  const now = new Date()
  const y = now.getFullYear()
  const m = String(now.getMonth() + 1).padStart(2, '0')
  const d = String(now.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function getCurrentTimeHHMM() {
  const now = new Date()
  const h = String(now.getHours()).padStart(2, '0')
  const m = String(now.getMinutes()).padStart(2, '0')
  return `${h}:${m}`
}

const clienteNome = ref('')
const clienteContato = ref('')
const marca = ref('')
const modelo = ref('')
const problemaInformado = ref('')
const servico = ref('')
const obsTecnicas = ref('')
const dataEmissao = ref(getTodayISO())
const horaEmissao = ref(getCurrentTimeHHMM())
const status = ref('Avaliação/Orçamento')
const valorTotal = ref<number>(100)
const pagamentoStatus = ref('Pendente')
const pagamentoForma = ref('Pendente')
const isSubmitting = ref(false)

const proximoNumeroOS = computed(() => configStore.config.proximo_numero_os || configStore.presets.os_proximo_numero || 1000)

const isExistingCliente = computed(() => {
  const q = clienteNome.value.trim().toLowerCase()
  if (!q) return false
  return clientesStore.clientes.some((c) => c.nome.trim().toLowerCase() === q)
})

const matchingClientes = computed(() => {
  const q = clienteNome.value.trim().toLowerCase()
  if (!q || isExistingCliente.value) return []
  return clientesStore.clientes
    .filter((c) => c.nome.toLowerCase().includes(q) || (c.contato && c.contato.toLowerCase().includes(q)))
    .slice(0, 6)
})

function selectCliente(nome: string) {
  clienteNome.value = nome
  clienteContato.value = ''
}

watch(
  () => nav.isNovaOSModalOpen,
  (open) => {
    if (open) {
      dataEmissao.value = getTodayISO()
      horaEmissao.value = getCurrentTimeHHMM()
      const estados = configStore.presets.estados_servico
      if (estados && estados.length > 0 && !estados.includes(status.value)) {
        status.value = estados[0]
      }
    }
  }
)

// Sugestões dinâmicas que aparecem APENAS quando o usuário digita
const matchingMarcas = computed(() => {
  const query = marca.value.trim().toLowerCase()
  if (!query || configStore.presets.marcas_visivel === false) return []
  return configStore.presets.marcas
    .filter((m) => m.toLowerCase().includes(query))
    .slice(0, 8)
})

const matchingModelos = computed(() => {
  const query = modelo.value.trim().toLowerCase()
  if (!query || configStore.presets.modelos_visivel === false) return []
  return configStore.presets.modelos
    .filter((mod) => mod.toLowerCase().includes(query))
    .slice(0, 8)
})

const matchingServicos = computed(() => {
  const query = servico.value.trim().toLowerCase()
  if (!query || configStore.presets.servicos_visivel === false) return []
  return configStore.presets.servicos
    .filter((s) => s.toLowerCase().includes(query))
    .slice(0, 6)
})

function selectMarca(m: string) {
  marca.value = m
}

function selectModelo(mod: string) {
  modelo.value = mod
}

function selectServico(s: string) {
  if (!servico.value) {
    servico.value = s
  } else if (!servico.value.includes(s)) {
    servico.value = `${servico.value} + ${s}`
  }
}

// Sincroniza pagamentoStatus com a seleção rápida
watch(pagamentoForma, (nova) => {
  if (nova === 'Pendente') {
    pagamentoStatus.value = 'Pendente'
  } else {
    pagamentoStatus.value = 'Pago'
  }
})

async function salvarOrdem(printAfter = false) {
  if (!clienteNome.value.trim() || !servico.value.trim()) return

  isSubmitting.value = true
  try {
    const equipFinal = [marca.value.trim(), modelo.value.trim()].filter(Boolean).join(' ') || 'N/A'

    const novaOS = await ordensStore.createOrdem({
      data_emissao: dataEmissao.value,
      hora_emissao: horaEmissao.value,
      cliente_nome: clienteNome.value.trim(),
      cliente_contato: isExistingCliente.value ? undefined : (clienteContato.value.trim() || undefined),
      equipamento: equipFinal,
      servico: servico.value.trim(),
      obs_tecnicas: mergeObsAndProblema(problemaInformado.value, obsTecnicas.value),
      status: status.value,
      valor_total: Number(valorTotal.value) || 0,
      pagamento_status: pagamentoStatus.value,
      pagamento_forma: pagamentoForma.value,
      is_vt: false,
    })

    // Atualiza listas, clientes e próximo número no background
    clientesStore.loadClientes()
    configStore.loadConfig()
    if (nav.currentView === 'dashboard') {
      dashboardStore.loadDashboard()
    }

    // Reset
    clienteNome.value = ''
    clienteContato.value = ''
    marca.value = ''
    modelo.value = ''
    problemaInformado.value = ''
    servico.value = ''
    obsTecnicas.value = ''
    dataEmissao.value = getTodayISO()
    horaEmissao.value = getCurrentTimeHHMM()
    valorTotal.value = 100
    pagamentoForma.value = 'Pendente'
    pagamentoStatus.value = 'Pendente'

    nav.closeNovaOS()

    if (printAfter) {
      ordensStore.selectedDocTypeForDoc = 'os'
      ordensStore.selectedOSForDoc = novaOS
      nav.navigate('reports')
    }
  } catch (err) {
    console.error('Erro ao criar OS:', err)
  } finally {
    isSubmitting.value = false
  }
}

async function handleSubmit() {
  await salvarOrdem(false)
}

async function handleSalvarEImprimir() {
  await salvarOrdem(true)
}
</script>

<template>
  <div
    v-if="nav.isNovaOSModalOpen"
    class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-end sm:items-center justify-center sm:p-4 z-50 no-print"
    @click.self="nav.closeNovaOS()"
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-t-2xl sm:rounded-2xl max-w-3xl lg:max-w-4xl w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden max-h-[92vh] flex flex-col"
    >
      <!-- Header do Modal -->
      <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center shrink-0">
        <div class="flex items-center gap-2">
          <Wrench class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">Novo Serviço #{{ proximoNumeroOS }}</h3>
        </div>
        <button
          @click="nav.closeNovaOS()"
          class="text-blue-200 hover:text-white p-1 rounded-lg transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Formulário com scroll -->
      <form @submit.prevent="handleSubmit" class="p-5 space-y-3.5 text-sm overflow-y-auto flex-1">
        <!-- Data e Hora de Entrada -->
        <div class="grid grid-cols-2 gap-3 bg-slate-50 dark:bg-slate-750 p-3 rounded-xl border border-slate-200 dark:border-slate-700">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Data de Entrada *
            </label>
            <input
              type="date"
              v-model="dataEmissao"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-1.5 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Hora de Entrada *
            </label>
            <input
              type="time"
              v-model="horaEmissao"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-1.5 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>
        </div>

        <!-- Cliente, Marca & Modelo/Equipamento -->
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Cliente *
            </label>
            <input
              type="text"
              v-model="clienteNome"
              required
              placeholder="Digite para buscar ou cadastrar novo..."
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
            <!-- Sugestões de clientes cadastrados ao digitar -->
            <div v-if="matchingClientes.length > 0" class="flex flex-wrap items-center gap-1.5 mt-1.5">
              <span class="text-[10px] text-slate-400 font-semibold">Cadastrados:</span>
              <button
                v-for="cli in matchingClientes"
                :key="cli.id"
                type="button"
                @click="selectCliente(cli.nome)"
                class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-slate-700 text-opti dark:text-blue-300 border border-blue-200 dark:border-slate-600 hover:bg-opti hover:text-white transition text-[11px] font-medium"
              >
                {{ cli.nome }}<span v-if="cli.contato" class="opacity-70 ml-1">({{ cli.contato }})</span>
              </button>
            </div>
            <!-- Campo Contato quando o cliente ainda não possui cadastro -->
            <div v-if="clienteNome.trim() && !isExistingCliente" class="mt-2.5 p-2.5 rounded-lg bg-slate-50 dark:bg-slate-700/50 border border-slate-200 dark:border-slate-600">
              <label class="block text-xs font-bold uppercase text-slate-600 dark:text-slate-300 mb-1">
                Contato (Novo Cliente)
              </label>
              <input
                type="text"
                v-model="clienteContato"
                placeholder="Telefone / WhatsApp do cliente..."
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
            </div>
          </div>

          <div
            v-if="configStore.presets.marcas_visivel !== false || configStore.presets.modelos_visivel !== false"
            class="grid grid-cols-1 sm:grid-cols-2 gap-3"
          >
            <!-- Caixa Fabricante -->
            <div v-if="configStore.presets.marcas_visivel !== false">
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                {{ configStore.presets.marcas_titulo || 'Fabricante' }}
              </label>
              <input
                type="text"
                v-model="marca"
                list="marcas-datalist"
                placeholder="Ex: Asus, Dell, Honda, Apple..."
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
              <datalist id="marcas-datalist">
                <option v-for="m in configStore.presets.marcas" :key="m" :value="m" />
              </datalist>

              <!-- Sugestões de Marcas que aparecem apenas ao digitar -->
              <div v-if="matchingMarcas.length > 0" class="flex flex-wrap items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-slate-400 font-semibold">Sugestões:</span>
                <button
                  v-for="m in matchingMarcas"
                  :key="m"
                  type="button"
                  @click="selectMarca(m)"
                  class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-slate-700 text-opti dark:text-blue-300 border border-blue-200 dark:border-slate-600 hover:bg-opti hover:text-white transition text-[11px]"
                >
                  {{ m }}
                </button>
              </div>
            </div>

            <!-- Caixa Modelo -->
            <div v-if="configStore.presets.modelos_visivel !== false" :class="{ 'sm:col-span-2': configStore.presets.marcas_visivel === false }">
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                {{ configStore.presets.modelos_titulo || 'Modelo' }}
              </label>
              <input
                type="text"
                v-model="modelo"
                list="modelos-datalist"
                placeholder="Ex: Notebook, Celular, Motor, Bolo..."
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
              <datalist id="modelos-datalist">
                <option v-for="mod in configStore.presets.modelos" :key="mod" :value="mod" />
              </datalist>

              <!-- Sugestões de Modelos que aparecem apenas ao digitar -->
              <div v-if="matchingModelos.length > 0" class="flex flex-wrap items-center gap-1.5 mt-1.5">
                <span class="text-[10px] text-slate-400 font-semibold">Sugestões:</span>
                <button
                  v-for="mod in matchingModelos"
                  :key="mod"
                  type="button"
                  @click="selectModelo(mod)"
                  class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-slate-700 text-opti dark:text-blue-300 border border-blue-200 dark:border-slate-600 hover:bg-opti hover:text-white transition text-[11px]"
                >
                  {{ mod }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Seções de Texto (lado a lado em telas grandes) -->
        <div
          class="grid grid-cols-1 gap-3.5"
          :class="configStore.presets.os_secao_problema_inf_ativa !== false ? 'lg:grid-cols-2' : ''"
        >
          <!-- Descrição do Serviço / Problema Informado -->
          <div v-if="configStore.presets.os_secao_problema_inf_ativa !== false">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              {{ configStore.presets.os_secao_problema_inf_rotulo || 'Descrição do Serviço' }}
            </label>
            <textarea
              v-model="problemaInformado"
              rows="3"
              placeholder="Sintoma ou relato informado pelo cliente..."
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            ></textarea>
          </div>

          <!-- Serviço / Caixa personalizada 2 -->
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              {{
                configStore.presets.os_secao_servico_exec_ativa === true
                  ? (configStore.presets.os_secao_servico_exec_rotulo || 'Caixa personalizada 2')
                  : 'Serviço'
              }} *
            </label>
            <textarea
              v-model="servico"
              rows="3"
              required
              placeholder="O que foi executado ou precisa ser orçado..."
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            ></textarea>

            <!-- Sugestões de Serviços Comuns (Aparecem apenas ao digitar) -->
            <div v-if="matchingServicos.length > 0" class="flex flex-wrap items-center gap-1.5 mt-1.5">
              <span class="text-[10px] text-slate-400 font-semibold">Sugestões:</span>
              <button
                v-for="s in matchingServicos"
                :key="s"
                type="button"
                @click="selectServico(s)"
                class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-slate-700 text-opti dark:text-blue-300 border border-blue-200 dark:border-slate-600 hover:bg-opti hover:text-white transition text-[11px] truncate max-w-xs"
                :title="s"
              >
                + {{ s }}
              </button>
            </div>
          </div>
        </div>

        <!-- Observações Adicionais -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Observações Adicionais
          </label>
          <textarea
            v-model="obsTecnicas"
            rows="3"
            placeholder="Ex: Senha, tensão, itens adicionais..."
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
          ></textarea>
        </div>

        <!-- Situação, Valor e Pagamento -->
        <div class="grid grid-cols-3 gap-2.5">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Situação
            </label>
            <select
              v-model="status"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2 py-2 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            >
              <option
                v-for="st in (configStore.presets.estados_servico && configStore.presets.estados_servico.length ? configStore.presets.estados_servico : ['Avaliação/Orçamento', 'Em Serviço', 'Aguardando', 'Concluído'])"
                :key="st"
                :value="st"
              >
                {{ st }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Valor (R$)
            </label>
            <input
              type="number"
              step="0.01"
              v-model.number="valorTotal"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 font-bold text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Pagamento
            </label>
            <select
              v-model="pagamentoForma"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2 py-2 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            >
              <option
                v-for="forma in (configStore.presets.formas_pagamento && configStore.presets.formas_pagamento.length ? configStore.presets.formas_pagamento : ['Pendente', 'Dinheiro', 'Cartão', 'Boleto', 'Pix'])"
                :key="forma"
                :value="forma"
              >
                {{ forma }}
              </option>
            </select>
          </div>
        </div>

        <!-- Rodapé do Formulário -->
        <div class="pt-3 border-t border-slate-200 dark:border-slate-700 flex flex-wrap items-center justify-end gap-2">
          <button
            type="button"
            @click="nav.closeNovaOS()"
            class="px-3.5 py-2 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition text-xs sm:text-sm font-semibold"
          >
            Cancelar
          </button>

          <button
            type="button"
            @click="handleSalvarEImprimir"
            :disabled="isSubmitting"
            class="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-bold transition flex items-center gap-1.5 shadow text-xs sm:text-sm disabled:opacity-50"
            title="Salva a OS e abre a folha de impressão em Relatórios de Serviço"
          >
            <Printer class="w-4 h-4" />
            <span>{{ isSubmitting ? 'Salvando...' : 'Salvar Serviço e Imprimir' }}</span>
          </button>

          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-4 py-2 rounded-lg bg-opti hover:bg-opti-dark text-white font-bold transition flex items-center gap-1.5 shadow text-xs sm:text-sm disabled:opacity-50"
          >
            <CheckCircle class="w-4 h-4" />
            <span>{{ isSubmitting ? 'Salvando...' : 'Salvar OS' }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
