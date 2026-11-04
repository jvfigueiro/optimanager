<script setup lang="ts">
import { ref, watch } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import type { OrdemServico, OSStatus, PagamentoStatus, PagamentoForma } from '../types'
import {
  X,
  Edit2,
  Trash2,
  Copy,
  Check,
  Calendar,
  DollarSign,
} from 'lucide-vue-next'

const props = defineProps<{
  ordem: OrdemServico | null
  isOpen: boolean
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

const clienteNome = ref('')
const equipamento = ref('')
const servico = ref('')
const obsTecnicas = ref('')
const status = ref<OSStatus>('Avaliação')
const dataEmissao = ref('')
const valorTotal = ref<number>(0)
const pagamentoStatus = ref<PagamentoStatus>('Pendente')
const pagamentoForma = ref<PagamentoForma>('Pix')

const isSubmitting = ref(false)
const isDeleting = ref(false)
const isDuplicating = ref(false)

watch(
  () => props.ordem,
  (o) => {
    if (o) {
      clienteNome.value = o.cliente_nome || ''
      equipamento.value = o.equipamento || ''
      servico.value = o.servico || ''
      obsTecnicas.value = o.obs_tecnicas || ''
      status.value = o.status
      dataEmissao.value = o.data_emissao ? o.data_emissao.substring(0, 10) : ''
      valorTotal.value = o.valor_total || 0
      pagamentoStatus.value = o.pagamento_status
      pagamentoForma.value = o.pagamento_forma
    }
  },
  { immediate: true }
)

function appendMarca(m: string) {
  if (!equipamento.value) {
    equipamento.value = m
  } else if (!equipamento.value.includes(m)) {
    equipamento.value = `${equipamento.value} ${m}`.trim()
  }
}

function appendModelo(mod: string) {
  if (!equipamento.value) {
    equipamento.value = mod
  } else if (!equipamento.value.includes(mod)) {
    equipamento.value = `${mod} ${equipamento.value}`.trim()
  }
}

function selectServico(s: string) {
  if (!servico.value) {
    servico.value = s
  } else {
    servico.value = `${servico.value} + ${s}`
  }
}

async function handleUpdate() {
  if (!props.ordem || !clienteNome.value.trim() || !servico.value.trim()) return

  isSubmitting.value = true
  try {
    const atualizada = await ordensStore.updateOrdem(props.ordem.id, {
      cliente_nome: clienteNome.value.trim(),
      equipamento: equipamento.value.trim() || 'N/A',
      servico: servico.value.trim(),
      obs_tecnicas: obsTecnicas.value.trim(),
      status: status.value,
      data_emissao: dataEmissao.value,
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
          <Edit2 class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">
            Editar Ordem de Serviço #{{ ordem.id }}
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
        <!-- Linha 1: Status & Data de Emissão -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 bg-slate-50 dark:bg-slate-750 p-3 rounded-xl border border-slate-200 dark:border-slate-700">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Situação / Status *
            </label>
            <select
              v-model="status"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm font-bold focus:ring-2 focus:ring-opti"
            >
              <option value="Avaliação">🟣 Avaliação</option>
              <option value="Orçamento">🟡 Orçamento</option>
              <option value="Em Serviço">🔵 Em Serviço</option>
              <option value="Concluído">🟢 Concluído</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Data de Emissão
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
        </div>

        <!-- Linha 2: Cliente & Equipamento -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
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

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Equipamento / Objeto
            </label>
            <input
              type="text"
              v-model="equipamento"
              placeholder="Ex: Notebook Dell Inspiron"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
            />
          </div>
        </div>

        <!-- Sugestões de Autocompletar: Marcas e Modelos -->
        <div class="bg-blue-50/50 dark:bg-slate-750/70 p-3 rounded-xl border border-blue-100 dark:border-slate-700 text-xs space-y-2">
          <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 font-bold uppercase text-[10px]">
            <span>Sugestões Rápidas de Modelo & Marca</span>
            <span class="font-normal text-slate-400">Clique para preencher</span>
          </div>

          <!-- Chips Modelos -->
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="mod in configStore.presets.modelos"
              :key="mod"
              type="button"
              @click="appendModelo(mod)"
              class="px-2 py-0.5 rounded-md bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-opti hover:text-opti transition text-[11px]"
            >
              + {{ mod }}
            </button>
          </div>

          <!-- Chips Marcas -->
          <div class="flex flex-wrap gap-1.5 pt-1 border-t border-slate-200/60 dark:border-slate-700/60">
            <button
              v-for="m in configStore.presets.marcas"
              :key="m"
              type="button"
              @click="appendMarca(m)"
              class="px-2 py-0.5 rounded-md bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:border-opti hover:text-opti transition text-[11px]"
            >
              + {{ m }}
            </button>
          </div>
        </div>

        <!-- Descrição do Serviço / Problema -->
        <div>
          <div class="flex justify-between items-center mb-1">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400">
              Descrição do Serviço / Problema *
            </label>
          </div>
          <textarea
            v-model="servico"
            rows="3"
            required
            class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-opti"
          ></textarea>

          <!-- Sugestões de Serviços Frequentes -->
          <div class="flex flex-wrap gap-1.5 mt-1.5">
            <button
              v-for="s in configStore.presets.servicos"
              :key="s"
              type="button"
              @click="selectServico(s)"
              class="px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-opti-light hover:text-opti transition text-[11px] truncate max-w-xs"
              :title="s"
            >
              + {{ s }}
            </button>
          </div>
        </div>

        <!-- Observações Técnicas -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Observações Técnicas / Senhas / Parâmetros
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
              <option value="Pix">Pix</option>
              <option value="Dinheiro">Dinheiro</option>
              <option value="Cartão">Cartão</option>
              <option value="Boleto">Boleto</option>
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
