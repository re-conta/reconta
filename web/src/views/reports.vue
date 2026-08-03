<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import CashFlowChart from "../components/charts/CashFlowChart.vue";
import CategoryExpenseChart from "../components/charts/CategoryExpenseChart.vue";
import MonthlyCashFlowChart from "../components/charts/MonthlyCashFlowChart.vue";
import { ApiError } from "../api/reports";
import { listPeriods, listTransactions } from "../api/transactions";
import type { ReportScopeKind } from "../types/report";
import type { Period, Transaction } from "../types/transaction";

const now = new Date();
const route = useRoute();
const router = useRouter();

function initialScope(): ReportScopeKind {
  const q = route.query.escopo;
  if (q === "month" || q === "year" || q === "range" || q === "all") return q;
  return "month";
}

function initialMonth(): number {
  const q = Number(route.query.mes);
  return q >= 1 && q <= 12 ? q : now.getMonth() + 1;
}

function initialYear(): number {
  const q = Number(route.query.ano);
  return q > 0 ? q : now.getFullYear();
}

function initialDateFrom(): string {
  const q = route.query.de;
  return typeof q === "string" && q ? q : now.toISOString().slice(0, 8) + "01";
}

function initialDateTo(): string {
  const q = route.query.ate;
  return typeof q === "string" && q ? q : now.toISOString().slice(0, 10);
}

const hasQueryPeriod = route.query.mes !== undefined || route.query.ano !== undefined;

const scopeKind = ref<ReportScopeKind>(initialScope());
const month = ref(initialMonth());
const year = ref(initialYear());
const dateFrom = ref(initialDateFrom());
const dateTo = ref(initialDateTo());

const periods = ref<Period[]>([]);
const years = computed(() => [...new Set(periods.value.map((p) => p.year))].sort((a, b) => b - a));
const monthsWithData = computed(
  () => new Set(periods.value.filter((p) => p.year === year.value).map((p) => p.month)),
);

const previewTransactions = ref<Transaction[]>([]);
const loadingPreview = ref(false);
const previewError = ref("");

const sortedPeriods = computed(() =>
  [...periods.value].sort((a, b) => a.year - b.year || a.month - b.month),
);

const currentPeriodIndex = computed(() =>
  sortedPeriods.value.findIndex((p) => p.month === month.value && p.year === year.value),
);

const canGoPrevPeriod = computed(() => currentPeriodIndex.value > 0);
const canGoNextPeriod = computed(
  () =>
    currentPeriodIndex.value !== -1 && currentPeriodIndex.value < sortedPeriods.value.length - 1,
);

function goToPrevPeriod() {
  if (!canGoPrevPeriod.value) return;
  const target = sortedPeriods.value[currentPeriodIndex.value - 1];
  month.value = target.month;
  year.value = target.year;
  loadPreview();
}

function goToNextPeriod() {
  if (!canGoNextPeriod.value) return;
  const target = sortedPeriods.value[currentPeriodIndex.value + 1];
  month.value = target.month;
  year.value = target.year;
  loadPreview();
}

const chartCount = computed(() => (scopeKind.value === "month" || scopeKind.value === "year" ? 2 : 1));

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

function buildQuery(): Record<string, string> {
  const query: Record<string, string> = { escopo: scopeKind.value };
  if (scopeKind.value === "month") {
    query.mes = String(month.value);
    query.ano = String(year.value);
  } else if (scopeKind.value === "year") {
    query.ano = String(year.value);
  } else if (scopeKind.value === "range") {
    query.de = dateFrom.value;
    query.ate = dateTo.value;
  }
  return query;
}

watch([scopeKind, month, year, dateFrom, dateTo], () => {
  router.replace({ query: buildQuery() });
});

onMounted(async () => {
  try {
    periods.value = await listPeriods();
    const hasCurrent = periods.value.some((p) => p.month === month.value && p.year === year.value);
    if (!hasQueryPeriod && !hasCurrent && sortedPeriods.value.length > 0) {
      const latest = sortedPeriods.value[sortedPeriods.value.length - 1];
      month.value = latest.month;
      year.value = latest.year;
    }
  } catch {
    // seletor de período funciona com os padrões mesmo sem histórico carregado
  }
  router.replace({ query: buildQuery() });
  await loadPreview();
});
</script>

<template>
  <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 px-2 py-2 md:px-6 md:py-4">
    <div>
      <h2 class="font-display text-xl font-bold text-ink-900">Relatórios</h2>
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
        <div class="flex items-center gap-1">
          <button
            type="button"
            :disabled="!canGoPrevPeriod"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-ink-200 text-sm text-ink-600 transition hover:bg-ink-100 disabled:cursor-not-allowed disabled:opacity-30"
            aria-label="Mês anterior"
            @click="goToPrevPeriod"
          >
            &lsaquo;
          </button>
          <button
            type="button"
            :disabled="!canGoNextPeriod"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-ink-200 text-sm text-ink-600 transition hover:bg-ink-100 disabled:cursor-not-allowed disabled:opacity-30"
            aria-label="Próximo mês"
            @click="goToNextPeriod"
          >
            &rsaquo;
          </button>
        </div>
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

    <div
      v-if="!loadingPreview"
      class="grid grid-cols-1 gap-4"
      :class="{ 'sm:grid-cols-2': chartCount > 1 }"
    >
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
