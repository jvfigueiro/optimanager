<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useNavigationStore } from '../stores/navigation'
import { useConfigStore } from '../stores/config'
import { formatBRL, formatDate } from '../utils/formatters'
import type { OrdemServico } from '../types'
import EditarOSModal from '../components/EditarOSModal.vue'
import {
  Search,
  FileText,
  MessageCircle,
  CheckCircle,
  Eye,
  Clock,
  Copy,
  Trash2,
  CreditCard,
  ArrowRightLeft,
  X,
  Layers,
} from 'lucide-vue-next'

const ordensStore = useOrdensStore()
const nav = useNavigationStore()
const configStore = useConfigStore()

const editingOS = ref<OrdemServico | null>(null)
const isEditModalOpen = ref(false)
const isConcludingModal = ref(false)

// Estado para Ações em Lote
const selectedOSIds = ref<number[]>([])
const isBatchBaixaModalOpen = ref(false)
const batchBaixaForma = ref('Pix')
const isBatchStatusModalOpen = ref(false)
const batchNovoStatus = ref('Concluído')
const isProcessingBatch = ref(false)

const isAllSelected = computed(() => {
  if (!ordensStore.ordens.length) return false
  return ordensStore.ordens.every((o) => selectedOSIds.value.includes(o.id))
})

const isSomeSelected = computed(() => {
  return selectedOSIds.value.length > 0 && !isAllSelected.value
})

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedOSIds.value = []
  } else {
    selectedOSIds.value = ordensStore.ordens.map((o) => o.id)
  }
}

function toggleSelect(id: number) {
  const idx = selectedOSIds.value.indexOf(id)
  if (idx !== -1) {
    selectedOSIds.value.splice(idx, 1)
  } else {
    selectedOSIds.value.push(id)
  }
}

function clearSelection() {
  selectedOSIds.value = []
}

function openBatchBaixa() {
  if (!selectedOSIds.value.length) return
  const available = configStore.presets.formas_pagamento || ['Pix', 'Dinheiro', 'Cartão de Crédito', 'Cartão de Débito', 'Boleto']
  batchBaixaForma.value = available[0] || 'Pix'
  isBatchBaixaModalOpen.value = true
}

async function handleConfirmBatchBaixa() {
  if (!selectedOSIds.value.length) return
  isProcessingBatch.value = true
  try {
    await ordensStore.batchBaixa(selectedOSIds.value, batchBaixaForma.value)
    isBatchBaixaModalOpen.value = false
    clearSelection()
  } finally {
    isProcessingBatch.value = false
  }
}

function openBatchStatus() {
  if (!selectedOSIds.value.length) return
  const available = configStore.presets.estados_servico || ['Avaliação', 'Orçamento', 'Em Serviço', 'Concluído']
  batchNovoStatus.value = available[0] || 'Concluído'
  isBatchStatusModalOpen.value = true
}

async function handleConfirmBatchStatus() {
  if (!selectedOSIds.value.length) return
  isProcessingBatch.value = true
  try {
    await ordensStore.batchStatus(selectedOSIds.value, batchNovoStatus.value)
    isBatchStatusModalOpen.value = false
    clearSelection()
  } finally {
    isProcessingBatch.value = false
  }
}

async function handleBatchDuplicar() {
  if (!selectedOSIds.value.length) return
  if (
    confirm(
      `Deseja realmente duplicar as ${selectedOSIds.value.length} ordens de serviço selecionadas? Serão geradas novas ordens com status de Orçamento para os mesmos clientes.`
    )
  ) {
    isProcessingBatch.value = true
    try {
      await ordensStore.batchDuplicar(selectedOSIds.value)
      clearSelection()
    } finally {
      isProcessingBatch.value = false
    }
  }
}

async function handleBatchDelete() {
  if (!selectedOSIds.value.length) return
  if (
    confirm(
      `Atenção: deseja realmente excluir permanentemente as ${selectedOSIds.value.length} ordens de serviço selecionadas? Esta ação é irreversível.`
    )
  ) {
    isProcessingBatch.value = true
    try {
      await ordensStore.batchDelete(selectedOSIds.value)
      clearSelection()
    } finally {
      isProcessingBatch.value = false
    }
  }
}

