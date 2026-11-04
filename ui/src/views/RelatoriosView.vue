<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useOrdensStore } from '../stores/ordens'
import { useConfigStore } from '../stores/config'
import { formatBRL, formatDate } from '../utils/formatters'
import {
  Printer,
  MessageCircle,
  FileText,
  ShieldCheck,
  SlidersHorizontal,
  Check,
  X,
} from 'lucide-vue-next'

const ordensStore = useOrdensStore()
const configStore = useConfigStore()

type DocType = 'relatorio' | 'os' | 'orcamento' | 'recibo'
const docType = ref<DocType>('relatorio')

const selectedOSId = ref<number | null>(null)
const docCliente = ref('')
const docCpfCnpj = ref('')
const docEquipamento = ref('')
const docValor = ref(0)
const docServico = ref('')
const docObs = ref('')
const docData = ref('')
const docHora = ref('')
const docEntrega = ref('')
const docContato = ref('')

// Cláusulas e termos customizáveis
const docTermoGarantia = ref('')
const docTermoResponsabilidade = ref('')

// Modal de Personalização
const isPersonalizarModalOpen = ref(false)
const tempGarantia = ref('')
const tempResponsabilidade = ref('')
const tempObs = ref('')
const isSavingCustomTerms = ref(false)

onMounted(async () => {
  if (!configStore.config.termo_garantia) {
    await configStore.loadConfig()
  }

  docTermoGarantia.value =
    configStore.config.termo_garantia ||
    'Garantia legal de 90 dias para serviços executados e peças aplicadas.'
  docTermoResponsabilidade.value =
    configStore.config.termo_responsabilidade ||
    'Garantia legal de 90 dias conforme Art. 26 do CDC. Equipamentos não retirados em 90 dias estarão sujeitos a custas de guarda.'

  if (!ordensStore.ordens.length) {
    await ordensStore.loadOrdens()
  }

  // Se uma OS foi selecionada anteriormente a partir da tabela de serviços
  if (ordensStore.selectedOSForDoc) {
    if (ordensStore.selectedDocTypeForDoc) {
      docType.value = ordensStore.selectedDocTypeForDoc
      ordensStore.selectedDocTypeForDoc = null
    }
    selectedOSId.value = ordensStore.selectedOSForDoc.id
    carregarOSNoGerador(ordensStore.selectedOSForDoc.id)
    ordensStore.selectedOSForDoc = null
  } else if (ordensStore.ordens.length) {
    selectedOSId.value = ordensStore.ordens[0].id
    carregarOSNoGerador(ordensStore.ordens[0].id)
  }
})

watch(
  () => ordensStore.selectedOSForDoc,
  (os) => {
    if (os) {
      if (ordensStore.selectedDocTypeForDoc) {
        docType.value = ordensStore.selectedDocTypeForDoc
        ordensStore.selectedDocTypeForDoc = null
      }
      selectedOSId.value = os.id
      carregarOSNoGerador(os.id)
      ordensStore.selectedOSForDoc = null
    }
  }
)

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
  docObs.value = os.obs_tecnicas || configStore.config.obs_padrao || ''
  docData.value = formatDate(os.data_emissao) || new Date().toLocaleDateString('pt-BR')
  docHora.value = os.hora_emissao || ''
  docEntrega.value = os.data_entrega
    ? `${formatDate(os.data_entrega)}${os.hora_entrega ? ' às ' + os.hora_entrega : ''}`
    : ''
  docContato.value = os.cliente_contato || ''
}

function openPersonalizarModal() {
  tempGarantia.value = docTermoGarantia.value
  tempResponsabilidade.value = docTermoResponsabilidade.value
  tempObs.value = docObs.value
  isPersonalizarModalOpen.value = true
}

function aplicarTermos() {
  docTermoGarantia.value = tempGarantia.value
  docTermoResponsabilidade.value = tempResponsabilidade.value
  docObs.value = tempObs.value
  isPersonalizarModalOpen.value = false
}

