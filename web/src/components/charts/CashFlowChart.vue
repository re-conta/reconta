<script setup lang="ts">
import { computed } from "vue";
import { Bar } from "vue-chartjs";
import {
  BarController,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  Tooltip,
} from "chart.js";
import type { Transaction } from "../../types/transaction";

ChartJS.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

const props = defineProps<{
  month: number;
  year: number;
  transactions: Transaction[];
}>();

const INCOME_COLOR = "#f2751f";
const EXPENSE_COLOR = "#d63163";
const GRID_COLOR = "#e6ded5";
const TEXT_COLOR = "#7c6d5c";

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", {
    style: "currency",
    currency: "BRL",
    maximumFractionDigits: 0,
  });
}

const dailyTotals = computed(() => {
  const daysInMonth = new Date(props.year, props.month, 0).getDate();
  const income = Array.from({ length: daysInMonth }, () => 0);
  const expense = Array.from({ length: daysInMonth }, () => 0);
  for (const tx of props.transactions) {
    const day = Number(tx.date.slice(8, 10));
    if (day < 1 || day > daysInMonth) continue;
    if (tx.type === "income") income[day - 1] += tx.amount;
    else expense[day - 1] += tx.amount;
  }
  return { daysInMonth, income, expense };
});

const chartData = computed(() => ({
  labels: Array.from({ length: dailyTotals.value.daysInMonth }, (_, i) => String(i + 1)),
  datasets: [
    {
      label: "Receitas",
      data: dailyTotals.value.income,
      backgroundColor: INCOME_COLOR,
      borderRadius: 4,
      borderSkipped: "bottom" as const,
      barThickness: "flex" as const,
      maxBarThickness: 12,
      categoryPercentage: 0.7,
      barPercentage: 0.9,
    },
    {
      label: "Despesas",
      data: dailyTotals.value.expense,
      backgroundColor: EXPENSE_COLOR,
      borderRadius: 4,
      borderSkipped: "bottom" as const,
      barThickness: "flex" as const,
      maxBarThickness: 12,
      categoryPercentage: 0.7,
      barPercentage: 0.9,
    },
  ],
}));

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: "index" as const, intersect: false },
  plugins: {
    legend: {
      position: "top" as const,
      align: "end" as const,
      labels: {
        usePointStyle: true,
        pointStyle: "circle" as const,
        boxWidth: 8,
        boxHeight: 8,
        color: TEXT_COLOR,
        font: { family: "Nunito", size: 12, weight: 600 as const },
        padding: 16,
      },
    },
    tooltip: {
      backgroundColor: "#1c1712",
      titleFont: { family: "Nunito", weight: 700 as const },
      bodyFont: { family: "Nunito" },
      padding: 10,
      cornerRadius: 10,
      displayColors: true,
      boxPadding: 4,
      callbacks: {
        label: (ctx: { dataset: { label?: string }; parsed: { y: number | null } }) =>
          `${ctx.dataset.label}: ${formatCurrency(ctx.parsed.y ?? 0)}`,
      },
    },
  },
  scales: {
    x: {
      stacked: false,
      grid: { display: false },
      border: { display: false },
      ticks: { color: TEXT_COLOR, font: { family: "Nunito", size: 10 }, maxTicksLimit: 15 },
    },
    y: {
      grid: { color: GRID_COLOR, drawTicks: false },
      border: { display: false },
      ticks: {
        color: TEXT_COLOR,
        font: { family: "Nunito", size: 10 },
        callback: (value: number | string) => formatCurrency(Number(value)),
        maxTicksLimit: 5,
      },
    },
  },
}));

const hasData = computed(() => props.transactions.length > 0);
</script>

<template>
  <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
    <div class="mb-4">
      <h2 class="font-display text-sm font-bold text-ink-900">Fluxo diário</h2>
      <p class="text-xs text-ink-500">Receitas e despesas por dia no período</p>
    </div>
    <div v-if="hasData" class="h-56 sm:h-64">
      <Bar :data="chartData" :options="chartOptions" />
    </div>
    <p v-else class="flex h-56 items-center justify-center text-sm text-ink-400 sm:h-64">
      Sem dados para exibir neste período
    </p>
  </div>
</template>
