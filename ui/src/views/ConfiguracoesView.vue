<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfigStore } from '../stores/config'
import { useAuthStore } from '../stores/auth'
import type { Usuario } from '../types'
import {
  Settings,
  Upload,
  Image,
  Palette,
  CheckCircle2,
  Users,
  UserPlus,
  Edit3,
  Trash2,
  ShieldCheck,
  User,
  X,
} from 'lucide-vue-next'

const configStore = useConfigStore()
const authStore = useAuthStore()

const form = ref({ ...configStore.config })
const isSaving = ref(false)

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
    <!-- Card 1: Gestão de Usuários e Acessos -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-6">
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
              <Image v-else class="w-8 h-8 text-slate-400" />
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
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 font-mono focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Responsável Técnico
            </label>
            <input
              type="text"
              v-model="form.responsavel"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
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
              placeholder="Rua, Número, Bairro"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Cidade - UF
            </label>
            <input
              type="text"
              v-model="form.cidade"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Valor Padrão Visita Técnica (VT R$)
            </label>
            <input
              type="number"
              step="0.01"
              v-model.number="form.vt_valor_padrao"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 font-bold focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div class="sm:col-span-2">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Termo de Garantia Padrão (Rodapé dos Relatórios)
            </label>
            <textarea
              v-model="form.termo_garantia"
              rows="2"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none text-xs"
            ></textarea>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-200 dark:border-slate-700 flex justify-end">
          <button
            type="submit"
            :disabled="isSaving"
            class="bg-opti hover:bg-opti-dark text-white font-bold px-6 py-2.5 rounded-xl transition shadow flex items-center gap-2 text-sm"
          >
            <CheckCircle2 class="w-4 h-4" />
            <span>{{ isSaving ? 'Salvando...' : 'Salvar Alterações' }}</span>
          </button>
        </div>
      </form>
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
