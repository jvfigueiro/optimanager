<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useConfigStore } from '../stores/config'
import { formatBRL, formatDate } from '../utils/formatters'
import { Printer, MessageCircle, FileText, ShieldCheck, Download } from 'lucide-vue-next'

const ordensStore = useOrdensStore()
const configStore = useConfigStore()

type DocType = 'relatorio' | 'orcamento' | 'recibo'
const docType = ref<DocType>('relatorio')

const selectedOSId = ref<number | null>(null)
const docCliente = ref('')
const docCpfCnpj = ref('')
const docEquipamento = ref('')
const docValor = ref(0)
const docServico = ref('')
const docObs = ref('')
const docData = ref('')
const docContato = ref('')

onMounted(async () => {
  if (!ordensStore.ordens.length) {
    await ordensStore.loadOrdens()
  }

  // Se uma OS foi selecionada anteriormente a partir da tabela de serviços
  if (ordensStore.selectedOSForDoc) {
    selectedOSId.value = ordensStore.selectedOSForDoc.id
    carregarOSNoGerador(ordensStore.selectedOSForDoc.id)
    ordensStore.selectedOSForDoc = null
  } else if (ordensStore.ordens.length) {
    selectedOSId.value = ordensStore.ordens[0].id
    carregarOSNoGerador(ordensStore.ordens[0].id)
  }
})

watch(selectedOSId, (novoId) => {
  if (novoId) {
    carregarOSNoGerador(novoId)
  }
})

function carregarOSNoGerador(id: number) {
  const os = ordensStore.ordens.find((o) => o.id === id)
  if (!os) return

  docCliente.value = os.cliente_nome || ''
  docCpfCnpj.value = os.cliente_doc || ''
  docEquipamento.value = os.equipamento || 'N/A'
  docValor.value = os.valor_total || 0
  docServico.value = os.servico || ''
  docObs.value = os.obs_tecnicas || ''
  docData.value = formatDate(os.data_emissao) || new Date().toLocaleDateString('pt-BR')
  docContato.value = os.cliente_contato || ''
}

const docTitulo = computed(() => {
  switch (docType.value) {
    case 'relatorio':
      return 'Relatório de Serviço'
    case 'orcamento':
      return 'Orçamento Técnico'
    case 'recibo':
      return 'Recibo de Pagamento'
  }
})

function dispararWhatsApp() {
  const cleanPhone = docContato.value.replace(/\D+/g, '')
  if (!cleanPhone) {
    alert('Cliente sem número de telefone cadastrado.')
    return
  }
  const fone = (cleanPhone.length === 10 || cleanPhone.length === 11) ? '55' + cleanPhone : cleanPhone

  const msg = `*${configStore.config.fantasia}* - ${docTitulo.value}\nRef: OS #${selectedOSId.value || 'N/A'}\n\n` +
    `*Cliente:* ${docCliente.value}\n` +
    `*Equipamento:* ${docEquipamento.value}\n` +
    `*Serviço:* ${docServico.value}\n` +
    `*Valor:* ${formatBRL(docValor.value)}\n\n` +
    `Chave Pix (CNPJ): ${configStore.config.cnpj}\n\n` +
    `Qualquer dúvida estamos à disposição!`

  window.open(`https://wa.me/${fone}?text=${encodeURIComponent(msg)}`, '_blank')
}

function imprimirDocumento() {
  window.print()
}

