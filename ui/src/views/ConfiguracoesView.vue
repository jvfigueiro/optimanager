<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfigStore } from '../stores/config'
import { useAuthStore } from '../stores/auth'
import type { Usuario, AutocompletePresets } from '../types'
import {
  Settings,
  Upload,
  Palette,
  CheckCircle2,
  Users,
  UserPlus,
  Edit3,
  Trash2,
  ShieldCheck,
  User,
  X,
  Sparkles,
  Info,
  Github,
  Globe,
  BookOpen,
} from 'lucide-vue-next'

const configStore = useConfigStore()
const authStore = useAuthStore()

const form = ref({ ...configStore.config })
const isSaving = ref(false)

const customPresets = ref<AutocompletePresets>({
  marcas_titulo: configStore.presets.marcas_titulo || 'Marcas / Fabricantes',
  marcas_visivel: configStore.presets.marcas_visivel !== false,
  marcas: [...configStore.presets.marcas],

  modelos_titulo: configStore.presets.modelos_titulo || 'Modelos / Tipos / Categorias de Item',
  modelos_visivel: configStore.presets.modelos_visivel !== false,
  modelos: [...configStore.presets.modelos],

  servicos_titulo: configStore.presets.servicos_titulo || 'Serviços Frequentes',
  servicos_visivel: configStore.presets.servicos_visivel !== false,
  servicos: [...configStore.presets.servicos],

  estados_servico: [...(configStore.presets.estados_servico || ['Avaliação/Orçamento', 'Em Serviço', 'Aguardando', 'Concluído'])],
  formas_pagamento: [...(configStore.presets.formas_pagamento || ['Pendente', 'Dinheiro', 'Cartão', 'Boleto', 'Pix'])],
})

const newMarca = ref('')
const newModelo = ref('')
const newServico = ref('')
const newEstado = ref('')
const newFormaPagamento = ref('')

function addMarca() {
  const val = newMarca.value.trim()
  if (val && !customPresets.value.marcas.includes(val)) {
    customPresets.value.marcas.push(val)
    newMarca.value = ''
  }
}

function removeMarca(idx: number) {
  customPresets.value.marcas.splice(idx, 1)
}

function addModelo() {
  const val = newModelo.value.trim()
  if (val && !customPresets.value.modelos.includes(val)) {
    customPresets.value.modelos.push(val)
    newModelo.value = ''
  }
}

function removeModelo(idx: number) {
  customPresets.value.modelos.splice(idx, 1)
}

function addServico() {
  const val = newServico.value.trim()
  if (val && !customPresets.value.servicos.includes(val)) {
    customPresets.value.servicos.push(val)
    newServico.value = ''
  }
}

function removeServico(idx: number) {
  customPresets.value.servicos.splice(idx, 1)
}

function addEstado() {
  const val = newEstado.value.trim()
  if (val) {
    if (!customPresets.value.estados_servico) customPresets.value.estados_servico = []
    if (!customPresets.value.estados_servico.includes(val)) {
      customPresets.value.estados_servico.push(val)
      newEstado.value = ''
    }
  }
}

function removeEstado(idx: number) {
  customPresets.value.estados_servico?.splice(idx, 1)
}

function addFormaPagamento() {
  const val = newFormaPagamento.value.trim()
  if (val) {
    if (!customPresets.value.formas_pagamento) customPresets.value.formas_pagamento = []
    if (!customPresets.value.formas_pagamento.includes(val)) {
      customPresets.value.formas_pagamento.push(val)
      newFormaPagamento.value = ''
    }
  }
}

function removeFormaPagamento(idx: number) {
  customPresets.value.formas_pagamento?.splice(idx, 1)
}

// Estado do modal de criar/editar usuário
const showUserModal = ref(false)
const isSavingUser = ref(false)
const editingUserId = ref<number | null>(null)
const userForm = ref({
  nome: '',
  email: '',
  senha: '',
  role: 'operador' as 'admin' | 'operador',
})

