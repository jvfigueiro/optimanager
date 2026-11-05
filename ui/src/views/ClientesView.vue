<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useClientesStore } from '../stores/clientes'
import { useConfigStore } from '../stores/config'
import type { Cliente } from '../types'
import {
  Users,
  UserPlus,
  Search,
  Phone,
  Trash2,
  Calendar,
  Edit2,
  X,
  MessageCircle,
} from 'lucide-vue-next'

const clientesStore = useClientesStore()
const configStore = useConfigStore()

const nome = ref('')
const doc = ref('')
const contato = ref('')
const isSubmitting = ref(false)

// Estado para Edição de Cliente
const editingClient = ref<Cliente | null>(null)
const editNome = ref('')
const editDoc = ref('')
const editContato = ref('')
const isUpdating = ref(false)

onMounted(() => {
  clientesStore.loadClientes()
})

function openClientWhatsApp(c: Cliente) {
  if (!c.contato) return
  const digits = c.contato.replace(/\D/g, '')
  if (!digits) return
  const phone = digits.length <= 11 ? `55${digits}` : digits
  const text = encodeURIComponent(`Olá ${c.nome}, tudo bem? Entramos em contato da ${configStore.config.fantasia || 'OptiManager'}.`)
  window.open(`https://wa.me/${phone}?text=${text}`, '_blank')
}

async function handleAddClient() {
  if (!nome.value.trim()) return

  isSubmitting.value = true
  try {
    await clientesStore.createCliente({
      nome: nome.value.trim(),
      doc: doc.value.trim(),
      contato: contato.value.trim(),
      desde_ano: new Date().getFullYear(),
    })
    nome.value = ''
    doc.value = ''
    contato.value = ''
  } catch (err) {
    console.error(err)
  } finally {
    isSubmitting.value = false
  }
}

function openEdit(c: Cliente) {
  editingClient.value = c
  editNome.value = c.nome
  editDoc.value = c.doc || ''
  editContato.value = c.contato || ''
}

function closeEdit() {
  editingClient.value = null
}

async function handleUpdateClient() {
  if (!editingClient.value || !editNome.value.trim()) return

  isUpdating.value = true
  try {
    await clientesStore.updateCliente(editingClient.value.id, {
      nome: editNome.value.trim(),
      doc: editDoc.value.trim(),
      contato: editContato.value.trim(),
    })
    closeEdit()
  } catch (err) {
    console.error(err)
  } finally {
    isUpdating.value = false
  }
}

