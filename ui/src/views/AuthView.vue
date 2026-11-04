<script setup lang="ts">
import { ref } from 'vue'
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
} from 'lucide-vue-next'

const authStore = useAuthStore()
const configStore = useConfigStore()
const toast = useToastStore()

const nome = ref('')
const email = ref('')
const senha = ref('')
const confirmaSenha = ref('')
const showPassword = ref(false)
const isSubmitting = ref(false)

const emit = defineEmits<{
  (e: 'authenticated'): void
}>()

async function handleSubmit() {
  if (authStore.needsSetup) {
    if (!nome.value.trim() || !email.value.trim() || !senha.value) {
      toast.show('Preencha todos os campos obrigatórios.', 'error')
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

    isSubmitting.value = true
    try {
      await authStore.setupAdmin(nome.value.trim(), email.value.trim(), senha.value)
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

    <!-- Cartão Central de Autenticação / Setup -->
    <div
      class="w-full max-w-md bg-white dark:bg-slate-800 rounded-2xl shadow-xl border border-slate-200 dark:border-slate-700 overflow-hidden"
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
          {{ configStore.config.fantasia || 'OptiManager' }}
        </h1>
        <p class="text-xs text-blue-100 mt-0.5">
          Sistema de Gestão de Serviços, OS e Documentos
        </p>
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
            <strong class="block font-bold text-sm mb-0.5">Bem-vindo ao Primeiro Acesso!</strong>
            Defina seu nome, e-mail e senha para criar a conta de <strong>Administrador</strong>. Depois você poderá gerenciar novos usuários nas Configurações.
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
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
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
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
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
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl pl-10 pr-10 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
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
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl pl-10 pr-3.5 py-2.5 focus:outline-none focus:ring-2 focus:ring-opti"
              />
            </div>
          </div>

          <!-- Botão de Ação -->
          <button
            type="submit"
            :disabled="isSubmitting"
            class="w-full bg-opti hover:bg-opti-dark text-white font-bold py-3 px-4 rounded-xl shadow-md transition flex items-center justify-center gap-2 text-sm mt-2"
          >
            <ShieldCheck v-if="authStore.needsSetup" class="w-5 h-5" />
            <LogIn v-else class="w-5 h-5" />
            <span>
              {{
                isSubmitting
                  ? 'Processando...'
                  : authStore.needsSetup
                  ? 'Criar Administrador e Acessar'
                  : 'Entrar no Sistema'
              }}
            </span>
          </button>
        </form>
      </div>

      <!-- Rodapé do Cartão -->
      <div
        class="px-6 py-3.5 bg-slate-50 dark:bg-slate-800/80 border-t border-slate-200 dark:border-slate-700 text-center text-[11px] text-slate-400"
      >
        OptiManager 3.0 • {{ configStore.config.cidade || 'Gestão Inteligente' }}
      </div>
    </div>

    <ToastContainer />
  </div>
</template>

