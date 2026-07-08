<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import {
  ApiError,
  createAccount,
  deleteAccount,
  listAccounts,
  updateAccount,
} from "../api/accounts";
import type { Account, AccountInput } from "../types/account";

const accounts = ref<Account[]>([]);
const errorMessage = ref("");
const loading = ref(true);
const submitting = ref(false);

const editingId = ref<number | null>(null);
const showForm = ref(false);
const form = reactive<AccountInput>({ name: "", type: "checking", balance: 0 });

const accountTypes = [
  { value: "checking", label: "Conta corrente" },
  { value: "savings", label: "Poupança" },
  { value: "credit", label: "Cartão de crédito" },
  { value: "investment", label: "Investimento" },
];

async function loadAccounts() {
  loading.value = true;
  errorMessage.value = "";
  try {
    accounts.value = await listAccounts();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar contas";
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  form.name = "";
  form.type = "checking";
  form.balance = 0;
  editingId.value = null;
  showForm.value = false;
}

function startCreate() {
  resetForm();
  showForm.value = true;
}

function startEdit(account: Account) {
  editingId.value = account.id;
  form.name = account.name;
  form.type = account.type;
  form.balance = account.balance;
  showForm.value = true;
}

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    if (editingId.value) {
      await updateAccount(editingId.value, { ...form });
    } else {
      await createAccount({ ...form });
    }
    resetForm();
    await loadAccounts();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar conta";
  } finally {
    submitting.value = false;
  }
}

async function handleDelete(id: number) {
  if (!confirm("Excluir esta conta?")) return;
  try {
    await deleteAccount(id);
    await loadAccounts();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir conta";
  }
}

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

onMounted(loadAccounts);
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Contas</h1>
        <p class="mt-0.5 text-sm text-ink-500">Contas bancárias, cartões e investimentos</p>
      </div>
      <button
        type="button"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="startCreate"
      >
        + Nova conta
      </button>
    </div>

    <form
      v-if="showForm"
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleSubmit"
    >
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Nome</span>
          <input
            v-model="form.name"
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
            <option v-for="t in accountTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Saldo</span>
          <input
            v-model.number="form.balance"
            type="number"
            step="0.01"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>
      <div class="flex gap-3">
        <button
          type="submit"
          :disabled="submitting"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ submitting ? "Salvando..." : "Salvar" }}
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
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span
          class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">
        {{ errorMessage }}
      </p>
      <div
        v-else-if="accounts.length === 0"
        class="flex flex-col items-center gap-1 p-12 text-center"
      >
        <p class="text-sm font-medium text-ink-600">Nenhuma conta cadastrada ainda</p>
        <p class="text-sm text-ink-400">Crie a primeira conta para começar.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li
          v-for="account in accounts"
          :key="account.id"
          class="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-ink-50/60"
        >
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-ink-900">{{ account.name }}</p>
            <p class="truncate text-xs text-ink-500">
              {{ accountTypes.find((t) => t.value === account.type)?.label ?? account.type }}
              &middot;
              {{ formatCurrency(account.balance) }}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              class="text-xs font-semibold text-brand-700 hover:text-brand-800"
              @click="startEdit(account)"
            >
              Editar
            </button>
            <button
              type="button"
              class="text-xs font-semibold text-coral-600 hover:text-coral-700"
              @click="handleDelete(account.id)"
            >
              Excluir
            </button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
