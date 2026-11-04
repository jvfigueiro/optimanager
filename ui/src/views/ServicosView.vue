<script setup lang="ts">
import { onMounted } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useNavigationStore } from '../stores/navigation'
import { formatBRL, formatDate } from '../utils/formatters'
import type { OrdemServico } from '../types'
import {
  Search,
  FileText,
  MessageCircle,
  CheckCircle,
} from 'lucide-vue-next'

const ordensStore = useOrdensStore()
const nav = useNavigationStore()

onMounted(() => {
  ordensStore.loadOrdens()
})

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
  nav.navigate('relatorios')
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
            🔧 Em Aberto (Bancada/Orçamento)
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
            ✅ Concluídos
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
            ⏳ Pagamento Pendente
          </button>
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
            >
              <td class="py-3.5 px-4 whitespace-nowrap">
                <span class="font-bold text-opti dark:text-blue-400">#{{ o.id }}</span>
                <span class="block text-xs text-slate-400">{{ formatDate(o.data_emissao) }}</span>
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
                  🔑 {{ o.obs_tecnicas }}
                </span>
              </td>
              <td class="py-3.5 px-4 text-right font-bold whitespace-nowrap text-slate-900 dark:text-white">
                {{ o.valor_total > 0 ? formatBRL(o.valor_total) : 'A definir' }}
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap">
                <button
                  v-if="o.pagamento_status === 'Pendente'"
                  @click="ordensStore.liquidarOS(o.id)"
                  class="px-2.5 py-1 text-xs font-bold rounded-lg bg-amber-100 dark:bg-amber-950/60 text-amber-800 dark:text-amber-300 hover:bg-emerald-600 hover:text-white transition flex items-center gap-1"
                  title="Clique para dar baixa no caixa"
                >
                  <CheckCircle class="w-3.5 h-3.5" />
                  <span>Pendente (Baixar)</span>
                </button>
                <span v-else class="text-xs font-semibold text-slate-600 dark:text-slate-300">
                  {{ o.pagamento_forma }} (Pago)
                </span>
              </td>
              <td class="py-3.5 px-4 whitespace-nowrap text-center">
                <div class="flex items-center justify-center gap-1.5">
                  <button
                    v-if="o.cliente_contato"
                    @click="ordensStore.openWhatsApp(o)"
                    title="Enviar WhatsApp com chave Pix"
                    class="p-2 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-slate-700 rounded-lg transition"
                  >
                    <MessageCircle class="w-4 h-4" />
                  </button>
                  <button
                    @click="gerarDocumento(o)"
                    class="px-3 py-1.5 text-xs bg-opti-light dark:bg-slate-700 text-opti dark:text-sky-300 hover:bg-opti dark:hover:bg-opti hover:text-white dark:hover:text-white border border-transparent dark:border-slate-600 rounded-xl font-bold transition flex items-center gap-1"
                  >
                    <FileText class="w-3.5 h-3.5" />
                    <span>Gerar Doc</span>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!ordensStore.ordens.length">
              <td colspan="8" class="py-12 text-center text-slate-400 dark:text-slate-500">
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
          class="p-4 space-y-2.5"
        >
          <div class="flex justify-between items-start">
            <div>
              <span class="text-xs font-bold text-opti dark:text-blue-400">#{{ o.id }} • {{ formatDate(o.data_emissao) }}</span>
              <h4 class="font-bold text-slate-900 dark:text-white text-base leading-tight">{{ o.cliente_nome }}</h4>
              <span class="text-xs text-slate-500 dark:text-slate-400">Equipamento: {{ o.equipamento }}</span>
            </div>
            <div class="text-right">
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
              <button
                v-if="o.pagamento_status === 'Pendente'"
                @click="ordensStore.liquidarOS(o.id)"
                class="font-bold text-amber-600 dark:text-amber-400 underline hover:text-emerald-600"
              >
                Pendente (Clique para Baixar)
              </button>
              <strong v-else class="text-slate-800 dark:text-slate-200">{{ o.pagamento_forma }}</strong>
            </div>

            <div class="flex items-center gap-2">
              <button
                v-if="o.cliente_contato"
                @click="ordensStore.openWhatsApp(o)"
                class="p-2 text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-slate-700 rounded-lg"
              >
                <MessageCircle class="w-4 h-4" />
              </button>
              <button
                @click="gerarDocumento(o)"
                class="px-3 py-1.5 bg-opti text-white rounded-xl font-bold flex items-center gap-1 shadow-xs"
              >
                <FileText class="w-3.5 h-3.5" />
                <span>Gerar Doc</span>
              </button>
            </div>
          </div>
        </div>

        <div v-if="!ordensStore.ordens.length" class="p-8 text-center text-slate-400">
          Nenhuma ordem de serviço encontrada.
        </div>
      </div>
    </div>
  </div>
</template>
