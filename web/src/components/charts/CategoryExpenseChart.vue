<script setup lang="ts">
import { computed } from "vue";
import { Bar } from "vue-chartjs";
import {
  BarController,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  LinearScale,
  Tooltip,
} from "chart.js";
import type { Transaction } from "../../types/transaction";

ChartJS.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip);

const props = defineProps<{
  transactions: Transaction[];
}>();

const OTHER_COLOR = "#a3937f";
const TEXT_COLOR = "#7c6d5c";
const MAX_SLICES = 7;

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

interface Slice {
  name: string;
  color: string;
  total: number;
}

const slices = computed<Slice[]>(() => {
  const byCategory = new Map<string, Slice>();
  for (const tx of props.transactions) {
    if (tx.type !== "expense") continue;
    const key = tx.categoryName ?? "Sem categoria";
    const entry = byCategory.get(key) ?? { name: key, color: tx.categoryColor ?? "#94a3b8", total: 0 };
    entry.total += tx.amount;
    byCategory.set(key, entry);
  }
  const sorted = [...byCategory.values()].sort((a, b) => b.total - a.total);
  if (sorted.length <= MAX_SLICES) return sorted;

  const top = sorted.slice(0, MAX_SLICES - 1);
  const rest = sorted.slice(MAX_SLICES - 1);
  const otherTotal = rest.reduce((sum, s) => sum + s.total, 0);
  top.push({ name: "Outras", color: OTHER_COLOR, total: otherTotal });
  return top;
});

const chartData = computed(() => ({
  labels: slices.value.map((s) => s.name),
  datasets: [
    {
      data: slices.value.map((s) => s.total),
      backgroundColor: slices.value.map((s) => s.color),
      borderRadius: 4,
      borderSkipped: "left" as const,
      barThickness: "flex" as const,
      maxBarThickness: 18,
      categoryPercentage: 0.7,
      barPercentage: 0.9,
    },
  ],
}));

const chartOptions = computed(() => ({
  indexAxis: "y" as const,
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: "#1c1712",
      titleFont: { family: "Nunito", weight: 700 as const },
      bodyFont: { family: "Nunito" },
      padding: 10,
      cornerRadius: 10,
      displayColors: true,
      boxPadding: 4,
      callbacks: {
        label: (ctx: { parsed: { x: number | null } }) => formatCurrency(ctx.parsed.x ?? 0),
      },
    },
  },
  scales: {
    x: {
      grid: { display: false },
      border: { display: false },
      ticks: {
        color: TEXT_COLOR,
        font: { family: "Nunito", size: 10 },
        callback: (value: number | string) => formatCurrency(Number(value)),
        maxTicksLimit: 4,
      },
    },
    y: {
      grid: { display: false },
      border: { display: false },
      ticks: { color: TEXT_COLOR, font: { family: "Nunito", size: 12, weight: 600 as const } },
    },
  },
}));

const hasData = computed(() => slices.value.length > 0);
const chartHeight = computed(() => Math.max(160, slices.value.length * 36));
</script>

<template>
  <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
    <div class="mb-4">
      <h2 class="font-display text-sm font-bold text-ink-900">Despesas por categoria</h2>
      <p class="text-xs text-ink-500">Ranking do período selecionado</p>
    </div>
    <div v-if="hasData" :style="{ height: `${chartHeight}px` }">
      <Bar :data="chartData" :options="chartOptions" />
    </div>
    <p v-else class="flex h-40 items-center justify-center text-sm text-ink-400">
      Sem despesas categorizadas neste período
    </p>
  </div>
</template>
