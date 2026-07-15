<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import CashFlowChart from "../components/charts/CashFlowChart.vue";
import CategoryExpenseChart from "../components/charts/CategoryExpenseChart.vue";
import { ApiError, downloadBlob, exportReport, importBackup } from "../api/reports";
import { listPeriods, listTransactions } from "../api/transactions";
import type { ChartImagePayload, ExportFormat, ReportScopeKind } from "../types/report";
import type { Period, Transaction } from "../types/transaction";

const now = new Date();

const scopeKind = ref<ReportScopeKind>("month");
const month = ref(now.getMonth() + 1);
const year = ref(now.getFullYear());
const dateFrom = ref(now.toISOString().slice(0, 8) + "01");
const dateTo = ref(now.toISOString().slice(0, 10));

const periods = ref<Period[]>([]);
const years = computed(() => [...new Set(periods.value.map((p) => p.year))].sort((a, b) => b - a));

const previewTransactions = ref<Transaction[]>([]);
const loadingPreview = ref(false);
const previewError = ref("");

const exportingFormat = ref<ExportFormat | null>(null);
const exportError = ref("");

const importing = ref(false);
const importMessage = ref("");
const importError = ref("");
const fileInput = ref<HTMLInputElement | null>(null);

const cashFlowRef = ref<InstanceType<typeof CashFlowChart>>();
const categoryChartRef = ref<InstanceType<typeof CategoryExpenseChart>>();

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

function toBase64(dataUrl: string | undefined): string | null {
  if (!dataUrl) return null;
  const parts = dataUrl.split(",");
  return parts.length > 1 ? parts[1] : null;
}

async function collectCharts(): Promise<ChartImagePayload[]> {
  const charts: ChartImagePayload[] = [];
  if (scopeKind.value === "month") {
    const img = toBase64(cashFlowRef.value?.toImage());
    if (img) charts.push({ title: "Fluxo diário", pngBase64: img });
  }
  const catImg = toBase64(categoryChartRef.value?.toImage());
  if (catImg) charts.push({ title: "Despesas por categoria", pngBase64: catImg });
  return charts;
}

async function handleExport(format: ExportFormat) {
  exportError.value = "";
  exportingFormat.value = format;
  try {
    const charts = format === "json" ? [] : await collectCharts();
    const { blob, filename } = await exportReport(
      format,
      {
        scope: scopeKind.value,
        month: month.value,
        year: year.value,
        dateFrom: dateFrom.value,
        dateTo: dateTo.value,
      },
      charts,
    );
    downloadBlob(blob, filename);
  } catch (err) {
    exportError.value = err instanceof ApiError ? err.message : "Falha ao gerar o relatório";
  } finally {
    exportingFormat.value = null;
  }
}

async function handleImport() {
  const file = fileInput.value?.files?.[0];
  if (!file) {
    importError.value = "Selecione um arquivo JSON de backup";
    return;
  }
  importError.value = "";
  importMessage.value = "";
  importing.value = true;
  try {
    const result = await importBackup(file);
    importMessage.value = `${result.imported} importado(s), ${result.skipped} ignorado(s) por duplicidade (de ${result.total}).`;
    if (fileInput.value) fileInput.value.value = "";
    await loadPreview();
  } catch (err) {
    importError.value = err instanceof ApiError ? err.message : "Falha ao importar backup";
  } finally {
    importing.value = false;
  }
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
      <h1 class="font-display text-2xl font-bold text-ink-900">Exportar</h1>
      <p class="mt-0.5 text-sm text-ink-500">
        Exporte um relatório de gastos do período escolhido em ODS, XLSX, PDF ou JSON, ou restaure
        um backup em JSON exportado anteriormente.
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
            <option v-for="m in 12" :key="m" :value="m">{{ String(m).padStart(2, "0") }}</option>
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

    <div v-if="!loadingPreview" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <CashFlowChart
        v-if="scopeKind === 'month'"
        ref="cashFlowRef"
        :month="month"
        :year="year"
        :transactions="previewTransactions"
      />
      <CategoryExpenseChart ref="categoryChartRef" :transactions="previewTransactions" />
    </div>

    <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
      <h2 class="font-display text-sm font-bold text-ink-900">Exportar relatório</h2>
      <p class="mt-0.5 text-xs text-ink-500">
        Baixe os lançamentos e gráficos do período no formato desejado.
      </p>
      <p v-if="exportError" class="mt-3 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ exportError }}
      </p>
      <div class="mt-4 flex flex-wrap gap-2">
        <button
          v-for="format in ['xlsx', 'ods', 'pdf', 'json'] as ExportFormat[]"
          :key="format"
          type="button"
          :disabled="exportingFormat !== null"
          class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
          @click="handleExport(format)"
        >
          {{ exportingFormat === format ? "Gerando..." : format.toUpperCase() }}
        </button>
      </div>
    </div>

    <div class="rounded-3xl border border-ink-200/70 bg-white p-5 shadow-sm">
      <h2 class="font-display text-sm font-bold text-ink-900">Importar backup</h2>
      <p class="mt-0.5 text-xs text-ink-500">
        Restaure lançamentos a partir de um arquivo JSON exportado anteriormente. Lançamentos
        duplicados (mesma data, descrição e valor) são ignorados automaticamente.
      </p>
      <p v-if="importError" class="mt-3 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ importError }}
      </p>
      <p v-if="importMessage" class="mt-3 rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700">
        {{ importMessage }}
      </p>
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <input
          ref="fileInput"
          type="file"
          accept="application/json"
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
        />
        <button
          type="button"
          :disabled="importing"
          class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
          @click="handleImport"
        >
          {{ importing ? "Importando..." : "Importar" }}
        </button>
      </div>
    </div>
  </div>
</template>