const paletasPredefinidas = [
  { nome: 'Opti Padrão', primaria: '#0B5694', secundaria: '#083f6e' },
  { nome: 'Esmeralda Tech', primaria: '#059669', secundaria: '#064e3b' },
  { nome: 'Índigo Moderno', primaria: '#4f46e5', secundaria: '#3730a3' },
  { nome: 'Grafite Elegante', primaria: '#334155', secundaria: '#1e293b' },
  { nome: 'Cobre Industrial', primaria: '#c2410c', secundaria: '#7c2d12' },
]

onMounted(async () => {
  await Promise.all([configStore.loadConfig(), authStore.loadUsuarios()])
  form.value = { ...configStore.config }
  customPresets.value = {
    marcas_titulo: configStore.presets.marcas_titulo || 'Marcas / Fabricantes',
    marcas_visivel: configStore.presets.marcas_visivel !== false,
    marcas: [...configStore.presets.marcas],

    modelos_titulo: configStore.presets.modelos_titulo || 'Modelos / Tipos / Categorias de Item',
    modelos_visivel: configStore.presets.modelos_visivel !== false,
    modelos: [...configStore.presets.modelos],

    servicos_titulo: configStore.presets.servicos_titulo || 'Serviços Frequentes',
    servicos_visivel: configStore.presets.servicos_visivel !== false,
    servicos: [...configStore.presets.servicos],

    estados_servico: [...(configStore.presets.estados_servico || ['Avaliação/Orçamento', 'Em Serviço', 'Aguardando', 'Concluído'])],
    formas_pagamento: [...(configStore.presets.formas_pagamento || ['Pendente', 'Dinheiro', 'Cartão', 'Boleto', 'Pix'])],
  }
})

function aplicarPaleta(primaria: string, secundaria: string) {
  form.value.cor_primaria = primaria
  form.value.cor_secundaria = secundaria
  configStore.applyColorTheme(primaria, secundaria)
}

function handleColorChange() {
  configStore.applyColorTheme(form.value.cor_primaria, form.value.cor_secundaria)
}

