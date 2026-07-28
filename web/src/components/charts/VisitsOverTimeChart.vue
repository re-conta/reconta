<script setup lang="ts">
import { computed } from "vue";
import { Line } from "vue-chartjs";
import {
  CategoryScale,
  Chart as ChartJS,
  Legend,
  LinearScale,
  LineController,
  LineElement,
  PointElement,
  Tooltip,
} from "chart.js";
import type { DayCount } from "../../types/analytics";

ChartJS.register(
  LineController,
  LineElement,
  PointElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend,
);

const props = defineProps<{ byDay: DayCount[] }>();

const VISITS_COLOR = "#f2751f";
const VISITORS_COLOR = "#d63163";
const GRID_COLOR = "#e6ded5";
const TEXT_COLOR = "#7c6d5c";

function formatDate(iso: string) {
  const [, month, day] = iso.split("-");
  return `${day}/${month}`;
}

const chartData = computed(() => ({
  labels: props.byDay.map((d) => formatDate(d.date)),
  datasets: [
    {
      label: "Visitas",
      data: props.byDay.map((d) => d.visits),
      borderColor: VISITS_COLOR,
      backgroundColor: VISITS_COLOR,
      tension: 0.3,
      pointRadius: 0,
      pointHoverRadius: 4,
      borderWidth: 2,
    },
    {
      label: "Visitantes únicos",
      data: props.byDay.map((d) => d.uniqueVisitors),
      borderColor: VISITORS_COLOR,
      backgroundColor: VISITORS_COLOR,
      tension: 0.3,
      pointRadius: 0,
      pointHoverRadius: 4,
      borderWidth: 2,
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
    },
  },
  scales: {
    x: {
      grid: { display: false },
      border: { display: false },
      ticks: { color: TEXT_COLOR, font: { family: "Nunito", size: 10 }, maxTicksLimit: 15 },
    },
    y: {
      beginAtZero: true,
      grid: { color: GRID_COLOR, drawTicks: false },
      border: { display: false },
      ticks: { color: TEXT_COLOR, font: { family: "Nunito", size: 10 }, maxTicksLimit: 5 },
    },
  },
}));

const hasData = computed(() => props.byDay.length > 0);
</script>

<template>
  <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
    <div class="mb-4">
      <h2 class="font-display text-sm font-bold text-ink-900">Visitas no período</h2>
      <p class="text-xs text-ink-500">Visitas totais e visitantes únicos por dia</p>
    </div>
    <div v-if="hasData" class="h-56 sm:h-64">
      <Line :data="chartData" :options="chartOptions" />
    </div>
    <p v-else class="flex h-56 items-center justify-center text-sm text-ink-400 sm:h-64">
      Sem dados para exibir neste período
    </p>
  </div>
</template>
