<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { Star } from "lucide-vue-next";
import { getFinancialHealth } from "../api/health";
import { healthLevelLabels, type HealthLevel, type HealthScore } from "../types/health";

const props = defineProps<{
  month: number;
  year: number;
}>();

const score = ref<HealthScore | null>(null);
const loading = ref(true);
const failed = ref(false);
const mascotMissing = ref(false);

watch(
  () => [props.month, props.year],
  async () => {
    loading.value = true;
    failed.value = false;
    try {
      score.value = await getFinancialHealth(props.month, props.year);
    } catch {
      failed.value = true;
      score.value = null;
    } finally {
      loading.value = false;
    }
  },
  { immediate: true },
);

const gradients: Record<HealthLevel, string> = {
  otima: "linear-gradient(135deg, #34d399 0%, #059669 100%)",
  boa: "linear-gradient(135deg, #2dd4bf 0%, #0d9488 100%)",
  estavel: "linear-gradient(135deg, #ffb42e 0%, #f2751f 100%)",
  ruim: "linear-gradient(135deg, #fb7185 0%, #d63163 100%)",
  pessima: "linear-gradient(135deg, #e8496b 0%, #8f1440 100%)",
};

const messages: Record<HealthLevel, string> = {
  otima: "Você está poupando muito bem. Continue assim!",
  boa: "Suas finanças vão bem, com folga no orçamento.",
  estavel: "Receitas e despesas estão equilibradas.",
  ruim: "As despesas estão ultrapassando as receitas.",
  pessima: "Atenção: as despesas superam muito as receitas.",
};

const level = computed<HealthLevel | null>(() =>
  score.value?.hasData && score.value.level ? (score.value.level as HealthLevel) : null,
);

const cardStyle = computed(() => ({
  background: level.value
    ? gradients[level.value]
    : "linear-gradient(135deg, #d3c6b8 0%, #7c6d5c 100%)",
}));

const savingsRateLabel = computed(() => {
  if (!score.value) return "";
  const rate = score.value.savingsRate;
  const rounded = Math.round(rate);
  return `${rounded > 0 ? "+" : ""}${rounded}% das receitas`;
});

// Largura da barra "despesas sobre receitas", limitada para nunca estourar.
const expenseRatioWidth = computed(() => {
  if (!score.value || score.value.income <= 0) return 100;
  return Math.min((score.value.expense / score.value.income) * 100, 100);
});

function formatCurrency(value: number) {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}
</script>

<template>
  <!-- Bloco desativado no painel de admin: não renderiza nada -->
  <div v-if="!loading && (failed || (score && !score.enabled))" class="hidden"></div>

  <div
    v-else
    class="relative overflow-hidden rounded-3xl p-5 text-white shadow-lg transition-[background] duration-500"
    :style="cardStyle"
  >
    <!-- Brilhos decorativos -->
    <div class="absolute -right-10 -top-12 h-40 w-40 rounded-full bg-white/15 blur-2xl"></div>
    <div class="absolute -bottom-14 -left-8 h-36 w-36 rounded-full bg-black/10 blur-2xl"></div>

    <!-- Carregando -->
    <div v-if="loading" class="relative flex animate-pulse flex-col gap-3">
      <div class="h-3 w-24 rounded-full bg-white/30"></div>
      <div class="h-6 w-32 rounded-full bg-white/40"></div>
      <div class="h-3 w-40 rounded-full bg-white/30"></div>
    </div>

    <template v-else-if="score">
      <div class="relative flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="text-[11px] font-semibold uppercase tracking-widest text-white/70">
            Saúde financeira
          </p>

          <!-- Sem lançamentos no mês -->
          <template v-if="!score.hasData">
            <p class="mt-2 font-display text-xl font-bold">Sem dados</p>
            <p class="mt-1 text-sm text-white/80">
              Nenhum lançamento neste mês ainda. Importe ou adicione transações para acompanhar sua
              saúde financeira.
            </p>
          </template>

          <template v-else>
            <div class="mt-2 flex items-center gap-1" role="img" :aria-label="`${score.stars} de 5 estrelas`">
              <Star
                v-for="i in 5"
                :key="i"
                :size="18"
                class="drop-shadow-sm"
                :class="i <= score.stars ? 'fill-white text-white' : 'fill-white/20 text-white/30'"
              />
            </div>
            <p class="mt-1.5 font-display text-2xl font-bold leading-tight">
              {{ level ? healthLevelLabels[level] : "" }}
            </p>
            <p class="mt-1 text-sm leading-snug text-white/85">
              {{ level ? messages[level] : "" }}
            </p>
          </template>
        </div>

        <!-- Mascote (porquinho) -->
        <img
          v-if="!mascotMissing"
          src="/images/mascot-pig.svg"
          alt=""
          class="pointer-events-none relative -mr-1 -mt-1 w-20 shrink-0 drop-shadow-lg sm:w-24"
          @error="mascotMissing = true"
        />
      </div>

      <template v-if="score.hasData">
        <!-- Barra de despesas sobre receitas -->
        <div class="relative mt-4 h-1.5 overflow-hidden rounded-full bg-white/25">
          <div
            class="h-full rounded-full bg-white/90 transition-all duration-500"
            :style="{ width: `${expenseRatioWidth}%` }"
          ></div>
        </div>

        <div class="relative mt-3 flex items-end justify-between gap-2 text-xs">
          <div>
            <p class="text-white/70">Receitas</p>
            <p class="font-semibold">{{ formatCurrency(score.income) }}</p>
          </div>
          <div class="text-right">
            <p class="text-white/70">Despesas</p>
            <p class="font-semibold">{{ formatCurrency(score.expense) }}</p>
          </div>
        </div>

        <span
          class="relative mt-3 inline-flex items-center rounded-full bg-white/20 px-2.5 py-1 text-[11px] font-semibold backdrop-blur-sm"
        >
          Saldo {{ formatCurrency(score.balance) }} &middot; {{ savingsRateLabel }}
        </span>
      </template>
    </template>
  </div>
</template>
