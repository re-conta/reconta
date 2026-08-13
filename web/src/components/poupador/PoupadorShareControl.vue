<script setup lang="ts">
import { Check, Copy, Link, LoaderCircle, TriangleAlert } from "lucide-vue-next";

defineProps<{
  shareUrl: string | null;
  status: "idle" | "saving" | "saved" | "error";
  error: string | null;
  copied: boolean;
}>();
defineEmits<{ copy: [] }>();
</script>

<template>
  <div
    class="w-full mt-5 rounded-2xl border border-brand-200 bg-brand-50 px-4 py-3 sm:flex sm:items-center sm:justify-between sm:gap-4"
  >
    <div class="flex items-center gap-2 text-sm text-brand-800">
      <LoaderCircle v-if="status === 'saving'" class="h-4 w-4 animate-spin" />
      <TriangleAlert v-else-if="status === 'error'" class="h-4 w-4" />
      <Check v-else-if="status === 'saved'" class="h-4 w-4" />
      <Link v-else class="h-4 w-4" />
      <p>
        <template v-if="status === 'saving'">Salvando resultado…</template>
        <template v-else-if="status === 'error'">{{ error }}</template>
        <template v-else-if="status === 'saved'">Resultado salvo. Compartilhe este link.</template>
        <template v-else
          >O link para compartilhar será criado ao salvar uma receita ou gasto.</template
        >
      </p>
    </div>
    <button
      v-if="shareUrl"
      type="button"
      class="mt-3 inline-flex items-center gap-2 rounded-full bg-ink-900 px-3.5 py-2 text-xs font-bold text-white transition hover:bg-ink-800 sm:mt-0"
      @click="$emit('copy')"
    >
      <Check v-if="copied" class="h-3.5 w-3.5" />
      <Copy v-else class="h-3.5 w-3.5" />
      {{ copied ? "Link copiado" : "Copiar link" }}
    </button>
  </div>
</template>
