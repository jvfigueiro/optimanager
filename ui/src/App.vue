<script setup lang="ts">
import { onMounted } from 'vue'
import AppLayout from './components/AppLayout.vue'
import { useConfigStore } from './stores/config'
import { useNavigationStore } from './stores/navigation'
import { useClientesStore } from './stores/clientes'

const configStore = useConfigStore()
const nav = useNavigationStore()
const clientesStore = useClientesStore()

onMounted(async () => {
  configStore.initTheme()
  nav.initFromHash()
  await configStore.loadConfig()
  clientesStore.loadClientes()
})
</script>

<template>
  <AppLayout />
</template>
