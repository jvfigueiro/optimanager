<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import type { OrdemServico, PagamentoStatus } from '../types'
import {
  X,
  Edit2,
  Trash2,
  Copy,
  Calendar,
  DollarSign,
  CheckCircle2,
} from 'lucide-vue-next'

const props = defineProps<{
  ordem: OrdemServico | null
  isOpen: boolean
  concluding?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated', os: OrdemServico): void
  (e: 'deleted', id: number): void
  (e: 'duplicated', os: OrdemServico): void
}>()

const ordensStore = useOrdensStore()
const clientesStore = useClientesStore()
const configStore = useConfigStore()

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
const marca = ref('')
const modelo = ref('')
const servico = ref('')
const obsTecnicas = ref('')
const status = ref<string>('Avaliação')
const dataEmissao = ref('')
const horaEmissao = ref('')
const dataEntrega = ref('')
const horaEntrega = ref('')
const valorTotal = ref<number>(0)
const pagamentoStatus = ref<PagamentoStatus>('Pendente')
const pagamentoForma = ref<string>('Pix')

const isSubmitting = ref(false)
const isDeleting = ref(false)
const isDuplicating = ref(false)

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

watch(
  () => [props.ordem, props.isOpen, props.concluding] as const,
  ([o, open, isConcluding]) => {
    if (o && open) {
      clienteNome.value = o.cliente_nome || ''
      const rawEq = o.equipamento || ''
      // Tenta separar em marca e modelo se houver lista conhecida
      let foundMarca = ''
      let foundModelo = rawEq
      for (const m of configStore.presets.marcas) {
        if (rawEq.toLowerCase().startsWith(m.toLowerCase())) {
          foundMarca = m
          foundModelo = rawEq.substring(m.length).trim()
          break
        }
      }
      marca.value = foundMarca
      modelo.value = foundModelo
      servico.value = o.servico || ''
      obsTecnicas.value = o.obs_tecnicas || ''
      dataEmissao.value = o.data_emissao ? o.data_emissao.substring(0, 10) : getTodayISO()
      horaEmissao.value = o.hora_emissao || ''
      dataEntrega.value = o.data_entrega || ''
      horaEntrega.value = o.hora_entrega || ''
      valorTotal.value = o.valor_total || 0
      pagamentoStatus.value = o.pagamento_status
      pagamentoForma.value = o.pagamento_forma

      if (isConcluding) {
        status.value = 'Concluído'
        if (!dataEntrega.value) dataEntrega.value = getTodayISO()
        if (!horaEntrega.value) horaEntrega.value = getCurrentTimeHHMM()
      } else {
        status.value = o.status
      }
    }
  },
  { immediate: true }
)

watch(status, (novoStatus) => {
  if (novoStatus === 'Concluído') {
    if (!dataEntrega.value) dataEntrega.value = getTodayISO()
    if (!horaEntrega.value) horaEntrega.value = getCurrentTimeHHMM()
  }
})

async function handleUpdate() {
  if (!props.ordem || !clienteNome.value.trim() || !servico.value.trim()) return

  isSubmitting.value = true
  try {
    const equipFinal = [marca.value.trim(), modelo.value.trim()].filter(Boolean).join(' ') || 'N/A'

    const atualizada = await ordensStore.updateOrdem(props.ordem.id, {
      cliente_nome: clienteNome.value.trim(),
      equipamento: equipFinal,
      servico: servico.value.trim(),
      obs_tecnicas: obsTecnicas.value.trim(),
      status: status.value,
      data_emissao: dataEmissao.value,
      hora_emissao: horaEmissao.value,
      data_entrega: dataEntrega.value,
      hora_entrega: horaEntrega.value,
      valor_total: Number(valorTotal.value) || 0,
      pagamento_status: pagamentoStatus.value,
      pagamento_forma: pagamentoForma.value,
    })
    emit('updated', atualizada)
    emit('close')
  } catch (err) {
    console.error('Erro ao atualizar OS:', err)
  } finally {
    isSubmitting.value = false
  }
}

async function handleDuplicate() {
  if (!props.ordem) return

  isDuplicating.value = true
  try {
    const duplicada = await ordensStore.duplicarOrdem(props.ordem)
    emit('duplicated', duplicada)
    emit('close')
  } catch (err) {
    console.error('Erro ao duplicar OS:', err)
  } finally {
    isDuplicating.value = false
  }
}

