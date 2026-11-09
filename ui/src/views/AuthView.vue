<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api'
import { useAuthStore } from '../stores/auth'
import { useConfigStore } from '../stores/config'
import { useToastStore } from '../stores/toast'
import ToastContainer from '../components/ToastContainer.vue'
import {
  ShieldCheck,
  Lock,
  Mail,
  User,
  Eye,
  EyeOff,
  LogIn,
  Sun,
  Moon,
  Sparkles,
  Building2,
  Phone,
  MapPin,
  FileText,
  ArrowRight,
  ArrowLeft,
  Globe,
  Code,
  BookOpen,
  Database,
  CheckCircle2,
  RotateCcw,
  ExternalLink,
  PartyPopper,
  X,
  FileCheck,
  Upload,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const configStore = useConfigStore()
const toast = useToastStore()

// Controle de etapas do Assistente de Primeiro Uso (1 a 6)
// 1 = Boas-Vindas
// 2 = Licença e Termos
// 3 = Opções de Instalação (Nova ou Backup)
// 4 = Conta de Administrador
// 5 = Dados da Empresa
// 6 = Você está pronto para começar
const setupStep = ref<1 | 2 | 3 | 4 | 5 | 6>(1)

// Passo 1: Modal de documentação rápida
const showDocModal = ref(false)

// Passo 2: Licença e Termos
const concordoLicenca = ref(false)
const concordoTermos = ref(false)

// Passo 3: Opção de Instalação ('nova' ou 'backup')
const opcaoInstalacao = ref<'nova' | 'backup'>('nova')
const isCheckingBackup = ref(false)
const selectedBackupFile = ref<File | null>(null)
const isRestoringBackup = ref(false)

function handleBackupFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    selectedBackupFile.value = target.files[0]
  } else {
    selectedBackupFile.value = null
  }
}

async function handleImportarBackupInicial() {
  if (!selectedBackupFile.value) {
    toast.show('Selecione um arquivo de backup (.zip ou .db) antes de restaurar.', 'error')
    return
  }

  isRestoringBackup.value = true
  try {
    await api.restoreInitialBackupFromFile(selectedBackupFile.value)
    await Promise.all([configStore.loadConfig(), authStore.checkStatus()])
    selectedBackupFile.value = null
    toast.show('Backup importado e restaurado com sucesso! Entre com suas credenciais.')
  } catch (err: any) {
    toast.show('Erro ao restaurar backup: ' + (err.message || 'Arquivo inválido'), 'error')
  } finally {
    isRestoringBackup.value = false
  }
}

// Passo 4: Administrador (e formulário de login padrão)
const nome = ref('')
const email = ref('')
const senha = ref('')
const confirmaSenha = ref('')
const showPassword = ref(false)
const rememberEmail = ref(localStorage.getItem('optimanager_remember_email') === 'true')

onMounted(() => {
  if (rememberEmail.value) {
    const saved = localStorage.getItem('optimanager_saved_email')
    if (saved) {
      email.value = saved
    }
  }
})

// Passo 5: Dados da Empresa
const fantasia = ref('')
const razaoSocial = ref('')
const cnpj = ref('')
const fones = ref('')
const emailContato = ref('')
const endereco = ref('')
const cidade = ref('')

const isSubmitting = ref(false)

const emit = defineEmits<{
  (e: 'authenticated'): void
}>()

// --- NAVEGAÇÃO ENTRE ETAPAS ---

function irParaPasso2() {
  setupStep.value = 2
}

function irParaPasso3() {
  if (!concordoLicenca.value || !concordoTermos.value) {
    toast.show('Você precisa marcar as duas opções de concordância para avançar.', 'error')
    return
  }
  setupStep.value = 3
}

function irParaPasso4() {
  if (opcaoInstalacao.value === 'nova') {
    setupStep.value = 4
  }
}

async function verificarBackup() {
  isCheckingBackup.value = true
  try {
    await authStore.checkStatus()
    if (!authStore.needsSetup) {
      toast.show('Banco de dados identificado com sucesso! Redirecionando para o login...')
    } else {
      toast.show('Nenhum banco de dados configurado foi detectado ainda. Verifique se o arquivo .db foi colocado no diretório correto e o container reiniciado.', 'error')
    }
  } catch (err: any) {
    toast.show('Erro ao verificar status do banco: ' + (err.message || 'Falha de conexão'), 'error')
  } finally {
    isCheckingBackup.value = false
  }
}

function avancarParaEmpresa() {
  if (!nome.value.trim() || !email.value.trim() || !senha.value) {
    toast.show('Preencha seu nome, e-mail e senha para avançar.', 'error')
    return
  }
  if (senha.value.length < 4) {
    toast.show('A senha deve ter pelo menos 4 caracteres.', 'error')
    return
  }
  if (senha.value !== confirmaSenha.value) {
    toast.show('A confirmação de senha não confere.', 'error')
    return
  }

  // Sugestões amigáveis de valores iniciais se ainda em branco
  if (!razaoSocial.value.trim()) {
    razaoSocial.value = nome.value.trim()
  }
  if (!emailContato.value.trim()) {
    emailContato.value = email.value.trim()
  }

  setupStep.value = 5
}

