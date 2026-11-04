<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
} from 'lucide-vue-next'

const authStore = useAuthStore()
const configStore = useConfigStore()
const toast = useToastStore()

// Controle de etapas do Assistente de Primeiro Acesso (1 = Admin, 2 = Empresa)
const setupStep = ref<1 | 2>(1)

// Dados do Administrador (Etapa 1 / Login)
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

// Dados da Empresa (Etapa 2 do Assistente de Primeiro Uso)
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

  // Sugere valores iniciais úteis caso ainda estejam em branco
  if (!razaoSocial.value.trim()) {
    razaoSocial.value = nome.value.trim()
  }
  if (!emailContato.value.trim()) {
    emailContato.value = email.value.trim()
  }

  setupStep.value = 2
}

async function handleSubmit() {
  if (authStore.needsSetup) {
    if (setupStep.value === 1) {
      avancarParaEmpresa()
      return
    }

    if (!fantasia.value.trim() || !razaoSocial.value.trim() || !cnpj.value.trim() || !cidade.value.trim()) {
      toast.show('Preencha Marca/Nome Fantasia, Razão Social, CNPJ e Cidade.', 'error')
      return
    }

    isSubmitting.value = true
    try {
      await authStore.setupAdmin({
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
      })
      emit('authenticated')
    } catch {
      // erro já exibido pelo toast da store
    } finally {
      isSubmitting.value = false
    }
  } else {
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
      // erro já exibido pelo toast da store
    } finally {
      isSubmitting.value = false
    }
  }
}
</script>

<template>
  <div
    class="min-h-screen w-full flex flex-col justify-center items-center p-4 bg-slate-100 dark:bg-slate-900 text-slate-800 dark:text-slate-100 relative overflow-y-auto"
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
      class="w-full bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 overflow-hidden transition-all"
      :class="authStore.needsSetup ? 'max-w-lg' : 'max-w-md'"
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
          <div
            v-else
            class="bg-white text-opti font-black px-4 py-2 rounded-xl text-2xl tracking-tight shadow-md"
          >
            OPTI
          </div>
        </div>
        <h1 class="font-black text-xl tracking-tight">
          {{ authStore.needsSetup ? (fantasia || 'OptiManager') : (configStore.config.fantasia || 'OptiManager') }}
        </h1>

        <!-- Indicador de Passos no Primeiro Acesso -->
        <div v-if="authStore.needsSetup" class="flex items-center justify-center gap-2 mt-4">
          <div
            class="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold transition"
            :class="setupStep === 1 ? 'bg-white text-opti shadow-xs' : 'bg-white/20 text-white'"
          >
            <span>1. Administrador</span>
          </div>
          <div class="w-4 h-0.5 bg-white/40"></div>
          <div
            class="flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold transition"
            :class="setupStep === 2 ? 'bg-white text-opti shadow-xs' : 'bg-white/20 text-white'"
          >
            <span>2. Dados da Empresa</span>
          </div>
        </div>
      </div>

      <!-- Corpo do Formulário -->
      <div class="p-6 sm:p-7 space-y-5">
        <!-- Banner de Primeiro Acesso -->
        <div
          v-if="authStore.needsSetup"
          class="bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/80 rounded-xl p-3.5 flex items-start gap-3"
        >
          <Sparkles class="w-5 h-5 text-emerald-600 dark:text-emerald-400 shrink-0 mt-0.5" />
          <div class="text-xs leading-relaxed text-emerald-900 dark:text-emerald-200">
            <template v-if="setupStep === 1">
              <strong class="block font-bold text-sm mb-0.5">Assistente de Configuração Inicial (1/2)</strong>
              Defina seu nome, e-mail e senha para criar a conta de <strong>Administrador</strong> do sistema.
            </template>
            <template v-else>
              <strong class="block font-bold text-sm mb-0.5">Dados da sua Empresa (2/2)</strong>
              Informe os dados da empresa que você irá gerenciar. Você poderá alterá-los a qualquer momento na tela de <strong>Configurações</strong>.
            </template>
          </div>
        </div>

        <div v-else class="text-center">
          <h2 class="font-bold text-lg text-slate-800 dark:text-white">
            Acessar o Painel
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Entre com suas credenciais para gerenciar ordens e clientes
          </p>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4 text-sm">
          <!-- ==================================================== -->
          <!-- MODO LOGIN OU PASSO 1 DO SETUP (CONTA ADMINISTRADOR) -->
          <!-- ==================================================== -->
          <template v-if="!authStore.needsSetup || setupStep === 1">
            <!-- Nome Completo (Apenas no Primeiro Acesso) -->
            <div v-if="authStore.needsSetup">
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Seu Nome Completo *
              </label>
              <div class="relative">
                <User class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
                <input
                  type="text"
                  v-model="nome"
                  required
                  placeholder="Ex: João Figueiró"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- E-mail -->
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                E-mail de Acesso *
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
                {{ authStore.needsSetup ? 'Definir Senha *' : 'Sua Senha *' }}
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

            <!-- Opção Lembrar E-mail no Login -->
            <div v-if="!authStore.needsSetup" class="flex items-center justify-between text-xs pt-0.5">
              <label class="flex items-center gap-2 cursor-pointer select-none text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white">
                <input
                  type="checkbox"
                  v-model="rememberEmail"
                  class="w-4 h-4 rounded text-opti focus:ring-opti border-slate-300 dark:border-slate-600 dark:bg-slate-700 cursor-pointer"
                />
                <span>Lembrar meu e-mail</span>
              </label>
            </div>

            <!-- Confirmar Senha (Apenas no Primeiro Acesso) -->
            <div v-if="authStore.needsSetup">
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

            <!-- Botão de Ação (Login ou Avançar para Passo 2) -->
            <button
              type="submit"
              :disabled="isSubmitting"
              class="w-full bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm mt-2"
            >
              <template v-if="authStore.needsSetup">
                <span>Próximo: Dados da Empresa</span>
                <ArrowRight class="w-4 h-4" />
              </template>
              <template v-else>
                <LogIn class="w-5 h-5" />
                <span>{{ isSubmitting ? 'Autenticando...' : 'Entrar no Sistema' }}</span>
              </template>
            </button>
          </template>

          <!-- ==================================================== -->
          <!-- PASSO 2 DO SETUP: DADOS DA EMPRESA GERENCIADA        -->
          <!-- ==================================================== -->
          <template v-else>
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
                    placeholder="Ex: Opti Tecnologia"
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
                    placeholder="Ex: João Vítor Figueiró MEI"
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
                    placeholder="(55) 99999-9999"
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
                    placeholder="Rua, Número, Bairro"
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
                  placeholder="Ex: Santa Maria - RS"
                  class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
                />
              </div>
            </div>

            <!-- Botões Voltar / Concluir -->
            <div class="flex items-center gap-2.5 pt-2">
              <button
                type="button"
                @click="setupStep = 1"
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
                <span>{{ isSubmitting ? 'Salvando...' : 'Concluir Configuração e Acessar' }}</span>
              </button>
            </div>
          </template>
        </form>
      </div>

      <!-- Rodapé do Cartão -->
      <div
        class="px-6 py-3.5 bg-slate-50 dark:bg-slate-800/80 border-t border-slate-200 dark:border-slate-700 text-center text-[11px] text-slate-400"
      >
        OptiManager • Ver 3.0
      </div>
    </div>

    <ToastContainer />
  </div>
</template>