function handleLogoUpload(event: Event) {
  const target = event.target as HTMLInputElement
  if (!target.files || !target.files[0]) return

  const file = target.files[0]
  if (file.size > 1.5 * 1024 * 1024) {
    alert('A imagem deve ter no máximo 1.5 MB')
    return
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    form.value.logo_base64 = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

function removerLogo() {
  form.value.logo_base64 = ''
}

async function handleSalvar() {
  isSaving.value = true
  try {
    form.value.custom_presets = JSON.stringify(customPresets.value)
    await configStore.saveConfig(form.value)
  } finally {
    isSaving.value = false
  }
}

function openCreateUserModal() {
  editingUserId.value = null
  userForm.value = {
    nome: '',
    email: '',
    senha: '',
    role: 'operador',
  }
  showUserModal.value = true
}

function openEditUserModal(u: Usuario) {
  editingUserId.value = u.id
  userForm.value = {
    nome: u.nome,
    email: u.email,
    senha: '',
    role: u.role,
  }
  showUserModal.value = true
}

async function handleSaveUser() {
  isSavingUser.value = true
  try {
    if (editingUserId.value) {
      await authStore.editUsuario(editingUserId.value, {
        nome: userForm.value.nome,
        email: userForm.value.email,
        role: userForm.value.role,
        senha: userForm.value.senha || undefined,
      })
    } else {
      await authStore.addUsuario({
        nome: userForm.value.nome,
        email: userForm.value.email,
        senha: userForm.value.senha,
        role: userForm.value.role,
      })
    }
    showUserModal.value = false
  } finally {
    isSavingUser.value = false
  }
}

async function handleDeleteUser(u: Usuario) {
  if (authStore.user?.id === u.id) {
    alert('Você não pode remover sua própria conta de usuário.')
    return
  }
  if (authStore.usuarios.length <= 1) {
    alert('Não é possível excluir o único usuário do sistema.')
    return
  }
  if (!confirm(`Deseja realmente remover o usuário "${u.nome}" (${u.email})?`)) {
    return
  }
  await authStore.removeUsuario(u.id)
}
</script>

<template>
  <div class="max-w-4xl space-y-6">
    <!-- Card 1: Gestão de Usuários e Acessos (Apenas Administradores) -->
    <div
      v-if="authStore.user?.role === 'admin'"
      class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-6"
    >
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
        <div>
          <div class="flex items-center gap-2.5">
            <Users class="w-5 h-5 text-opti" />
            <h3 class="font-bold text-lg text-slate-800 dark:text-white">
              Usuários & Controle de Acesso
            </h3>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Gerencie os operadores e administradores com permissão de acesso ao OptiManager.
          </p>
        </div>

        <button
          type="button"
          @click="openCreateUserModal"
          class="bg-opti hover:bg-opti-dark text-white font-bold px-4 py-2 rounded-xl transition shadow-xs flex items-center gap-2 text-xs self-start sm:self-auto"
        >
          <UserPlus class="w-4 h-4" />
          <span>Adicionar Usuário</span>
        </button>
      </div>

      <!-- Lista de Usuários -->
      <div class="divide-y divide-slate-200 dark:divide-slate-700 border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
        <div
          v-for="u in authStore.usuarios"
          :key="u.id"
          class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-750 transition"
        >
          <div class="flex items-center gap-3 min-w-0">
            <div
              class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
              :class="
                u.role === 'admin'
                  ? 'bg-opti/10 text-opti dark:bg-opti/20 dark:text-sky-300'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300'
              "
            >
              <ShieldCheck v-if="u.role === 'admin'" class="w-5 h-5" />
              <User v-else class="w-5 h-5" />
            </div>

            <div class="min-w-0">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-bold text-sm text-slate-800 dark:text-white truncate">
                  {{ u.nome }}
                </span>
                <span
                  class="text-[10px] font-bold uppercase px-2 py-0.5 rounded-full"
                  :class="
                    u.role === 'admin'
                      ? 'bg-opti/15 text-opti dark:bg-opti/30 dark:text-sky-300'
                      : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
                  "
                >
                  {{ u.role === 'admin' ? 'Administrador' : 'Operador' }}
                </span>
                <span
                  v-if="authStore.user?.id === u.id"
                  class="text-[10px] font-semibold bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300 px-2 py-0.5 rounded-full"
                >
                  Você
                </span>
              </div>
              <p class="text-xs text-slate-500 dark:text-slate-400 truncate">
                {{ u.email }}
              </p>
            </div>
          </div>

          <div class="flex items-center gap-2 self-end sm:self-auto">
            <button
              type="button"
              @click="openEditUserModal(u)"
              class="px-3 py-1.5 rounded-lg text-xs font-semibold border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition flex items-center gap-1.5"
              title="Editar usuário"
            >
              <Edit3 class="w-3.5 h-3.5" />
              <span>Editar</span>
            </button>
            <button
              v-if="authStore.user?.id !== u.id"
              type="button"
              @click="handleDeleteUser(u)"
              :disabled="authStore.usuarios.length <= 1"
              class="px-3 py-1.5 rounded-lg text-xs font-semibold border border-red-200 dark:border-red-900/50 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/40 disabled:opacity-40 disabled:cursor-not-allowed transition flex items-center gap-1.5"
              title="Remover usuário"
            >
              <Trash2 class="w-3.5 h-3.5" />
              <span>Remover</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Card 2: Configurações da Empresa & Personalização -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-6">
      <div class="flex items-center gap-2.5 mb-2">
        <Settings class="w-5 h-5 text-opti" />
        <h3 class="font-bold text-lg text-slate-800 dark:text-white">
          Configurações da Empresa & Personalização
        </h3>
      </div>
      <p class="text-xs text-slate-500 dark:text-slate-400 mb-6">
        Estes dados são exibidos no rodapé do sistema, cabeçalhos dos documentos A4 e mensagens de cobrança via WhatsApp.
      </p>

      <form @submit.prevent="handleSalvar" class="space-y-6">
        <!-- 1. Identidade Visual: Logotipo & Cores -->
        <div class="p-5 bg-slate-50 dark:bg-slate-750 rounded-2xl border border-slate-200 dark:border-slate-700 space-y-4">
          <div class="flex items-center gap-2">
            <Palette class="w-4 h-4 text-opti" />
            <h4 class="font-bold text-sm text-slate-800 dark:text-white">Identidade Visual & Cores</h4>
          </div>

          <!-- Logotipo -->
          <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4 pt-1">
            <div
              class="w-20 h-20 rounded-xl border-2 border-dashed border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 flex items-center justify-center overflow-hidden shrink-0"
            >
              <img
                v-if="form.logo_base64"
                :src="form.logo_base64"
                alt="Logo Preview"
                class="w-full h-full object-contain"
              />
              <img
                v-else
                src="/icon.png"
                alt="Ícone Padrão"
                class="w-full h-full object-contain p-2 opacity-70"
                title="Ícone padrão da aplicação"
              />
            </div>

            <div class="space-y-2">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                Logotipo para Relatórios e Impressão
              </label>
              <div class="flex items-center gap-2">
                <label
                  class="cursor-pointer bg-white dark:bg-slate-700 hover:bg-slate-100 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-600 px-3.5 py-1.5 rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-xs"
                >
                  <Upload class="w-3.5 h-3.5" />
                  <span>Escolher Imagem (PNG/JPG)</span>
                  <input
                    type="file"
                    accept="image/*"
                    @change="handleLogoUpload"
                    class="hidden"
                  />
                </label>
                <button
                  v-if="form.logo_base64"
                  type="button"
                  @click="removerLogo"
                  class="text-xs text-red-500 hover:underline px-2 py-1"
                >
                  Remover
                </button>
              </div>
              <span class="text-[11px] text-slate-400 block">
                Tamanho recomendado: 200x200px (Fundo transparente ou branco).
              </span>
            </div>
          </div>

          <!-- Paletas e Cores Primárias -->
          <div class="pt-3 border-t border-slate-200 dark:border-slate-700 space-y-3">
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
              Paletas de Cores Rápidas
            </label>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="p in paletasPredefinidas"
                :key="p.nome"
                type="button"
                @click="aplicarPaleta(p.primaria, p.secundaria)"
                class="flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-medium border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 hover:border-slate-400 dark:hover:border-slate-500 transition"
              >
                <span class="w-3.5 h-3.5 rounded-full" :style="{ backgroundColor: p.primaria }"></span>
                <span class="text-slate-700 dark:text-slate-200">{{ p.nome }}</span>
              </button>
            </div>

            <!-- Seletores Hex Manuais -->
            <div class="grid grid-cols-2 gap-3 pt-2">
              <div>
                <label class="block text-[11px] font-bold text-slate-500 dark:text-slate-400 mb-1">Cor Primária (Hex)</label>
                <div class="flex items-center gap-2">
                  <input
                    type="color"
                    v-model="form.cor_primaria"
                    @input="handleColorChange"
                    class="w-8 h-8 rounded cursor-pointer border-0 p-0 bg-transparent"
                  />
                  <input
                    type="text"
                    v-model="form.cor_primaria"
                    @input="handleColorChange"
                    class="text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-white rounded-lg px-2.5 py-1.5 font-mono w-28 uppercase"
                  />
                </div>
              </div>

              <div>
                <label class="block text-[11px] font-bold text-slate-500 dark:text-slate-400 mb-1">Cor Secundária (Escura)</label>
                <div class="flex items-center gap-2">
                  <input
                    type="color"
                    v-model="form.cor_secundaria"
                    @input="handleColorChange"
                    class="w-8 h-8 rounded cursor-pointer border-0 p-0 bg-transparent"
                  />
                  <input
                    type="text"
                    v-model="form.cor_secundaria"
                    @input="handleColorChange"
                    class="text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-800 dark:text-white rounded-lg px-2.5 py-1.5 font-mono w-28 uppercase"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Dados Cadastrais da Empresa -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs sm:text-sm">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Nome Fantasia / Marca
            </label>
            <input
              type="text"
              v-model="form.fantasia"
              required
              placeholder="Ex: Tech Soluções & Serviços"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Razão Social / Titular
            </label>
            <input
              type="text"
              v-model="form.razao_social"
              required
              placeholder="Ex: Soluções em Tecnologia & Serviços Ltda"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              CNPJ (ou CPF)
            </label>
            <input
              type="text"
              v-model="form.cnpj"
              required
              placeholder="00.000.000/0001-00"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 font-mono focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Telefones / WhatsApp
            </label>
            <input
              type="text"
              v-model="form.fones"
              required
              placeholder="(11) 98765-4321"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              E-mail Comercial / Contato
            </label>
            <input
              type="email"
              v-model="form.email_contato"
              placeholder="contato@empresa.com.br"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Endereço Comercial
            </label>
            <input
              type="text"
              v-model="form.endereco"
              placeholder="Ex: Av. Paulista, 1000 - Bela Vista"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div class="sm:col-span-2">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Cidade - UF
            </label>
            <input
              type="text"
              v-model="form.cidade"
              required
              placeholder="Ex: São Paulo - SP"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div class="sm:col-span-2">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Informações de Garantia e Observações Legais
            </label>
            <textarea
              v-model="form.termo_garantia"
              rows="2"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none text-xs"
            ></textarea>
          </div>

          <div class="sm:col-span-2">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Informações Adicionais
            </label>
            <textarea
              v-model="form.obs_padrao"
              rows="2"
              placeholder="Ex: Aparelho recebido com marcas normais de uso. Backup de dados sob responsabilidade do cliente..."
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none text-xs"
            ></textarea>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-200 dark:border-slate-700 flex justify-end">
          <button
            v-if="authStore.user?.role === 'admin'"
            type="submit"
            :disabled="isSaving"
            class="bg-opti hover:bg-opti-dark text-white font-bold px-6 py-2.5 rounded-xl transition shadow flex items-center gap-2 text-sm"
          >
            <CheckCircle2 class="w-4 h-4" />
            <span>{{ isSaving ? 'Salvando...' : 'Salvar Alterações' }}</span>
          </button>
          <span v-else class="text-xs text-slate-400 italic">
            Alterações restritas a administradores.
          </span>
        </div>
      </form>
    </div>

    <!-- Card 4: Segmento de Atuação & Autocompletar Customizável -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-6 space-y-6">
      <div class="border-b border-slate-200 dark:border-slate-700 pb-4">
        <div class="flex items-center gap-2.5">
          <Sparkles class="w-5 h-5 text-opti" />
          <h3 class="font-bold text-lg text-slate-800 dark:text-white">
            Segmento de Atuação & Autocompletar
          </h3>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Personalize as categorias, renomeie os títulos, oculte ou exiba seções, defina estados de serviço e formas de pagamento conforme a necessidade do seu negócio.
        </p>
      </div>

      <!-- 1. Marcas / Fabricantes -->
      <div class="space-y-3 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-750/50">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div class="flex-1 max-w-sm">
            <label class="block text-[11px] font-bold uppercase text-slate-400 mb-1">Título da Categoria</label>
            <input
              type="text"
              v-model="customPresets.marcas_titulo"
              placeholder="Ex: Marcas / Fabricantes"
              class="w-full text-xs font-bold border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-opti"
            />
          </div>
          <label class="flex items-center gap-2 text-xs font-semibold text-slate-700 dark:text-slate-300 cursor-pointer self-start sm:self-end pb-1.5">
            <input
              type="checkbox"
              v-model="customPresets.marcas_visivel"
              class="rounded text-opti focus:ring-opti w-4 h-4"
            />
            <span>Exibir esta categoria</span>
          </label>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            v-model="newMarca"
            @keyup.enter.prevent="addMarca"
            placeholder="Adicionar novo item... (Enter)"
            class="flex-1 text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 focus:ring-2 focus:ring-opti"
          />
          <button
            type="button"
            @click="addMarca"
            class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition"
          >
            + Adicionar
          </button>
        </div>

        <div class="flex flex-wrap gap-1.5 pt-1">
          <span
            v-for="(m, idx) in customPresets.marcas"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 shadow-2xs"
          >
            <span>{{ m }}</span>
            <button
              type="button"
              @click="removeMarca(idx)"
              class="text-slate-400 hover:text-red-500 font-bold ml-0.5"
            >
              ×
            </button>
          </span>
          <span v-if="!customPresets.marcas.length" class="text-xs text-slate-400 italic">
            Nenhum item cadastrado nesta categoria.
          </span>
        </div>
      </div>

      <!-- 2. Modelos / Tipos / Categorias de Item -->
      <div class="space-y-3 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-750/50">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div class="flex-1 max-w-sm">
            <label class="block text-[11px] font-bold uppercase text-slate-400 mb-1">Título da Categoria</label>
            <input
              type="text"
              v-model="customPresets.modelos_titulo"
              placeholder="Ex: Modelos / Tipos / Categorias de Item"
              class="w-full text-xs font-bold border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-opti"
            />
          </div>
          <label class="flex items-center gap-2 text-xs font-semibold text-slate-700 dark:text-slate-300 cursor-pointer self-start sm:self-end pb-1.5">
            <input
              type="checkbox"
              v-model="customPresets.modelos_visivel"
              class="rounded text-opti focus:ring-opti w-4 h-4"
            />
            <span>Exibir esta categoria</span>
          </label>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            v-model="newModelo"
            @keyup.enter.prevent="addModelo"
            placeholder="Adicionar novo modelo ou categoria... (Enter)"
            class="flex-1 text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 focus:ring-2 focus:ring-opti"
          />
          <button
            type="button"
            @click="addModelo"
            class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition"
          >
            + Adicionar
          </button>
        </div>

        <div class="flex flex-wrap gap-1.5 pt-1">
          <span
            v-for="(mod, idx) in customPresets.modelos"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 shadow-2xs"
          >
            <span>{{ mod }}</span>
            <button
              type="button"
              @click="removeModelo(idx)"
              class="text-slate-400 hover:text-red-500 font-bold ml-0.5"
            >
              ×
            </button>
          </span>
          <span v-if="!customPresets.modelos.length" class="text-xs text-slate-400 italic">
            Nenhum item cadastrado nesta categoria.
          </span>
        </div>
      </div>

      <!-- 3. Serviços Frequentes -->
      <div class="space-y-3 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-750/50">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div class="flex-1 max-w-sm">
            <label class="block text-[11px] font-bold uppercase text-slate-400 mb-1">Título da Categoria</label>
            <input
              type="text"
              v-model="customPresets.servicos_titulo"
              placeholder="Ex: Serviços Frequentes"
              class="w-full text-xs font-bold border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-lg px-3 py-1.5 focus:ring-2 focus:ring-opti"
            />
          </div>
          <label class="flex items-center gap-2 text-xs font-semibold text-slate-700 dark:text-slate-300 cursor-pointer self-start sm:self-end pb-1.5">
            <input
              type="checkbox"
              v-model="customPresets.servicos_visivel"
              class="rounded text-opti focus:ring-opti w-4 h-4"
            />
            <span>Exibir esta categoria</span>
          </label>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            v-model="newServico"
            @keyup.enter.prevent="addServico"
            placeholder="Adicionar novo serviço frequente... (Enter)"
            class="flex-1 text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 focus:ring-2 focus:ring-opti"
          />
          <button
            type="button"
            @click="addServico"
            class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition"
          >
            + Adicionar
          </button>
        </div>

        <div class="flex flex-wrap gap-1.5 pt-1">
          <span
            v-for="(s, idx) in customPresets.servicos"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 shadow-2xs"
          >
            <span>{{ s }}</span>
            <button
              type="button"
              @click="removeServico(idx)"
              class="text-slate-400 hover:text-red-500 font-bold ml-0.5"
            >
              ×
            </button>
          </span>
          <span v-if="!customPresets.servicos.length" class="text-xs text-slate-400 italic">
            Nenhum serviço cadastrado nesta categoria.
          </span>
        </div>
      </div>

      <!-- 4. Estados de Serviço (Status da OS) -->
      <div class="space-y-3 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-750/50">
        <div>
          <label class="block text-xs font-bold uppercase text-slate-700 dark:text-slate-300">
            Estados de Serviço (Status da OS)
          </label>
          <p class="text-[11px] text-slate-500 dark:text-slate-400">
            Situações disponíveis para as Ordens de Serviço (ex: Avaliação/Orçamento, Em Serviço, Aguardando, Concluído).
          </p>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            v-model="newEstado"
            @keyup.enter.prevent="addEstado"
            placeholder="Ex: Em Espera, Aguardando Peça, Entregue... (Enter)"
            class="flex-1 text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 focus:ring-2 focus:ring-opti"
          />
          <button
            type="button"
            @click="addEstado"
            class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition"
          >
            + Adicionar
          </button>
        </div>

        <div class="flex flex-wrap gap-1.5 pt-1">
          <span
            v-for="(st, idx) in (customPresets.estados_servico || [])"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 shadow-2xs"
          >
            <span>{{ st }}</span>
            <button
              type="button"
              @click="removeEstado(idx)"
              class="text-slate-400 hover:text-red-500 font-bold ml-0.5"
            >
              ×
            </button>
          </span>
        </div>
      </div>

      <!-- 5. Formas de Pagamento -->
      <div class="space-y-3 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-750/50">
        <div>
          <label class="block text-xs font-bold uppercase text-slate-700 dark:text-slate-300">
            Formas de Pagamento
          </label>
          <p class="text-[11px] text-slate-500 dark:text-slate-400">
            Opções de pagamento permitidas (ex: Pendente, Dinheiro, Cartão, Boleto, Pix).
          </p>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            v-model="newFormaPagamento"
            @keyup.enter.prevent="addFormaPagamento"
            placeholder="Ex: Carnê, Transferência Bancária, Vale... (Enter)"
            class="flex-1 text-xs border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3 py-2 focus:ring-2 focus:ring-opti"
          />
          <button
            type="button"
            @click="addFormaPagamento"
            class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition"
          >
            + Adicionar
          </button>
        </div>

        <div class="flex flex-wrap gap-1.5 pt-1">
          <span
            v-for="(fp, idx) in (customPresets.formas_pagamento || [])"
            :key="idx"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-600 shadow-2xs"
          >
            <span>{{ fp }}</span>
            <button
              type="button"
              @click="removeFormaPagamento(idx)"
              class="text-slate-400 hover:text-red-500 font-bold ml-0.5"
            >
              ×
            </button>
          </span>
        </div>
      </div>

      <div class="pt-4 border-t border-slate-200 dark:border-slate-700 flex justify-end">
        <button
          v-if="authStore.user?.role === 'admin'"
          type="button"
          @click="handleSalvar"
          :disabled="isSaving"
          class="bg-opti hover:bg-opti-dark text-white font-bold px-6 py-2.5 rounded-xl transition shadow flex items-center gap-2 text-sm"
        >
          <CheckCircle2 class="w-4 h-4" />
          <span>{{ isSaving ? 'Salvando...' : 'Salvar Autocompletar' }}</span>
        </button>
        <span v-else class="text-xs text-slate-400 italic">
          Alterações restritas a administradores.
        </span>
      </div>
    </div>

    <!-- Card 5: Sobre o Sistema & Dados da Empresa -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-6 space-y-5">
      <div class="border-b border-slate-200 dark:border-slate-700 pb-4">
        <div class="flex items-center gap-2.5">
          <Info class="w-5 h-5 text-opti" />
          <div>
            <h3 class="font-bold text-lg text-slate-800 dark:text-white leading-tight">
              Sobre
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Informações da plataforma e dados cadastrais registrados
            </p>
          </div>
        </div>
      </div>

      <!-- Bloco Institucional da Aplicação -->
      <div class="p-5 rounded-2xl bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-750 dark:to-slate-700/60 border border-slate-200 dark:border-slate-700 space-y-3">
        <div>
          <h4 class="text-xl font-black text-opti dark:text-blue-400 tracking-tight">
            OptiManager
          </h4>
          <p class="text-xs text-slate-600 dark:text-slate-300 font-semibold mt-0.5">
            Versão 3.0.0
          </p>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
            Copyright &copy; {{ new Date().getFullYear() }} João Figueiró
          </p>
        </div>

        <!-- Botões de Links Institucionais -->
        <div class="flex flex-wrap items-center gap-2 pt-2 border-t border-slate-200 dark:border-slate-600/80">
          <a
            href="https://github.com/jvfigueiro/optimanager"
            target="_blank"
            rel="noopener noreferrer"
            class="px-3.5 py-1.5 rounded-xl border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-650 text-xs font-semibold transition flex items-center gap-1.5 shadow-2xs"
            title="Repositório no GitHub"
          >
            <Github class="w-3.5 h-3.5" />
            <span>Repositório do GitHub</span>
          </a>
          <a
            href="https://jvfigueiro.com/projects/optimanager"
            target="_blank"
            rel="noopener noreferrer"
            class="px-3.5 py-1.5 rounded-xl border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-650 text-xs font-semibold transition flex items-center gap-1.5 shadow-2xs"
            title="Site Oficial do OptiManager"
          >
            <Globe class="w-3.5 h-3.5 text-opti dark:text-blue-400" />
            <span>Site oficial</span>
          </a>
          <a
            href="https://github.com/jvfigueiro/optimanager#readme"
            target="_blank"
            rel="noopener noreferrer"
            class="px-3.5 py-1.5 rounded-xl border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-650 text-xs font-semibold transition flex items-center gap-1.5 shadow-2xs"
            title="Documentação do Projeto"
          >
            <BookOpen class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" />
            <span>Documentação</span>
          </a>
          <a
            href="https://github.com/jvfigueiro/optimanager/blob/master/LICENSE"
            target="_blank"
            rel="noopener noreferrer"
            class="px-3.5 py-1.5 rounded-xl border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-650 text-xs font-semibold transition flex items-center gap-1.5 shadow-2xs"
            title="Licença GNU GPLv3"
          >
            <ShieldCheck class="w-3.5 h-3.5 text-amber-500" />
            <span>Licença GNU GPLv3</span>
          </a>
        </div>
      </div>

      <!-- Informações da empresa definidas pelo usuário administrador -->
      <div class="space-y-2.5 pt-2">
        <h5 class="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
          Informações da empresa definidas pelo usuário administrador
        </h5>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-xs">
          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Marca / Nome Fantasia</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.fantasia || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Razão Social</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.razao_social || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">CNPJ</span>
            <strong class="text-sm font-mono text-slate-800 dark:text-white">{{ configStore.config.cnpj || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Operador Atual</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ authStore.user?.nome || 'Não identificado' }} ({{ authStore.user?.role === 'admin' ? 'Administrador' : 'Operador' }})</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Telefones / WhatsApp</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.fones || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">E-mail de Contato</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.email_contato || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Endereço Comercial</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.endereco || 'Não informado' }}</strong>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-750 border border-slate-200 dark:border-slate-700">
            <span class="text-[10px] font-bold uppercase text-slate-400 block mb-0.5">Cidade - UF</span>
            <strong class="text-sm text-slate-800 dark:text-white">{{ configStore.config.cidade || 'Não informado' }}</strong>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal de Criar / Editar Usuário -->
    <div
      v-if="showUserModal"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs z-50 flex items-center justify-center p-4"
      @click.self="showUserModal = false"
    >
      <div class="bg-white dark:bg-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl border border-slate-200 dark:border-slate-700 space-y-4">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-700 pb-3">
          <h4 class="font-bold text-base text-slate-800 dark:text-white flex items-center gap-2">
            <Users class="w-4 h-4 text-opti" />
            <span>{{ editingUserId ? 'Editar Usuário' : 'Novo Usuário' }}</span>
          </h4>
          <button
            type="button"
            @click="showUserModal = false"
            class="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <form @submit.prevent="handleSaveUser" class="space-y-4 text-xs sm:text-sm">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Nome Completo
            </label>
            <input
              type="text"
              v-model="userForm.nome"
              required
              placeholder="Ex: Maria Oliveira"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              E-mail de Acesso
            </label>
            <input
              type="email"
              v-model="userForm.email"
              required
              placeholder="usuario@empresa.com.br"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Perfil de Permissão
            </label>
            <select
              v-model="userForm.role"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            >
              <option value="operador">Operador</option>
              <option value="admin">Administrador</option>
            </select>
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              {{ editingUserId ? 'Nova Senha (deixe em branco para manter)' : 'Senha de Acesso (mín. 6 caracteres)' }}
            </label>
            <input
              type="password"
              v-model="userForm.senha"
              :required="!editingUserId"
              minlength="6"
              placeholder="••••••••"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div class="flex justify-end gap-2 pt-3 border-t border-slate-200 dark:border-slate-700">
            <button
              type="button"
              @click="showUserModal = false"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
            >
              Cancelar
            </button>
            <button
              type="submit"
              :disabled="isSavingUser"
              class="bg-opti hover:bg-opti-dark text-white font-bold px-5 py-2 rounded-xl transition shadow-xs text-xs"
            >
              {{ isSavingUser ? 'Salvando...' : editingUserId ? 'Salvar Alterações' : 'Criar Usuário' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