async function handleDeleteClient(c: Cliente) {
  if (confirm(`Deseja realmente remover o cliente "${c.nome}"?`)) {
    await clientesStore.deleteCliente(c.id)
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Formulário de Cadastro Rápido de Cliente -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5">
      <div class="flex items-center gap-2 mb-1">
        <UserPlus class="w-5 h-5 text-opti" />
        <h3 class="font-bold text-base text-slate-800 dark:text-white">
          Adicionar Cliente
        </h3>
      </div>
      <p class="text-xs text-slate-500 dark:text-slate-400 mb-4">
        Os clientes salvos aqui preenchem automaticamente o telefone, endereço e CPF/CNPJ ao criar uma OS ou Relatório.
      </p>

      <form @submit.prevent="handleAddClient" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 text-sm">
        <input
          type="text"
          v-model="nome"
          required
          placeholder="Ex: Mariana Silva ou Mercado Central"
          class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 lg:col-span-2 focus:outline-none focus:ring-2 focus:ring-opti"
        />
        <input
          type="text"
          v-model="doc"
          placeholder="000.000.000-00 ou CNPJ"
          class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
        />
        <input
          type="text"
          v-model="contato"
          placeholder="(11) 98765-4321"
          class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
        />
        <button
          type="submit"
          :disabled="isSubmitting"
          class="bg-opti hover:bg-opti-dark text-white font-bold rounded-xl px-4 py-2 transition shadow-xs flex items-center justify-center gap-1.5"
        >
          <span>{{ isSubmitting ? 'Salvando...' : '+ Salvar Cliente' }}</span>
        </button>
      </form>
    </div>

    <!-- Lista de Clientes -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5">
      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 mb-4">
        <div class="flex items-center gap-2">
          <Users class="w-5 h-5 text-opti" />
          <h3 class="font-bold text-base text-slate-800 dark:text-white">
            Clientes Cadastrados ({{ clientesStore.clientes.length }})
          </h3>
        </div>

        <div class="relative w-full sm:w-64">
          <Search class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
          <input
            type="text"
            v-model="clientesStore.searchQuery"
            @input="clientesStore.setSearch(clientesStore.searchQuery)"
            placeholder="Filtrar por nome ou doc..."
            class="text-xs border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl pl-9 pr-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-opti"
          />
        </div>
      </div>

      <!-- VISÃO DESKTOP: Layout de Lista (Tabela Limpa) -->
      <div class="hidden md:block overflow-x-auto">
        <table class="w-full text-left border-collapse text-sm">
          <thead>
            <tr class="bg-slate-50 dark:bg-slate-750 text-slate-500 dark:text-slate-400 text-xs uppercase border-b border-slate-200 dark:border-slate-700">
              <th class="py-3 px-4">ID</th>
              <th class="py-3 px-4">Cliente / Nome</th>
              <th class="py-3 px-4">CPF / CNPJ</th>
              <th class="py-3 px-4">Contato / WhatsApp</th>
              <th class="py-3 px-4">Cliente Desde</th>
              <th class="py-3 px-4 text-center">Ações</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-700/60">
            <tr
              v-for="c in clientesStore.clientes"
              :key="c.id"
              class="hover:bg-slate-50/80 dark:hover:bg-slate-750 transition"
            >
              <td class="py-3 px-4 font-mono text-xs text-opti dark:text-blue-400 font-bold">
                {{ c.id }}
              </td>
              <td class="py-3 px-4 font-bold text-slate-800 dark:text-slate-100">
                {{ c.nome }}
              </td>
              <td class="py-3 px-4 font-mono text-xs text-slate-600 dark:text-slate-300">
                {{ c.doc || 'Não informado' }}
              </td>
              <td class="py-3 px-4 text-slate-600 dark:text-slate-300">
                <span v-if="c.contato" class="flex items-center gap-1.5 text-xs">
                  <Phone class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
                  <span>{{ c.contato }}</span>
                </span>
                <span v-else class="text-xs text-slate-400">Sem telefone</span>
              </td>
              <td class="py-3 px-4 text-xs text-slate-500 dark:text-slate-400">
                <span class="flex items-center gap-1">
                  <Calendar class="w-3.5 h-3.5 text-slate-400" />
                  <span>Desde {{ c.desde_ano }}</span>
                </span>
              </td>
              <td class="py-3 px-4 text-center whitespace-nowrap">
                <div class="flex items-center justify-center gap-1">
                  <!-- Botão WhatsApp (Do lado do botão de editar) -->
                  <button
                    v-if="c.contato"
                    @click="openClientWhatsApp(c)"
                    title="Enviar mensagem via WhatsApp"
                    class="p-1.5 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 rounded-lg transition"
                  >
                    <MessageCircle class="w-4 h-4" />
                  </button>
                  <span v-else class="w-7 h-7 inline-block"></span>

                  <!-- Botão Editar -->
                  <button
                    @click="openEdit(c)"
                    title="Editar cadastro do cliente"
                    class="p-1.5 text-slate-500 hover:text-opti dark:text-slate-400 dark:hover:text-blue-400 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition"
                  >
                    <Edit2 class="w-4 h-4" />
                  </button>

                  <!-- Botão Excluir -->
                  <button
                    @click="handleDeleteClient(c)"
                    title="Remover cliente"
                    class="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-950/40 rounded-lg transition"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!clientesStore.clientes.length">
              <td colspan="6" class="py-10 text-center text-slate-400">
                Nenhum cliente cadastrado no momento.
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- VISÃO MOBILE: Lista Vertical com Linhas -->
      <div class="md:hidden divide-y divide-slate-200 dark:divide-slate-700">
        <div
          v-for="c in clientesStore.clientes"
          :key="c.id"
          class="py-3.5 flex items-center justify-between gap-3"
        >
          <div class="min-w-0">
            <div class="flex items-center gap-1.5">
              <span class="text-xs font-mono font-bold text-opti dark:text-blue-400">{{ c.id }}</span>
              <h4 class="font-bold text-slate-800 dark:text-white text-sm truncate">{{ c.nome }}</h4>
            </div>
            <p v-if="c.doc" class="text-xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
              {{ c.doc }}
            </p>
            <p v-if="c.contato" class="text-xs text-slate-600 dark:text-slate-300 flex items-center gap-1 mt-0.5">
              <Phone class="w-3 h-3 text-emerald-500" />
              <span>{{ c.contato }}</span>
            </p>
          </div>

          <div class="flex items-center gap-1 shrink-0">
            <!-- Botão WhatsApp -->
            <button
              v-if="c.contato"
              @click="openClientWhatsApp(c)"
              title="WhatsApp"
              class="p-2 text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/40 rounded-xl"
            >
              <MessageCircle class="w-4 h-4" />
            </button>

            <!-- Botão Editar -->
            <button
              @click="openEdit(c)"
              title="Editar"
              class="p-2 text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 rounded-xl"
            >
              <Edit2 class="w-4 h-4" />
            </button>

            <!-- Botão Excluir -->
            <button
              @click="handleDeleteClient(c)"
              title="Excluir"
              class="p-2 text-slate-400 hover:text-red-500 bg-slate-100 dark:bg-slate-700 rounded-xl"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div v-if="!clientesStore.clientes.length" class="py-10 text-center text-slate-400 text-xs">
          Nenhum cliente cadastrado no momento.
        </div>
      </div>
    </div>

    <!-- Modal Editar Cliente -->
    <div
      v-if="editingClient"
      class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-center justify-center p-4 z-50 no-print"
      @click.self="closeEdit"
    >
      <div
        class="bg-white dark:bg-slate-800 rounded-2xl max-w-md w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden"
      >
        <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center">
          <div class="flex items-center gap-2">
            <Edit2 class="w-5 h-5" />
            <h3 class="font-bold text-sm sm:text-base">Editar Cliente {{ editingClient.id }}</h3>
          </div>
          <button
            @click="closeEdit"
            class="text-blue-200 hover:text-white p-1 rounded-lg transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <form @submit.prevent="handleUpdateClient" class="p-5 space-y-4 text-sm">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Nome / Empresa *
            </label>
            <input
              type="text"
              v-model="editNome"
              required
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              CPF ou CNPJ
            </label>
            <input
              type="text"
              v-model="editDoc"
              placeholder="000.000.000-00 ou CNPJ"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              WhatsApp / Telefone
            </label>
            <input
              type="text"
              v-model="editContato"
              placeholder="(11) 98765-4321"
              class="w-full border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
            />
          </div>

          <div class="pt-2 flex justify-end gap-2">
            <button
              type="button"
              @click="closeEdit"
              class="px-4 py-2 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-xl text-xs font-semibold transition"
            >
              Cancelar
            </button>
            <button
              type="submit"
              :disabled="isUpdating"
              class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-xs disabled:opacity-50"
            >
              <Check class="w-3.5 h-3.5" />
              <span>{{ isUpdating ? 'Salvando...' : 'Salvar Alterações' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

