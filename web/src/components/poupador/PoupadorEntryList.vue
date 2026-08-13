<script setup lang="ts">
import { computed } from "vue";
import { Pencil, Trash2 } from "lucide-vue-next";
import type { PoupadorEntry, PoupadorKind } from "../../types/poupador";

const props = defineProps<{ entries: PoupadorEntry[]; kind: PoupadorKind }>();
defineEmits<{ edit: [entry: PoupadorEntry]; remove: [id: string] }>();

const frequencyLabels = {
  monthly: "Mensal",
  weekly: "Semanal",
  biweekly: "Quinzenal",
  yearly: "Anual",
  "one-time": "Único",
};
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const sortedEntries = computed(() => [...props.entries].sort((a, b) => b.amount - a.amount));
</script>

<template>
  <div
    class="w-full min-w-0 overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm"
  >
    <div class="flex items-center justify-between border-b border-ink-100 px-5 py-4">
      <h2 class="min-w-0 truncate font-display text-base font-bold text-ink-900">
        {{ kind === "income" ? "Suas receitas" : "Seus gastos" }}
      </h2>
      <span class="rounded-full bg-ink-100 px-2.5 py-1 text-xs font-bold text-ink-500">{{
        entries.length
      }}</span>
    </div>
    <ul v-if="entries.length" class="divide-y divide-ink-100">
      <li
        v-for="entry in sortedEntries"
        :key="entry.id"
        class="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-3 px-4 py-3.5 sm:px-5"
      >
        <div class="flex min-w-0 items-center gap-3">
          <span
            :class="
              kind === 'income' ? 'bg-brand-100 text-brand-700' : 'bg-coral-100 text-coral-700'
            "
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-lg font-bold"
            >{{ kind === "income" ? "+" : "−" }}</span
          >
          <div class="min-w-0">
            <p class="truncate text-sm font-bold text-ink-800">{{ entry.name }}</p>
            <p class="truncate text-xs text-ink-500">
              {{ frequencyLabels[entry.frequency]
              }}<template v-if="entry.frequency === 'one-time'"> · mês {{ entry.month }}</template>
            </p>
          </div>
        </div>
        <div class="flex shrink-0 items-center gap-1 sm:gap-2">
          <span class="text-xs font-bold text-ink-800 sm:text-sm">{{
            currency.format(entry.amount)
          }}</span>
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
        </div>
      </li>
    </ul>
    <p v-else class="px-5 py-8 text-center text-sm text-ink-400">Nenhuma fonte adicionada ainda.</p>
  </div>
</template>