function baixarPDFNativo() {
  if (!selectedOSId.value) return
  window.open(`/api/ordens/${selectedOSId.value}/pdf?tipo=${docType.value}`, '_blank')
}
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
    <!-- PAINEL DE CONFIGURAÇÃO (ESQUERDA) -->
    <div class="lg:col-span-5 bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5 space-y-4 no-print">
      <div>
        <h3 class="font-bold text-base text-slate-800 dark:text-white flex items-center gap-2">
          <FileText class="w-5 h-5 text-opti" />
          <span>Gerador de Documentos</span>
        </h3>
        <p class="text-xs text-slate-500 dark:text-slate-400">
          Selecione uma Ordem de Serviço para carregar os dados ou edite os campos em tempo real.
        </p>
      </div>

      <!-- 1. Seletor de OS -->
      <div>
        <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
          1. Puxar dados da Ordem de Serviço
        </label>
        <select
          v-model="selectedOSId"
          class="w-full border border-slate-300 dark:border-slate-600 rounded-xl px-3 py-2 text-sm bg-slate-50 dark:bg-slate-700 text-slate-800 dark:text-white font-semibold focus:outline-none focus:ring-2 focus:ring-opti"
        >
          <option
            v-for="o in ordensStore.ordens"
            :key="o.id"
            :value="o.id"
          >
            OS #{{ o.id }} - {{ o.cliente_nome }} ({{ formatDate(o.data_emissao) }})
          </option>
        </select>
      </div>

      <!-- 2. Tipo de Documento -->
      <div>
        <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
          2. Tipo de Documento
        </label>
        <div class="grid grid-cols-3 gap-2 text-xs font-bold">
          <button
            type="button"
            @click="docType = 'relatorio'"
            class="py-2.5 px-2 rounded-xl transition border text-center"
            :class="
              docType === 'relatorio'
                ? 'bg-opti text-white border-opti shadow-xs'
                : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border-slate-300 dark:border-slate-600 hover:bg-slate-50'
            "
          >
            Relatório / OS
          </button>
          <button
            type="button"
            @click="docType = 'orcamento'"
            class="py-2.5 px-2 rounded-xl transition border text-center"
            :class="
              docType === 'orcamento'
                ? 'bg-opti text-white border-opti shadow-xs'
                : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border-slate-300 dark:border-slate-600 hover:bg-slate-50'
            "
          >
            Orçamento
          </button>
          <button
            type="button"
            @click="docType = 'recibo'"
            class="py-2.5 px-2 rounded-xl transition border text-center"
            :class="
              docType === 'recibo'
                ? 'bg-opti text-white border-opti shadow-xs'
                : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border-slate-300 dark:border-slate-600 hover:bg-slate-50'
            "
          >
            Recibo Pgto
          </button>
        </div>
      </div>

      <!-- 3. Campos Editáveis em Tempo Real -->
      <div class="space-y-3 pt-2 border-t border-slate-200 dark:border-slate-700 text-xs">
        <div class="grid grid-cols-2 gap-2">
          <div>
            <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">Cliente</label>
            <input
              type="text"
              v-model="docCliente"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-opti"
            />
          </div>
          <div>
            <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">CPF / CNPJ</label>
            <input
              type="text"
              v-model="docCpfCnpj"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-opti"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-2">
          <div>
            <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">Equipamento</label>
            <input
              type="text"
              v-model="docEquipamento"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-opti"
            />
          </div>
          <div>
            <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">Valor Total (R$)</label>
            <input
              type="number"
              step="0.01"
              v-model.number="docValor"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 font-bold focus:outline-none focus:ring-1 focus:ring-opti"
            />
          </div>
        </div>

        <div>
          <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">
            Descritivo Detalhado / Serviços Realizados
          </label>
          <textarea
            v-model="docServico"
            rows="3"
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-opti"
          ></textarea>
        </div>

        <div>
          <label class="font-semibold text-slate-600 dark:text-slate-300 block mb-1">
            Parâmetros Técnicos / Observações
          </label>
          <textarea
            v-model="docObs"
            rows="2"
            placeholder="Ex: IP do roteador, senhas de Wi-Fi ou detalhes de peças..."
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-opti"
          ></textarea>
        </div>
      </div>

      <!-- Botões de Ação -->
      <div class="space-y-2 pt-2">
        <button
          @click="imprimirDocumento"
          class="w-full bg-emerald-600 hover:bg-emerald-700 text-white font-bold py-3 rounded-xl shadow transition flex items-center justify-center gap-2 text-sm"
        >
          <Printer class="w-4 h-4" />
          <span>Imprimir / Salvar como PDF (Ctrl+P)</span>
        </button>

        <button
          @click="baixarPDFNativo"
          class="w-full bg-opti hover:bg-opti-dark text-white font-bold py-2.5 rounded-xl transition flex items-center justify-center gap-2 text-xs shadow-xs"
        >
          <Download class="w-4 h-4" />
          <span>Baixar PDF Oficial (Maroto Go)</span>
        </button>

        <button
          @click="dispararWhatsApp"
          class="w-full bg-white dark:bg-slate-700 hover:bg-slate-50 dark:hover:bg-slate-600 text-emerald-700 dark:text-emerald-400 border border-emerald-300 dark:border-emerald-700 font-bold py-2.5 rounded-xl transition flex items-center justify-center gap-2 text-xs"
        >
          <MessageCircle class="w-4 h-4" />
          <span>Enviar Resumo pelo WhatsApp</span>
        </button>
      </div>
    </div>

    <!-- FOLHA A4 DE PRÉ-VISUALIZAÇÃO AO VIVO (DIREITA) -->
    <div class="lg:col-span-7">
      <div
        id="a4-preview"
        class="bg-white text-slate-800 p-6 sm:p-10 rounded-2xl shadow-md border border-slate-200 text-xs sm:text-sm min-h-[640px] flex flex-col justify-between"
      >
        <!-- CABEÇALHO DA EMPRESA -->
        <div>
          <div class="flex justify-between items-start border-b-2 border-opti pb-4 mb-5">
            <div class="flex items-center gap-3.5">
              <img
                v-if="configStore.config.logo_base64"
                :src="configStore.config.logo_base64"
                alt="Logo"
                class="w-14 h-14 object-contain rounded"
              />
              <div v-else class="bg-opti text-white font-black px-3 py-2 rounded-lg text-lg tracking-tight">
                OPTI
              </div>
              <div>
                <h2 class="text-xl font-black text-opti leading-tight">
                  {{ configStore.config.fantasia }}
                </h2>
                <p class="text-xs text-slate-600">CNPJ: {{ configStore.config.cnpj }}</p>
                <p class="text-xs text-slate-600">Fones: {{ configStore.config.fones }}</p>
                <p class="text-xs text-slate-600">Responsável: {{ configStore.config.responsavel }}</p>
              </div>
            </div>

            <div class="text-right">
              <span class="bg-opti text-white font-bold px-3 py-1.5 rounded-lg text-xs uppercase tracking-wider inline-block">
                {{ docTitulo }}
              </span>
              <p class="text-xs text-slate-500 mt-2 font-bold">
                Cód. #{{ selectedOSId || '---' }}
              </p>
            </div>
          </div>

          <!-- MODELO 1: RELATÓRIO DE SERVIÇO -->
          <div v-if="docType === 'relatorio'" class="space-y-4">
            <div class="bg-opti text-white text-xs font-bold px-3 py-1.5 rounded-sm uppercase tracking-wide">
              Dados do Atendimento
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs border-b border-slate-200 pb-4">
              <div><strong>Data de Emissão:</strong> {{ docData }}</div>
              <div><strong>Cidade:</strong> {{ configStore.config.cidade }}</div>
              <div><strong>Cliente:</strong> {{ docCliente || 'N/A' }}</div>
              <div><strong>CPF/CNPJ:</strong> {{ docCpfCnpj || 'Não informado' }}</div>
              <div class="col-span-2"><strong>Equipamento:</strong> {{ docEquipamento }}</div>
            </div>

            <div class="bg-opti text-white text-xs font-bold px-3 py-1.5 rounded-sm uppercase tracking-wide">
              Descrição dos Serviços Realizados
            </div>
            <div class="p-4 bg-slate-50 border border-slate-200 rounded-lg text-xs leading-relaxed space-y-2">
              <p class="whitespace-pre-line">{{ docServico }}</p>
              <div v-if="docObs" class="pt-2 border-t border-slate-300 font-semibold text-opti">
                Parâmetros / Observações Técnicas: {{ docObs }}
              </div>
            </div>

            <div v-if="configStore.config.termo_garantia" class="flex items-start gap-2 p-3 bg-blue-50/50 border border-blue-200/80 rounded-lg text-[11px] text-slate-600">
              <ShieldCheck class="w-4 h-4 text-opti shrink-0 mt-0.5" />
              <span>{{ configStore.config.termo_garantia }}</span>
            </div>
          </div>

          <!-- MODELO 2: ORÇAMENTO -->
          <div v-else-if="docType === 'orcamento'" class="space-y-4">
            <div class="grid grid-cols-2 gap-2 text-xs border-b border-slate-200 pb-4">
              <div><strong>Cliente:</strong> {{ docCliente || 'N/A' }}</div>
              <div><strong>Data de Emissão:</strong> {{ docData }} (Validade: 30 dias)</div>
              <div><strong>Equipamento:</strong> {{ docEquipamento }}</div>
              <div><strong>Cidade:</strong> {{ configStore.config.cidade }}</div>
            </div>

            <div class="bg-opti text-white text-xs font-bold px-3 py-1.5 rounded-sm uppercase tracking-wide">
              Descritivo do Orçamento
            </div>
            <div class="p-4 bg-slate-50 border border-slate-200 rounded-lg text-xs leading-relaxed">
              <p class="whitespace-pre-line">{{ docServico }}</p>
              <div class="text-right text-base font-black text-opti mt-4 pt-3 border-t border-slate-200">
                Valor Total: {{ formatBRL(docValor) }}
              </div>
            </div>

            <div v-if="docObs" class="p-3 bg-blue-50/60 border border-blue-200 rounded-lg text-xs text-slate-700">
              <strong>Observações adicionais:</strong> {{ docObs }}
            </div>
          </div>

          <!-- MODELO 3: RECIBO DE PAGAMENTO -->
          <div v-else class="py-6 space-y-6">
            <div class="flex justify-between items-center bg-slate-100 p-4 rounded-xl border border-slate-200">
              <div>
                <span class="text-xs uppercase font-bold text-slate-500 block">Data</span>
                <span class="font-bold text-slate-800 text-sm">{{ docData }}</span>
              </div>
              <div class="text-right">
                <span class="text-xs uppercase font-bold text-slate-500 block">Valor Recebido</span>
                <span class="text-xl font-black text-opti">{{ formatBRL(docValor) }}</span>
              </div>
            </div>

            <p class="text-justify leading-loose text-sm text-slate-700">
              Eu, <strong>{{ configStore.config.razao_social }}</strong>, inscrito sob o CNPJ <strong>{{ configStore.config.cnpj }}</strong>,
              recebi de <strong>{{ docCliente }}</strong>
              <span v-if="docCpfCnpj"> (CPF/CNPJ: {{ docCpfCnpj }})</span>, a importância de
              <strong>{{ formatBRL(docValor) }}</strong>, referente a:
              <em>{{ docServico }}</em>, e para clareza e fins legais firmo o presente recibo.
            </p>

            <p class="text-right text-xs text-slate-600">
              {{ configStore.config.cidade }}, {{ docData }}.
            </p>
          </div>
        </div>

        <!-- RODAPÉ & ASSINATURAS DO DOCUMENTO -->
        <div class="pt-8">
          <div class="grid grid-cols-2 gap-8 mb-6 text-xs text-center text-slate-600">
            <div class="border-t border-slate-400 pt-1.5">
              {{ configStore.config.responsavel }} ({{ configStore.config.fantasia }})
            </div>
            <div class="border-t border-slate-400 pt-1.5">
              {{ docCliente || 'Assinatura do Cliente' }}
            </div>
          </div>
          <p class="text-center text-[10px] font-bold text-opti border-t border-slate-200 pt-2 tracking-wide">
            Emitido via OptiManager • {{ configStore.config.fantasia }} • {{ configStore.config.cidade }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