async function salvarComoPadrao() {
  isSavingCustomTerms.value = true
  try {
    await configStore.saveConfig({
      termo_garantia: tempGarantia.value,
      termo_responsabilidade: tempResponsabilidade.value,
      obs_padrao: tempObs.value,
    })
    aplicarTermos()
  } catch (err) {
    console.error('Erro ao salvar termos padrão:', err)
  } finally {
    isSavingCustomTerms.value = false
  }
}

const docTitulo = computed(() => {
  switch (docType.value) {
    case 'relatorio':
      return 'Relatório de Serviço'
    case 'os':
      return 'Ordem de Serviço (OS)'
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
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs font-bold">
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
            Relatório
          </button>
          <button
            type="button"
            @click="docType = 'os'"
            class="py-2.5 px-2 rounded-xl transition border text-center"
            :class="
              docType === 'os'
                ? 'bg-opti text-white border-opti shadow-xs'
                : 'bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border-slate-300 dark:border-slate-600 hover:bg-slate-50'
            "
          >
            Ordem de Serviço
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
          type="button"
          @click="openPersonalizarModal"
          class="w-full bg-slate-100 hover:bg-slate-200 dark:bg-slate-700 dark:hover:bg-slate-650 text-slate-700 dark:text-slate-200 font-bold py-2.5 rounded-xl transition flex items-center justify-center gap-2 text-xs border border-slate-300 dark:border-slate-600 shadow-2xs"
        >
          <SlidersHorizontal class="w-4 h-4 text-opti dark:text-blue-400" />
          <span>Personalizar Termos & Cláusulas</span>
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
        class="bg-white text-slate-900 p-6 sm:p-8 rounded-2xl shadow-md border border-slate-200 text-xs sm:text-sm min-h-[640px]"
      >
        <!-- LAYOUT DUPLICADO NA MESMA FOLHA: ORDEM DE SERVIÇO (OS) -->
        <div v-if="docType === 'os'" class="space-y-4">
          <!-- 1ª VIA -->
          <div class="border border-slate-400 rounded-xl p-4 bg-white space-y-2.5">
            <div class="flex justify-between items-start border-b border-opti pb-2">
              <div class="flex items-center gap-2.5">
                <img
                  v-if="configStore.config.logo_base64"
                  :src="configStore.config.logo_base64"
                  alt="Logo"
                  class="w-9 h-9 object-contain rounded"
                />
                <div v-else class="bg-white text-opti border-2 border-opti font-black px-2 py-0.5 rounded text-xs tracking-tight">
                  OPTI
                </div>
                <div>
                  <h3 class="text-sm font-black text-opti leading-tight">
                    {{ configStore.config.fantasia }}
                  </h3>
                  <p class="text-[10px] text-slate-700 leading-tight">
                    CNPJ: {{ configStore.config.cnpj }} • Tel: {{ configStore.config.fones }}
                  </p>
                  <p class="text-[10px] text-slate-600 leading-tight">
                    <span v-if="configStore.config.endereco">{{ configStore.config.endereco }} - </span>{{ configStore.config.cidade }}
                  </p>
                </div>
              </div>

              <div class="text-right">
                <div class="inline-flex items-center gap-1.5">
                  <span class="bg-white text-opti border-2 border-opti font-black px-2.5 py-0.5 rounded text-[10px] uppercase tracking-wider">
                    ORDEM DE SERVIÇO
                  </span>
                </div>
                <p class="text-xs text-slate-800 mt-0.5 font-bold">
                  OS #{{ selectedOSId || '---' }} • <span class="text-slate-600 font-normal">{{ docData }}<span v-if="docHora"> às {{ docHora }}</span></span>
                </p>
              </div>
            </div>

            <!-- Dados Cliente & Equipamento -->
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-1.5 text-[11px] bg-white p-2 rounded-lg border border-slate-300">
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Cliente:</span>
                <span class="font-bold text-slate-900">{{ docCliente || 'N/A' }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">CPF / CNPJ:</span>
                <span class="font-mono text-slate-800">{{ docCpfCnpj || 'Não informado' }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Contato:</span>
                <span class="text-slate-800">{{ docContato || 'Não informado' }}</span>
              </div>
              <div class="sm:col-span-2">
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Equipamento:</span>
                <span class="font-semibold text-slate-900">{{ docEquipamento }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Valor Total:</span>
                <span class="font-black text-opti text-xs">{{ formatBRL(docValor) }}</span>
              </div>
            </div>

            <!-- Serviço Solicitado -->
            <div class="text-[11px] space-y-0.5">
              <span class="font-bold text-slate-800 uppercase text-[9px] tracking-wide block">
                Serviço / Descrição Técnica:
              </span>
              <div class="p-2 bg-white border border-slate-300 rounded text-slate-900 leading-snug whitespace-pre-line text-xs">
                {{ docServico }}
                <div v-if="docObs" class="pt-1 mt-1 border-t border-slate-300 text-opti font-semibold text-[11px]">
                  Obs: {{ docObs }}
                </div>
              </div>
            </div>

            <!-- Assinaturas -->
            <div class="pt-1">
              <div class="grid grid-cols-2 gap-6 text-[10px] text-center text-slate-700 mb-1">
                <div class="border-t border-slate-500 pt-1">
                  {{ configStore.config.responsavel }} (Responsável)
                </div>
                <div class="border-t border-slate-500 pt-1">
                  {{ docCliente || 'Assinatura do Cliente' }}
                </div>
              </div>
              <p class="text-center text-[9px] text-slate-600 leading-tight">
                {{ docTermoResponsabilidade }}
              </p>
            </div>
          </div>

          <!-- LINHA DE CORTE APENAS TRACEJADA (ECONOMIA DE SUPRIMENTOS) -->
          <div class="py-1 select-none">
            <div class="border-t-2 border-dashed border-slate-400 w-full"></div>
          </div>

          <!-- 2ª VIA -->
          <div class="border border-slate-400 rounded-xl p-4 bg-white space-y-2.5">
            <div class="flex justify-between items-start border-b border-opti pb-2">
              <div class="flex items-center gap-2.5">
                <img
                  v-if="configStore.config.logo_base64"
                  :src="configStore.config.logo_base64"
                  alt="Logo"
                  class="w-9 h-9 object-contain rounded"
                />
                <div v-else class="bg-white text-opti border-2 border-opti font-black px-2 py-0.5 rounded text-xs tracking-tight">
                  OPTI
                </div>
                <div>
                  <h3 class="text-sm font-black text-opti leading-tight">
                    {{ configStore.config.fantasia }}
                  </h3>
                  <p class="text-[10px] text-slate-700 leading-tight">
                    CNPJ: {{ configStore.config.cnpj }} • Tel: {{ configStore.config.fones }}
                  </p>
                  <p class="text-[10px] text-slate-600 leading-tight">
                    <span v-if="configStore.config.endereco">{{ configStore.config.endereco }} - </span>{{ configStore.config.cidade }}
                  </p>
                </div>
              </div>

              <div class="text-right">
                <div class="inline-flex items-center gap-1.5">
                  <span class="bg-white text-opti border-2 border-opti font-black px-2.5 py-0.5 rounded text-[10px] uppercase tracking-wider">
                    ORDEM DE SERVIÇO
                  </span>
                </div>
                <p class="text-xs text-slate-800 mt-0.5 font-bold">
                  OS #{{ selectedOSId || '---' }} • <span class="text-slate-600 font-normal">{{ docData }}<span v-if="docHora"> às {{ docHora }}</span></span>
                </p>
              </div>
            </div>

            <!-- Dados Cliente & Equipamento -->
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-1.5 text-[11px] bg-white p-2 rounded-lg border border-slate-300">
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Cliente:</span>
                <span class="font-bold text-slate-900">{{ docCliente || 'N/A' }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">CPF / CNPJ:</span>
                <span class="font-mono text-slate-800">{{ docCpfCnpj || 'Não informado' }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Contato:</span>
                <span class="text-slate-800">{{ docContato || 'Não informado' }}</span>
              </div>
              <div class="sm:col-span-2">
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Equipamento:</span>
                <span class="font-semibold text-slate-900">{{ docEquipamento }}</span>
              </div>
              <div>
                <span class="text-slate-600 font-semibold block text-[9px] uppercase">Valor Total:</span>
                <span class="font-black text-opti text-xs">{{ formatBRL(docValor) }}</span>
              </div>
            </div>

            <!-- Serviço Solicitado -->
            <div class="text-[11px] space-y-0.5">
              <span class="font-bold text-slate-800 uppercase text-[9px] tracking-wide block">
                Serviço / Descrição Técnica:
              </span>
              <div class="p-2 bg-white border border-slate-300 rounded text-slate-900 leading-snug whitespace-pre-line text-xs">
                {{ docServico }}
                <div v-if="docObs" class="pt-1 mt-1 border-t border-slate-300 text-opti font-semibold text-[11px]">
                  Obs: {{ docObs }}
                </div>
              </div>
            </div>

            <!-- Assinaturas -->
            <div class="pt-1">
              <div class="grid grid-cols-2 gap-6 text-[10px] text-center text-slate-700 mb-1">
                <div class="border-t border-slate-500 pt-1">
                  {{ configStore.config.responsavel }} (Responsável)
                </div>
                <div class="border-t border-slate-500 pt-1">
                  {{ docCliente || 'Assinatura do Cliente' }}
                </div>
              </div>
              <p class="text-center text-[9px] text-slate-600 leading-tight">
                {{ docTermoResponsabilidade }}
              </p>
            </div>
          </div>
        </div>

        <!-- LAYOUT PADRÃO EM PÁGINA INTEIRA: RELATÓRIO, ORÇAMENTO, RECIBO -->
        <div v-else class="flex flex-col justify-between min-h-[580px]">
          <div>
            <!-- CABEÇALHO DA EMPRESA -->
            <div class="flex justify-between items-start border-b-2 border-opti pb-4 mb-5">
              <div class="flex items-center gap-3.5">
                <img
                  v-if="configStore.config.logo_base64"
                  :src="configStore.config.logo_base64"
                  alt="Logo"
                  class="w-14 h-14 object-contain rounded"
                />
                <div v-else class="bg-white text-opti border-2 border-opti font-black px-3 py-2 rounded-lg text-lg tracking-tight">
                  OPTI
                </div>
                <div>
                  <h2 class="text-xl font-black text-opti leading-tight">
                    {{ configStore.config.fantasia }}
                  </h2>
                  <p class="text-xs text-slate-700">CNPJ: {{ configStore.config.cnpj }}</p>
                  <p class="text-xs text-slate-700">
                    Fones: {{ configStore.config.fones }}
                    <span v-if="configStore.config.email_contato"> • {{ configStore.config.email_contato }}</span>
                  </p>
                  <p class="text-xs text-slate-700">
                    <span v-if="configStore.config.endereco">{{ configStore.config.endereco }} - </span>{{ configStore.config.cidade }}
                  </p>
                </div>
              </div>

              <div class="text-right">
                <span class="bg-white text-opti border-2 border-opti font-black px-3 py-1.5 rounded-lg text-xs uppercase tracking-wider inline-block">
                  {{ docTitulo }}
                </span>
                <p class="text-xs text-slate-700 mt-2 font-bold">
                  Cód. #{{ selectedOSId || '---' }}
                </p>
              </div>
            </div>

            <!-- MODELO: RELATÓRIO DE SERVIÇO -->
            <div v-if="docType === 'relatorio'" class="space-y-4">
              <div class="bg-white text-opti border-l-4 border-opti border-y border-r border-slate-300 text-xs font-black px-3 py-1.5 rounded-r-sm uppercase tracking-wide">
                Dados do Atendimento
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs border-b border-slate-300 pb-4">
                <div><strong>Data de Entrada:</strong> {{ docData }}<span v-if="docHora"> às {{ docHora }}</span></div>
                <div v-if="docEntrega"><strong>Data de Entrega:</strong> {{ docEntrega }}</div>
                <div v-else><strong>Cidade:</strong> {{ configStore.config.cidade }}</div>
                <div><strong>Cliente:</strong> {{ docCliente || 'N/A' }}</div>
                <div><strong>CPF/CNPJ:</strong> {{ docCpfCnpj || 'Não informado' }}</div>
                <div class="col-span-2"><strong>Equipamento:</strong> {{ docEquipamento }}</div>
              </div>

              <div class="bg-white text-opti border-l-4 border-opti border-y border-r border-slate-300 text-xs font-black px-3 py-1.5 rounded-r-sm uppercase tracking-wide">
                Descrição dos Serviços Realizados
              </div>
              <div class="p-4 bg-white border border-slate-300 rounded-lg text-xs leading-relaxed space-y-2">
                <p class="whitespace-pre-line text-slate-900">{{ docServico }}</p>
                <div v-if="docObs" class="pt-2 border-t border-slate-300 font-semibold text-opti">
                  Parâmetros / Observações Técnicas: {{ docObs }}
                </div>
              </div>

              <div v-if="docTermoGarantia" class="flex items-start gap-2 p-3 bg-white border border-opti/60 rounded-lg text-[11px] text-slate-800">
                <ShieldCheck class="w-4 h-4 text-opti shrink-0 mt-0.5" />
                <span>{{ docTermoGarantia }}</span>
              </div>
            </div>

            <!-- MODELO: ORÇAMENTO -->
            <div v-else-if="docType === 'orcamento'" class="space-y-4">
              <div class="grid grid-cols-2 gap-2 text-xs border-b border-slate-300 pb-4">
                <div><strong>Cliente:</strong> {{ docCliente || 'N/A' }}</div>
                <div><strong>Data de Emissão:</strong> {{ docData }} (Validade: 30 dias)</div>
                <div><strong>Equipamento:</strong> {{ docEquipamento }}</div>
                <div><strong>Cidade:</strong> {{ configStore.config.cidade }}</div>
              </div>

              <div class="bg-white text-opti border-l-4 border-opti border-y border-r border-slate-300 text-xs font-black px-3 py-1.5 rounded-r-sm uppercase tracking-wide">
                Descritivo do Orçamento
              </div>
              <div class="p-4 bg-white border border-slate-300 rounded-lg text-xs leading-relaxed">
                <p class="whitespace-pre-line text-slate-900">{{ docServico }}</p>
                <div class="text-right text-base font-black text-opti mt-4 pt-3 border-t border-slate-300">
                  Valor Total: {{ formatBRL(docValor) }}
                </div>
              </div>

              <div v-if="docObs" class="p-3 bg-white border border-opti/60 rounded-lg text-xs text-slate-800">
                <strong>Observações adicionais:</strong> {{ docObs }}
              </div>
            </div>

            <!-- MODELO: RECIBO DE PAGAMENTO -->
            <div v-else class="py-6 space-y-6">
              <div class="flex justify-between items-center bg-white p-4 rounded-xl border-2 border-opti">
                <div>
                  <span class="text-xs uppercase font-bold text-slate-600 block">Data</span>
                  <span class="font-bold text-slate-900 text-sm">{{ docData }}</span>
                </div>
                <div class="text-right">
                  <span class="text-xs uppercase font-bold text-slate-600 block">Valor Recebido</span>
                  <span class="text-xl font-black text-opti">{{ formatBRL(docValor) }}</span>
                </div>
              </div>

              <p class="text-justify leading-loose text-sm text-slate-900">
                Eu, <strong>{{ configStore.config.razao_social }}</strong>, inscrito sob o CNPJ <strong>{{ configStore.config.cnpj }}</strong>,
                recebi de <strong>{{ docCliente }}</strong>
                <span v-if="docCpfCnpj"> (CPF/CNPJ: {{ docCpfCnpj }})</span>, a importância de
                <strong>{{ formatBRL(docValor) }}</strong>, referente a:
                <em>{{ docServico }}</em>, e para clareza e fins legais firmo o presente recibo.
              </p>

              <p class="text-right text-xs text-slate-700">
                {{ configStore.config.cidade }}, {{ docData }}.
              </p>
            </div>
          </div>

          <!-- RODAPÉ & ASSINATURAS DO DOCUMENTO (PÁGINA INTEIRA) -->
          <div class="pt-8">
            <div class="grid grid-cols-2 gap-8 mb-6 text-xs text-center text-slate-700">
              <div class="border-t border-slate-500 pt-1.5">
                {{ configStore.config.responsavel }} ({{ configStore.config.fantasia }})
              </div>
              <div class="border-t border-slate-500 pt-1.5">
                {{ docCliente || 'Assinatura do Cliente' }}
              </div>
            </div>
            <p class="text-center text-[10px] font-bold text-opti border-t border-slate-300 pt-2 tracking-wide">
              Emitido via OptiManager • {{ configStore.config.fantasia }} • {{ configStore.config.cidade }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal Personalizar Termos, Cláusulas & Guarda -->
    <div
      v-if="isPersonalizarModalOpen"
      class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-center justify-center p-4 z-50 no-print"
      @click.self="isPersonalizarModalOpen = false"
    >
      <div
        class="bg-white dark:bg-slate-800 rounded-2xl max-w-lg w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden flex flex-col max-h-[90vh]"
      >
        <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center shrink-0">
          <div class="flex items-center gap-2">
            <SlidersHorizontal class="w-5 h-5" />
            <h3 class="font-bold text-sm sm:text-base">Personalizar Termos & Cláusulas</h3>
          </div>
          <button
            @click="isPersonalizarModalOpen = false"
            class="text-blue-200 hover:text-white p-1 rounded-lg transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form @submit.prevent="aplicarTermos" class="p-5 space-y-4 text-xs sm:text-sm overflow-y-auto flex-1">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Informação adicional de garantia (opcional)
            </label>
            <textarea
              v-model="tempGarantia"
              rows="3"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 text-xs focus:ring-2 focus:ring-opti leading-relaxed"
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Observações adicionais 1 (ex. Termo de responsabilidade)
            </label>
            <textarea
              v-model="tempResponsabilidade"
              rows="3"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 text-xs focus:ring-2 focus:ring-opti leading-relaxed"
            ></textarea>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Informações Adicionais
            </label>
            <textarea
              v-model="tempObs"
              rows="2"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 text-xs focus:ring-2 focus:ring-opti"
            ></textarea>
          </div>

          <div class="pt-3 border-t border-slate-200 dark:border-slate-700 flex flex-wrap items-center justify-between gap-2">
            <button
              type="button"
              @click="salvarComoPadrao"
              :disabled="isSavingCustomTerms"
              class="px-3.5 py-2 border border-blue-200 dark:border-blue-900 bg-blue-50 dark:bg-blue-950/40 text-opti dark:text-sky-300 rounded-xl text-xs font-bold hover:bg-blue-100 transition flex items-center gap-1.5"
            >
              <Check class="w-4 h-4" />
              <span>{{ isSavingCustomTerms ? 'Salvando...' : 'Salvar como Padrão do Sistema' }}</span>
            </button>

            <div class="flex items-center gap-2">
              <button
                type="button"
                @click="isPersonalizarModalOpen = false"
                class="px-3.5 py-2 border border-slate-300 dark:border-slate-600 rounded-xl text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
              >
                Cancelar
              </button>
              <button
                type="submit"
                class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition shadow-xs"
              >
                Aplicar ao Documento
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
