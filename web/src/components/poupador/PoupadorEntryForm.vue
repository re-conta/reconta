<script setup lang="ts">
import { computed, reactive } from "vue";
import { Plus } from "lucide-vue-next";
import type {
  PoupadorEntry,
  PoupadorEntryDraft,
  PoupadorFrequency,
  PoupadorKind,
} from "../../types/poupador";

const props = defineProps<{ kind: PoupadorKind; entry?: PoupadorEntry; months: string[] }>();
const emit = defineEmits<{ save: [draft: PoupadorEntryDraft]; cancel: [] }>();

const isIncome = computed(() => props.kind === "income");
const form = reactive<PoupadorEntryDraft>({
  name: props.entry?.name ?? "",
  amount: props.entry?.amount ?? 0,
  frequency: props.entry?.frequency ?? "monthly",
  month: props.entry?.month ?? new Date().getMonth() + 1,
});

const frequencies: Array<{ value: PoupadorFrequency; label: string }> = [
  { value: "monthly", label: "Mensal" },
  { value: "weekly", label: "Semanal" },
  { value: "biweekly", label: "Quinzenal" },
  { value: "yearly", label: "Anual" },
  { value: "one-time", label: "Único" },
];

function submit() {
  const name = form.name.trim();
  const amount = Number(String(form.amount).trim());
  if (!name || !Number.isFinite(amount) || amount <= 0) return;

  form.name = name;
  form.amount = amount;
  emit("save", { ...form });
  if (!props.entry) {
    form.name = "";
    form.amount = 0;
    form.frequency = "monthly";
  }
}
</script>

<template>
  <form
    class="w-full min-w-0 rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm sm:p-5"
    @submit.prevent="submit"
  >
    <div class="mb-4 flex items-center gap-2">
      <span
        :class="isIncome ? 'bg-brand-100 text-brand-700' : 'bg-coral-100 text-coral-700'"
        class="rounded-full px-2.5 py-1 text-xs font-bold"
      >
        {{ isIncome ? "Receita" : "Gasto" }}
      </span>
      <p class="text-sm text-ink-500">
        {{ entry ? "Atualize os dados da fonte" : "Adicione uma nova fonte" }}
      </p>
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label class="min-w-0 text-xs font-bold text-ink-600"
        >Nome
        <input
          v-model.trim="form.name"
          required
          maxlength="60"
          placeholder="Ex.: Salário"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-ink-200 bg-ink-50 px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
        />
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600"
        >Valor
        <input
          v-model.number="form.amount"
          required
          min="0.01"
          step="0.01"
          type="number"
          inputmode="decimal"
          placeholder="0,00"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-ink-200 bg-ink-50 px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
        />
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600"
        >Frequência
        <select
          v-model="form.frequency"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-ink-200 bg-ink-50 px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
        >
          <option v-for="frequency in frequencies" :key="frequency.value" :value="frequency.value">
            {{ frequency.label }}
          </option>
        </select>
      </label>
      <label v-if="form.frequency === 'one-time'" class="min-w-0 text-xs font-bold text-ink-600"
        >Mês de ocorrência
        <select
          v-model.number="form.month"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-ink-200 bg-ink-50 px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
        >
          <option v-for="(month, index) in months" :key="month" :value="index + 1">
            {{ month }}
          </option>
        </select>
      </label>
    </div>
    <div class="mt-4 flex flex-wrap items-center gap-3">
      <button
        type="submit"
        :class="isIncome ? 'bg-ink-900 hover:bg-ink-800' : 'bg-coral-600 hover:bg-coral-700'"
        class="inline-flex items-center gap-2 rounded-full px-4 py-2.5 text-sm font-bold text-white transition"
      >
        <Plus class="h-4 w-4" />
        {{ entry ? "Salvar alteração" : `Adicionar ${isIncome ? "receita" : "gasto"}` }}
      </button>
      <button
        v-if="entry"
        type="button"
        class="text-sm font-bold text-ink-500 transition hover:text-ink-900"
        @click="emit('cancel')"
      >
        Cancelar
      </button>
    </div>
  </form>
</template>
