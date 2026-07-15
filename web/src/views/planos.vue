<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { BadgeCheck, Check, ShieldCheck, Sparkles, Zap } from "lucide-vue-next";
import { ApiError, getSubscription, listPlans } from "../api/billing";
import SubscribeModal from "../components/modals/SubscribeModal.vue";
import { useAuth } from "../composables/useAuth";
import { formatPrice } from "../types/billing";
import type { BillingCycle, Plan, SubscriptionInfo } from "../types/billing";

const router = useRouter();
const { currentUser, initialized, init } = useAuth();

const plans = ref<Plan[]>([]);
const info = ref<SubscriptionInfo | null>(null);
const loading = ref(true);
const error = ref("");

const cycle = ref<BillingCycle>("monthly");
const selectedPlan = ref<Plan | null>(null);

const yearlyDiscount = computed(() => {
  const paid = plans.value.find((p) => p.priceMonthly > 0 && p.priceYearly > 0);
  if (!paid) return 0;
  return Math.round((1 - paid.priceYearly / (paid.priceMonthly * 12)) * 100);
});

function priceFor(plan: Plan) {
  return cycle.value === "yearly" ? plan.priceYearly : plan.priceMonthly;
}

function monthlyEquivalent(plan: Plan) {
  return plan.priceYearly / 12;
}

function isCurrentPlan(plan: Plan) {
  if (!info.value) return plan.priceMonthly === 0 && !!currentUser.value;
  const sub = info.value.subscription;
  if (info.value.planCode === plan.code && plan.priceMonthly === 0) return true;
  return !!sub && sub.status === "active" && sub.planCode === plan.code;
}

function ctaLabel(plan: Plan) {
  if (plan.priceMonthly === 0) return isCurrentPlan(plan) ? "Seu plano atual" : "Começar grátis";
  if (isCurrentPlan(plan)) {
    return info.value?.subscription?.cycle === cycle.value ? "Renovar agora" : "Assinar";
  }
  return "Assinar";
}

function handleCta(plan: Plan) {
  if (!currentUser.value) {
    router.push({ name: "Login", query: { redirect: "/planos" } });
    return;
  }
  if (plan.priceMonthly === 0) {
    if (!isCurrentPlan(plan)) router.push("/configuracoes");
    return;
  }
  selectedPlan.value = plan;
}

async function load() {
  loading.value = true;
  error.value = "";
  try {
    if (!initialized.value) await init();
    const requests: [Promise<Plan[]>, Promise<SubscriptionInfo> | null] = [
      listPlans(),
      currentUser.value ? getSubscription() : null,
    ];
    plans.value = await requests[0];
    if (requests[1]) info.value = await requests[1];
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao carregar os planos";
  } finally {
    loading.value = false;
  }
}

async function handleSubscribed() {
  if (currentUser.value) {
    try {
      info.value = await getSubscription();
    } catch {
      // O modal já mostrou o sucesso; a página atualiza no próximo load.
    }
  }
}

onMounted(load);
</script>

