<script setup lang="ts">
import { onMounted, onUnmounted, shallowRef, watch } from "vue";
import { Calculator, CircleHelp } from "lucide-vue-next";
import PoupadorCharts from "../components/poupador/PoupadorCharts.vue";
import PoupadorEntryForm from "../components/poupador/PoupadorEntryForm.vue";
import PoupadorEntryList from "../components/poupador/PoupadorEntryList.vue";
import PoupadorSummary from "../components/poupador/PoupadorSummary.vue";
import PoupadorShareControl from "../components/poupador/PoupadorShareControl.vue";
import { usePoupador } from "../composables/usePoupador";
import type { PoupadorEntry, PoupadorEntryDraft, PoupadorKind } from "../types/poupador";
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();
const isPoupadorHost = window.location.hostname === "poupa.reconta.app";
const snapshotID =
  typeof route.params.snapshotId === "string" ? route.params.snapshotId : undefined;
const poupadorTitle = "Poupador - ReConta";
const poupadorDescription =
  "Organize receitas e gastos, acompanhe o saldo mensal e projete suas finanças do ano com o Poupador da ReConta.";
const originalTitle = document.title;
const metadata = [
  { selector: 'meta[name="description"]', content: poupadorDescription },
  { selector: 'meta[property="og:site_name"]', content: "ReConta" },
  { selector: 'meta[property="og:title"]', content: poupadorTitle },
  { selector: 'meta[property="og:description"]', content: poupadorDescription },
  { selector: 'meta[name="twitter:title"]', content: poupadorTitle },
  { selector: 'meta[name="twitter:description"]', content: poupadorDescription },
];
const originalMetadata = metadata.map(({ selector }) => ({
  selector,
  content: document.querySelector<HTMLMetaElement>(selector)?.content,
}));

const {
  months,
  incomes,
  expenses,
  incomeMonthlyTotal,
  expenseMonthlyTotal,
  monthlyBalance,
  yearlyBalance,
  monthlySeries,
  loadedSnapshotID,
  shareURL,
  saveStatus,
  saveError,
  isLoadingSnapshot,
  snapshotLoadError,
  saveEntry,
  removeEntry,
} = usePoupador(snapshotID);
const editing = shallowRef<{ kind: PoupadorKind; entry: PoupadorEntry } | null>(null);
const copied = shallowRef(false);

onMounted(() => {
  document.title = poupadorTitle;
  for (const { selector, content } of metadata) {
    document.querySelector<HTMLMetaElement>(selector)?.setAttribute("content", content);
  }
});

onUnmounted(() => {
  document.title = originalTitle;
  for (const { selector, content } of originalMetadata) {
    if (content !== undefined) {
      document.querySelector<HTMLMetaElement>(selector)?.setAttribute("content", content);
    }
  }
});

watch(loadedSnapshotID, (id) => {
  if (!id) return;
  const path = isPoupadorHost ? `/${id}` : `/poupa/${id}`;
  if (route.path !== path) void router.replace(path);
});

function save(kind: PoupadorKind, draft: PoupadorEntryDraft) {
  saveEntry(kind, draft, editing.value?.kind === kind ? editing.value.entry.id : undefined);
  editing.value = null;
}
function edit(kind: PoupadorKind, entry: PoupadorEntry) {
  editing.value = { kind, entry };
}

async function copyShareURL() {
  if (!shareURL.value) return;
  try {
    await navigator.clipboard.writeText(shareURL.value);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2_000);
  } catch {
    // O link continua visível na barra do navegador caso a área de transferência seja bloqueada.
  }
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
        Veja, com clareza, quanto do seu dinheiro pode sobrar. Cada alteração salva um resultado
        compartilhável.
      </p>
      <PoupadorShareControl
        :share-url="shareURL"
        :status="saveStatus"
        :error="saveError"
        :copied="copied"
        @copy="copyShareURL"
      />
    </section>
    <div
      v-if="isLoadingSnapshot"
      class="rounded-3xl border border-ink-200 bg-white p-8 text-center text-sm text-ink-500"
    >
      Carregando resultado compartilhado…
    </div>
    <div
      v-else-if="snapshotLoadError"
      class="rounded-3xl border border-coral-200 bg-coral-50 p-8 text-center text-sm text-coral-700"
    >
      {{ snapshotLoadError }}
    </div>
    <template v-else>
      <PoupadorSummary
        :income="incomeMonthlyTotal"
        :expense="expenseMonthlyTotal"
        :balance="monthlyBalance"
        :yearly-balance="yearlyBalance"
      />
      <section class="mt-7 grid gap-5 lg:grid-cols-2">
        <div class="min-w-0">
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
        <div class="min-w-0">
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
            >Recorrências são convertidas em média mensal; por exemplo, um valor anual é dividido
            por 12.</span
          >
        </div>
        <PoupadorCharts :incomes="incomes" :expenses="expenses" :monthly-series="monthlySeries" />
      </section>
    </template>
  </div>
</template>
