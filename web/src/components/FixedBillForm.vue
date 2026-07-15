<script setup lang="ts">
import { reactive, ref, watch } from "vue";
import { X } from "lucide-vue-next";
import { ApiError, createFixedBill, updateFixedBill } from "../api/fixedBills";
import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { FixedBill, FixedBillInput, FixedBillPeriodicity } from "../types/fixedBill";
import { PERIODICITY_LABELS } from "../types/fixedBill";

const props = defineProps<{
  bill?: FixedBill | null;
  categories: Category[];
  accounts: Account[];
}>();

const emit = defineEmits<{ saved: [FixedBill]; cancel: [] }>();

const periodicityOptions = Object.entries(PERIODICITY_LABELS) as [FixedBillPeriodicity, string][];

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

function blankForm(): FixedBillInput {
  return {
    name: "",
    amount: 0,
    categoryId: null,
    accountId: null,
    periodicity: "monthly",
    dueDate: todayISO(),
    notes: null,
  };
}

const form = reactive<FixedBillInput>(blankForm());
const errorMessage = ref("");
const submitting = ref(false);

watch(
  () => props.bill,
  (bill) => {
    errorMessage.value = "";
    if (bill) {
      form.name = bill.name;
      form.amount = bill.amount;
      form.categoryId = bill.categoryId;
      form.accountId = bill.accountId;
      form.periodicity = bill.periodicity;
      form.dueDate = bill.dueDate;
      form.notes = bill.notes;
    } else {
      Object.assign(form, blankForm());
    }
  },
  { immediate: true },
);

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    const payload: FixedBillInput = { ...form, notes: form.notes || null };
    const saved = props.bill
      ? await updateFixedBill(props.bill.id, payload)
      : await createFixedBill(payload);
    emit("saved", saved);
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar a conta fixa";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div
    class="fixed inset-x-0 top-0 z-50 flex h-dvh items-end justify-center bg-ink-900/40 p-0 sm:items-center sm:p-4"
  >
    <div
      class="flex max-h-full w-full max-w-lg flex-col overflow-y-auto rounded-t-3xl bg-white p-6 shadow-xl sm:rounded-3xl"
    >
      <div class="mb-4 flex items-center justify-between">
        <h2 class="font-display text-lg font-bold text-ink-900">
          {{ bill ? "Editar conta fixa" : "Nova conta fixa" }}
        </h2>
        <button
          type="button"
          class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
          @click="emit('cancel')"
        >
          <X class="h-5 w-5" />
        </button>
      </div>

      <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Nome</span>
          <input
            v-model="form.name"
            type="text"
            required
            placeholder="Ex.: Energia elétrica"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>

        <div class="grid gap-4 sm:grid-cols-2">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Valor estimado</span>
            <input
              v-model.number="form.amount"
              type="number"
              step="0.01"
              min="0.01"
              required
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Próximo vencimento</span>
            <input
              v-model="form.dueDate"
              type="date"
              required
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
        </div>

        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Periodicidade</span>
          <select
            v-model="form.periodicity"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option v-for="[value, label] in periodicityOptions" :key="value" :value="value">
              {{ label }}
            </option>
          </select>
        </label>

        <div class="grid gap-4 sm:grid-cols-2">
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
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Conta pagadora padrão</span>
            <select
              v-model="form.accountId"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            >
              <option :value="null">Sem conta padrão</option>
              <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </label>
        </div>

        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Notas</span>
          <textarea
            v-model="form.notes"
            rows="2"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          ></textarea>
        </label>

        <p v-if="errorMessage" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>

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
            @click="emit('cancel')"
          >
            Cancelar
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
