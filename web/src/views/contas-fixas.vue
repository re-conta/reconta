<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { Bell, PauseCircle, Pencil, PlayCircle, StopCircle, Trash2 } from "lucide-vue-next";
import { listAccounts } from "../api/accounts";
import { listCategories } from "../api/categories";
import {
  ApiError,
  closeFixedBill,
  deleteFixedBill,
  freezeFixedBill,
  listFixedBills,
  reactivateFixedBill,
} from "../api/fixedBills";
import { useNotifications } from "../composables/useNotifications";
import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { FixedBill, PayFixedBillResult } from "../types/fixedBill";
import { PERIODICITY_LABELS, STATUS_LABELS } from "../types/fixedBill";
import FixedBillForm from "../components/FixedBillForm.vue";
import PayBillModal from "../components/PayBillModal.vue";

const router = useRouter();
const { unreadCount } = useNotifications();

const bills = ref<FixedBill[]>([]);
const categories = ref<Category[]>([]);
const accounts = ref<Account[]>([]);
const loading = ref(true);
const errorMessage = ref("");

const editingBill = ref<FixedBill | null>(null);
const showForm = ref(false);
const payingBill = ref<FixedBill | null>(null);

const statusOrder: Record<string, number> = { active: 0, frozen: 1, closed: 2 };
const sortedBills = computed(() =>
  [...bills.value].sort((a, b) => statusOrder[a.status] - statusOrder[b.status]),
);

async function loadAll() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const [billsData, categoriesData, accountsData] = await Promise.all([
      listFixedBills(),
      listCategories(),
      listAccounts(),
    ]);
    bills.value = billsData;
    categories.value = categoriesData;
    accounts.value = accountsData;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar contas fixas";
  } finally {
    loading.value = false;
  }
}

function startCreate() {
  editingBill.value = null;
  showForm.value = true;
}

function startEdit(bill: FixedBill) {
  editingBill.value = bill;
  showForm.value = true;
}

function onSaved(bill: FixedBill) {
  const idx = bills.value.findIndex((b) => b.id === bill.id);
  if (idx >= 0) bills.value[idx] = bill;
  else bills.value.push(bill);
  showForm.value = false;
  editingBill.value = null;
}

function onPaid(result: PayFixedBillResult) {
  const idx = bills.value.findIndex((b) => b.id === result.bill.id);
  if (idx >= 0) bills.value[idx] = result.bill;
  payingBill.value = null;
}

async function handleDelete(bill: FixedBill) {
  if (!confirm(`Excluir a conta fixa "${bill.name}"? O histórico de pagamentos será perdido.`)) return;
  try {
    await deleteFixedBill(bill.id);
    bills.value = bills.value.filter((b) => b.id !== bill.id);
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir a conta fixa";
  }
}

async function handleFreeze(bill: FixedBill) {
  await runStatusAction(() => freezeFixedBill(bill.id));
}
async function handleReactivate(bill: FixedBill) {
  await runStatusAction(() => reactivateFixedBill(bill.id));
}
async function handleClose(bill: FixedBill) {
  if (!confirm(`Encerrar definitivamente "${bill.name}"?`)) return;
  await runStatusAction(() => closeFixedBill(bill.id));
}

async function runStatusAction(action: () => Promise<FixedBill>) {
  try {
    const updated = await action();
    const idx = bills.value.findIndex((b) => b.id === updated.id);
    if (idx >= 0) bills.value[idx] = updated;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao atualizar a conta fixa";
  }
}

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}

function formatDate(value: string) {
  const [year, month, day] = value.split("-");
  return `${day}/${month}/${year}`;
}

function isOverdue(bill: FixedBill) {
  if (bill.status !== "active") return false;
  return bill.dueDate < new Date().toISOString().slice(0, 10);
}

function statusBadgeClass(status: string) {
  if (status === "frozen") return "bg-brand-100 text-brand-700";
  if (status === "closed") return "bg-ink-200 text-ink-600";
  return "bg-emerald-100 text-emerald-700";
}