onMounted(() => {
  ordensStore.loadOrdens()
})

function openEditOS(os: OrdemServico) {
  editingOS.value = os
  isConcludingModal.value = false
  isEditModalOpen.value = true
}

function openConcludeOS(os: OrdemServico) {
  editingOS.value = os
  isConcludingModal.value = true
  isEditModalOpen.value = true
}

function closeEditOS() {
  isEditModalOpen.value = false
  isConcludingModal.value = false
  editingOS.value = null
}

function getStatusBadgeClass(status: string) {
  switch (status) {
    case 'Concluído':
      return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
    case 'Em Serviço':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-950/60 dark:text-blue-300 border-blue-200 dark:border-blue-800'
    case 'Avaliação':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-950/60 dark:text-purple-300 border-purple-200 dark:border-purple-800'
    case 'Orçamento':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 border-amber-200 dark:border-amber-800'
    default:
      return 'bg-slate-100 text-slate-800 dark:bg-slate-800 dark:text-slate-300 border-slate-200'
  }
}

function gerarDocumento(os: OrdemServico) {
  ordensStore.selectedOSForDoc = os
  nav.navigate('reports')
}
</script>

<template>
  <div class="space-y-4">
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs overflow-hidden">
      <!-- Barra Superior: Filtros & Busca -->
      <div class="p-4 border-b border-slate-200 dark:border-slate-700 flex flex-col md:flex-row justify-between items-stretch md:items-center gap-3 bg-slate-50/50 dark:bg-slate-800/50">
        <!-- Chips de Filtro Deslizantes -->
        <div class="flex gap-2 overflow-x-auto no-scrollbar text-xs font-semibold py-0.5">
          <button
            @click="ordensStore.setFiltro('TODOS')"
            class="px-3.5 py-2 rounded-xl transition whitespace-nowrap"
            :class="
              ordensStore.currentFiltro === 'TODOS'
                ? 'bg-opti text-white shadow-xs'
                : 'bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-600'
            "
          >
            Todos
          </button>
          <button
            @click="ordensStore.setFiltro('ABERTOS')"
            class="px-3.5 py-2 rounded-xl transition whitespace-nowrap"
            :class="
              ordensStore.currentFiltro === 'ABERTOS'
                ? 'bg-opti text-white shadow-xs'
                : 'bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-600'
            "
          >
            Aberto
          </button>
          <button
            @click="ordensStore.setFiltro('CONCLUIDOS')"
            class="px-3.5 py-2 rounded-xl transition whitespace-nowrap"
            :class="
              ordensStore.currentFiltro === 'CONCLUIDOS'
                ? 'bg-opti text-white shadow-xs'
                : 'bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-600'
            "
          >
            Concluídos
          </button>
          <button
            @click="ordensStore.setFiltro('PENDENTE')"
            class="px-3.5 py-2 rounded-xl transition whitespace-nowrap"
            :class="
              ordensStore.currentFiltro === 'PENDENTE'
                ? 'bg-opti text-white shadow-xs'
                : 'bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-300 hover:bg-amber-100'
            "
          >
            Pagamento Pendente
          </button>

          <!-- Chips dinâmicos para cada Estado de Serviço personalizado -->
          <template v-for="st in (configStore.presets.estados_servico || [])" :key="st">
            <button
              v-if="st !== 'Concluído'"
              @click="ordensStore.setFiltro(st)"
              class="px-3.5 py-2 rounded-xl transition whitespace-nowrap"
              :class="
                ordensStore.currentFiltro === st
                  ? 'bg-opti text-white shadow-xs'
                  : 'bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-600'
              "
            >
              {{ st }}
            </button>
          </template>
        </div>

        <!-- Barra de Busca Instantânea -->
        <div class="relative w-full md:w-80">
          <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
          <input
            type="text"
            v-model="ordensStore.searchQuery"
            @input="ordensStore.setSearch(ordensStore.searchQuery)"
            placeholder="Buscar cliente, equipamento ou serviço..."
            class="text-sm border border-slate-300 dark:border-slate-600 rounded-xl pl-9 pr-3 py-2 w-full bg-white dark:bg-slate-700 text-slate-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-opti"
          />
        </div>
      </div>

      <!-- VISÃO DESKTOP: Tabela de Alta Densidade -->
      <div class="hidden md:block overflow-x-auto">
        <table class="w-full text-left border-collapse text-sm">
          <thead>
            <tr class="bg-slate-100 dark:bg-slate-750 text-slate-600 dark:text-slate-300 text-xs uppercase border-b border-slate-200 dark:border-slate-700">
              <th class="py-3 px-3 w-10 text-center">
                <input
                  type="checkbox"
                  :checked="isAllSelected"
                  :indeterminate.prop="isSomeSelected"
                  @change="toggleSelectAll"
                  title="Selecionar / Desmarcar todos visíveis"
                  class="rounded border-slate-300 dark:border-slate-600 text-opti focus:ring-opti w-4 h-4 cursor-pointer"
                />
              </th>
              <th class="py-3 px-4">OS / Data</th>
              <th class="py-3 px-4">Situação</th>
              <th class="py-3 px-4">Cliente</th>
              <th class="py-3 px-4">Equipamento</th>
              <th class="py-3 px-4">Descrição do Serviço</th>
              <th class="py-3 px-4 text-right">Valor</th>
              <th class="py-3 px-4">Pagamento</th>
              <th class="py-3 px-4 text-center">Ações</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-700/60">
            <tr
              v-for="o in ordensStore.ordens"
              :key="o.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-750 transition"
              :class="selectedOSIds.includes(o.id) ? 'bg-blue-50/70 dark:bg-blue-950/40' : ''"
            >
              <td class="py-3.5 px-3 text-center whitespace-nowrap">
                <input
                  type="checkbox"
                  :checked="selectedOSIds.includes(o.id)"
                  @change="toggleSelect(o.id)"
                  class="rounded border-slate-300 dark:border-slate-600 text-opti focus:ring-opti w-4 h-4 cursor-pointer"
                />
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap">
                <span class="font-bold text-opti dark:text-blue-400">{{ o.id }}</span>
                <span class="block text-xs text-slate-400">
                  {{ formatDate(o.data_emissao) }}<span v-if="o.hora_emissao"> às {{ o.hora_emissao }}</span>
                </span>
                <span v-if="o.data_entrega" class="block text-[10px] text-emerald-600 dark:text-emerald-400 font-semibold mt-0.5">
                  Entregue: {{ formatDate(o.data_entrega) }}<span v-if="o.hora_entrega">&nbsp;às {{ o.hora_entrega }}</span>
                </span>
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap">
                <span class="px-2.5 py-0.5 text-[11px] font-bold rounded-full border" :class="getStatusBadgeClass(o.status)">
                  {{ o.status }}
                </span>
              </td>
              <td class="py-3.5 px-4 font-bold text-slate-800 dark:text-slate-100 whitespace-nowrap">
                {{ o.cliente_nome }}
              </td>
              <td class="py-3.5 px-4 text-slate-600 dark:text-slate-300 whitespace-nowrap">
                {{ o.equipamento }}
              </td>
              <td class="py-3.5 px-4 max-w-sm">
                <p class="text-slate-700 dark:text-slate-200 line-clamp-2">{{ o.servico }}</p>
                <span v-if="o.obs_tecnicas" class="inline-block mt-1 text-[11px] bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 px-2 py-0.5 rounded">
                  Obs: {{ o.obs_tecnicas }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-right font-bold whitespace-nowrap text-slate-900 dark:text-white">
                {{ o.valor_total > 0 ? formatBRL(o.valor_total) : 'A definir' }}
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap">
                <span
                  v-if="o.pagamento_status === 'Pendente'"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-800/60 inline-flex items-center gap-1"
                >
                  <Clock class="w-3.5 h-3.5" />
                  <span>Pendente</span>
                </span>
                <span v-else class="text-xs font-semibold text-slate-600 dark:text-slate-300">
                  {{ o.pagamento_forma }} (Pago)
                </span>
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <!-- Concluir OS -->
                  <button
                    v-if="o.status !== 'Concluído'"
                    @click="openConcludeOS(o)"
                    title="Marcar serviço como concluído e registrar data/hora de entrega"
                    class="px-2.5 py-1.5 text-xs bg-emerald-50 hover:bg-emerald-600 text-emerald-700 hover:text-white dark:bg-emerald-950/60 dark:text-emerald-300 dark:hover:bg-emerald-600 dark:hover:text-white rounded-xl font-bold transition flex items-center gap-1 border border-emerald-200 dark:border-emerald-800/70"
                  >
                    <CheckCircle class="w-3.5 h-3.5" />
                    <span>Concluir</span>
                  </button>

                  <!-- Visualizar OS -->
                  <button
                    @click="openEditOS(o)"
                    title="Visualizar e editar detalhes da OS"
                    class="px-2.5 py-1.5 text-xs bg-slate-100 hover:bg-opti-light text-slate-700 dark:text-slate-200 dark:bg-slate-700 dark:hover:bg-slate-600 rounded-xl font-bold transition flex items-center gap-1 border border-slate-200 dark:border-slate-600"
                  >
                    <Eye class="w-3.5 h-3.5 text-opti dark:text-blue-400" />
                    <span>Visualizar</span>
                  </button>

                  <!-- Gerar Documento / Impressão -->
                  <button
                    @click="gerarDocumento(o)"
                    title="Imprimir ou gerar PDF"
                    class="px-2.5 py-1.5 text-xs bg-opti-light dark:bg-slate-700 text-opti dark:text-sky-300 hover:bg-opti dark:hover:bg-opti hover:text-white dark:hover:text-white border border-transparent dark:border-slate-600 rounded-xl font-bold transition flex items-center gap-1"
                  >
                    <FileText class="w-3.5 h-3.5" />
                    <span>Doc</span>
                  </button>

                  <!-- WhatsApp -->
                  <button
                    v-if="o.cliente_contato"
                    @click="ordensStore.openWhatsApp(o)"
                    title="Enviar WhatsApp com chave Pix"
                    class="p-1.5 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-slate-700 rounded-lg transition"
                  >
                    <MessageCircle class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!ordensStore.ordens.length">
              <td colspan="9" class="py-12 text-center text-slate-400 dark:text-slate-500">
                Nenhuma ordem de serviço encontrada com os filtros atuais.
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- VISÃO MOBILE: Cards Verticais Touch-friendly -->
      <div class="md:hidden divide-y divide-slate-200 dark:divide-slate-700/60">
        <div
          v-for="o in ordensStore.ordens"
          :key="o.id"
          class="p-4 space-y-2.5 transition"
          :class="selectedOSIds.includes(o.id) ? 'bg-blue-50/70 dark:bg-blue-950/40' : ''"
        >
          <div class="flex justify-between items-start gap-2">
            <div class="flex items-start gap-2.5 min-w-0">
              <input
                type="checkbox"
                :checked="selectedOSIds.includes(o.id)"
                @change="toggleSelect(o.id)"
                class="rounded border-slate-300 dark:border-slate-600 text-opti focus:ring-opti w-4 h-4 mt-0.5 cursor-pointer shrink-0"
              />
              <div class="min-w-0">
                <span class="text-xs font-bold text-opti dark:text-blue-400">
                  {{ o.id }} • {{ formatDate(o.data_emissao) }}<span v-if="o.hora_emissao"> às {{ o.hora_emissao }}</span>
                </span>
                <span v-if="o.data_entrega" class="block text-[10px] text-emerald-600 dark:text-emerald-400 font-semibold">
                  Entregue: {{ formatDate(o.data_entrega) }}<span v-if="o.hora_entrega">&nbsp;às {{ o.hora_entrega }}</span>
                </span>
                <h4 class="font-bold text-slate-900 dark:text-white text-base leading-tight">{{ o.cliente_nome }}</h4>
                <span class="text-xs text-slate-500 dark:text-slate-400">Equipamento: {{ o.equipamento }}</span>
              </div>
            </div>
            <div class="text-right shrink-0">
              <span class="px-2.5 py-0.5 text-[11px] font-bold rounded-full border inline-block" :class="getStatusBadgeClass(o.status)">
                {{ o.status }}
              </span>
              <p class="font-black text-sm mt-1 text-slate-900 dark:text-white">
                {{ o.valor_total > 0 ? formatBRL(o.valor_total) : 'A definir' }}
              </p>
            </div>
          </div>

          <p class="text-xs text-slate-700 dark:text-slate-300 bg-slate-50 dark:bg-slate-750 p-3 rounded-xl border border-slate-200 dark:border-slate-700 leading-relaxed">
            {{ o.servico }}
          </p>

          <div class="flex justify-between items-center pt-1 text-xs">
            <div>
              <span class="text-slate-500 dark:text-slate-400">Pgto: </span>
              <span
                v-if="o.pagamento_status === 'Pendente'"
                class="font-bold text-amber-600 dark:text-amber-400"
              >
                Pendente
              </span>
              <strong v-else class="text-slate-800 dark:text-slate-200">{{ o.pagamento_forma }} (Pago)</strong>
            </div>

            <div class="flex items-center gap-1.5 flex-wrap justify-end">
              <!-- Concluir -->
              <button
                v-if="o.status !== 'Concluído'"
                @click="openConcludeOS(o)"
                class="px-2.5 py-1.5 bg-emerald-50 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800/70 rounded-xl font-bold flex items-center gap-1"
              >
                <CheckCircle class="w-3.5 h-3.5" />
                <span>Concluir</span>
              </button>

              <!-- Visualizar -->
              <button
                @click="openEditOS(o)"
                class="px-2.5 py-1.5 bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 rounded-xl font-bold flex items-center gap-1"
              >
                <Eye class="w-3.5 h-3.5 text-opti dark:text-blue-400" />
                <span>Visualizar</span>
              </button>

              <!-- Gerar Doc -->
              <button
                @click="gerarDocumento(o)"
                class="px-2.5 py-1.5 bg-opti text-white rounded-xl font-bold flex items-center gap-1 shadow-xs"
              >
                <FileText class="w-3.5 h-3.5" />
                <span>Doc</span>
              </button>

              <!-- WhatsApp -->
              <button
                v-if="o.cliente_contato"
                @click="ordensStore.openWhatsApp(o)"
                title="WhatsApp"
                class="p-1.5 text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-slate-700 rounded-lg"
              >
                <MessageCircle class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <div v-if="!ordensStore.ordens.length" class="p-8 text-center text-slate-400">
          Nenhuma ordem de serviço encontrada.
        </div>
      </div>
    </div>

    <!-- Barra Flutuante de Ações em Lote para Serviços -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="translate-y-12 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="translate-y-12 opacity-0"
    >
      <div
        v-if="selectedOSIds.length > 0"
        class="fixed bottom-6 left-1/2 -translate-x-1/2 z-40 bg-slate-900/95 dark:bg-slate-800/95 text-white backdrop-blur shadow-2xl rounded-2xl px-4 py-2.5 border border-slate-700 flex items-center gap-2 sm:gap-3 flex-wrap max-w-[95vw] sm:max-w-max"
      >
        <div class="flex items-center gap-2 pr-2 border-r border-slate-700 text-xs font-bold whitespace-nowrap">
          <Layers class="w-4 h-4 text-blue-400" />
          <span>{{ selectedOSIds.length }} selecionada(s)</span>
        </div>

        <button
          @click="openBatchBaixa"
          :disabled="isProcessingBatch"
          class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-xs font-bold flex items-center gap-1.5 transition shadow-xs"
          title="Marcar ordens selecionadas como pagas"
        >
          <CreditCard class="w-3.5 h-3.5" />
          <span>Dar Baixa</span>
        </button>

        <button
          @click="openBatchStatus"
          :disabled="isProcessingBatch"
          class="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-bold flex items-center gap-1.5 transition shadow-xs"
          title="Alterar situação das ordens selecionadas"
        >
          <ArrowRightLeft class="w-3.5 h-3.5" />
          <span>Mudar Estado</span>
        </button>

        <button
          @click="handleBatchDuplicar"
          :disabled="isProcessingBatch"
          class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-xs font-bold flex items-center gap-1.5 transition shadow-xs"
          title="Duplicar ordens selecionadas"
        >
          <Copy class="w-3.5 h-3.5" />
          <span>Duplicar</span>
        </button>

        <button
          @click="handleBatchDelete"
          :disabled="isProcessingBatch"
          class="px-3 py-1.5 bg-red-600/90 hover:bg-red-600 text-white rounded-xl text-xs font-bold flex items-center gap-1.5 transition shadow-xs"
          title="Excluir ordens selecionadas"
        >
          <Trash2 class="w-3.5 h-3.5" />
          <span>Excluir</span>
        </button>

        <button
          @click="clearSelection"
          class="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition"
          title="Limpar seleção"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
    </Transition>

    <!-- Modal: Dar Baixa em Lote -->
    <div
      v-if="isBatchBaixaModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 w-full max-w-md p-5 space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-700">
          <h3 class="font-bold text-base text-slate-800 dark:text-white flex items-center gap-2">
            <CreditCard class="w-5 h-5 text-emerald-500" />
            <span>Dar Baixa em {{ selectedOSIds.length }} OS</span>
          </h3>
          <button @click="isBatchBaixaModalOpen = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
            <X class="w-5 h-5" />
          </button>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300">
          Selecione a forma de pagamento aplicada. Todas as <strong>{{ selectedOSIds.length }}</strong> ordens selecionadas serão marcadas como <strong>Concluído</strong> e <strong>Pago</strong> na data de hoje.
        </p>

        <div>
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
            Forma de Pagamento
          </label>
          <select
            v-model="batchBaixaForma"
            class="w-full border border-slate-300 dark:border-slate-600 rounded-xl px-3 py-2 text-sm bg-slate-50 dark:bg-slate-700 text-slate-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-opti"
          >
            <option
              v-for="fp in (configStore.presets.formas_pagamento || ['Pix', 'Dinheiro', 'Cartão de Crédito', 'Cartão de Débito', 'Boleto'])"
              :key="fp"
              :value="fp"
            >
              {{ fp }}
            </option>
          </select>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button
            type="button"
            @click="isBatchBaixaModalOpen = false"
            class="px-4 py-2 rounded-xl text-xs font-bold border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700"
          >
            Cancelar
          </button>
          <button
            type="button"
            :disabled="isProcessingBatch"
            @click="handleConfirmBatchBaixa"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-xs font-bold transition shadow-xs"
          >
            {{ isProcessingBatch ? 'Processando...' : 'Confirmar Baixa em Lote' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal: Mudar Estado em Lote -->
    <div
      v-if="isBatchStatusModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 w-full max-w-md p-5 space-y-4">
        <div class="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-700">
          <h3 class="font-bold text-base text-slate-800 dark:text-white flex items-center gap-2">
            <ArrowRightLeft class="w-5 h-5 text-blue-500" />
            <span>Mudar Estado de {{ selectedOSIds.length }} OS</span>
          </h3>
          <button @click="isBatchStatusModalOpen = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
            <X class="w-5 h-5" />
          </button>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300">
          Selecione a nova situação para as <strong>{{ selectedOSIds.length }}</strong> ordens selecionadas (conforme os estados configurados no sistema).
        </p>

        <div>
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
            Nova Situação / Estado
          </label>
          <select
            v-model="batchNovoStatus"
            class="w-full border border-slate-300 dark:border-slate-600 rounded-xl px-3 py-2 text-sm bg-slate-50 dark:bg-slate-700 text-slate-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-opti"
          >
            <option
              v-for="st in (configStore.presets.estados_servico || ['Avaliação', 'Orçamento', 'Em Serviço', 'Concluído'])"
              :key="st"
              :value="st"
            >
              {{ st }}
            </option>
          </select>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button
            type="button"
            @click="isBatchStatusModalOpen = false"
            class="px-4 py-2 rounded-xl text-xs font-bold border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700"
          >
            Cancelar
          </button>
          <button
            type="button"
            :disabled="isProcessingBatch"
            @click="handleConfirmBatchStatus"
            class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-bold transition shadow-xs"
          >
            {{ isProcessingBatch ? 'Processando...' : 'Aplicar Novo Estado' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Editar Ordem de Serviço -->
    <EditarOSModal
      :ordem="editingOS"
      :is-open="isEditModalOpen"
      :concluding="isConcludingModal"
      @close="closeEditOS"
      @updated="() => ordensStore.loadOrdens()"
      @deleted="() => ordensStore.loadOrdens()"
      @duplicated="() => ordensStore.loadOrdens()"
    />
  </div>
</template>
