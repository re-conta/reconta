<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { listAccounts } from "../api/accounts";
import {
  acceptShare,
  ApiError,
  cancelShare,
  createShare,
  createSharedTransaction,
  deleteSharedTransaction,
  getShareCategories,
  listReceivedShares,
  listSentShares,
  listSharedTransactions,
  rejectShare,
  updateSharedTransaction,
} from "../api/shares";
import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { Share } from "../types/share";
import type { Transaction, TransactionInput } from "../types/transaction";

type Tab = "sent" | "received";

const activeTab = ref<Tab>("sent");

const tabs: { id: Tab; label: string }[] = [
  { id: "sent", label: "Compartilhamentos" },
  { id: "received", label: "Compartilhado comigo" },
];

function switchTab(tab: Tab) {
  activeTab.value = tab;
}

const errorMessage = ref("");

// --- Compartilhamentos enviados ---

const sentShares = ref<Share[]>([]);
const accounts = ref<Account[]>([]);
const loadingSent = ref(true);
const submitting = ref(false);
const showForm = ref(false);

const form = reactive({
  recipientEmail: "",
  accountIds: [] as number[],
  canEdit: false,
  includeFuture: false,
  periodStart: "",
  periodEnd: "",
});

const statusLabels: Record<Share["status"], string> = {
  pending: "Aguardando resposta",
  accepted: "Aceito",
  rejected: "Rejeitado",
  cancelled: "Cancelado",
};

const statusClasses: Record<Share["status"], string> = {
  pending: "bg-brand-100 text-brand-700",
  accepted: "bg-emerald-100 text-emerald-700",
  rejected: "bg-coral-100 text-coral-700",
  cancelled: "bg-ink-100 text-ink-500",
};

async function loadSent() {
  loadingSent.value = true;
  errorMessage.value = "";
  try {
    const [sharesResult, accountsResult] = await Promise.all([listSentShares(), listAccounts()]);
    sentShares.value = sharesResult;
    accounts.value = accountsResult;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar compartilhamentos";
  } finally {
    loadingSent.value = false;
  }
}

function resetForm() {
  form.recipientEmail = "";
  form.accountIds = [];
  form.canEdit = false;
  form.includeFuture = false;
  form.periodStart = "";
  form.periodEnd = "";
  showForm.value = false;
}

function toggleAccount(id: number) {
  const idx = form.accountIds.indexOf(id);
  if (idx === -1) form.accountIds.push(id);
  else form.accountIds.splice(idx, 1);
}

async function handleSubmit() {
  errorMessage.value = "";
  if (form.accountIds.length === 0) {
    errorMessage.value = "Selecione ao menos uma conta para compartilhar";
    return;
  }
  if (!form.includeFuture && !form.periodEnd) {
    errorMessage.value = "Informe o período final ou habilite o compartilhamento de transações futuras";
    return;
  }

  submitting.value = true;
  try {
    await createShare({
      recipientEmail: form.recipientEmail,
      accountIds: form.accountIds,
      canEdit: form.canEdit,
      includeFuture: form.includeFuture,
      periodStart: form.periodStart || null,
      periodEnd: form.includeFuture ? null : form.periodEnd || null,
    });
    resetForm();
    await loadSent();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao criar compartilhamento";
  } finally {
    submitting.value = false;
  }
}

async function handleCancel(id: number) {
  if (!confirm("Cancelar este compartilhamento? O acesso do convidado será revogado.")) return;
  try {
    await cancelShare(id);
    await loadSent();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao cancelar compartilhamento";
  }
}

function formatDate(value: string | null) {
  if (!value) return "sem limite";
  return new Date(`${value}T00:00:00`).toLocaleDateString("pt-BR");
}

// --- Compartilhado comigo ---

const receivedShares = ref<Share[]>([]);
const loadingReceived = ref(true);
const respondingId = ref<number | null>(null);

const pending = computed(() => receivedShares.value.filter((s) => s.status === "pending"));
const accepted = computed(() => receivedShares.value.filter((s) => s.status === "accepted"));

const selectedShareId = ref<number | null>(null);
const selectedShare = computed(() => accepted.value.find((s) => s.id === selectedShareId.value) ?? null);

const transactions = ref<Transaction[]>([]);
const categories = ref<Category[]>([]);
const loadingTransactions = ref(false);

