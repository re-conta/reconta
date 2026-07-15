<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import CashFlowChart from "../components/charts/CashFlowChart.vue";
import CategoryExpenseChart from "../components/charts/CategoryExpenseChart.vue";
import MonthlyCashFlowChart from "../components/charts/MonthlyCashFlowChart.vue";
import { ApiError } from "../api/reports";
import { listPeriods, listTransactions } from "../api/transactions";
import type { ReportScopeKind } from "../types/report";
import type { Period, Transaction } from "../types/transaction";

const now = new Date();

const scopeKind = ref<ReportScopeKind>("month");
const month = ref(now.getMonth() + 1);
const year = ref(now.getFullYear());
const dateFrom = ref(now.toISOString().slice(0, 8) + "01");
const dateTo = ref(now.toISOString().slice(0, 10));

const periods = ref<Period[]>([]);
const years = computed(() => [...new Set(periods.value.map((p) => p.year))].sort((a, b) => b - a));
const monthsWithData = computed(
  () => new Set(periods.value.filter((p) => p.year === year.value).map((p) => p.month)),
);

const previewTransactions = ref<Transaction[]>([]);
const loadingPreview = ref(false);
const previewError = ref("");

function scopeLabel() {
  if (scopeKind.value === "month") {
    return `${String(month.value).padStart(2, "0")}/${year.value}`;
  }
  if (scopeKind.value === "year") return String(year.value);
  if (scopeKind.value === "range") return `${dateFrom.value} a ${dateTo.value}`;
  return "Todo o período";
}

function inRange(date: string): boolean {
  if (scopeKind.value === "all") return true;
  if (scopeKind.value === "month") {
    const start = `${year.value}-${String(month.value).padStart(2, "0")}-01`;
    const end = `${year.value}-${String(month.value).padStart(2, "0")}-31`;
    return date >= start && date <= end;
  }
  if (scopeKind.value === "year") {
    return date >= `${year.value}-01-01` && date <= `${year.value}-12-31`;
  }
  return date >= dateFrom.value && date <= dateTo.value;
}

async function loadPreview() {
  loadingPreview.value = true;
  previewError.value = "";
  try {
    if (scopeKind.value === "month") {
      const result = await listTransactions({ month: month.value, year: year.value, limit: 5000 });
      previewTransactions.value = result.data;
    } else {
      const result = await listTransactions({ limit: 5000 });
      previewTransactions.value = result.data.filter((tx) => inRange(tx.date));
    }
  } catch (err) {
    previewError.value = err instanceof ApiError ? err.message : "Falha ao carregar transações";
    previewTransactions.value = [];
  } finally {
    loadingPreview.value = false;
  }
}

const totals = computed(() => {
  let income = 0;
  let expense = 0;
  for (const tx of previewTransactions.value) {
    if (tx.type === "income") income += tx.amount;
    else expense += tx.amount;
  }
  return { income, expense, balance: income - expense, count: previewTransactions.value.length };
});

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

onMounted(async () => {
  try {
    periods.value = await listPeriods();
  } catch {
    // seletor de período funciona com os padrões mesmo sem histórico carregado
  }
  await loadPreview();
});
</script>

<template>
  <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Relatórios</h1>
      <p class="mt-0.5 text-sm text-ink-500">
        Acompanhe receitas, despesas e saldo do período escolhido.
      </p>
    </div>

    <div
      class="flex flex-wrap items-end gap-3 rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm"
    >
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Período
        <select
          v-model="scopeKind"
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
          @change="loadPreview"
        >
          <option value="month">Mês</option>
          <option value="year">Ano</option>
          <option value="range">Intervalo personalizado</option>
          <option value="all">Tudo</option>
        </select>
      </label>

      <template v-if="scopeKind === 'month'">
        <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
          Mês
          <select
            v-model.number="month"
            class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
            @change="loadPreview"
          >
            <option v-if="!monthsWithData.has(month)" :value="month">
              {{ String(month).padStart(2, "0") }}
            </option>
            <template v-for="m in 12" :key="m">
              <option v-if="monthsWithData.has(m)" :value="m">
                {{ String(m).padStart(2, "0") }}
              </option>
            </template>
          </select>
        </label>
        <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
          Ano
          <select
            v-model.number="year"
            class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
            @change="loadPreview"
          >
            <option v-if="!years.includes(year)" :value="year">{{ year }}</option>
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </label>
      </template>

      <label
        v-else-if="scopeKind === 'year'"
        class="flex flex-col gap-1 text-xs font-medium text-ink-600"
      >
        Ano
        <select
          v-model.number="year"
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
          @change="loadPreview"
        >
          <option v-if="!years.includes(year)" :value="year">{{ year }}</option>
          <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
        </select>
      </label>

      <template v-else-if="scopeKind === 'range'">
        <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
          De
          <input
            v-model="dateFrom"
            type="date"
            class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
            @change="loadPreview"
          />
        </label>
        <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
          Até
          <input
            v-model="dateTo"
            type="date"
            class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
            @change="loadPreview"
          />
        </label>
      </template>

      <p class="ml-auto text-xs text-ink-500">{{ scopeLabel() }}</p>
    </div>

    <p v-if="previewError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
      {{ previewError }}
    </p>

    <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
        <p class="text-xs text-ink-500">Receitas</p>
        <p class="font-display text-lg font-bold text-brand-700">
          {{ formatCurrency(totals.income) }}
        </p>
      </div>
      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
        <p class="text-xs text-ink-500">Despesas</p>
        <p class="font-display text-lg font-bold text-coral-700">
          {{ formatCurrency(totals.expense) }}
        </p>
      </div>
      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
        <p class="text-xs text-ink-500">Saldo</p>
        <p class="font-display text-lg font-bold text-ink-900">
          {{ formatCurrency(totals.balance) }}
        </p>
      </div>
      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
        <p class="text-xs text-ink-500">Lançamentos</p>
        <p class="font-display text-lg font-bold text-ink-900">{{ totals.count }}</p>
      </div>
    </div>

    <div v-if="!loadingPreview" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <CashFlowChart
        v-if="scopeKind === 'month'"
        :month="month"
        :year="year"
        :transactions="previewTransactions"
      />
      <MonthlyCashFlowChart v-else-if="scopeKind === 'year'" :transactions="previewTransactions" />
      <CategoryExpenseChart :transactions="previewTransactions" />
    </div>
  </div>
</template>
