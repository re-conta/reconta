<script setup lang="ts">
import { reactive, ref } from "vue";
import { X } from "lucide-vue-next";
import { ApiError, payFixedBill } from "../api/fixedBills";
import type { Account } from "../types/account";
import type { FixedBill, PayFixedBillInput, PayFixedBillResult } from "../types/fixedBill";

const props = defineProps<{ bill: FixedBill; accounts: Account[] }>();
const emit = defineEmits<{ paid: [PayFixedBillResult]; cancel: [] }>();

const detailed = ref(false);
const errorMessage = ref("");
const submitting = ref(false);

function todayISO() {
  return new Date().toISOString().slice(0, 10);
}

const paymentMethods = [
  { value: "pix", label: "Pix" },
  { value: "boleto", label: "Boleto" },
  { value: "debit_card", label: "Cartão de débito" },
  { value: "credit_card", label: "Cartão de crédito" },
  { value: "cash", label: "Dinheiro" },
  { value: "bank_transfer", label: "Transferência" },
];

const form = reactive({
  bank: "",
  paymentMethod: "",
  paidAt: todayISO(),
  amountPaid: props.bill.amount,
  accountId: props.bill.accountId,
  notes: "",
});

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    const input: PayFixedBillInput = detailed.value
      ? {
          bank: form.bank || null,
          paymentMethod: form.paymentMethod || null,
          paidAt: form.paidAt || null,
          amountPaid: form.amountPaid || null,
          accountId: form.accountId,
          notes: form.notes || null,
        }
      : {};
    const result = await payFixedBill(props.bill.id, input);
    emit("paid", result);
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao registrar o pagamento";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div
    class="fixed inset-0 z-50 flex items-end justify-center bg-ink-900/40 p-0 sm:items-center sm:p-4"
  >
    <div
      class="flex max-h-full w-full max-w-lg flex-col overflow-y-auto rounded-t-3xl bg-white p-6 shadow-xl sm:rounded-3xl"
    >
      <div class="mb-4 flex items-center justify-between">
        <div>
          <h2 class="font-display text-lg font-bold text-ink-900">Marcar como paga</h2>
          <p class="text-sm text-ink-500">{{ bill.name }}</p>
        </div>
        <button
          type="button"
          class="rounded-full p-1.5 text-ink-400 transition hover:bg-ink-100 hover:text-ink-700"
          @click="emit('cancel')"
        >
          <X class="h-5 w-5" />
        </button>
      </div>

      <div class="mb-4 flex rounded-full bg-ink-100 p-1 text-sm font-medium">
        <button
          type="button"
          class="flex-1 rounded-full px-3 py-1.5 transition"
          :class="!detailed ? 'bg-white text-ink-900 shadow-sm' : 'text-ink-500'"
          @click="detailed = false"
        >
          Pagamento simples
        </button>
        <button
          type="button"
          class="flex-1 rounded-full px-3 py-1.5 transition"
          :class="detailed ? 'bg-white text-ink-900 shadow-sm' : 'text-ink-500'"
          @click="detailed = true"
        >
          Informar detalhes
        </button>
      </div>

      <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
        <template v-if="detailed">
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Data do pagamento</span>
              <input
                v-model="form.paidAt"
                type="date"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Valor pago</span>
              <input
                v-model.number="form.amountPaid"
                type="number"
                step="0.01"
                min="0.01"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Banco pagador</span>
              <input
                v-model="form.bank"
                type="text"
                placeholder="Ex.: Nubank"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Método de pagamento</span>
              <select
                v-model="form.paymentMethod"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              >
                <option value="">Não informado</option>
                <option v-for="m in paymentMethods" :key="m.value" :value="m.value">
                  {{ m.label }}
                </option>
              </select>
            </label>
          </div>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Conta utilizada</span>
            <select
              v-model="form.accountId"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            >
              <option :value="null">Sem conta</option>
              <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Notas</span>
            <textarea
              v-model="form.notes"
              rows="2"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            ></textarea>
          </label>
        </template>
        <p v-else class="text-sm text-ink-500">
          Registra o pagamento hoje, no valor estimado da conta ({{
            bill.amount.toLocaleString("pt-BR", { style: "currency", currency: "BRL" })
          }}), sem detalhes adicionais.
        </p>

        <p v-if="errorMessage" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>

        <div class="flex gap-3">
          <button
            type="submit"
            :disabled="submitting"
            class="rounded-full bg-emerald-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-emerald-700 disabled:opacity-50"
          >
            {{ submitting ? "Registrando..." : "Confirmar pagamento" }}
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
