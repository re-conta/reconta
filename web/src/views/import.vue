<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listAccounts } from "../api/accounts";
import { listCategories } from "../api/categories";
import { ApiError, confirmStatementImport, previewStatementImport } from "../api/statement";
import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { Bank, ParsedTransaction } from "../types/statement";

const banks: Bank[] = [
  { key: "", label: "Detectar automaticamente" },
  { key: "bb", label: "Banco do Brasil" },
  { key: "sicredi", label: "Sicredi" },
  { key: "nubank", label: "Nubank" },
  { key: "mercadopago", label: "Mercado Pago" },
  { key: "itau", label: "Itaú" },
  { key: "generic", label: "Outro / genérico" },
];

const accounts = ref<Account[]>([]);
const categories = ref<Category[]>([]);

const selectedFile = ref<File | null>(null);
const selectedBank = ref("");
const accountId = ref<number | "">("");

const analyzing = ref(false);
const importing = ref(false);
const errorMessage = ref("");
const resultMessage = ref("");

const bankLabel = ref("");
const bankKey = ref("");
const rows = ref<(ParsedTransaction & { include: boolean })[]>([]);

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  selectedFile.value = input.files?.[0] ?? null;
}

async function handleAnalyze() {
  if (!selectedFile.value) {
    errorMessage.value = "Selecione um arquivo PDF do extrato";
    return;
  }
  errorMessage.value = "";
  resultMessage.value = "";
  analyzing.value = true;
  try {
    const preview = await previewStatementImport(
      selectedFile.value,
      selectedBank.value || undefined,
    );
    bankKey.value = preview.bank;
    bankLabel.value = preview.bankLabel;
    rows.value = preview.transactions.map((t) => ({ ...t, include: !t.duplicate }));
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao analisar o extrato";
    rows.value = [];
  } finally {
    analyzing.value = false;
  }
}

async function handleImport() {
  const selected = rows.value.filter((r) => r.include);
  if (selected.length === 0) {
    errorMessage.value = "Selecione ao menos um lançamento para importar";
    return;
  }
  errorMessage.value = "";
  resultMessage.value = "";
  importing.value = true;
  try {
    const res = await confirmStatementImport(
      bankKey.value,
      accountId.value === "" ? null : accountId.value,
      selected.map((r) => ({
        date: r.date,
        description: r.description,
        amount: r.amount,
        type: r.type,
        categoryId: r.categoryId ?? null,
        pixBeneficiary: r.pixBeneficiary ?? null,
      })),
    );
    resultMessage.value = `${res.imported} de ${res.total} lançamentos importados.`;
    rows.value = rows.value.filter((r) => !r.include);
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao importar lançamentos";
  } finally {
    importing.value = false;
  }
}

function toggleAll(value: boolean) {
  rows.value = rows.value.map((r) => ({ ...r, include: value }));
}

onMounted(async () => {
  try {
    const [a, c] = await Promise.all([listAccounts(), listCategories()]);
    accounts.value = a;
    categories.value = c;
  } catch {
    // filtros de conta/categoria são opcionais; segue sem eles em caso de falha
  }
});
</script>

<template>
  <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Importar extrato</h1>
      <p class="mt-0.5 text-sm text-ink-500">
        Envie o PDF do extrato do banco para reconhecer os lançamentos automaticamente. Suporta
        Banco do Brasil, Sicredi, Nubank, Mercado Pago, Itaú e outros formatos genéricos.
      </p>
    </div>

    <p v-if="errorMessage" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
      {{ errorMessage }}
    </p>
    <p v-if="resultMessage" class="rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700">
      {{ resultMessage }}
    </p>

    <div
      class="flex flex-wrap items-end gap-3 rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm"
    >
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Arquivo PDF
        <input
          type="file"
          accept="application/pdf"
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
          @change="handleFileChange"
        />
      </label>
      <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
        Banco
        <select
          v-model="selectedBank"
          class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
        >
          <option v-for="b in banks" :key="b.key" :value="b.key">{{ b.label }}</option>
        </select>
      </label>
      <button
        type="button"
        :disabled="analyzing || !selectedFile"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        @click="handleAnalyze"
      >
        {{ analyzing ? "Analisando..." : "Analisar extrato" }}
      </button>
    </div>

    <template v-if="rows.length > 0">
      <div
        class="flex flex-wrap items-center justify-between gap-3 rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm"
      >
        <p class="text-sm text-ink-600">
          Banco identificado:
          <span class="font-semibold text-ink-900">{{ bankLabel }}</span> &middot;
          {{ rows.length }} lançamento(s) encontrado(s)
        </p>
        <label class="flex flex-col gap-1 text-xs font-medium text-ink-600">
          Conta de destino
          <select
            v-model="accountId"
            class="rounded-lg border border-ink-200 px-2.5 py-1.5 text-sm"
          >
            <option value="">Sem conta</option>
            <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
          </select>
        </label>
      </div>

      <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
        <div class="flex items-center gap-3 border-b border-ink-100 px-5 py-3 text-xs text-ink-500">
          <button type="button" class="font-semibold text-brand-700" @click="toggleAll(true)">
            Marcar todos
          </button>
          <button type="button" class="font-semibold text-ink-500" @click="toggleAll(false)">
            Desmarcar todos
          </button>
        </div>
        <ul class="divide-y divide-ink-100">
          <li
            v-for="(row, i) in rows"
            :key="i"
            class="flex items-center gap-3 px-5 py-4 transition hover:bg-ink-50/60"
            :class="row.duplicate ? 'bg-amber-50/50' : ''"
          >
            <input type="checkbox" v-model="row.include" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-semibold text-ink-900">{{ row.description }}</p>
              <p class="truncate text-xs text-ink-500">
                {{ row.date }}
                <span v-if="row.pixBeneficiary"> &middot; PIX: {{ row.pixBeneficiary }}</span>
                <span
                  v-if="row.duplicate"
                  class="ml-1 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-semibold text-amber-700"
                >
                  possível duplicata
                </span>
              </p>
            </div>
            <select v-model="row.type" class="rounded-lg border border-ink-200 px-2 py-1 text-xs">
              <option value="income">Receita</option>
              <option value="expense">Despesa</option>
            </select>
            <select
              v-model="row.categoryId"
              class="rounded-lg border border-ink-200 px-2 py-1 text-xs"
            >
              <option :value="null">Sem categoria</option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <input
              v-model.number="row.amount"
              type="number"
              step="0.01"
              class="w-24 rounded-lg border border-ink-200 px-2 py-1 text-right text-xs"
            />
          </li>
        </ul>
      </div>

      <div class="flex justify-end">
        <button
          type="button"
          :disabled="importing"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
          @click="handleImport"
        >
          {{
            importing
              ? "Importando..."
              : `Importar ${rows.filter((r) => r.include).length} selecionado(s)`
          }}
        </button>
      </div>
    </template>
  </div>
</template>