onMounted(loadAll);
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-6 px-2 md:px-4 py-4 md:py-8">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Contas Fixas</h1>
        <p class="mt-0.5 text-sm text-ink-500">Despesas recorrentes e seus vencimentos</p>
      </div>
      <div class="flex items-center gap-3">
        <button
          type="button"
          class="relative rounded-full border border-ink-200 bg-white p-2.5 text-ink-700 shadow-sm transition hover:bg-ink-50"
          aria-label="Ver notificações"
          @click="router.push('/notificacoes')"
        >
          <Bell class="h-5 w-5" />
          <span
            v-if="unreadCount > 0"
            class="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-coral-500 px-1 text-[11px] font-bold text-white"
          >
            {{ unreadCount > 99 ? "99+" : unreadCount }}
          </span>
        </button>
        <button
          type="button"
          class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          @click="startCreate"
        >
          + Nova conta fixa
        </button>
      </div>
    </div>

    <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">{{ errorMessage }}</p>
      <div v-else-if="bills.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">Nenhuma conta fixa cadastrada ainda</p>
        <p class="text-sm text-ink-400">Cadastre luz, internet, aluguel e outras despesas recorrentes.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li
          v-for="bill in sortedBills"
          :key="bill.id"
          class="flex flex-col gap-3 px-5 py-4 transition hover:bg-ink-50/60 sm:flex-row sm:items-center sm:justify-between"
        >
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <p class="truncate text-sm font-semibold text-ink-900">{{ bill.name }}</p>
              <span class="rounded-full px-2 py-0.5 text-[11px] font-semibold" :class="statusBadgeClass(bill.status)">
                {{ STATUS_LABELS[bill.status] }}
              </span>
              <span v-if="isOverdue(bill)" class="rounded-full bg-coral-100 px-2 py-0.5 text-[11px] font-semibold text-coral-700">
                Vencida
              </span>
            </div>
            <p class="mt-0.5 truncate text-xs text-ink-500">
              {{ formatCurrency(bill.amount) }} &middot; {{ PERIODICITY_LABELS[bill.periodicity] }} &middot;
              vence em {{ formatDate(bill.dueDate) }}
              <template v-if="bill.categoryName"> &middot; {{ bill.categoryName }}</template>
            </p>
          </div>

          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <button
              v-if="bill.status === 'active'"
              type="button"
              class="rounded-full bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-emerald-700"
              @click="payingBill = bill"
            >
              Marcar como paga
            </button>
            <button
              type="button"
              class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
              title="Editar"
              @click="startEdit(bill)"
            >
              <Pencil class="h-4 w-4" />
            </button>
            <button
              v-if="bill.status === 'active'"
              type="button"
              class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
              title="Congelar"
              @click="handleFreeze(bill)"
            >
              <PauseCircle class="h-4 w-4" />
            </button>
            <button
              v-if="bill.status !== 'active'"
              type="button"
              class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
              title="Reativar"
              @click="handleReactivate(bill)"
            >
              <PlayCircle class="h-4 w-4" />
            </button>
            <button
              v-if="bill.status !== 'closed'"
              type="button"
              class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
              title="Encerrar"
              @click="handleClose(bill)"
            >
              <StopCircle class="h-4 w-4" />
            </button>
            <button
              type="button"
              class="rounded-full p-1.5 text-coral-500 transition hover:bg-coral-50"
              title="Excluir"
              @click="handleDelete(bill)"
            >
              <Trash2 class="h-4 w-4" />
            </button>
          </div>
        </li>
      </ul>
    </div>

    <FixedBillForm
      v-if="showForm"
      :bill="editingBill"
      :categories="categories"
      :accounts="accounts"
      @saved="onSaved"
      @cancel="showForm = false"
    />
    <PayBillModal
      v-if="payingBill"
      :bill="payingBill"
      :accounts="accounts"
      @paid="onPaid"
      @cancel="payingBill = null"
    />
  </div>
</template>
