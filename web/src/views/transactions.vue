<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { Pencil, Trash2 } from "lucide-vue-next";
import { listAccounts } from "../api/accounts";
import { listCategories } from "../api/categories";
import { listTags } from "../api/tags";
import CashFlowChart from "../components/charts/CashFlowChart.vue";
import CategoryExpenseChart from "../components/charts/CategoryExpenseChart.vue";
import TransactionCalendar from "../components/TransactionCalendar.vue";
import {
  ApiError,
  autoCategorize,
  bulkDeleteTransactions,
  bulkUpdateTransactions,
  createTransaction,
  deleteTransaction,
  getOpeningBalance,
  listPeriods,
  listTransactions,
  setOpeningBalance,
  updateTransaction,
} from "../api/transactions";
import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { Tag } from "../types/tag";
import type { Period, Transaction, TransactionInput } from "../types/transaction";

const now = new Date();
const filters = reactive({
  month: now.getMonth() + 1,
  year: now.getFullYear(),
  type: "" as "" | "income" | "expense",
  categoryId: "" as number | "",
  tagId: "" as number | "",
  search: "",
  page: 1,
});

const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const accounts = ref<Account[]>([]);
const periods = ref<Period[]>([]);

const transactions = ref<Transaction[]>([]);
const totals = ref({ income: 0, expense: 0, balance: 0, count: 0 });
const pagination = ref({ page: 1, limit: 50, total: 0 });

const loading = ref(true);
const errorMessage = ref("");
const submitting = ref(false);

const editingId = ref<number | null>(null);
const showForm = ref(false);
const emptyForm = (): TransactionInput => ({
  date: new Date().toISOString().slice(0, 10),
  description: "",
  amount: 0,
  type: "expense",
  categoryId: null,
  accountId: null,
  notes: null,
  tagIds: [],
});
const form = reactive<TransactionInput>(emptyForm());

const selectedIds = ref<Set<number>>(new Set());
const bulkCategoryId = ref<number | "_none" | "">("");

const openingBalance = ref<number | null>(null);
const editingOpeningBalance = ref(false);
const openingBalanceInput = ref(0);

const autoCategorizeMessage = ref("");

const selectedDate = ref<string | null>(null);

const sortedPeriods = computed(() =>
  [...periods.value].sort((a, b) => a.year - b.year || a.month - b.month),
);

const currentPeriodIndex = computed(() =>
  sortedPeriods.value.findIndex((p) => p.month === filters.month && p.year === filters.year),
);

const canGoPrevPeriod = computed(() => currentPeriodIndex.value > 0);
const canGoNextPeriod = computed(
  () => currentPeriodIndex.value !== -1 && currentPeriodIndex.value < sortedPeriods.value.length - 1,
);

function goToPrevPeriod() {
  if (!canGoPrevPeriod.value) return;
  const target = sortedPeriods.value[currentPeriodIndex.value - 1];
  filters.month = target.month;
  filters.year = target.year;
}

function goToNextPeriod() {
  if (!canGoNextPeriod.value) return;
  const target = sortedPeriods.value[currentPeriodIndex.value + 1];
  filters.month = target.month;
  filters.year = target.year;
}

const displayedTransactions = computed(() =>
  selectedDate.value ? transactions.value.filter((tx) => tx.date === selectedDate.value) : transactions.value,
);

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function accountName(accountId: number | null) {
  return accounts.value.find((a) => a.id === accountId)?.name ?? null;
}

async function loadReferenceData() {
  const [c, t, a, p] = await Promise.all([listCategories(), listTags(), listAccounts(), listPeriods()]);
  categories.value = c;
  tags.value = t;
  accounts.value = a;
  periods.value = p;

  if (p.length > 0 && !p.some((period) => period.month === filters.month && period.year === filters.year)) {
    filters.month = p[0].month;
    filters.year = p[0].year;
  }
}

