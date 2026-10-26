<script setup lang="ts">
import { ref } from 'vue'
import { useDashboardStore } from '../stores/dashboard'
import { X, Link2 } from 'lucide-vue-next'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const dashboardStore = useDashboardStore()

const titulo = ref('')
const descricao = ref('')
const url = ref('')
const isSubmitting = ref(false)

async function handleSubmit() {
  if (!titulo.value.trim() || !url.value.trim()) return

  isSubmitting.value = true
  try {
    await dashboardStore.addShortcut({
      titulo: titulo.value.trim(),
      descricao: descricao.value.trim(),
      url: url.value.trim(),
      ordem: dashboardStore.data.atalhos.length + 1,
    })
    emit('close')
  } catch (err) {
    console.error('Erro ao adicionar atalho:', err)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    class="fixed inset-0 bg-slate-900/60 dark:bg-black/80 backdrop-blur-xs flex items-center justify-center p-4 z-50 no-print"
    @click.self="emit('close')"
  >
    <div
      class="bg-white dark:bg-slate-800 rounded-2xl max-w-md w-full shadow-2xl border border-slate-200 dark:border-slate-700 overflow-hidden"
    >
      <div class="bg-opti text-white px-5 py-3.5 flex justify-between items-center">
        <div class="flex items-center gap-2">
          <Link2 class="w-5 h-5" />
          <h3 class="font-bold text-sm sm:text-base">Novo Atalho Rápido</h3>
        </div>
        <button
          @click="emit('close')"
          class="text-blue-200 hover:text-white p-1 rounded-lg transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="p-5 space-y-3.5 text-sm">
        <div>
          <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
            Nome do Portal / Site *
          </label>
          <input
            type="text"
            v-model="titulo"
            required
            placeholder="Ex: Portal Nacional NFS-e"
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
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
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
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
            class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-opti"
          />
        </div>

        <div class="pt-2 flex justify-end gap-2">
          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-2 rounded-lg border border-slate-300 dark:border-slate-600 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
          >
            Cancelar
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2 rounded-lg bg-opti hover:bg-opti-dark text-white font-bold transition shadow"
          >
            {{ isSubmitting ? 'Salvando...' : 'Adicionar Atalho' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
