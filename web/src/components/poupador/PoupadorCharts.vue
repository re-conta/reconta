<script setup lang="ts">
import { computed } from "vue";
import { Bar, Doughnut } from "vue-chartjs";
import {
  ArcElement,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  Tooltip,
} from "chart.js";
import type { ChartOptions } from "chart.js";
import type { PoupadorEntry } from "../../types/poupador";

ChartJS.register(ArcElement, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

const props = defineProps<{
  incomes: PoupadorEntry[];
  expenses: PoupadorEntry[];
  monthlySeries: Array<{ label: string; income: number; expense: number; balance: number }>;
}>();

const currency = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
  maximumFractionDigits: 0,
});
const colors = ["#f2751f", "#ffb42e", "#e8496b", "#d63163", "#9b4419", "#7c6d5c"];
function recurringMonthlyAmount(entry: PoupadorEntry) {
  if (entry.frequency === "one-time") return entry.amount / 12;
  return (
    entry.amount *
    { monthly: 1, weekly: 52 / 12, biweekly: 26 / 12, yearly: 1 / 12 }[entry.frequency]
  );
}
function breakdown(entries: PoupadorEntry[]) {
  return entries.map((entry, index) => ({
    label: entry.name,
    value: recurringMonthlyAmount(entry),
    color: colors[index % colors.length],
  }));
}
const incomeBreakdown = computed(() => breakdown(props.incomes));
const expenseBreakdown = computed(() => breakdown(props.expenses));
const annualData = computed(() => ({
  labels: props.monthlySeries.map((item) => item.label),
  datasets: [
    {
      label: "Receitas",
      data: props.monthlySeries.map((item) => item.income),
      backgroundColor: "#f2751f",
      borderRadius: 6,
      borderSkipped: false,
    },
    {
      label: "Gastos",
      data: props.monthlySeries.map((item) => item.expense),
      backgroundColor: "#d63163",
      borderRadius: 6,
      borderSkipped: false,
    },
  ],
}));
const chartOptions: ChartOptions<"bar"> = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "top" as const,
      align: "end" as const,
      labels: {
        usePointStyle: true,
        boxWidth: 8,
        color: "#5c5044",
        font: { family: "Nunito", size: 12, weight: 700 as const },
      },
    },
    tooltip: {
      backgroundColor: "#1c1712",
      padding: 10,
      cornerRadius: 10,
      callbacks: {
        label: (context) => `${context.dataset.label}: ${currency.format(context.parsed.y ?? 0)}`,
      },
    },
  },
  scales: {
    x: { grid: { display: false }, border: { display: false }, ticks: { color: "#7c6d5c" } },
    y: {
      grid: { color: "#e6ded5" },
      border: { display: false },
      ticks: {
        color: "#7c6d5c",
        callback: (value: number | string) => currency.format(Number(value)),
        maxTicksLimit: 5,
      },
    },
  },
};
const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: "68%",
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: "#1c1712",
      callbacks: {
        label: (context: { label: string; parsed: number }) =>
          `${context.label}: ${currency.format(context.parsed)}`,
      },
    },
  },
};
function doughnutData(items: Array<{ label: string; value: number; color: string }>) {
  return {
    labels: items.map((item) => item.label),
    datasets: [
      {
        data: items.map((item) => item.value),
        backgroundColor: items.map((item) => item.color),
        borderWidth: 0,
        hoverOffset: 4,
      },
    ],
  };
}
</script>

<template>
  <section class="grid gap-5 xl:grid-cols-3">
    <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm xl:col-span-2">
      <div class="mb-4">
        <h2 class="font-display text-base font-bold text-ink-900">Projeção anual</h2>
        <p class="text-xs text-ink-500">Receitas e gastos estimados em cada mês</p>
      </div>
      <div class="h-64 sm:h-80"><Bar :data="annualData" :options="chartOptions" /></div>
    </div>
    <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
      <h2 class="font-display text-base font-bold text-ink-900">Fontes de receitas</h2>
      <p class="mb-4 text-xs text-ink-500">Participação na média mensal</p>
      <div v-if="incomeBreakdown.length" class="h-40">
        <Doughnut :data="doughnutData(incomeBreakdown)" :options="doughnutOptions" />
      </div>
      <p v-else class="flex h-40 items-center justify-center text-sm text-ink-400">
        Adicione receitas para ver o gráfico.
      </p>
      <ul v-if="incomeBreakdown.length" class="mt-4 space-y-2">
        <li
          v-for="item in incomeBreakdown"
          :key="item.label"
          class="flex items-center justify-between gap-3 text-xs"
        >
          <span class="flex min-w-0 items-center gap-2 text-ink-600"
            ><i
              class="h-2.5 w-2.5 shrink-0 rounded-full"
              :style="{ backgroundColor: item.color }"
            ></i
            ><span class="truncate">{{ item.label }}</span></span
          ><span class="font-bold text-ink-800">{{ currency.format(item.value) }}</span>
        </li>
      </ul>
    </div>
    <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm xl:col-start-3">
      <h2 class="font-display text-base font-bold text-ink-900">Fontes de gastos</h2>
      <p class="mb-4 text-xs text-ink-500">Participação na média mensal</p>
      <div v-if="expenseBreakdown.length" class="h-40">
        <Doughnut :data="doughnutData(expenseBreakdown)" :options="doughnutOptions" />
      </div>
      <p v-else class="flex h-40 items-center justify-center text-sm text-ink-400">
        Adicione gastos para ver o gráfico.
      </p>
      <ul v-if="expenseBreakdown.length" class="mt-4 space-y-2">
        <li
          v-for="item in expenseBreakdown"
          :key="item.label"
          class="flex items-center justify-between gap-3 text-xs"
        >
          <span class="flex min-w-0 items-center gap-2 text-ink-600"
            ><i
              class="h-2.5 w-2.5 shrink-0 rounded-full"
              :style="{ backgroundColor: item.color }"
            ></i
            ><span class="truncate">{{ item.label }}</span></span
          ><span class="font-bold text-ink-800">{{ currency.format(item.value) }}</span>
        </li>
      </ul>
    </div>
  </section>
</template>