async function handleSalvarEmpresa() {
  if (!fantasia.value.trim() || !razaoSocial.value.trim() || !cnpj.value.trim() || !cidade.value.trim()) {
    toast.show('Preencha Marca/Nome Fantasia, Razão Social, CNPJ e Cidade.', 'error')
    return
  }

  isSubmitting.value = true
  try {
    // Passa autoAuthenticate = false para segurar a transição até o usuário clicar em "Vamos lá!" no Passo 6
    await authStore.setupAdmin(
      {
        nome: nome.value.trim(),
        email: email.value.trim(),
        senha: senha.value,
        fantasia: fantasia.value.trim(),
        razao_social: razaoSocial.value.trim(),
        cnpj: cnpj.value.trim(),
        fones: fones.value.trim(),
        email_contato: emailContato.value.trim(),
        endereco: endereco.value.trim(),
        cidade: cidade.value.trim(),
      },
      false
    )
    setupStep.value = 6
  } catch {
    // erro exibido pelo toast da store
  } finally {
    isSubmitting.value = false
  }
}

function finalizarEIniciar() {
  authStore.completeSetup()
  emit('authenticated')
}

async function handleLogin() {
  if (!email.value.trim() || !senha.value) {
    toast.show('Informe seu e-mail e senha.', 'error')
    return
  }

  isSubmitting.value = true
  try {
    await authStore.login(email.value.trim(), senha.value)
    if (rememberEmail.value) {
      localStorage.setItem('optimanager_remember_email', 'true')
      localStorage.setItem('optimanager_saved_email', email.value.trim())
    } else {
      localStorage.removeItem('optimanager_remember_email')
      localStorage.removeItem('optimanager_saved_email')
    }
    emit('authenticated')
  } catch {
    // erro exibido pelo toast da store
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    class="min-h-screen w-full flex flex-col items-center py-8 sm:py-12 px-4 sm:px-6 bg-slate-100 dark:bg-slate-900 text-slate-800 dark:text-slate-100 relative"
  >
    <!-- Botão Alternar Tema no Topo Direito -->
    <div class="fixed top-4 right-4 z-10">
      <button
        @click="configStore.toggleTheme()"
        class="p-2.5 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 shadow-xs transition"
        :title="configStore.isDarkMode ? 'Mudar para tema claro' : 'Mudar para tema escuro'"
      >
        <Sun v-if="configStore.isDarkMode" class="w-5 h-5 text-amber-400" />
        <Moon v-else class="w-5 h-5 text-slate-600" />
      </button>
    </div>

    <!-- Cartão Central de Autenticação / Assistente de Setup -->
    <div
      class="w-full bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 overflow-hidden transition-all my-auto"
      :class="authStore.needsSetup ? 'max-w-2xl' : 'max-w-md'"
    >
      <!-- Cabeçalho Institucional -->
      <div class="bg-opti text-white p-6 text-center relative">
        <div class="flex justify-center mb-3">
          <img
            v-if="configStore.config.logo_base64"
            :src="configStore.config.logo_base64"
            alt="Logo"
            class="w-16 h-16 object-contain rounded-xl bg-white p-1 shadow"
          />
          <img
            v-else
            src="/icon.png"
            alt="OptiManager"
            class="w-16 h-16 object-contain rounded-xl bg-white/95 p-1 shadow"
          />
        </div>
        <h1 class="font-black text-xl tracking-tight">
          {{ authStore.needsSetup ? (fantasia || 'OptiManager') : (configStore.config.fantasia || 'OptiManager') }}
        </h1>

        <!-- Indicador de Passos no Assistente de Primeiro Uso (1 a 6) -->
        <div v-if="authStore.needsSetup" class="mt-4">
          <div class="flex items-center justify-between text-xs text-white/80 font-medium px-2 mb-2">
            <span>Passo {{ setupStep }} de 6</span>
            <span class="font-bold text-white">
              <template v-if="setupStep === 1">Boas-Vindas</template>
              <template v-else-if="setupStep === 2">Licença & Termos</template>
              <template v-else-if="setupStep === 3">Instalação</template>
              <template v-else-if="setupStep === 4">Administrador</template>
              <template v-else-if="setupStep === 5">Empresa</template>
              <template v-else-if="setupStep === 6">Conclusão</template>
            </span>
          </div>

          <!-- Barra de Progresso com 6 Segmentos -->
          <div class="grid grid-cols-6 gap-1.5 px-1">
            <div
              v-for="stepNum in 6"
              :key="stepNum"
              class="h-1.5 rounded-full transition-all duration-300"
              :class="stepNum <= setupStep ? 'bg-white shadow-xs' : 'bg-white/25'"
            ></div>
          </div>
        </div>
      </div>

      <!-- Corpo do Cartão -->
      <div class="p-6 sm:p-7 space-y-5">
        <!-- ==================================================== -->
        <!-- MODO SETUP: ETAPAS DE PRIMEIRO USO (1 A 6)           -->
        <!-- ==================================================== -->
        <div v-if="authStore.needsSetup" class="space-y-5">

          <!-- -------------------------------------------------- -->
          <!-- PASSO 1: BOAS-VINDAS                               -->
          <!-- -------------------------------------------------- -->
          <div v-if="setupStep === 1" class="space-y-6">
            <!-- Saudação Principal -->
            <div class="text-center space-y-3">
              <div class="inline-flex p-3 bg-blue-50 dark:bg-blue-950/40 text-opti dark:text-blue-400 rounded-2xl shadow-xs">
                <Sparkles class="w-8 h-8" />
              </div>
              <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-tight">
                Bem-vindo ao OptiManager!
              </h2>
              <div class="text-sm text-slate-600 dark:text-slate-300 max-w-lg mx-auto space-y-2 leading-relaxed">
                <p>
                  O <strong>OptiManager</strong> é uma aplicação web auto-hospedada que tem como objetivo facilitar a vida de você, pequeno empreendedor!
                </p>
                <p class="text-xs text-slate-500 dark:text-slate-400">
                  O Assistente de Boas-Vindas irá lhe auxiliar na configuração do aplicativo nos passos seguintes.
                </p>
              </div>
            </div>

            <!-- Botões Institucionais Solicitados: Site Oficial | GitHub | Docs | Licença -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5 pt-2">
              <a
                href="https://jvfigueiro.com/projects/optimanager"
                target="_blank"
                rel="noopener noreferrer"
                class="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-opti dark:hover:border-opti hover:bg-slate-50 dark:hover:bg-slate-750 transition text-center group"
              >
                <Globe class="w-5 h-5 text-slate-500 group-hover:text-opti transition mb-1.5" />
                <span class="text-xs font-bold text-slate-700 dark:text-slate-200">Site Oficial</span>
                <span class="text-[10px] text-slate-400 flex items-center gap-0.5 mt-0.5">
                  Web <ExternalLink class="w-2.5 h-2.5" />
                </span>
              </a>

              <a
                href="https://github.com/jvfigueiro/optimanager"
                target="_blank"
                rel="noopener noreferrer"
                class="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-opti dark:hover:border-opti hover:bg-slate-50 dark:hover:bg-slate-750 transition text-center group"
              >
                <Code class="w-5 h-5 text-slate-500 group-hover:text-opti transition mb-1.5" />
                <span class="text-xs font-bold text-slate-700 dark:text-slate-200">GitHub</span>
                <span class="text-[10px] text-slate-400 flex items-center gap-0.5 mt-0.5">
                  Código <ExternalLink class="w-2.5 h-2.5" />
                </span>
              </a>

              <button
                type="button"
                @click="showDocModal = true"
                class="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-opti dark:hover:border-opti hover:bg-slate-50 dark:hover:bg-slate-750 transition text-center group"
              >
                <BookOpen class="w-5 h-5 text-slate-500 group-hover:text-opti transition mb-1.5" />
                <span class="text-xs font-bold text-slate-700 dark:text-slate-200">Documentação</span>
                <span class="text-[10px] text-slate-400 mt-0.5">Guia Rápido</span>
              </button>

              <button
                type="button"
                @click="irParaPasso2"
                class="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 dark:border-slate-700 hover:border-opti dark:hover:border-opti hover:bg-slate-50 dark:hover:bg-slate-750 transition text-center group"
              >
                <FileCheck class="w-5 h-5 text-slate-500 group-hover:text-opti transition mb-1.5" />
                <span class="text-xs font-bold text-slate-700 dark:text-slate-200">Licença</span>
                <span class="text-[10px] text-slate-400 mt-0.5">GPLv3 / Termos</span>
              </button>
            </div>

            <!-- Botão Principal: Vamos lá! -->
            <div class="pt-4">
              <button
                type="button"
                @click="irParaPasso2"
                class="w-full bg-opti hover:bg-opti-dark text-white font-black py-3.5 px-6 rounded-xl shadow-md hover:shadow-lg transition flex items-center justify-center gap-2 text-base"
              >
                <span>Vamos lá!</span>
                <ArrowRight class="w-5 h-5" />
              </button>
            </div>
          </div>

          <!-- -------------------------------------------------- -->
          <!-- PASSO 2: LICENÇA, TERMOS E CONDIÇÕES DE USO        -->
          <!-- -------------------------------------------------- -->
          <div v-else-if="setupStep === 2" class="space-y-4">
            <div>
              <h2 class="text-lg font-bold text-slate-800 dark:text-white">
                Licença, Termos e Condições de Uso
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Revise os termos e diretrizes de utilização da aplicação para prosseguir.
              </p>
            </div>

            <!-- Espaço para Licença e Termos de Uso (Scrollable) -->
            <div
              class="max-h-60 overflow-y-auto p-4 bg-slate-50 dark:bg-slate-900/70 rounded-xl border border-slate-200 dark:border-slate-700 text-xs text-slate-700 dark:text-slate-300 space-y-3 font-sans leading-relaxed shadow-inner"
            >
              <div>
                <h3 class="font-bold text-slate-900 dark:text-white uppercase tracking-wider text-[11px] mb-1">
                  1. Licença de Uso do Software (GNU General Public License v3 - GPLv3)
                </h3>
                <p>
                  O OptiManager é um software livre distribuído sob os termos da <strong>GNU General Public License v3 (GPLv3)</strong>. Você tem a liberdade de executar, estudar, modificar e redistribuir esta aplicação, desde que quaisquer cópias, modificações ou trabalhos derivados continuem sob os mesmos termos da licença GNU GPLv3 com o respectivo código-fonte aberto e acessível.
                </p>
                <p class="italic text-[11px] text-slate-500 dark:text-slate-400 mt-1">
                  "ESTE PROGRAMA É DISTRIBUÍDO NA ESPERANÇA DE QUE SEJA ÚTIL, MAS SEM QUALQUER GARANTIA; NEM MESMO A GARANTIA IMPLÍCITA DE COMERCIABILIDADE OU ADEQUAÇÃO A UM DETERMINADO PROPÓSITO."
                </p>
              </div>

              <div>
                <h3 class="font-bold text-slate-900 dark:text-white uppercase tracking-wider text-[11px] mb-1">
                  2. Auto-Hospedagem e Responsabilidade
                </h3>
                <p>
                  O OptiManager foi projetado para execução em ambiente auto-hospedado (on-premises / Docker). O gerenciamento da infraestrutura de servidores, portas de rede, firewalls, certificados SSL/TLS e integridade do hardware é de responsabilidade integral do usuário ou administrador do sistema.
                </p>
              </div>

              <div>
                <h3 class="font-bold text-slate-900 dark:text-white uppercase tracking-wider text-[11px] mb-1">
                  3. Privacidade e Armazenamento Local
                </h3>
                <p>
                  Todos os dados da sua empresa, clientes, faturamentos, ordens de serviço e registros de auditoria residem exclusivamente no seu banco de dados local (SQLite). Não há envio, telemetria ou compartilhamento não autorizado de dados com servidores externos.
                </p>
              </div>

              <div>
                <h3 class="font-bold text-slate-900 dark:text-white uppercase tracking-wider text-[11px] mb-1">
                  4. Políticas de Backup e Continuidade
                </h3>
                <p>
                  Recomenda-se veementemente a configuração de rotinas automáticas de backup e a exportação periódica de cópias para volumes ou mídias externas à máquina hospedeira para prevenir perdas ocasionadas por danos físicos ao disco local.
                </p>
              </div>
            </div>

            <!-- Duas Checkboxes Obrigatórias -->
            <div class="space-y-2.5 pt-1">
              <label class="flex items-start gap-3 p-3 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/60 dark:bg-slate-900/40 cursor-pointer hover:bg-slate-100/70 dark:hover:bg-slate-900/70 transition">
                <input
                  type="checkbox"
                  v-model="concordoLicenca"
                  class="mt-0.5 w-4 h-4 rounded text-opti focus:ring-opti border-slate-300 dark:border-slate-600 dark:bg-slate-700 cursor-pointer shrink-0"
                />
                <span class="text-xs text-slate-700 dark:text-slate-300 leading-snug">
                  Li e concordo com os termos da <strong>Licença de Uso do Software (GNU GPLv3)</strong>.
                </span>
              </label>

              <label class="flex items-start gap-3 p-3 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/60 dark:bg-slate-900/40 cursor-pointer hover:bg-slate-100/70 dark:hover:bg-slate-900/70 transition">
                <input
                  type="checkbox"
                  v-model="concordoTermos"
                  class="mt-0.5 w-4 h-4 rounded text-opti focus:ring-opti border-slate-300 dark:border-slate-600 dark:bg-slate-700 cursor-pointer shrink-0"
                />
                <span class="text-xs text-slate-700 dark:text-slate-300 leading-snug">
                  Li e concordo com os <strong>Termos e Condições de Uso</strong> e as diretrizes de auto-hospedagem.
                </span>
              </label>
            </div>

            <!-- Botões Voltar / Avançar -->
            <div class="flex items-center gap-3 pt-2">
              <button
                type="button"
                @click="setupStep = 1"
                class="px-4 py-3 rounded-xl border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 font-bold text-xs transition flex items-center gap-1.5"
              >
                <ArrowLeft class="w-4 h-4" />
                <span>Voltar</span>
              </button>

              <button
                type="button"
                :disabled="!concordoLicenca || !concordoTermos"
                @click="irParaPasso3"
                class="flex-1 bg-opti hover:bg-opti-dark disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
              >
                <span>Avançar</span>
                <ArrowRight class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- -------------------------------------------------- -->
          <!-- PASSO 3: OPÇÕES DE INSTALAÇÃO                      -->
          <!-- -------------------------------------------------- -->
          <div v-else-if="setupStep === 3" class="space-y-4">
            <div>
              <h2 class="text-lg font-bold text-slate-800 dark:text-white">
                Opções de Instalação
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Selecione como deseja inicializar o seu ambiente OptiManager:
              </p>
            </div>

            <!-- Seleção de Modo: Nova Instalação ou Restaurar Backup -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
              <!-- Opção A: Iniciar pela primeira vez -->
              <div
                @click="opcaoInstalacao = 'nova'"
                class="p-4 rounded-xl border-2 cursor-pointer transition flex flex-col justify-between"
                :class="opcaoInstalacao === 'nova' ? 'border-opti bg-blue-50/50 dark:bg-blue-950/20 shadow-xs' : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <div class="space-y-2">
                  <div class="flex items-center justify-between">
                    <div class="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/40 text-opti dark:text-blue-400">
                      <Sparkles class="w-5 h-5" />
                    </div>
                    <span
                      class="w-4 h-4 rounded-full border flex items-center justify-center"
                      :class="opcaoInstalacao === 'nova' ? 'border-opti bg-opti text-white' : 'border-slate-300 dark:border-slate-600'"
                    >
                      <span v-if="opcaoInstalacao === 'nova'" class="w-1.5 h-1.5 rounded-full bg-white"></span>
                    </span>
                  </div>
                  <h3 class="font-bold text-sm text-slate-800 dark:text-white">
                    Iniciar pela primeira vez
                  </h3>
                  <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                    Configuração limpa: crie sua conta de Administrador e cadastre os dados da sua empresa.
                  </p>
                </div>
                <div class="mt-3">
                  <span class="inline-block text-[11px] font-bold text-opti dark:text-blue-400 bg-blue-50 dark:bg-blue-950/60 px-2 py-0.5 rounded">
                    Recomendado para novas instâncias
                  </span>
                </div>
              </div>

              <!-- Opção B: Restaurar a partir de um backup -->
              <div
                @click="opcaoInstalacao = 'backup'"
                class="p-4 rounded-xl border-2 cursor-pointer transition flex flex-col justify-between"
                :class="opcaoInstalacao === 'backup' ? 'border-opti bg-blue-50/50 dark:bg-blue-950/20 shadow-xs' : 'border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <div class="space-y-2">
                  <div class="flex items-center justify-between">
                    <div class="p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/40 text-emerald-600 dark:text-emerald-400">
                      <Database class="w-5 h-5" />
                    </div>
                    <span
                      class="w-4 h-4 rounded-full border flex items-center justify-center"
                      :class="opcaoInstalacao === 'backup' ? 'border-opti bg-opti text-white' : 'border-slate-300 dark:border-slate-600'"
                    >
                      <span v-if="opcaoInstalacao === 'backup'" class="w-1.5 h-1.5 rounded-full bg-white"></span>
                    </span>
                  </div>
                  <h3 class="font-bold text-sm text-slate-800 dark:text-white">
                    Restaurar um backup
                  </h3>
                  <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
                    Já possui um arquivo de backup compactado (.zip) ou banco de dados (.db) de uma instalação anterior e deseja restaurá-lo.
                  </p>
                </div>
                <div class="mt-3">
                  <span class="inline-block text-[11px] font-bold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/60 px-2 py-0.5 rounded">
                    Migração / Restauração
                  </span>
                </div>
              </div>
            </div>

            <!-- Painel de Importação Direta de Backup -->
            <div
              v-if="opcaoInstalacao === 'backup'"
              class="p-4 rounded-xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-700 space-y-4"
            >
              <div class="flex items-start gap-2.5">
                <Database class="w-5 h-5 text-opti shrink-0 mt-0.5" />
                <div class="text-xs text-slate-700 dark:text-slate-300 leading-relaxed">
                  <strong class="block text-slate-900 dark:text-white font-bold mb-0.5">
                    Importar arquivo de backup (.zip ou .db)
                  </strong>
                  <p class="text-slate-600 dark:text-slate-400">
                    Selecione um arquivo de backup gerado anteriormente pelo OptiManager. O sistema extrairá o pacote compactado, validará a integridade do banco e restaurará automaticamente seus usuários, clientes, ordens de serviço, logotipo e configurações.
                  </p>
                </div>
              </div>

              <!-- Seletor de Arquivo .zip / .db -->
              <label
                class="flex flex-col items-center justify-center p-4 border-2 border-dashed rounded-xl cursor-pointer transition text-center"
                :class="
                  selectedBackupFile
                    ? 'border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20'
                    : 'border-slate-300 dark:border-slate-600 hover:border-opti dark:hover:border-opti bg-white dark:bg-slate-800/70'
                "
              >
                <Upload class="w-6 h-6 mb-1.5" :class="selectedBackupFile ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-400'" />
                <span v-if="selectedBackupFile" class="text-xs font-bold text-emerald-700 dark:text-emerald-300">
                  {{ selectedBackupFile.name }} ({{ (selectedBackupFile.size / 1024).toFixed(1) }} KB)
                </span>
                <span v-else class="text-xs font-bold text-slate-700 dark:text-slate-200">
                  Clique para escolher o arquivo de backup (.zip ou .db)
                </span>
                <span class="text-[11px] text-slate-400 mt-0.5">
                  Aceita pacotes compactados (.zip com banco + logo) ou arquivos SQLite (.db)
                </span>
                <input
                  type="file"
                  accept=".zip,.db,.sqlite,.sqlite3,application/zip,application/vnd.sqlite3,application/octet-stream"
                  @change="handleBackupFileChange"
                  class="hidden"
                />
              </label>

              <div class="flex flex-col sm:flex-row gap-2">
                <button
                  type="button"
                  @click="handleImportarBackupInicial"
                  :disabled="!selectedBackupFile || isRestoringBackup"
                  class="flex-1 bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold py-2.5 px-4 rounded-xl text-xs flex items-center justify-center gap-2 transition shadow-xs"
                >
                  <Upload class="w-4 h-4" :class="{ 'animate-bounce': isRestoringBackup }" />
                  <span>{{ isRestoringBackup ? 'Validando e restaurando backup...' : 'Importar e Restaurar Backup' }}</span>
                </button>

                <button
                  type="button"
                  @click="verificarBackup"
                  :disabled="isCheckingBackup || isRestoringBackup"
                  class="px-3.5 py-2.5 rounded-xl border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 font-semibold text-xs flex items-center justify-center gap-1.5 transition"
                  title="Caso tenha restaurado manualmente via volume Docker"
                >
                  <RotateCcw class="w-3.5 h-3.5" :class="{ 'animate-spin': isCheckingBackup }" />
                  <span>{{ isCheckingBackup ? 'Verificando...' : 'Verificar Volume' }}</span>
                </button>
              </div>
            </div>

            <!-- Botões Voltar / Avançar -->
            <div class="flex items-center gap-3 pt-2">
              <button
                type="button"
                @click="setupStep = 2"
                class="px-4 py-3 rounded-xl border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 font-bold text-xs transition flex items-center gap-1.5"
              >
                <ArrowLeft class="w-4 h-4" />
                <span>Voltar</span>
              </button>

              <button
                v-if="opcaoInstalacao === 'nova'"
                type="button"
                @click="irParaPasso4"
                class="flex-1 bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
              >
                <span>Avançar: Conta de Administrador</span>
                <ArrowRight class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- -------------------------------------------------- -->
          <!-- PASSO 4: CONTA DE ADMINISTRADOR                    -->
          <!-- -------------------------------------------------- -->
          <form v-else-if="setupStep === 4" @submit.prevent="avancarParaEmpresa" class="space-y-4 text-sm">
            <div>
              <h2 class="text-lg font-bold text-slate-800 dark:text-white">
                Conta de Administrador
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Defina as credenciais de acesso do usuário principal da plataforma.
              </p>
            </div>

            <!-- Nome Completo -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Seu Nome Completo *
              </label>
              <div class="relative">
                <User class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  v-model="nome"
                  required
                  placeholder="Ex: Carlos Eduardo Silva"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- E-mail -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                E-mail de Login *
              </label>
              <div class="relative">
                <Mail class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  type="email"
                  v-model="email"
                  required
                  placeholder="admin@empresa.com.br"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- Senha -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Definir Senha *
              </label>
              <div class="relative">
                <Lock class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  :type="showPassword ? 'text' : 'password'"
                  v-model="senha"
                  required
                  placeholder="Mínimo 4 caracteres"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-10 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1"
                  tabindex="-1"
                >
                  <EyeOff v-if="showPassword" class="w-4 h-4" />
                  <Eye v-else class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Confirmar Senha -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Confirmar Senha *
              </label>
              <div class="relative">
                <ShieldCheck class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  :type="showPassword ? 'text' : 'password'"
                  v-model="confirmaSenha"
                  required
                  placeholder="Repita a senha definida"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- Botões Voltar / Avançar -->
            <div class="flex items-center gap-3 pt-2">
              <button
                type="button"
                @click="setupStep = 3"
                class="px-4 py-3 rounded-xl border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 font-bold text-xs transition flex items-center gap-1.5"
              >
                <ArrowLeft class="w-4 h-4" />
                <span>Voltar</span>
              </button>

              <button
                type="submit"
                class="flex-1 bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
              >
                <span>Avançar: Dados da Empresa</span>
                <ArrowRight class="w-4 h-4" />
              </button>
            </div>
          </form>

          <!-- -------------------------------------------------- -->
          <!-- PASSO 5: DADOS DA EMPRESA                          -->
          <!-- -------------------------------------------------- -->
          <form v-else-if="setupStep === 5" @submit.prevent="handleSalvarEmpresa" class="space-y-4 text-sm">
            <div>
              <h2 class="text-lg font-bold text-slate-800 dark:text-white">
                Dados da Empresa
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                Preencha os dados cadastrais da sua empresa. Você poderá alterá-los a qualquer momento em Configurações.
              </p>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
              <!-- Marca / Nome Fantasia -->
              <div class="sm:col-span-2">
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  Marca / Nome Fantasia *
                </label>
                <div class="relative">
                  <Building2 class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    v-model="fantasia"
                    required
                    placeholder="Ex: Tech Soluções & Serviços"
                    class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                  />
                </div>
              </div>

              <!-- Razão Social -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  Razão Social / Titular *
                </label>
                <div class="relative">
                  <FileText class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    v-model="razaoSocial"
                    required
                    placeholder="Ex: Silva & Santos Serviços Ltda"
                    class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                  />
                </div>
              </div>

              <!-- CNPJ -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  CNPJ (ou CPF) *
                </label>
                <input
                  type="text"
                  v-model="cnpj"
                  required
                  placeholder="00.000.000/0001-00"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 font-mono focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>

              <!-- Telefone / WhatsApp -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  Telefone / WhatsApp *
                </label>
                <div class="relative">
                  <Phone class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    v-model="fones"
                    required
                    placeholder="(11) 98765-4321"
                    class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                  />
                </div>
              </div>

              <!-- E-mail de Contato -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  E-mail Comercial / Contato
                </label>
                <div class="relative">
                  <Mail class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                  <input
                    type="email"
                    v-model="emailContato"
                    placeholder="contato@empresa.com.br"
                    class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                  />
                </div>
              </div>

              <!-- Endereço -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  Endereço Comercial
                </label>
                <div class="relative">
                  <MapPin class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                  <input
                    type="text"
                    v-model="endereco"
                    placeholder="Ex: Av. Paulista, 1000"
                    class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                  />
                </div>
              </div>

              <!-- Cidade - UF -->
              <div>
                <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                  Cidade - UF *
                </label>
                <input
                  type="text"
                  v-model="cidade"
                  required
                  placeholder="Ex: São Paulo - SP"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- Botões Voltar / Concluir -->
            <div class="flex items-center gap-2.5 pt-2">
              <button
                type="button"
                @click="setupStep = 4"
                class="px-4 py-3 rounded-xl border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 font-bold text-xs transition flex items-center gap-1.5"
              >
                <ArrowLeft class="w-4 h-4" />
                <span>Voltar</span>
              </button>

              <button
                type="submit"
                :disabled="isSubmitting"
                class="flex-1 bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm"
              >
                <ShieldCheck class="w-5 h-5" />
                <span>{{ isSubmitting ? 'Salvando dados...' : 'Salvar e Concluir' }}</span>
              </button>
            </div>
          </form>

          <!-- -------------------------------------------------- -->
          <!-- PASSO 6: VOCÊ ESTÁ PRONTO PARA COMEÇAR             -->
          <!-- -------------------------------------------------- -->
          <div v-else-if="setupStep === 6" class="text-center space-y-6 py-2">
            <div class="space-y-3">
              <div class="inline-flex p-4 bg-emerald-100 dark:bg-emerald-950/50 text-emerald-600 dark:text-emerald-400 rounded-full shadow-md ring-8 ring-emerald-50 dark:ring-emerald-900/20">
                <PartyPopper class="w-10 h-10" />
              </div>

              <h2 class="text-2xl font-black text-slate-800 dark:text-white tracking-tight">
                Você está pronto para começar!
              </h2>

              <p class="text-sm text-slate-600 dark:text-slate-300 max-w-md mx-auto leading-relaxed">
                Sua conta de Administrador e as configurações da sua empresa foram registradas com sucesso. Agora você tem o controle total do seu negócio!
              </p>
            </div>

            <!-- Cartão com Resumo do Ambiente Criado -->
            <div class="bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-700 rounded-2xl p-4 text-left max-w-md mx-auto space-y-2.5">
              <div class="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400 font-bold uppercase tracking-wider">
                <CheckCircle2 class="w-4 h-4 text-emerald-500" />
                <span>Ambiente Configurado</span>
              </div>
              <div class="grid grid-cols-2 gap-3 text-xs">
                <div>
                  <span class="block text-slate-400 text-[11px]">Empresa</span>
                  <span class="font-bold text-slate-800 dark:text-slate-200 truncate block">{{ fantasia || 'OptiManager' }}</span>
                </div>
                <div>
                  <span class="block text-slate-400 text-[11px]">Administrador</span>
                  <span class="font-bold text-slate-800 dark:text-slate-200 truncate block">{{ email }}</span>
                </div>
              </div>
            </div>

            <!-- Botão Principal: Vamos lá! -->
            <div class="pt-2">
              <button
                type="button"
                @click="finalizarEIniciar"
                class="w-full bg-emerald-600 hover:bg-emerald-700 text-white font-black py-3.5 px-6 rounded-xl shadow-md hover:shadow-lg transition flex items-center justify-center gap-2 text-base"
              >
                <span>Vamos lá!</span>
                <ArrowRight class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- MODO LOGIN REGULAR (QUANDO JÁ CONFIGURADO)           -->
        <!-- ==================================================== -->
        <template v-else>
          <div class="text-center">
            <h2 class="font-bold text-lg text-slate-800 dark:text-white">
              Acessar o Painel
            </h2>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Entre com suas credenciais para gerenciar ordens e clientes
            </p>
          </div>

          <form @submit.prevent="handleLogin" class="space-y-4 text-sm">
            <!-- E-mail -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                E-mail *
              </label>
              <div class="relative">
                <Mail class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  type="email"
                  v-model="email"
                  required
                  placeholder="admin@empresa.com.br"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- Senha -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Senha *
              </label>
              <div class="relative">
                <Lock class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  :type="showPassword ? 'text' : 'password'"
                  v-model="senha"
                  required
                  placeholder="••••••••"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-10 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1"
                  tabindex="-1"
                >
                  <EyeOff v-if="showPassword" class="w-4 h-4" />
                  <Eye v-else class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Opção Lembrar E-mail -->
            <div class="flex items-center justify-between text-xs pt-0.5">
              <label class="flex items-center gap-2 cursor-pointer select-none text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:white">
                <input
                  type="checkbox"
                  v-model="rememberEmail"
                  class="w-4 h-4 rounded text-opti focus:ring-opti border-slate-300 dark:border-slate-600 dark:bg-slate-700 cursor-pointer"
                />
                <span>Lembrar meu e-mail</span>
              </label>
            </div>

            <!-- Botão Entrar -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="w-full bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm mt-2"
            >
              <LogIn class="w-5 h-5" />
              <span>{{ isSubmitting ? 'Autenticando...' : 'Entrar' }}</span>
            </button>
          </form>
        </template>
      </div>

      <!-- Rodapé do Cartão -->
      <div
        class="px-6 py-3.5 bg-slate-50 dark:bg-slate-800/80 border-t border-slate-200 dark:border-slate-700 text-center text-[11px] text-slate-400"
      >
        OptiManager • Ver 3.0
      </div>
    </div>

    <!-- Modal Informativo de Documentação (Passo 1) -->
    <div
      v-if="showDocModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-xs overflow-y-auto"
    >
      <div class="bg-white dark:bg-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl border border-slate-200 dark:border-slate-700 space-y-4 my-auto">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <BookOpen class="w-5 h-5 text-opti" />
            <h3 class="font-bold text-base text-slate-800 dark:text-white">
              Guia & Documentação OptiManager
            </h3>
          </div>
          <button
            @click="showDocModal = false"
            class="p-1 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="text-xs text-slate-600 dark:text-slate-300 space-y-3 leading-relaxed max-h-72 overflow-y-auto pr-1">
          <p>
            O <strong>OptiManager</strong> foi arquitetado com foco em simplicidade, velocidade e auto-suficiência para pequenos empreendedores e prestadores de serviço:
          </p>
          <div class="p-3 bg-slate-50 dark:bg-slate-900/60 rounded-xl space-y-2 border border-slate-100 dark:border-slate-700">
            <div>
              <strong class="text-slate-800 dark:text-white font-bold block">Ordens de Serviço & Relatórios:</strong>
              Acompanhe ordens em andamento, concluídas ou pendentes, gerando relatórios de serviço para envio ao cliente.
            </div>
            <div>
              <strong class="text-slate-800 dark:text-white font-bold block">Base de Clientes com Autocompletar:</strong>
              Ao cadastrar clientes frequentes, o sistema preenche dados de contato automaticamente ao abrir novas ordens.
            </div>
            <div>
              <strong class="text-slate-800 dark:text-white font-bold block">Banco Local SQLite & Backups:</strong>
              Todos os dados ficam sob seu controle. Na aba Configurações, configure rotinas automáticas de backup para tranquilidade total.
            </div>
          </div>
        </div>

        <div class="pt-2 flex justify-end">
          <button
            type="button"
            @click="showDocModal = false"
            class="bg-opti hover:bg-opti-dark text-white font-bold text-xs py-2 px-4 rounded-xl transition"
          >
            Entendi, fechar
          </button>
        </div>
      </div>
    </div>

    <ToastContainer />
  </div>
</template>