async function loadTransactions() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const result = await listTransactions({
      month: filters.month,
      year: filters.year,
      type: filters.type || undefined,
      categoryId: filters.categoryId || undefined,
      tagId: filters.tagId || undefined,
      search: filters.search || undefined,
      page: filters.page,
      limit: 50,
    });
    transactions.value = result.data;
    totals.value = result.totals;
    pagination.value = result.pagination;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar transações";
  } finally {
    loading.value = false;
  }
}

async function loadOpeningBalance() {
  try {
    const res = await getOpeningBalance(filters.month, filters.year);
    openingBalance.value = res.amount;
  } catch {
    openingBalance.value = null;
  }
}

function resetForm() {
  Object.assign(form, emptyForm());
  editingId.value = null;
  showForm.value = false;
}

function startCreate() {
  resetForm();
  showForm.value = true;
}

function startEdit(tx: Transaction) {
  editingId.value = tx.id;
  form.date = tx.date;
  form.description = tx.description;
  form.amount = tx.amount;
  form.type = tx.type;
  form.categoryId = tx.categoryId;
  form.accountId = tx.accountId;
  form.notes = tx.notes;
  form.tagIds = tx.tags.map((t) => t.id);
  showForm.value = true;
}

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    if (editingId.value) {
      await updateTransaction(editingId.value, { ...form, tagIds: [...form.tagIds] });
    } else {
      await createTransaction({ ...form, tagIds: [...form.tagIds] });
    }
    resetForm();
    await loadTransactions();
    await loadOpeningBalance();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar transação";
  } finally {
    submitting.value = false;
  }
}

async function handleDelete(id: number) {
  if (!confirm("Excluir esta transação?")) return;
  try {
    await deleteTransaction(id);
    selectedIds.value.delete(id);
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir transação";
  }
}

function toggleSelected(id: number) {
  if (selectedIds.value.has(id)) {
    selectedIds.value.delete(id);
  } else {
    selectedIds.value.add(id);
  }
  selectedIds.value = new Set(selectedIds.value);
}

async function applyBulkCategory() {
  if (selectedIds.value.size === 0 || bulkCategoryId.value === "") return;
  try {
    await bulkUpdateTransactions([...selectedIds.value], { categoryId: bulkCategoryId.value });
    selectedIds.value = new Set();
    bulkCategoryId.value = "";
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao editar em lote";
  }
}

async function handleBulkDeleteMonth() {
  if (!confirm(`Excluir todas as transações de ${filters.month}/${filters.year}?`)) return;
  try {
    await bulkDeleteTransactions("month", filters.month, filters.year);
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir transações";
  }
}

async function runAutoCategorize() {
  autoCategorizeMessage.value = "";
  try {
    const res = await autoCategorize();
    autoCategorizeMessage.value = `${res.updated} de ${res.checked} transações categorizadas.`;
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao auto-categorizar";
  }
}

function startEditOpeningBalance() {
  openingBalanceInput.value = openingBalance.value ?? 0;
  editingOpeningBalance.value = true;
}

async function saveOpeningBalance() {
  try {
    const res = await setOpeningBalance(filters.month, filters.year, openingBalanceInput.value);
    openingBalance.value = res.amount;
    editingOpeningBalance.value = false;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar saldo de abertura";
  }
}

const monthLabel = (month: number) => new Date(2000, month - 1, 1).toLocaleDateString("pt-BR", { month: "long" });

const yearOptions = computed(() => {
  if (periods.value.length === 0) return [now.getFullYear()];
  return [...new Set(periods.value.map((p) => p.year))].sort((a, b) => b - a);
});

const monthOptions = computed(() => {
  if (periods.value.length === 0) {
    return Array.from({ length: 12 }, (_, i) => ({ value: i + 1, label: monthLabel(i + 1) }));
  }
  return [...new Set(periods.value.filter((p) => p.year === filters.year).map((p) => p.month))]
    .sort((a, b) => a - b)
    .map((value) => ({ value, label: monthLabel(value) }));
});

