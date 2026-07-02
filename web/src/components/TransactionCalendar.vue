<script setup lang="ts">
import { computed } from "vue";
import type { Transaction } from "../types/transaction";

const props = defineProps<{
  month: number;
  year: number;
  transactions: Transaction[];
  selectedDate: string | null;
  canGoPrev: boolean;
  canGoNext: boolean;
}>();

const emit = defineEmits<{
  prev: [];
  next: [];
  "select-date": [date: string | null];
}>();

const weekdayLabels = ["D", "S", "T", "Q", "Q", "S", "S"];

const monthLabel = computed(() =>
  new Date(props.year, props.month - 1, 1).toLocaleDateString("pt-BR", { month: "long", year: "numeric" }),
);

interface DayCell {
  day: number;
  date: string;
  hasTransactions: boolean;
  income: number;
  expense: number;
  isToday: boolean;
}

const todayIso = new Date().toISOString().slice(0, 10);

const dayMap = computed(() => {
  const map = new Map<string, { income: number; expense: number }>();
  for (const tx of props.transactions) {
    const entry = map.get(tx.date) ?? { income: 0, expense: 0 };
    if (tx.type === "income") entry.income += tx.amount;
    else entry.expense += tx.amount;
    map.set(tx.date, entry);
  }
  return map;
});

const weeks = computed(() => {
  const firstOfMonth = new Date(props.year, props.month - 1, 1);
  const startOffset = firstOfMonth.getDay();
  const daysInMonth = new Date(props.year, props.month, 0).getDate();

  const cells: (DayCell | null)[] = [];
  for (let i = 0; i < startOffset; i++) cells.push(null);
  for (let day = 1; day <= daysInMonth; day++) {
    const date = `${props.year}-${String(props.month).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
    const stats = dayMap.value.get(date);
    cells.push({
      day,
      date,
      hasTransactions: !!stats,
      income: stats?.income ?? 0,
      expense: stats?.expense ?? 0,
      isToday: date === todayIso,
    });
  }
  while (cells.length % 7 !== 0) cells.push(null);

  const rows: (DayCell | null)[][] = [];
  for (let i = 0; i < cells.length; i += 7) rows.push(cells.slice(i, i + 7));
  return rows;
});

function toggleDay(cell: DayCell) {
  if (!cell.hasTransactions) return;
  emit("select-date", props.selectedDate === cell.date ? null : cell.date);
}
</script>

<template>
  <div class="rounded-2xl border border-ink-200/70 bg-white p-3 shadow-sm">
    <div class="flex items-center justify-between gap-1">
      <button
        type="button"
        :disabled="!canGoPrev"
        class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-ink-200 text-xs text-ink-600 transition hover:bg-ink-100 disabled:cursor-not-allowed disabled:opacity-30"
        aria-label="Mês anterior"
        @click="emit('prev')"
      >
        &lsaquo;
      </button>
      <p class="font-display text-xs font-bold capitalize text-ink-900">{{ monthLabel }}</p>
      <button
        type="button"
        :disabled="!canGoNext"
        class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-ink-200 text-xs text-ink-600 transition hover:bg-ink-100 disabled:cursor-not-allowed disabled:opacity-30"
        aria-label="Próximo mês"
        @click="emit('next')"
      >
        &rsaquo;
      </button>
    </div>

    <div class="mt-2 grid grid-cols-7 gap-0.5 text-center text-[9px] font-semibold text-ink-400">
      <span v-for="(w, i) in weekdayLabels" :key="i">{{ w }}</span>
    </div>

    <div class="mt-1 flex flex-col gap-0.5">
      <div v-for="(week, wi) in weeks" :key="wi" class="grid grid-cols-7 gap-0.5">
        <button
          v-for="(cell, ci) in week"
          :key="ci"
          type="button"
          :disabled="!cell || !cell.hasTransactions"
          class="group relative flex aspect-square flex-col items-center justify-center rounded-lg text-[10px] transition"
          :class="[
            !cell ? 'pointer-events-none' : '',
            cell && cell.hasTransactions
              ? 'cursor-pointer font-semibold text-ink-900 hover:bg-brand-50'
              : cell
                ? 'text-ink-300'
                : '',
            cell && selectedDate === cell.date ? 'bg-brand-500 !text-white hover:bg-brand-500' : '',
            cell && cell.isToday && selectedDate !== cell.date ? 'ring-1 ring-inset ring-brand-300' : '',
          ]"
          @click="cell && toggleDay(cell)"
        >
          <span v-if="cell">{{ cell.day }}</span>
          <span
            v-if="cell && cell.hasTransactions"
            class="mt-0.5 flex gap-0.5"
          >
            <span
              v-if="cell.income > 0"
              class="h-0.5 w-0.5 rounded-full"
              :class="selectedDate === cell.date ? 'bg-white' : 'bg-brand-500'"
            ></span>
            <span
              v-if="cell.expense > 0"
              class="h-0.5 w-0.5 rounded-full"
              :class="selectedDate === cell.date ? 'bg-white' : 'bg-coral-500'"
            ></span>
          </span>
        </button>
      </div>
    </div>

    <div class="mt-3 flex flex-wrap items-center gap-2 text-[10px] text-ink-500">
      <span class="flex items-center gap-1"><span class="h-1.5 w-1.5 rounded-full bg-brand-500"></span> Receita</span>
      <span class="flex items-center gap-1"><span class="h-1.5 w-1.5 rounded-full bg-coral-500"></span> Despesa</span>
      <button
        v-if="selectedDate"
        type="button"
        class="ml-auto font-semibold text-brand-700 hover:text-brand-800"
        @click="emit('select-date', null)"
      >
        Limpar
      </button>
    </div>
  </div>
</template>