const showTransactionForm = ref(false);
const editingId = ref<number | null>(null);
const transactionForm = reactive<TransactionInput>({
  date: "",
  description: "",
  amount: 0,
  type: "expense",
  categoryId: null,
  accountId: null,
  notes: null,
  tagIds: [],
});

async function loadReceived() {
  loadingReceived.value = true;
  errorMessage.value = "";
  try {
    receivedShares.value = await listReceivedShares();
    if (!selectedShareId.value && accepted.value.length > 0) {
      selectedShareId.value = accepted.value[0].id;
    }
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar compartilhamentos";
  } finally {
    loadingReceived.value = false;
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
    await loadReceived();
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
    await loadReceived();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao rejeitar convite";
  } finally {
    respondingId.value = null;
  }
}

function resetTransactionForm() {
  transactionForm.date = "";
  transactionForm.description = "";
  transactionForm.amount = 0;
  transactionForm.type = "expense";
  transactionForm.categoryId = null;
  transactionForm.accountId = selectedShare.value?.accountIds[0] ?? null;
  transactionForm.notes = null;
  transactionForm.tagIds = [];
  editingId.value = null;
  showTransactionForm.value = false;
}

function startCreate() {
  resetTransactionForm();
  transactionForm.accountId = selectedShare.value?.accountIds[0] ?? null;
  showTransactionForm.value = true;
}

function startEdit(tx: Transaction) {
  editingId.value = tx.id;
  transactionForm.date = tx.date;
  transactionForm.description = tx.description;
  transactionForm.amount = tx.amount;
  transactionForm.type = tx.type;
  transactionForm.categoryId = tx.categoryId;
  transactionForm.accountId = tx.accountId;
  transactionForm.notes = tx.notes;
  transactionForm.tagIds = tx.tags.map((t) => t.id);
  showTransactionForm.value = true;
}

