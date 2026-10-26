<script setup lang="ts">
import { useToastStore } from '../stores/toast'

const toastStore = useToastStore()
</script>

<template>
  <div class="fixed bottom-20 lg:bottom-6 right-4 sm:right-6 z-50 flex flex-col gap-2 pointer-events-none no-print">
    <TransitionGroup
      enter-active-class="transform ease-out duration-300 transition"
      enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
      enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
      leave-active-class="transition ease-in duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        class="pointer-events-auto px-4 py-3 rounded-xl shadow-lg border text-xs sm:text-sm font-medium flex items-center gap-2.5 max-w-sm backdrop-blur-md"
        :class="[
          toast.type === 'error'
            ? 'bg-rose-900/90 text-white border-rose-700'
            : toast.type === 'info'
            ? 'bg-sky-900/90 text-white border-sky-700'
            : 'bg-slate-900/95 text-white border-emerald-500/80'
        ]"
      >
        <span class="flex-1">{{ toast.text }}</span>
        <button
          @click="toastStore.remove(toast.id)"
          class="text-slate-400 hover:text-white text-lg leading-none p-1"
        >
          &times;
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

