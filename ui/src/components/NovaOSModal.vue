<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useNavigationStore } from '../stores/navigation'
import { useOrdensStore } from '../stores/ordens'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import { useDashboardStore } from '../stores/dashboard'
import { X, Wrench, CheckCircle } from 'lucide-vue-next'

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
const equipamento = ref('')
const servico = ref('')
const obsTecnicas = ref('')
const dataEmissao = ref(getTodayISO())
const horaEmissao = ref(getCurrentTimeHHMM())
const status = ref('Avaliação/Orçamento')
const valorTotal = ref<number>(100)
const pagamentoStatus = ref('Pendente')
const pagamentoForma = ref('Pix')
const isSubmitting = ref(false)

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
const matchingEquipamentos = computed(() => {
  const query = equipamento.value.trim().toLowerCase()
  if (!query) return []
  const pool: string[] = []
  if (configStore.presets.modelos_visivel !== false) {
    pool.push(...configStore.presets.modelos)
  }
  if (configStore.presets.marcas_visivel !== false) {
    pool.push(...configStore.presets.marcas)
  }
  return pool
    .filter((item) => item.toLowerCase().includes(query))
    .slice(0, 8)
})

const matchingServicos = computed(() => {
  const query = servico.value.trim().toLowerCase()
  if (!query || configStore.presets.servicos_visivel === false) return []
  return configStore.presets.servicos
    .filter((s) => s.toLowerCase().includes(query))
    .slice(0, 6)
})

function appendItem(item: string) {
  if (!equipamento.value) {
    equipamento.value = item
  } else if (!equipamento.value.includes(item)) {
    equipamento.value = `${equipamento.value} ${item}`.trim()
  }
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

async function handleSubmit() {
  if (!clienteNome.value.trim() || !servico.value.trim()) return

  isSubmitting.value = true
  try {
    await ordensStore.createOrdem({
      data_emissao: dataEmissao.value,
      hora_emissao: horaEmissao.value,
      cliente_nome: clienteNome.value.trim(),
      equipamento: equipamento.value.trim() || 'N/A',
      servico: servico.value.trim(),
      obs_tecnicas: obsTecnicas.value.trim(),
      status: status.value,
      valor_total: Number(valorTotal.value) || 0,
      pagamento_status: pagamentoStatus.value,
      pagamento_forma: pagamentoForma.value,
      is_vt: false,
    })

    // Atualiza listas e clientes no background
    clientesStore.loadClientes()
    if (nav.currentView === 'dashboard') {
      dashboardStore.loadDashboard()
    }

    // Reset
    clienteNome.value = ''
    equipamento.value = ''
    servico.value = ''
    obsTecnicas.value = ''
    dataEmissao.value = getTodayISO()
    horaEmissao.value = getCurrentTimeHHMM()
    valorTotal.value = 100
    pagamentoForma.value = 'Pix'
    pagamentoStatus.value = 'Pendente'

    nav.closeNovaOS()
  } catch (err) {
    console.error('Erro ao criar OS:', err)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    v-if="nav.isNovaOSModalOpen"
    class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-end sm:items-center justify-center sm:p-4 z-50 no-print"
    @click.self="nav.closeNovaOS()"
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-t-2xl sm:rounded-2xl max-w-lg w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden max-h-[92vh] flex flex-col"
    >
      <!-- Header do Modal -->
      <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center shrink-0">
        <div class="flex items-center gap-2">
          <Wrench class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">Registrar Serviço / Orçamento</h3>
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

        <!-- Cliente & Equipamento -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Cliente (Autocompletar) *
            </label>
            <input
              type="text"
              v-model="clienteNome"
              list="clientes-datalist"
              required
              placeholder="Digite para buscar ou criar..."
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
            <datalist id="clientes-datalist">
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
              {{ configStore.presets.modelos_titulo || 'Equipamento / Objeto' }}
            </label>
            <input
              type="text"
              v-model="equipamento"
              list="equipamentos-datalist"
              placeholder="Ex: Notebook Dell Inspiron"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
            <datalist id="equipamentos-datalist">
              <option v-for="mod in configStore.presets.modelos" :key="mod" :value="mod" />
              <option v-for="m in configStore.presets.marcas" :key="m" :value="m" />
            </datalist>

            <!-- Sugestões que aparecem apenas ao digitar -->
            <div v-if="matchingEquipamentos.length > 0" class="flex flex-wrap items-center gap-1.5 mt-1.5">
              <span class="text-[10px] text-slate-400 font-semibold">Sugestões:</span>
              <button
                v-for="item in matchingEquipamentos"
                :key="item"
                type="button"
                @click="appendItem(item)"
                class="px-2 py-0.5 rounded-md bg-blue-50 dark:bg-slate-700 text-opti dark:text-blue-300 border border-blue-200 dark:border-slate-600 hover:bg-opti hover:text-white transition text-[11px]"
              >
                + {{ item }}
              </button>
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
            placeholder="O que foi executado ou precisa ser orçado..."
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
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
            placeholder="Ex: Senha de acesso, detalhes específicos, prazos..."
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-opti"
          />
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
        <div class="pt-3 border-t border-slate-200 dark:border-slate-700 flex justify-end gap-2">
          <button
            type="button"
            @click="nav.closeNovaOS()"
            class="px-4 py-2 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
          >
            Cancelar
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2 rounded-lg bg-opti hover:bg-opti-dark text-white font-bold transition flex items-center gap-1.5 shadow"
          >
            <CheckCircle class="w-4 h-4" />
            <span>{{ isSubmitting ? 'Salvando...' : 'Salvar OS' }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
