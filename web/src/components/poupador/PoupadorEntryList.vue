<script setup lang="ts">
import { Pencil, Trash2 } from "lucide-vue-next";
import type { PoupadorEntry, PoupadorKind } from "../../types/poupador";

defineProps<{ entries: PoupadorEntry[]; kind: PoupadorKind }>();
defineEmits<{ edit: [entry: PoupadorEntry]; remove: [id: string] }>();

const frequencyLabels = {
  monthly: "Mensal",
  weekly: "Semanal",
  biweekly: "Quinzenal",
  yearly: "Anual",
  "one-time": "Único",
};
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
</script>

<template>
  <div class="rounded-3xl border border-ink-200/70 bg-white shadow-sm">
    <div class="flex items-center justify-between border-b border-ink-100 px-5 py-4">
      <h2 class="font-display text-base font-bold text-ink-900">
        {{ kind === "income" ? "Suas receitas" : "Seus gastos" }}
      </h2>
      <span class="rounded-full bg-ink-100 px-2.5 py-1 text-xs font-bold text-ink-500">{{
        entries.length
      }}</span>
    </div>
    <ul v-if="entries.length" class="divide-y divide-ink-100">
      <li v-for="entry in entries" :key="entry.id" class="flex items-center gap-3 px-5 py-3.5">
        <span
          :class="kind === 'income' ? 'bg-brand-100 text-brand-700' : 'bg-coral-100 text-coral-700'"
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-lg font-bold"
          >{{ kind === "income" ? "+" : "−" }}</span
        >
        <div class="min-w-0 grow">
          <p class="truncate text-sm font-bold text-ink-800">{{ entry.name }}</p>
          <p class="text-xs text-ink-500">
            {{ frequencyLabels[entry.frequency]
            }}<template v-if="entry.frequency === 'one-time'"> · mês {{ entry.month }}</template>
          </p>
        </div>
        <span class="text-sm font-bold text-ink-800">{{ currency.format(entry.amount) }}</span>
        <button
          :aria-label="`Editar ${entry.name}`"
          class="rounded-lg p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-800"
          @click="$emit('edit', entry)"
        >
          <Pencil class="h-4 w-4" />
        </button>
        <button
          :aria-label="`Excluir ${entry.name}`"
          class="rounded-lg p-1.5 text-ink-400 transition hover:bg-coral-50 hover:text-coral-600"
          @click="$emit('remove', entry.id)"
        >
          <Trash2 class="h-4 w-4" />
        </button>
      </li>
    </ul>
    <p v-else class="px-5 py-8 text-center text-sm text-ink-400">Nenhuma fonte adicionada ainda.</p>
  </div>
</template>