async function handleTransactionSubmit() {
  if (!selectedShareId.value) return;
  errorMessage.value = "";
  try {
    if (editingId.value) {
      await updateSharedTransaction(selectedShareId.value, editingId.value, { ...transactionForm });
    } else {
      await createSharedTransaction(selectedShareId.value, { ...transactionForm });
    }
    resetTransactionForm();
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

function formatTxDate(value: string) {
  return new Date(`${value}T00:00:00`).toLocaleDateString("pt-BR");
}

watch(selectedShareId, () => {
  resetTransactionForm();
  loadTransactions();
});

onMounted(() => {
  loadSent();
  loadReceived();
});
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div class="flex items-start justify-between">
      <div>
        <h1 class="font-display text-base sm:text-lg md:text-2xl font-bold text-ink-900">Compartilhamentos</h1>
        <p class="mt-0.5 text-xs md:text-sm text-ink-500">
          Compartilhe suas transações e veja o que compartilharam com você
        </p>
      </div>
      <button
        v-if="activeTab === 'sent'"
        type="button"
        class="shrink-0 rounded-full bg-ink-900 px-2 md:px-4 py-2 text-xs md:text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="showForm = !showForm"
      >
        + Novo <span class="hidden sm:inline"> compartilhamento</span>
      </button>
    </div>

    <div class="flex gap-1 rounded-full border border-ink-200/70 bg-white p-1 shadow-sm w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="rounded-full px-2 md:px-4 py-1.5 text-xs md:text-sm font-semibold transition"
        :class="
          activeTab === tab.id
            ? 'bg-ink-900 text-white'
            : 'text-ink-600 hover:bg-ink-100'
        "
        @click="switchTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <p v-if="errorMessage" class="rounded-2xl border border-coral-200 bg-coral-50 p-4 text-sm text-coral-600">
      {{ errorMessage }}
    </p>

    <template v-if="activeTab === 'sent'">
      <form
        v-if="showForm"
        class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
        @submit.prevent="handleSubmit"
      >
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">E-mail do convidado</span>
          <input
            v-model="form.recipientEmail"
            type="email"
            required
            placeholder="pessoa@exemplo.com"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
          <span class="text-xs text-ink-400">A pessoa precisa já ter uma conta no Reconta.</span>
        </label>

        <div class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Contas bancárias</span>
          <div class="flex flex-wrap gap-2">
            <label
              v-for="account in accounts"
              :key="account.id"
              class="flex cursor-pointer items-center gap-2 rounded-xl border border-ink-200 px-3 py-2 text-sm transition"
              :class="form.accountIds.includes(account.id) ? 'border-brand-400 bg-brand-50 text-brand-700' : 'text-ink-700 hover:bg-ink-50'"
            >
              <input
                type="checkbox"
                class="accent-brand-500"
                :checked="form.accountIds.includes(account.id)"
                @change="toggleAccount(account.id)"
              />
              {{ account.name }}
            </label>
          </div>
          <p v-if="accounts.length === 0" class="text-xs text-ink-400">
            Você ainda não tem contas cadastradas.
          </p>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Período inicial (opcional)</span>
            <input
              v-model="form.periodStart"
              type="date"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">
              Período final {{ form.includeFuture ? "(ignorado)" : "" }}
            </span>
            <input
              v-model="form.periodEnd"
              type="date"
              :disabled="form.includeFuture"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100 disabled:opacity-50"
            />
          </label>
        </div>

        <label class="flex items-center gap-2 text-sm text-ink-700">
          <input v-model="form.includeFuture" type="checkbox" class="accent-brand-500" />
          Incluir transações futuras automaticamente
        </label>
        <label class="flex items-center gap-2 text-sm text-ink-700">
          <input v-model="form.canEdit" type="checkbox" class="accent-brand-500" />
          Permitir que o convidado crie, edite e exclua transações
        </label>

        <div class="flex gap-3">
          <button
            type="submit"
            :disabled="submitting"
            class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Enviando..." : "Enviar convite" }}
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
        <div v-if="loadingSent" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
          <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
          Carregando...
        </div>
        <div v-else-if="sentShares.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
          <p class="text-sm font-medium text-ink-600">Nenhum compartilhamento criado ainda</p>
          <p class="text-sm text-ink-400">Convide alguém para ver suas transações.</p>
        </div>
        <ul v-else class="divide-y divide-ink-100">
          <li v-for="share in sentShares" :key="share.id" class="flex flex-col gap-2 px-5 py-4">
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0">
                <p class="truncate text-sm font-semibold text-ink-900">{{ share.recipientName }}</p>
                <p class="truncate text-xs text-ink-500">{{ share.accountNames.join(", ") }}</p>
              </div>
              <div class="flex shrink-0 items-center gap-2">
                <span
                  class="rounded-full px-2.5 py-1 text-xs font-semibold"
                  :class="statusClasses[share.status]"
                >
                  {{ statusLabels[share.status] }}
                </span>
                <button
                  v-if="share.status !== 'cancelled'"
                  type="button"
                  class="text-xs font-semibold text-coral-600 hover:text-coral-700"
                  @click="handleCancel(share.id)"
                >
                  Cancelar
                </button>
              </div>
            </div>
            <p class="text-xs text-ink-400">
              {{ formatDate(share.periodStart) }} até {{ share.includeFuture ? "sempre (inclui futuras)" : formatDate(share.periodEnd) }}
              &middot;
              {{ share.canEdit ? "pode editar" : "somente leitura" }}
            </p>
          </li>
        </ul>
      </div>
    </template>

    <template v-else>
      <div v-if="loadingReceived" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
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
            v-if="showTransactionForm"
            class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
            @submit.prevent="handleTransactionSubmit"
          >
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Data</span>
                <input
                  v-model="transactionForm.date"
                  type="date"
                  required
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                />
              </label>
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Valor</span>
                <input
                  v-model.number="transactionForm.amount"
                  type="number"
                  step="0.01"
                  required
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                />
              </label>
              <label class="flex flex-col gap-1.5 sm:col-span-2">
                <span class="text-sm font-medium text-ink-700">Descrição</span>
                <input
                  v-model="transactionForm.description"
                  type="text"
                  required
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                />
              </label>
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Tipo</span>
                <select
                  v-model="transactionForm.type"
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                >
                  <option value="expense">Despesa</option>
                  <option value="income">Receita</option>
                </select>
              </label>
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Categoria</span>
                <select
                  v-model="transactionForm.categoryId"
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
                @click="resetTransactionForm"
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
                    {{ formatTxDate(tx.date) }}
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
    </template>
  </div>
</template>
