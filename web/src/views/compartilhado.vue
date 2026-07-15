<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  acceptShare,
  ApiError,
  createSharedTransaction,
  deleteSharedTransaction,
  getShareCategories,
  listReceivedShares,
  listSharedTransactions,
  rejectShare,
  updateSharedTransaction,
} from "../api/shares";
import type { Category } from "../types/category";
import type { Share } from "../types/share";
import type { Transaction, TransactionInput } from "../types/transaction";

const shares = ref<Share[]>([]);
const loading = ref(true);
const errorMessage = ref("");
const respondingId = ref<number | null>(null);

const pending = computed(() => shares.value.filter((s) => s.status === "pending"));
const accepted = computed(() => shares.value.filter((s) => s.status === "accepted"));

const selectedShareId = ref<number | null>(null);
const selectedShare = computed(() => accepted.value.find((s) => s.id === selectedShareId.value) ?? null);

const transactions = ref<Transaction[]>([]);
const categories = ref<Category[]>([]);
const loadingTransactions = ref(false);

const showForm = ref(false);
const editingId = ref<number | null>(null);
const form = reactive<TransactionInput>({
  date: "",
  description: "",
  amount: 0,
  type: "expense",
  categoryId: null,
  accountId: null,
  notes: null,
  tagIds: [],
});

async function loadShares() {
  loading.value = true;
  errorMessage.value = "";
  try {
    shares.value = await listReceivedShares();
    if (!selectedShareId.value && accepted.value.length > 0) {
      selectedShareId.value = accepted.value[0].id;
    }
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar compartilhamentos";
  } finally {
    loading.value = false;
  }
}

async function loadTransactions() {
  if (!selectedShareId.value) {
    transactions.value = [];
    return;
  }
  loadingTransactions.value = true;
  try {
    const [result, cats] = await Promise.all([
      listSharedTransactions(selectedShareId.value, { limit: 200 }),
      getShareCategories(selectedShareId.value),
    ]);
    transactions.value = result.data;
    categories.value = cats;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar transações compartilhadas";
  } finally {
    loadingTransactions.value = false;
  }
}

async function handleAccept(id: number) {
  respondingId.value = id;
  try {
    await acceptShare(id);
    await loadShares();
    selectedShareId.value = id;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao aceitar convite";
  } finally {
    respondingId.value = null;
  }
}

async function handleReject(id: number) {
  respondingId.value = id;
  try {
    await rejectShare(id);
    await loadShares();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao rejeitar convite";
  } finally {
    respondingId.value = null;
  }
}

function resetForm() {
  form.date = "";
  form.description = "";
  form.amount = 0;
  form.type = "expense";
  form.categoryId = null;
  form.accountId = selectedShare.value?.accountIds[0] ?? null;
  form.notes = null;
  form.tagIds = [];
  editingId.value = null;
  showForm.value = false;
}

function startCreate() {
  resetForm();
  form.accountId = selectedShare.value?.accountIds[0] ?? null;
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
  if (!selectedShareId.value) return;
  errorMessage.value = "";
  try {
    if (editingId.value) {
      await updateSharedTransaction(selectedShareId.value, editingId.value, { ...form });
    } else {
      await createSharedTransaction(selectedShareId.value, { ...form });
    }
    resetForm();
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar transação";
  }
}

async function handleDelete(tx: Transaction) {
  if (!selectedShareId.value) return;
  if (!confirm("Excluir esta transação?")) return;
  try {
    await deleteSharedTransaction(selectedShareId.value, tx.id);
    await loadTransactions();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir transação";
  }
}

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(value: string) {
  return new Date(`${value}T00:00:00`).toLocaleDateString("pt-BR");
}

watch(selectedShareId, () => {
  resetForm();
  loadTransactions();
});