watch(
  () => filters.year,
  () => {
    if (!monthOptions.value.some((m) => m.value === filters.month) && monthOptions.value.length > 0) {
      filters.month = monthOptions.value[0].value;
    }
  },
);

const totalPages = computed(() => Math.max(1, Math.ceil(pagination.value.total / pagination.value.limit)));

watch(
  () => [filters.month, filters.year, filters.type, filters.categoryId, filters.tagId, filters.page],
  () => {
    loadTransactions();
  },
);

watch(
  () => [filters.month, filters.year],
  () => {
    loadOpeningBalance();
    editingOpeningBalance.value = false;
    selectedDate.value = null;
  },
);

let searchDebounce: ReturnType<typeof setTimeout> | undefined;
watch(
  () => filters.search,
  () => {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(() => {
      filters.page = 1;
      loadTransactions();
    }, 300);
  },
);

onMounted(async () => {
  await loadReferenceData();
  await loadTransactions();
  await loadOpeningBalance();
});
</script>

<template>
  <div class="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-10 sm:py-14">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Transações</h1>
        <p class="mt-0.5 text-sm text-ink-500">Lançamentos de receitas e despesas</p>
      </div>
      <div class="flex gap-2">
        <button
          type="button"
          class="rounded-full border border-ink-200 bg-white px-4 py-2 text-sm font-semibold text-ink-700 transition hover:bg-ink-100"
          @click="runAutoCategorize"
        >
          Auto-categorizar
        </button>
        <button
          type="button"
          class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          @click="startCreate"
        >
          + Nova transação
        </button>
      </div>
    </div>

    <p v-if="autoCategorizeMessage" class="rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700">
      {{ autoCategorizeMessage }}
    </p>

    <div class="flex flex-col gap-6 md:flex-row md:items-start">
    <!-- Calendário -->
    <div class="order-first md:sticky md:top-20 md:order-2 md:w-64 md:shrink-0 xl:w-72">
      <TransactionCalendar
        :month="filters.month"
        :year="filters.year"
        :transactions="transactions"
        :selected-date="selectedDate"
        :can-go-prev="canGoPrevPeriod"
        :can-go-next="canGoNextPeriod"
        @prev="goToPrevPeriod"
        @next="goToNextPeriod"
        @select-date="(d) => (selectedDate = d)"
      />
    </div>

    <div class="flex min-w-0 flex-1 flex-col gap-6 md:order-1">
    <!-- Filtros -->
    <div class="flex flex-wrap items-end gap-3 rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Mês
        <select v-model.number="filters.month" class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm">
          <option v-for="m in monthOptions" :key="m.value" :value="m.value">{{ m.label }}</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Ano
        <select v-model.number="filters.year" class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm">
          <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Tipo
        <select v-model="filters.type" class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm">
          <option value="">Todos</option>
          <option value="income">Receita</option>
          <option value="expense">Despesa</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Categoria
        <select v-model="filters.categoryId" class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm">
          <option value="">Todas</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </label>
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Tag
        <select v-model="filters.tagId" class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm">
          <option value="">Todas</option>
          <option v-for="t in tags" :key="t.id" :value="t.id">{{ t.name }}</option>
        </select>
      </label>
      <label class="flex flex-1 flex-col gap-1 text-xs font-medium text-ink-600">
        Buscar
        <input
          v-model="filters.search"
          type="text"
          placeholder="Descrição..."
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
        />
      </label>
      <button
        type="button"
        class="rounded-full border border-coral-200 px-3 py-1.5 text-xs font-semibold text-coral-600 transition hover:bg-coral-50"
        @click="handleBulkDeleteMonth"
      >
        Excluir mês
      </button>
    </div>

    <!-- Saldo de abertura -->
    <div class="flex items-center justify-between rounded-3xl border border-ink-200/70 bg-white px-5 py-3 shadow-sm">
      <p class="text-sm text-ink-600">
        Saldo de abertura ({{ filters.month }}/{{ filters.year }}):
        <span class="font-semibold text-ink-900">{{ openingBalance !== null ? formatCurrency(openingBalance) : "-" }}</span>
      </p>
      <div v-if="editingOpeningBalance" class="flex items-center gap-2">
        <input v-model.number="openingBalanceInput" type="number" step="0.01" class="w-32 rounded-lg border border-ink-200 px-2 py-1 text-sm" />
        <button type="button" class="text-xs font-semibold text-brand-700" @click="saveOpeningBalance">Salvar</button>
        <button type="button" class="text-xs text-ink-400" @click="editingOpeningBalance = false">Cancelar</button>
      </div>
      <button v-else type="button" class="text-xs font-semibold text-brand-700 hover:text-brand-800" @click="startEditOpeningBalance">
        Ajustar
      </button>
    </div>

    <!-- Totais -->
    <div class="grid grid-cols-3 gap-3">
      <div class="rounded-2xl border border-ink-200/70 bg-white p-4 text-center shadow-sm">
        <p class="text-xs font-medium text-ink-500">Receitas</p>
        <p class="mt-1 font-display text-lg font-bold text-brand-600">{{ formatCurrency(totals.income) }}</p>
      </div>
      <div class="rounded-2xl border border-ink-200/70 bg-white p-4 text-center shadow-sm">
        <p class="text-xs font-medium text-ink-500">Despesas</p>
        <p class="mt-1 font-display text-lg font-bold text-coral-600">{{ formatCurrency(totals.expense) }}</p>
      </div>
      <div class="rounded-2xl border border-ink-200/70 bg-white p-4 text-center shadow-sm">
        <p class="text-xs font-medium text-ink-500">Saldo</p>
        <p class="mt-1 font-display text-lg font-bold text-ink-900">{{ formatCurrency(totals.balance) }}</p>
      </div>
    </div>

    <!-- Gráficos -->
    <div class="grid gap-6 lg:grid-cols-2">
      <CashFlowChart :month="filters.month" :year="filters.year" :transactions="transactions" />
      <CategoryExpenseChart :transactions="transactions" />
    </div>

    <!-- Formulário -->
    <form
      v-if="showForm"
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleSubmit"
    >
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Data</span>
          <input v-model="form.date" type="date" required class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm" />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Descrição</span>
          <input v-model="form.description" type="text" required class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm" />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Valor</span>
          <input v-model.number="form.amount" type="number" step="0.01" required class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm" />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Tipo</span>
          <select v-model="form.type" class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm">
            <option value="expense">Despesa</option>
            <option value="income">Receita</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Categoria</span>
          <select v-model="form.categoryId" class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm">
            <option :value="null">Sem categoria</option>
            <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Conta</span>
          <select v-model="form.accountId" class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm">
            <option :value="null">Sem conta</option>
            <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Notas</span>
          <input v-model="form.notes" type="text" class="rounded-xl border border-ink-200 px-3.5 py-2.5 text-sm" />
        </label>
      </div>
      <div class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-ink-700">Tags</span>
        <div class="flex flex-wrap gap-2">
          <label
            v-for="t in tags"
            :key="t.id"
            class="flex cursor-pointer items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium"
            :class="form.tagIds.includes(t.id) ? 'border-brand-400 bg-brand-50 text-brand-700' : 'border-ink-200 text-ink-600'"
          >
            <input type="checkbox" class="hidden" :value="t.id" v-model="form.tagIds" />
            {{ t.name }}
          </label>
        </div>
      </div>
      <div class="flex gap-3">
        <button
          type="submit"
          :disabled="submitting"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ submitting ? "Salvando..." : "Salvar" }}
        </button>
        <button type="button" class="rounded-full border border-ink-200 px-4 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-100" @click="resetForm">
          Cancelar
        </button>
      </div>
    </form>

    <!-- Ações em lote -->
    <div v-if="selectedIds.size > 0" class="flex items-center gap-3 rounded-2xl bg-ink-900 px-4 py-3 text-sm text-white">
      <span>{{ selectedIds.size }} selecionada(s)</span>
      <select v-model="bulkCategoryId" class="rounded-lg border border-ink-700 bg-ink-800 px-2 py-1 text-sm">
        <option value="">Definir categoria...</option>
        <option value="_none">Sem categoria</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
      <button type="button" class="rounded-full bg-white px-3 py-1 text-xs font-semibold text-ink-900" @click="applyBulkCategory">
        Aplicar
      </button>
      <button type="button" class="ml-auto text-xs text-ink-300 hover:text-white" @click="selectedIds = new Set()">
        Limpar seleção
      </button>
    </div>

    <!-- Lista -->
    <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">{{ errorMessage }}</p>
      <div v-else-if="displayedTransactions.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">
          {{ selectedDate ? "Nenhuma transação neste dia" : "Nenhuma transação neste período" }}
        </p>
        <p class="text-sm text-ink-400">
          {{ selectedDate ? "Escolha outro dia no calendário." : "Lance a primeira transação para começar." }}
        </p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li v-for="tx in displayedTransactions" :key="tx.id" class="flex items-center gap-3 px-5 py-4 transition hover:bg-ink-50/60">
          <input type="checkbox" :checked="selectedIds.has(tx.id)" @change="toggleSelected(tx.id)" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold text-ink-900">{{ tx.description }}</p>
            <div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-ink-500">
              <span>{{ tx.date }}</span>
              <span v-if="accountName(tx.accountId)">&middot; {{ accountName(tx.accountId) }}</span>
              <span
                v-if="tx.categoryName"
                class="inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-[10px] font-medium"
                :style="{ backgroundColor: `${tx.categoryColor ?? '#94a3b8'}1a`, color: tx.categoryColor ?? '#64748b' }"
              >
                <span class="h-1.5 w-1.5 rounded-full" :style="{ backgroundColor: tx.categoryColor ?? '#94a3b8' }"></span>
                {{ tx.categoryName }}
              </span>
              <span
                v-for="t in tx.tags"
                :key="t.id"
                class="rounded-full px-1.5 py-0.5 text-[10px] font-medium"
                :style="{ backgroundColor: `${t.color}1a`, color: t.color }"
              >
                {{ t.name }}
              </span>
              <span
                v-if="tx.importedFrom"
                class="rounded-full bg-ink-100 px-1.5 py-0.5 text-[10px] text-ink-500"
                :title="tx.pixBeneficiary ? `Beneficiário PIX: ${tx.pixBeneficiary}` : undefined"
              >
                importado{{ tx.bank ? ` · ${tx.bank}` : "" }}
              </span>
            </div>
            <p v-if="tx.notes" class="mt-1 truncate text-xs italic text-ink-400" :title="tx.notes">{{ tx.notes }}</p>
          </div>
          <p class="shrink-0 text-sm font-semibold" :class="tx.type === 'income' ? 'text-brand-600' : 'text-coral-600'">
            {{ tx.type === "income" ? "+" : "-" }}{{ formatCurrency(tx.amount) }}
          </p>
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              class="text-brand-700 hover:text-brand-800"
              title="Editar"
              @click="startEdit(tx)"
            >
              <Pencil class="h-4 w-4" />
            </button>
            <button
              type="button"
              class="text-coral-600 hover:text-coral-700"
              title="Excluir"
              @click="handleDelete(tx.id)"
            >
              <Trash2 class="h-4 w-4" />
            </button>
          </div>
        </li>
      </ul>
      <div v-if="pagination.total > pagination.limit" class="flex items-center justify-between border-t border-ink-100 px-5 py-3 text-sm text-ink-500">
        <button type="button" :disabled="filters.page <= 1" class="disabled:opacity-30" @click="filters.page--">Anterior</button>
        <span>Página {{ pagination.page }} de {{ totalPages }}</span>
        <button type="button" :disabled="filters.page >= totalPages" class="disabled:opacity-30" @click="filters.page++">Próxima</button>
      </div>
    </div>
    </div>
    </div>
  </div>
</template>
