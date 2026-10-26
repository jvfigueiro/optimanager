<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfigStore } from '../stores/config'
import { Settings, Upload, Image, Palette, CheckCircle2 } from 'lucide-vue-next'

const configStore = useConfigStore()

const form = ref({ ...configStore.config })
const isSaving = ref(false)

const paletasPredefinidas = [
  { nome: 'Opti Padrão', primaria: '#0B5694', secundaria: '#083f6e' },
  { nome: 'Esmeralda Tech', primaria: '#059669', secundaria: '#064e3b' },
  { nome: 'Índigo Moderno', primaria: '#4f46e5', secundaria: '#3730a3' },
  { nome: 'Grafite Elegante', primaria: '#334155', secundaria: '#1e293b' },
  { nome: 'Cobre Industrial', primaria: '#c2410c', secundaria: '#7c2d12' },
]

onMounted(async () => {
  await configStore.loadConfig()
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
</script>

<template>
  <div class="max-w-4xl space-y-6">
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
                class="flex items-center gap-2 px-3 py-1.5 rounded-xl text-xs font-medium border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-700 hover:border-slate-400 transition"
              >
                <span class="w-3.5 h-3.5 rounded-full" :style="{ backgroundColor: p.primaria }"></span>
                <span class="text-slate-700 dark:text-slate-200">{{ p.nome }}</span>
              </button>
            </div>

            <!-- Seletores Hex Manuais -->
            <div class="grid grid-cols-2 gap-3 pt-2">
              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Cor Primária (Hex)</label>
                <div class="flex items-center gap-2">
                  <input
                    type="color"
                    v-model="form.cor_primaria"
                    @input="handleColorChange"
                    class="w-8 h-8 rounded cursor-pointer border-0 p-0"
                  />
                  <input
                    type="text"
                    v-model="form.cor_primaria"
                    @input="handleColorChange"
                    class="text-xs border border-slate-300 dark:border-slate-600 dark:bg-slate-700 rounded-lg px-2.5 py-1.5 font-mono w-28 uppercase"
                  />
                </div>
              </div>

              <div>
                <label class="block text-[11px] font-bold text-slate-500 mb-1">Cor Secundária (Escura)</label>
                <div class="flex items-center gap-2">
                  <input
                    type="color"
                    v-model="form.cor_secundaria"
                    @input="handleColorChange"
                    class="w-8 h-8 rounded cursor-pointer border-0 p-0"
                  />
                  <input
                    type="text"
                    v-model="form.cor_secundaria"
                    @input="handleColorChange"
                    class="text-xs border border-slate-300 dark:border-slate-600 dark:bg-slate-700 rounded-lg px-2.5 py-1.5 font-mono w-28 uppercase"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Dados Cadastrais do MEI/Empresa -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs sm:text-sm">
          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Nome Fantasia / Marca
            </label>
            <input
              type="text"
              v-model="form.fantasia"
              required
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Razão Social / Titular MEI
            </label>
            <input
              type="text"
              v-model="form.razao_social"
              required
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div>
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              CNPJ
            </label>
            <input
              type="text"
              v-model="form.cnpj"
              required
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 font-mono focus:ring-2 focus:ring-opti outline-none"
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
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
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
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
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
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none"
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
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 font-bold focus:ring-2 focus:ring-opti outline-none"
            />
          </div>

          <div class="sm:col-span-2">
            <label class="block text-xs font-bold uppercase text-slate-500 dark:text-slate-400 mb-1">
              Termo de Garantia Padrão (Rodapé dos Relatórios)
            </label>
            <textarea
              v-model="form.termo_garantia"
              rows="2"
              class="w-full border border-slate-300 dark:border-slate-600 dark:bg-slate-700 dark:text-white rounded-xl px-3.5 py-2 focus:ring-2 focus:ring-opti outline-none text-xs"
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
  </div>
</template>