onMounted(loadShares);
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Compartilhado comigo</h1>
      <p class="mt-0.5 text-sm text-ink-500">Transações que outras pessoas compartilharam com você</p>
    </div>

    <p v-if="errorMessage" class="rounded-2xl border border-coral-200 bg-coral-50 p-4 text-sm text-coral-600">
      {{ errorMessage }}
    </p>

    <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
      <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
      Carregando...
    </div>

    <template v-else>
      <div v-if="pending.length > 0" class="flex flex-col gap-3">
        <h2 class="text-xs font-semibold uppercase tracking-wide text-ink-400">Convites pendentes</h2>
        <div
          v-for="share in pending"
          :key="share.id"
          class="flex items-center justify-between gap-3 rounded-3xl border border-brand-200 bg-brand-50 p-5 shadow-sm"
        >
          <div class="min-w-0">
            <p class="text-sm font-semibold text-ink-900">{{ share.ownerName }} quer compartilhar transações</p>
            <p class="mt-0.5 text-xs text-ink-500">
              {{ share.accountNames.join(", ") }} &middot;
              {{ share.canEdit ? "você poderá editar" : "somente leitura" }}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              :disabled="respondingId === share.id"
              class="rounded-full bg-ink-900 px-3.5 py-2 text-xs font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
              @click="handleAccept(share.id)"
            >
              Aceitar
            </button>
            <button
              type="button"
              :disabled="respondingId === share.id"
              class="rounded-full border border-ink-200 px-3.5 py-2 text-xs font-semibold text-ink-700 transition hover:bg-ink-100 disabled:opacity-50"
              @click="handleReject(share.id)"
            >
              Rejeitar
            </button>
          </div>
        </div>
      </div>

      <div v-if="accepted.length === 0 && pending.length === 0" class="flex flex-col items-center gap-1 rounded-3xl border border-ink-200/70 bg-white p-12 text-center shadow-sm">
        <p class="text-sm font-medium text-ink-600">Ninguém compartilhou transações com você ainda</p>
      </div>

      <div v-if="accepted.length > 0" class="flex flex-col gap-4">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Quem compartilhou</span>
          <select
            v-model.number="selectedShareId"
            class="w-full max-w-sm rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option v-for="share in accepted" :key="share.id" :value="share.id">
              {{ share.ownerName }} ({{ share.accountNames.join(", ") }})
            </option>
          </select>
        </label>

        <div v-if="selectedShare" class="flex items-center justify-between">
          <p class="text-xs text-ink-400">
            {{ selectedShare.canEdit ? "Você pode criar, editar e excluir transações nesta conta." : "Você tem acesso somente leitura." }}
          </p>
          <button
            v-if="selectedShare.canEdit"
            type="button"
            class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
            @click="startCreate"
          >
            + Nova transação
          </button>
        </div>

        <form
          v-if="showForm"
          class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
          @submit.prevent="handleSubmit"
        >
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Data</span>
              <input
                v-model="form.date"
                type="date"
                required
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Valor</span>
              <input
                v-model.number="form.amount"
                type="number"
                step="0.01"
                required
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5 sm:col-span-2">
              <span class="text-sm font-medium text-ink-700">Descrição</span>
              <input
                v-model="form.description"
                type="text"
                required
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Tipo</span>
              <select
                v-model="form.type"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              >
                <option value="expense">Despesa</option>
                <option value="income">Receita</option>
              </select>
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Categoria</span>
              <select
                v-model="form.categoryId"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              >
                <option :value="null">Sem categoria</option>
                <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </label>
          </div>
          <div class="flex gap-3">
            <button
              type="submit"
              class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
            >
              Salvar
            </button>
            <button
              type="button"
              class="rounded-full border border-ink-200 px-4 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-100"
              @click="resetForm"
            >
              Cancelar
            </button>
          </div>
        </form>

        <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
          <div v-if="loadingTransactions" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
            <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
            Carregando...
          </div>
          <div v-else-if="transactions.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
            <p class="text-sm font-medium text-ink-600">Nenhuma transação neste período</p>
          </div>
          <ul v-else class="divide-y divide-ink-100">
            <li
              v-for="tx in transactions"
              :key="tx.id"
              class="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-ink-50/60"
            >
              <div class="min-w-0">
                <p class="truncate text-sm font-semibold text-ink-900">{{ tx.description }}</p>
                <p class="truncate text-xs text-ink-500">
                  {{ formatDate(tx.date) }}
                  <template v-if="tx.categoryName"> &middot; {{ tx.categoryName }}</template>
                </p>
              </div>
              <div class="flex shrink-0 items-center gap-3">
                <span
                  class="text-sm font-semibold"
                  :class="tx.type === 'income' ? 'text-emerald-600' : 'text-coral-600'"
                >
                  {{ tx.type === "income" ? "+" : "-" }}{{ formatCurrency(tx.amount) }}
                </span>
                <template v-if="selectedShare?.canEdit">
                  <button type="button" class="text-xs font-semibold text-brand-700 hover:text-brand-800" @click="startEdit(tx)">
                    Editar
                  </button>
                  <button type="button" class="text-xs font-semibold text-coral-600 hover:text-coral-700" @click="handleDelete(tx)">
                    Excluir
                  </button>
                </template>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </template>
  </div>
</template>
