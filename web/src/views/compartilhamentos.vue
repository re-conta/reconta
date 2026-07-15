<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { listAccounts } from "../api/accounts";
import { ApiError, cancelShare, createShare, listSentShares } from "../api/shares";
import type { Account } from "../types/account";
import type { Share } from "../types/share";

const shares = ref<Share[]>([]);
const accounts = ref<Account[]>([]);
const loading = ref(true);
const errorMessage = ref("");
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

async function loadAll() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const [sharesResult, accountsResult] = await Promise.all([listSentShares(), listAccounts()]);
    shares.value = sharesResult;
    accounts.value = accountsResult;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar compartilhamentos";
  } finally {
    loading.value = false;
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
    await loadAll();
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
    await loadAll();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao cancelar compartilhamento";
  }
}

function formatDate(value: string | null) {
  if (!value) return "sem limite";
  return new Date(`${value}T00:00:00`).toLocaleDateString("pt-BR");
}

onMounted(loadAll);
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Compartilhamentos</h1>
        <p class="mt-0.5 text-sm text-ink-500">
          Compartilhe suas transações e relatórios com outros usuários cadastrados
        </p>
      </div>
      <button
        type="button"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="showForm = !showForm"
      >
        + Novo compartilhamento
      </button>
    </div>

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

    <p v-if="errorMessage" class="rounded-2xl border border-coral-200 bg-coral-50 p-4 text-sm text-coral-600">
      {{ errorMessage }}
    </p>

    <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
        Carregando...
      </div>
      <div v-else-if="shares.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">Nenhum compartilhamento criado ainda</p>
        <p class="text-sm text-ink-400">Convide alguém para ver suas transações.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li v-for="share in shares" :key="share.id" class="flex flex-col gap-2 px-5 py-4">
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
  </div>
</template>