async function handleDelete() {
  if (!props.ordem) return

  if (confirm(`Deseja realmente EXCLUIR permanentemente a OS #${props.ordem.id} do cliente "${props.ordem.cliente_nome}"?`)) {
    isDeleting.value = true
    try {
      await ordensStore.deleteOrdem(props.ordem.id)
      emit('deleted', props.ordem.id)
      emit('close')
    } catch (err) {
      console.error('Erro ao excluir OS:', err)
    } finally {
      isDeleting.value = false
    }
  }
}
</script>

<template>
  <div
    v-if="isOpen && ordem"
    class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-end sm:items-center justify-center sm:p-4 z-50 no-print"
    @click.self="emit('close')"
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-t-2xl sm:rounded-2xl max-w-2xl w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden max-h-[92vh] flex flex-col"
    >
      <!-- Header do Modal -->
      <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center shrink-0">
        <div class="flex items-center gap-2">
          <CheckCircle2 v-if="concluding" class="w-5 h-5 text-emerald-300" />
          <Edit2 v-else class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">
            {{ concluding ? `Concluir Serviço & Registrar Entrega — OS #${ordem.id}` : `Editar Serviço #${ordem.id}` }}
          </h3>
        </div>
        <button
          @click="emit('close')"
          class="text-blue-200 hover:text-white p-1 rounded-lg transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Formulário com Scroll -->
      <form @submit.prevent="handleUpdate" class="p-5 space-y-4 text-sm overflow-y-auto flex-1">
        <!-- Banner informativo quando concluindo OS -->
        <div
          v-if="concluding"
          class="bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 rounded-xl p-3 text-xs text-emerald-800 dark:text-emerald-200 flex items-center gap-2.5"
        >
          <CheckCircle2 class="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0" />
          <span>
            Confirme a <strong>data e hora de entrega</strong> do equipamento ao cliente e ajuste o <strong>valor final</strong> ou forma de pagamento, se necessário.
          </span>
        </div>

        <!-- Linha 1: Status, Data/Hora de Entrada e Data/Hora de Entrega -->
        <div class="space-y-3 bg-slate-50 dark:bg-slate-750 p-3.5 rounded-xl border border-slate-200 dark:border-slate-700">
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Situação / Status *
              </label>
              <select
                v-model="status"
                required
                class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm font-bold focus:ring-2 focus:ring-opti"
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
                Data de Entrada
              </label>
              <div class="relative">
                <Calendar class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  type="date"
                  v-model="dataEmissao"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg pl-9 pr-3 py-2 text-sm focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Hora de Entrada
              </label>
              <input
                type="time"
                v-model="horaEmissao"
                class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
              />
            </div>
          </div>

          <!-- Data e Hora de Entrega (Visível sempre, e destacada quando Concluído) -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2 border-t border-slate-200 dark:border-slate-700">
            <div>
              <label class="block text-xs font-bold uppercase text-emerald-700 dark:text-emerald-400 mb-1">
                Data de Entrega / Saída
              </label>
              <input
                type="date"
                v-model="dataEntrega"
                class="w-full border border-emerald-300 dark:border-emerald-700 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-emerald-500"
              />
            </div>

            <div>
              <label class="block text-xs font-bold uppercase text-emerald-700 dark:text-emerald-400 mb-1">
                Hora de Entrega / Saída
              </label>
              <input
                type="time"
                v-model="horaEntrega"
                class="w-full border border-emerald-300 dark:border-emerald-700 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-emerald-500"
              />
            </div>
          </div>
        </div>

        <!-- Linha 2: Cliente, Marca & Modelo/Equipamento -->
        <div class="space-y-3">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Cliente *
            </label>
            <input
              type="text"
              v-model="clienteNome"
              list="edit-clientes-datalist"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
            />
            <datalist id="edit-clientes-datalist">
              <option
                v-for="cli in clientesStore.clientes"
                :key="cli.id"
                :value="cli.nome"
              >
                {{ cli.doc ? `CPF/CNPJ: ${cli.doc}` : '' }}
              </option>
            </datalist>
          </div>

          <div
            v-if="configStore.presets.marcas_visivel !== false || configStore.presets.modelos_visivel !== false"
            class="grid grid-cols-1 sm:grid-cols-2 gap-3"
          >
            <!-- Caixa Marcas / Fabricantes -->
            <div v-if="configStore.presets.marcas_visivel !== false">
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                {{ configStore.presets.marcas_titulo || 'Marcas / Fabricantes' }}
              </label>
              <input
                type="text"
                v-model="marca"
                list="edit-marcas-datalist"
                placeholder="Ex: Asus, Dell, Honda, Apple..."
                class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
              />
              <datalist id="edit-marcas-datalist">
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

            <!-- Caixa Modelos / Tipos / Categorias de Item -->
            <div v-if="configStore.presets.modelos_visivel !== false" :class="{ 'sm:col-span-2': configStore.presets.marcas_visivel === false }">
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                {{ configStore.presets.modelos_titulo || 'Modelos / Tipos / Categorias de Item' }}
              </label>
              <input
                type="text"
                v-model="modelo"
                list="edit-modelos-datalist"
                placeholder="Ex: Notebook, Celular, Motor, Bolo..."
                class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
              />
              <datalist id="edit-modelos-datalist">
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

        <!-- Descrição do Serviço -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Descrição do Serviço *
          </label>
          <textarea
            v-model="servico"
            rows="3"
            required
            class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
          ></textarea>

          <!-- Sugestões de Serviços Frequentes (Aparecem apenas ao digitar) -->
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

        <!-- Observações Adicionais -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Observações Adicionais
          </label>
          <input
            type="text"
            v-model="obsTecnicas"
            placeholder="Ex: Senha: 1234, peças trocadas, detalhes..."
            class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
          />
        </div>

        <!-- Linha 3: Valor & Pagamento -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 bg-slate-50 dark:bg-slate-750 p-3.5 rounded-xl border border-slate-200 dark:border-slate-700">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Valor Total (R$)
            </label>
            <div class="relative">
              <DollarSign class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <input
                type="number"
                step="0.01"
                v-model.number="valorTotal"
                required
                class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg pl-9 pr-3 py-2 text-sm font-bold focus:ring-2 focus:ring-opti"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Status Pagamento
            </label>
            <select
              v-model="pagamentoStatus"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm font-semibold focus:ring-2 focus:ring-opti"
            >
              <option value="Pendente">⏳ Pendente</option>
              <option value="Pago">✅ Pago</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Forma de Pagamento
            </label>
            <select
              v-model="pagamentoForma"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm font-semibold focus:ring-2 focus:ring-opti"
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

        <!-- Botões de Ação do Rodapé -->
        <div class="pt-3 border-t border-slate-200 dark:border-slate-700 flex flex-wrap items-center justify-between gap-2">
          <!-- Ações Secundárias (Excluir & Duplicar) -->
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="handleDelete"
              :disabled="isDeleting"
              class="px-3 py-2 rounded-xl border border-red-200 dark:border-red-900/60 bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-300 hover:bg-red-100 text-xs font-bold transition flex items-center gap-1.5"
              title="Excluir OS permanentemente"
            >
              <Trash2 class="w-4 h-4" />
              <span>{{ isDeleting ? 'Excluindo...' : 'Excluir OS' }}</span>
            </button>

            <button
              type="button"
              @click="handleDuplicate"
              :disabled="isDuplicating"
              class="px-3 py-2 rounded-xl border border-slate-300 dark:border-slate-600 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 text-slate-700 dark:text-slate-200 text-xs font-bold transition flex items-center gap-1.5"
              title="Criar uma nova OS clonando os dados desta"
            >
              <Copy class="w-4 h-4" />
              <span>{{ isDuplicating ? 'Duplicando...' : 'Duplicar OS' }}</span>
            </button>
          </div>

          <!-- Ações Primárias (Cancelar & Salvar) -->
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="emit('close')"
              class="px-4 py-2 border border-slate-300 dark:border-slate-600 rounded-xl text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 text-xs font-semibold transition"
            >
              Cancelar
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-5 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-xs disabled:opacity-50"
            >
              <Check class="w-4 h-4" />
              <span>{{ isSubmitting ? 'Salvando...' : 'Salvar Alterações' }}</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>
