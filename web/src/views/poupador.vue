<script setup lang="ts">
import { shallowRef } from "vue";
import { Calculator, CircleHelp } from "lucide-vue-next";
import PoupadorCharts from "../components/poupador/PoupadorCharts.vue";
import PoupadorEntryForm from "../components/poupador/PoupadorEntryForm.vue";
import PoupadorEntryList from "../components/poupador/PoupadorEntryList.vue";
import PoupadorSummary from "../components/poupador/PoupadorSummary.vue";
import { usePoupador } from "../composables/usePoupador";
import type { PoupadorEntry, PoupadorEntryDraft, PoupadorKind } from "../types/poupador";

const {
  months,
  incomes,
  expenses,
  incomeMonthlyTotal,
  expenseMonthlyTotal,
  monthlyBalance,
  yearlyBalance,
  monthlySeries,
  saveEntry,
  removeEntry,
} = usePoupador();
const editing = shallowRef<{ kind: PoupadorKind; entry: PoupadorEntry } | null>(null);

function save(kind: PoupadorKind, draft: PoupadorEntryDraft) {
  saveEntry(kind, draft, editing.value?.kind === kind ? editing.value.entry.id : undefined);
  editing.value = null;
}
function edit(kind: PoupadorKind, entry: PoupadorEntry) {
  editing.value = { kind, entry };
}
</script>

<template>
  <div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
    <section class="mb-7 max-w-3xl">
      <div
        class="mb-3 inline-flex items-center gap-2 rounded-full border border-brand-300/70 bg-brand-100/70 px-3 py-1 text-xs font-bold text-brand-800"
      >
        <Calculator class="h-3.5 w-3.5" /> Planejador financeiro
      </div>
      <h1 class="font-display text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">
        Poupador
      </h1>
      <p class="mt-2 text-sm leading-relaxed text-ink-500 sm:text-base">
        Veja, com clareza, quanto do seu dinheiro pode sobrar. Seus dados ficam salvos somente neste
        dispositivo.
      </p>
    </section>
    <PoupadorSummary
      :income="incomeMonthlyTotal"
      :expense="expenseMonthlyTotal"
      :balance="monthlyBalance"
      :yearly-balance="yearlyBalance"
    />
    <section class="mt-7 grid gap-5 lg:grid-cols-2">
      <div>
        <PoupadorEntryForm
          v-if="!editing || editing.kind === 'income'"
          :key="editing?.kind === 'income' ? editing.entry.id : 'income-new'"
          kind="income"
          :entry="editing?.kind === 'income' ? editing.entry : undefined"
          :months="months"
          @save="save('income', $event)"
          @cancel="editing = null"
        />
        <div class="mt-5">
          <PoupadorEntryList
            :entries="incomes"
            kind="income"
            @edit="edit('income', $event)"
            @remove="removeEntry('income', $event)"
          />
        </div>
      </div>
      <div>
        <PoupadorEntryForm
          v-if="!editing || editing.kind === 'expense'"
          :key="editing?.kind === 'expense' ? editing.entry.id : 'expense-new'"
          kind="expense"
          :entry="editing?.kind === 'expense' ? editing.entry : undefined"
          :months="months"
          @save="save('expense', $event)"
          @cancel="editing = null"
        />
        <div class="mt-5">
          <PoupadorEntryList
            :entries="expenses"
            kind="expense"
            @edit="edit('expense', $event)"
            @remove="removeEntry('expense', $event)"
          />
        </div>
      </div>
    </section>
    <section class="mt-7">
      <div class="mb-4 flex items-center gap-2 text-xs text-ink-500">
        <CircleHelp class="h-4 w-4 text-brand-700" /><span
          >Recorrências são convertidas em média mensal; por exemplo, um valor anual é dividido por
          12.</span
        >
      </div>
      <PoupadorCharts :incomes="incomes" :expenses="expenses" :monthly-series="monthlySeries" />
    </section>
  </div>
</template>
