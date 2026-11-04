<script setup lang="ts">
import { ref } from 'vue'
import { useDashboardStore } from '../stores/dashboard'
import type { Atalho } from '../types'
import { X, Link2, Plus, Edit2, Trash2, Check, RotateCcw } from 'lucide-vue-next'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const dashboardStore = useDashboardStore()

const editingId = ref<number | null>(null)
const titulo = ref('')
const descricao = ref('')
const url = ref('')
const isSubmitting = ref(false)

function startEdit(a: Atalho) {
  editingId.value = a.id
  titulo.value = a.titulo
  descricao.value = a.descricao || ''
  url.value = a.url
}

function cancelEdit() {
  editingId.value = null
  titulo.value = ''
  descricao.value = ''
  url.value = ''
}

async function handleSubmit() {
  if (!titulo.value.trim() || !url.value.trim()) return

  isSubmitting.value = true
  try {
    if (editingId.value) {
      await dashboardStore.updateShortcut(editingId.value, {
        titulo: titulo.value.trim(),
        descricao: descricao.value.trim(),
        url: url.value.trim(),
      })
      cancelEdit()
    } else {
      await dashboardStore.addShortcut({
        titulo: titulo.value.trim(),
        descricao: descricao.value.trim(),
        url: url.value.trim(),
        ordem: dashboardStore.data.atalhos.length + 1,
      })
      cancelEdit()
    }
  } catch (err) {
    console.error('Erro ao salvar atalho:', err)
  } finally {
    isSubmitting.value = false
  }
}

async function handleDelete(id: number) {
  if (confirm('Deseja realmente remover este atalho?')) {
    if (editingId.value === id) {
      cancelEdit()
    }
    await dashboardStore.removeShortcut(id)
  }
}
</script>

<template>
  <div
    class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-center justify-center p-4 z-50 no-print"
    @click.self="emit('close')"
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-2xl max-w-lg w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden flex flex-col max-h-[90vh]"
    >
      <!-- Header -->
      <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center shrink-0">
        <div class="flex items-center gap-2">
          <Link2 class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">Gerenciar Atalhos Rápidos</h3>
        </div>
        <button
          @click="emit('close')"
          class="text-blue-200 hover:text-white p-1 rounded-lg transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Conteúdo com Scroll -->
      <div class="p-5 space-y-5 overflow-y-auto">
        <!-- Formulário Adicionar / Editar -->
        <div class="bg-slate-50 dark:bg-slate-750 p-4 rounded-xl border border-slate-200 dark:border-slate-700">
          <div class="flex items-center justify-between mb-3">
            <h4 class="font-bold text-xs uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-1.5">
              <span v-if="editingId" class="text-amber-600 dark:text-amber-400">✏️ Editar Atalho</span>
              <span v-else class="text-opti dark:text-blue-400">➕ Novo Atalho</span>
            </h4>
            <button
              v-if="editingId"
              type="button"
              @click="cancelEdit"
              class="text-xs text-slate-500 hover:text-slate-700 dark:text-slate-400 flex items-center gap-1"
            >
              <RotateCcw class="w-3 h-3" />
              Cancelar Edição
            </button>
          </div>

          <form @submit.prevent="handleSubmit" class="space-y-3 text-sm">
            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Nome do Portal / Site *
              </label>
              <input
                type="text"
                v-model="titulo"
                required
                placeholder="Ex: Portal Nacional NFS-e"
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
            </div>

            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                Descrição Curta
              </label>
              <input
                type="text"
                v-model="descricao"
                placeholder="Ex: Emissão de notas e consulta de pendências"
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
            </div>

            <div>
              <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
                URL (Link Completo) *
              </label>
              <input
                type="url"
                v-model="url"
                required
                placeholder="https://..."
                class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
              />
            </div>

            <div class="pt-1 flex justify-end gap-2">
              <button
                type="submit"
                :disabled="isSubmitting"
                class="px-4 py-2 bg-opti hover:bg-opti-dark text-white rounded-lg text-xs font-bold transition flex items-center gap-1.5 shadow-xs disabled:opacity-50"
              >
                <Check v-if="editingId" class="w-3.5 h-3.5" />
                <Plus v-else class="w-3.5 h-3.5" />
                <span>{{ editingId ? 'Salvar Alterações' : 'Adicionar Atalho' }}</span>
              </button>
            </div>
          </form>
        </div>

        <!-- Lista de Atalhos Atuais -->
        <div>
          <h4 class="font-bold text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2.5">
            Atalhos Cadastrados ({{ dashboardStore.data.atalhos.length }})
          </h4>

          <div
            v-if="dashboardStore.data.atalhos.length === 0"
            class="text-center py-6 text-xs text-slate-400 dark:text-slate-500 border border-dashed border-slate-200 dark:border-slate-700 rounded-xl"
          >
            Nenhum atalho cadastrado ainda. Use o formulário acima para adicionar.
          </div>

          <div v-else class="divide-y divide-slate-100 dark:divide-slate-700 border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden bg-white dark:bg-slate-800">
            <div
              v-for="a in dashboardStore.data.atalhos"
              :key="a.id"
              class="p-3 flex items-center justify-between gap-3 hover:bg-slate-50 dark:hover:bg-slate-750 transition"
              :class="editingId === a.id ? 'bg-amber-50/60 dark:bg-amber-950/30' : ''"
            >
              <div class="min-w-0 flex-1">
                <div class="font-bold text-sm text-slate-800 dark:text-slate-100 truncate">
                  {{ a.titulo }}
                </div>
                <div class="text-xs text-slate-500 dark:text-slate-400 truncate">
                  {{ a.descricao || a.url }}
                </div>
              </div>

              <div class="flex items-center gap-1 shrink-0">
                <button
                  @click="startEdit(a)"
                  class="p-1.5 text-slate-500 hover:text-opti dark:hover:text-blue-400 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition"
                  title="Editar atalho"
                >
                  <Edit2 class="w-4 h-4" />
                </button>
                <button
                  @click="handleDelete(a.id)"
                  class="p-1.5 text-slate-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/40 rounded-lg transition"
                  title="Excluir atalho"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-slate-200 dark:border-slate-700 flex justify-end shrink-0 bg-slate-50 dark:bg-slate-850">
        <button
          type="button"
          @click="emit('close')"
          class="px-4 py-2 border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-xl text-xs font-semibold transition"
        >
          Concluir
        </button>
      </div>
    </div>
  </div>
</template>

