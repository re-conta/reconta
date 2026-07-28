<script setup lang="ts">
import { computed } from "vue";
import { Doughnut } from "vue-chartjs";
import { ArcElement, Chart as ChartJS, DoughnutController, Legend, Tooltip } from "chart.js";
import type { NamedCount } from "../../types/analytics";

ChartJS.register(DoughnutController, ArcElement, Tooltip, Legend);

const props = defineProps<{ title: string; items: NamedCount[] }>();

const COLORS = ["#f2751f", "#d63163", "#7c6d5c", "#f2b705", "#4c6ef5", "#37b24d", "#ae3ec9"];
const TEXT_COLOR = "#7c6d5c";

const chartData = computed(() => ({
  labels: props.items.map((i) => i.name),
  datasets: [
    {
      data: props.items.map((i) => i.visits),
      backgroundColor: props.items.map((_, idx) => COLORS[idx % COLORS.length]),
      borderWidth: 0,
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "right" as const,
      labels: {
        usePointStyle: true,
        pointStyle: "circle" as const,
        boxWidth: 8,
        boxHeight: 8,
        color: TEXT_COLOR,
        font: { family: "Nunito", size: 11, weight: 600 as const },
        padding: 10,
      },
    },
    tooltip: {
      backgroundColor: "#1c1712",
      titleFont: { family: "Nunito", weight: 700 as const },
      bodyFont: { family: "Nunito" },
      padding: 10,
      cornerRadius: 10,
    },
  },
};

const hasData = computed(() => props.items.length > 0);
</script>

<template>
  <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
    <h2 class="mb-3 font-display text-sm font-bold text-ink-900">{{ title }}</h2>
    <div v-if="hasData" class="h-40">
      <Doughnut :data="chartData" :options="chartOptions" />
    </div>
    <p v-else class="flex h-40 items-center justify-center text-sm text-ink-400">Sem dados</p>
  </div>
</template>
