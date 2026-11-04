<script setup lang="ts">
import { ref, watch } from 'vue'
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

const clienteNome = ref('')
const equipamento = ref('')
const servico = ref('')
const obsTecnicas = ref('')
const status = ref('Concluído')
const valorTotal = ref<number>(100)
const pagamentoStatus = ref('Pendente')
const pagamentoForma = ref('Pix')
const isSubmitting = ref(false)

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
    status.value = 'Concluído'
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
              Equipamento / Objeto
            </label>
            <input
              type="text"
              v-model="equipamento"
              placeholder="Ex: Notebook Dell Inspiron"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>
        </div>

        <!-- Sugestões de Autocompletar: Modelos & Marcas -->
        <div class="bg-blue-50/50 dark:bg-slate-750/70 p-3 rounded-xl border border-blue-100 dark:border-slate-700 text-xs space-y-2">
          <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 font-bold uppercase text-[10px]">
            <span>Sugestões Rápidas: Modelos & Marcas</span>
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

        <!-- Descrição do Serviço -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Descrição do Serviço / Problema *
          </label>
          <textarea
            v-model="servico"
            rows="3"
            required
            placeholder="O que foi executado ou precisa ser orçado..."
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
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

        <!-- Parâmetros Técnicos / Obs -->
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Parâmetros / Credenciais (Opcional)
          </label>
          <input
            type="text"
            v-model="obsTecnicas"
            placeholder="Ex: IP: 192.168.70.1 | Wi-Fi: Escritorio | Senha: 123"
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
              <option value="Concluído">Concluído</option>
              <option value="Em Serviço">Em Serviço</option>
              <option value="Avaliação">Avaliação</option>
              <option value="Orçamento">Orçamento</option>
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
              <option value="Pendente">⏳ Pendente</option>
              <option value="Pix">Pix (Quitado)</option>
              <option value="Dinheiro">Dinheiro</option>
              <option value="Cartão">Cartão</option>
              <option value="Boleto">Boleto</option>
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
