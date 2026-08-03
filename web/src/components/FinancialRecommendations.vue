<script setup lang="ts">
import { Lightbulb, Lock, Scissors, Sparkles, TrendingUp } from "lucide-vue-next";
import { onBeforeUnmount, ref, watch } from "vue";
import { getRecommendations } from "../api/health";
import type { Recommendation, RecommendationsResponse } from "../types/health";

const props = defineProps<{
  month: number;
  year: number;
  isPayingUser: boolean;
}>();

const data = ref<RecommendationsResponse | null>(null);
const loading = ref(true);
const failed = ref(false);

// Enquanto o back-end estiver processando a análise (a fila do Groq é
// deliberadamente lenta, para nunca estourar o plano gratuito), a tela faz um
// polling espaçado até o resultado ficar pronto ou desistimos depois de um
// tempo — sem martelar o servidor.
let pollTimer: ReturnType<typeof setTimeout> | null = null;
const POLL_INTERVAL_MS = 20_000;
const MAX_POLL_ATTEMPTS = 12;

function clearPoll() {
  if (pollTimer) {
    clearTimeout(pollTimer);
    pollTimer = null;
  }
}

async function load(attempt = 0) {
  clearPoll();
  try {
    const res = await getRecommendations(props.month, props.year);
    data.value = res;
    failed.value = false;
    if (res.status === "pending" && attempt < MAX_POLL_ATTEMPTS) {
      pollTimer = setTimeout(() => load(attempt + 1), POLL_INTERVAL_MS);
    }
  } catch {
    failed.value = true;
    data.value = null;
  } finally {
    loading.value = false;
  }
}

watch(
  () => [props.month, props.year, props.isPayingUser],
  () => {
    if (!props.isPayingUser) {
      loading.value = false;
      return;
    }
    loading.value = true;
    load();
  },
  { immediate: true },
);

onBeforeUnmount(clearPoll);

function icon(kind: Recommendation["kind"]) {
  return kind === "invest" ? TrendingUp : Scissors;
}

function badgeLabel(kind: Recommendation["kind"]) {
  return kind === "invest" ? "Investir" : "Cortar gasto";
}
</script>

<template>
  <div v-if="!isPayingUser" class="rounded-3xl border border-black/5 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-neutral-900">
    <div class="flex items-center gap-2">
      <Sparkles :size="18" class="text-indigo-500" />
      <h3 class="font-display text-base font-bold text-neutral-800 dark:text-neutral-100">
        Recomendações da IA
      </h3>
    </div>

    <div class="mt-4 flex flex-col items-center gap-3 rounded-2xl border border-dashed border-indigo-200 bg-indigo-50/50 p-6 text-center dark:border-indigo-500/20 dark:bg-indigo-500/5">
      <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-indigo-100 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
        <Lock :size="18" />
      </div>
      <p class="text-sm font-medium text-neutral-700 dark:text-neutral-200">
        Recomendações personalizadas de IA são exclusivas para assinantes.
      </p>
      <p class="text-sm text-neutral-500 dark:text-neutral-400">
        Assine um plano para receber sugestões de investimento e cortes de gastos com base nas suas transações.
      </p>
      <RouterLink
        to="/planos"
        class="mt-1 rounded-full bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-700"
      >
        Ver planos
      </RouterLink>
    </div>
  </div>

  <div v-else-if="!loading && (failed || !data || data.status === 'disabled' || data.status === 'no_data')" class="hidden"></div>

  <div v-else class="rounded-3xl border border-black/5 bg-white p-5 shadow-sm dark:border-white/10 dark:bg-neutral-900">
    <div class="flex items-center gap-2">
      <Sparkles :size="18" class="text-indigo-500" />
      <h3 class="font-display text-base font-bold text-neutral-800 dark:text-neutral-100">
        Recomendações da IA
      </h3>
    </div>

    <div v-if="loading" class="mt-4 animate-pulse space-y-3">
      <div class="h-14 rounded-2xl bg-neutral-100 dark:bg-neutral-800"></div>
      <div class="h-14 rounded-2xl bg-neutral-100 dark:bg-neutral-800"></div>
    </div>

    <div v-else-if="data?.status === 'pending'" class="mt-4 flex items-start gap-3 rounded-2xl bg-indigo-50 p-4 text-sm text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-300">
      <Lightbulb :size="18" class="mt-0.5 shrink-0" />
      <p>
        Estamos analisando suas transações deste mês para gerar recomendações personalizadas.
        Isso pode levar alguns minutos — volte a esta seção em instantes.
      </p>
    </div>

    <ul v-else-if="data?.recommendations?.length" class="mt-4 space-y-3">
      <li
        v-for="(rec, i) in data.recommendations"
        :key="i"
        class="flex gap-3 rounded-2xl border border-black/5 p-4 dark:border-white/10"
      >
        <div
          class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
          :class="rec.kind === 'invest' ? 'bg-emerald-100 text-emerald-600 dark:bg-emerald-500/10 dark:text-emerald-400' : 'bg-amber-100 text-amber-600 dark:bg-amber-500/10 dark:text-amber-400'"
        >
          <component :is="icon(rec.kind)" :size="18" />
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <p class="font-semibold text-neutral-800 dark:text-neutral-100">{{ rec.title }}</p>
            <span
              class="rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide"
              :class="rec.kind === 'invest' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-500/10 dark:text-amber-400'"
            >
              {{ badgeLabel(rec.kind) }}
            </span>
          </div>
          <p class="mt-1 text-sm text-neutral-600 dark:text-neutral-400">{{ rec.description }}</p>
          <p v-if="rec.impact" class="mt-1.5 text-xs font-medium text-neutral-500 dark:text-neutral-500">
            {{ rec.impact }}
          </p>
        </div>
      </li>
    </ul>

    <p class="mt-3 text-[11px] text-neutral-400 dark:text-neutral-600">
      Gerado por IA a partir das suas transações do mês. Pode conter imprecisões — use como ponto de partida, não como conselho financeiro definitivo.
    </p>
  </div>
</template>