<template>
  <div class="mx-auto flex w-full max-w-5xl flex-col gap-10 px-4 py-10 md:px-6 md:py-14">
    <!-- Hero -->
    <div class="flex flex-col items-center gap-4 text-center">
      <span
        class="flex items-center gap-1.5 rounded-full border border-brand-200 bg-brand-50 px-3.5 py-1.5 text-xs font-semibold text-brand-700"
      >
        <Sparkles class="h-3.5 w-3.5" />
        Planos e preços
      </span>
      <h1 class="font-display text-3xl font-bold text-ink-900 sm:text-4xl">
        Escolha o plano ideal para as suas finanças
      </h1>
      <p class="max-w-xl text-sm text-ink-500 sm:text-base">
        Comece grátis e evolua quando precisar. Pague com PIX, boleto ou cartão — e cancele a
        qualquer momento, com reembolso proporcional ao tempo não usado.
      </p>

      <!-- Toggle mensal/anual -->
      <div
        class="mt-2 flex items-center gap-1 rounded-full border border-ink-200/70 bg-white p-1 shadow-sm"
      >
        <button
          type="button"
          class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
          :class="cycle === 'monthly' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'"
          @click="cycle = 'monthly'"
        >
          Mensal
        </button>
        <button
          type="button"
          class="flex items-center gap-1.5 rounded-full px-4 py-1.5 text-sm font-semibold transition"
          :class="cycle === 'yearly' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'"
          @click="cycle = 'yearly'"
        >
          Anual
          <span
            v-if="yearlyDiscount > 0"
            class="rounded-full px-1.5 py-0.5 text-[10px] font-bold"
            :class="cycle === 'yearly' ? 'bg-emerald-400 text-emerald-950' : 'bg-emerald-100 text-emerald-700'"
          >
            -{{ yearlyDiscount }}%
          </span>
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex flex-col items-center gap-2 py-16 text-sm text-ink-400">
      <span
        class="h-6 w-6 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
      ></span>
      Carregando planos...
    </div>
    <p v-else-if="error" class="rounded-xl bg-coral-50 px-4 py-3 text-center text-sm text-coral-700">
      {{ error }}
    </p>

    <!-- Cards de planos -->
    <div v-else class="grid gap-5 md:grid-cols-3">
      <div
        v-for="plan in plans"
        :key="plan.id"
        class="relative flex flex-col gap-5 rounded-3xl border bg-white p-6 shadow-sm transition hover:shadow-md sm:p-7"
        :class="
          plan.highlight
            ? 'border-brand-400 ring-4 ring-brand-100 md:-my-2 md:py-9'
            : 'border-ink-200/70'
        "
      >
        <span
          v-if="plan.highlight"
          class="absolute -top-3 left-1/2 flex -translate-x-1/2 items-center gap-1 rounded-full bg-linear-to-r from-brand-500 to-coral-500 px-3 py-1 text-[11px] font-bold uppercase tracking-wide text-white shadow-sm"
        >
          <Zap class="h-3 w-3" />
          Mais popular
        </span>

        <div>
          <h2 class="font-display text-lg font-bold text-ink-900">{{ plan.name }}</h2>
          <p class="mt-0.5 text-sm text-ink-500">{{ plan.description }}</p>
        </div>

        <div class="flex items-baseline gap-1.5">
          <span class="font-display text-3xl font-bold text-ink-900">
            {{ priceFor(plan) === 0 ? "R$ 0" : formatPrice(priceFor(plan)) }}
          </span>
          <span v-if="priceFor(plan) > 0" class="text-sm text-ink-400">
            /{{ cycle === "yearly" ? "ano" : "mês" }}
          </span>
        </div>
        <p
          v-if="cycle === 'yearly' && plan.priceYearly > 0"
          class="-mt-3 text-xs text-emerald-600"
        >
          equivale a {{ formatPrice(monthlyEquivalent(plan)) }}/mês
        </p>

        <button
          type="button"
          :disabled="isCurrentPlan(plan) && plan.priceMonthly === 0"
          class="rounded-full px-5 py-2.5 text-sm font-semibold shadow-sm transition disabled:cursor-default disabled:opacity-60"
          :class="
            plan.highlight
              ? 'bg-linear-to-r from-brand-500 to-coral-500 text-white hover:opacity-90'
              : 'bg-ink-900 text-white hover:bg-ink-800'
          "
          @click="handleCta(plan)"
        >
          {{ ctaLabel(plan) }}
        </button>
        <p
          v-if="isCurrentPlan(plan) && plan.priceMonthly > 0 && info?.subscription?.currentPeriodEnd"
          class="-mt-2 text-center text-xs text-ink-400"
        >
          Ativo até
          {{ new Date(info.subscription.currentPeriodEnd).toLocaleDateString("pt-BR") }}
        </p>

        <ul class="flex flex-col gap-2.5 border-t border-ink-100 pt-5">
          <li
            v-for="benefit in plan.benefits"
            :key="benefit"
            class="flex items-start gap-2 text-sm text-ink-600"
          >
            <Check class="mt-0.5 h-4 w-4 shrink-0 text-brand-600" />
            {{ benefit }}
          </li>
        </ul>
      </div>
    </div>

    <!-- Garantias -->
    <div class="grid gap-4 sm:grid-cols-3">
      <div
        class="flex items-start gap-3 rounded-2xl border border-ink-200/70 bg-white p-4 shadow-sm"
      >
        <ShieldCheck class="h-5 w-5 shrink-0 text-brand-600" />
        <div>
          <p class="text-sm font-semibold text-ink-900">Pagamento seguro</p>
          <p class="text-xs text-ink-500">
            Processado pelo Mercado Pago. Não guardamos os dados do seu cartão.
          </p>
        </div>
      </div>
      <div
        class="flex items-start gap-3 rounded-2xl border border-ink-200/70 bg-white p-4 shadow-sm"
      >
        <BadgeCheck class="h-5 w-5 shrink-0 text-brand-600" />
        <div>
          <p class="text-sm font-semibold text-ink-900">Cancele quando quiser</p>
          <p class="text-xs text-ink-500">
            Reembolso proporcional ao tempo não usado, ou acesso até o fim do ciclo.
          </p>
        </div>
      </div>
      <div
        class="flex items-start gap-3 rounded-2xl border border-ink-200/70 bg-white p-4 shadow-sm"
      >
        <Sparkles class="h-5 w-5 shrink-0 text-brand-600" />
        <div>
          <p class="text-sm font-semibold text-ink-900">Lembretes de renovação</p>
          <p class="text-xs text-ink-500">
            Avisamos no site e por e-mail antes da sua assinatura vencer.
          </p>
        </div>
      </div>
    </div>

    <SubscribeModal
      v-if="selectedPlan"
      :plan="selectedPlan"
      :cycle="cycle"
      @close="selectedPlan = null"
      @subscribed="handleSubscribed"
    />
  </div>
</template>
