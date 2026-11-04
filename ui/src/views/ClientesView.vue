<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useClientesStore } from '../stores/clientes'
import { Users, UserPlus, Search, Phone, Trash2, Calendar } from 'lucide-vue-next'

const clientesStore = useClientesStore()

const nome = ref('')
const doc = ref('')
const contato = ref('')
const isSubmitting = ref(false)

onMounted(() => {
  clientesStore.loadClientes()
})

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
</script>

<template>
  <div class="space-y-6">
    <!-- Formulário de Cadastro Rápido de Cliente -->
    <div class="bg-white dark:bg-slate-800 rounded-2xl border border-slate-200 dark:border-slate-700 shadow-xs p-5">
      <div class="flex items-center gap-2 mb-1">
        <UserPlus class="w-5 h-5 text-opti" />
        <h3 class="font-bold text-base text-slate-800 dark:text-white">
          Adicionar Cliente à Base de Autocompletar
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
          placeholder="Nome / Empresa *"
          class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 lg:col-span-2 focus:outline-none focus:ring-2 focus:ring-opti"
        />
        <input
          type="text"
          v-model="doc"
          placeholder="CPF ou CNPJ (opcional)"
          class="border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-white rounded-xl px-3.5 py-2 focus:outline-none focus:ring-2 focus:ring-opti"
        />
        <input
          type="text"
          v-model="contato"
          placeholder="WhatsApp / Telefone"
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

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
        <div
          v-for="c in clientesStore.clientes"
          :key="c.id"
          class="group p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-750 flex flex-col justify-between hover:border-blue-300 dark:hover:border-slate-600 transition"
        >
          <div>
            <div class="flex justify-between items-start">
              <h4 class="font-bold text-slate-800 dark:text-white text-base">{{ c.nome }}</h4>
              <button
                @click="clientesStore.deleteCliente(c.id)"
                title="Excluir cliente"
                class="opacity-0 group-hover:opacity-100 text-slate-400 hover:text-red-500 p-1 transition"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
              CPF/CNPJ: <span class="font-mono">{{ c.doc || 'Não informado' }}</span>
            </p>
            <div class="flex items-center gap-1.5 text-xs text-slate-600 dark:text-slate-300 mt-1">
              <Phone class="w-3.5 h-3.5 text-emerald-500 shrink-0" />
              <span>{{ c.contato || 'Sem telefone' }}</span>
            </div>
          </div>

          <div class="flex justify-between items-center pt-3 mt-3 border-t border-slate-200/80 dark:border-slate-700 text-[11px] text-slate-400">
            <span class="flex items-center gap-1">
              <Calendar class="w-3 h-3" />
              <span>Desde {{ c.desde_ano }}</span>
            </span>
            <span class="text-opti font-bold text-xs">#{{ c.id }}</span>
          </div>
        </div>

        <div v-if="!clientesStore.clientes.length" class="col-span-full py-12 text-center text-slate-400">
          Nenhum cliente cadastrado no momento.
        </div>
      </div>
    </div>
  </div>
</template>

