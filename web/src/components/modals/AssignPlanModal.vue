<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { CreditCard } from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";
import {
  ApiError,
  adminCancelSubscription,
  adminGetUserSubscription,
  adminGrantSubscription,
} from "../../api/billing";
import type { BillingCycle, Plan, SubscriptionInfo } from "../../types/billing";
import { PlanFree } from "../../types/billing";
import type { User } from "../../types/user";

const props = defineProps<{
  user: User;
  plans: Plan[];
}>();

const emit = defineEmits<{ close: []; updated: [info: SubscriptionInfo] }>();

const loading = ref(true);
const info = ref<SubscriptionInfo | null>(null);
const error = ref("");
const submitting = ref(false);

const paidPlans = computed(() => props.plans.filter((p) => p.code !== PlanFree));
const selectedPlanCode = ref("");
const selectedCycle = ref<BillingCycle>("monthly");

async function load() {
  loading.value = true;
  error.value = "";
  try {
    info.value = await adminGetUserSubscription(props.user.id);
    selectedPlanCode.value =
      info.value.planCode !== PlanFree ? info.value.planCode : (paidPlans.value[0]?.code ?? "");
    if (info.value.subscription) selectedCycle.value = info.value.subscription.cycle;
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao carregar plano do usuário";
  } finally {
    loading.value = false;
  }
}

onMounted(load);

async function handleGrant() {
  if (!selectedPlanCode.value) return;
  submitting.value = true;
  error.value = "";
  try {
    const updated = await adminGrantSubscription(props.user.id, {
      planCode: selectedPlanCode.value,
      cycle: selectedCycle.value,
    });
    emit("updated", updated);
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao conceder o plano";
  } finally {
    submitting.value = false;
  }
}

async function handleRemove() {
  submitting.value = true;
  error.value = "";
  try {
    const updated = await adminCancelSubscription(props.user.id);
    emit("updated", updated);
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao remover o plano";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <BaseModal
    title="Plano do usuário"
    :subtitle="user.name"
    :icon="CreditCard"
    @close="emit('close')"
  >
    <div v-if="loading" class="flex flex-col items-center gap-2 p-8 text-sm text-ink-400">
      <span
        class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
      ></span>
      Carregando...
    </div>
    <div v-else class="flex flex-col gap-4">
      <p class="text-sm text-ink-600">
        Plano atual:
        <span class="font-semibold text-ink-900">{{
          info?.subscription?.planName ?? "Gratuito"
        }}</span>
        <span v-if="info?.subscription?.currentPeriodEnd" class="text-ink-500">
          &middot; válido até
          {{ new Date(info.subscription.currentPeriodEnd).toLocaleDateString("pt-BR") }}
        </span>
      </p>

      <div class="flex flex-col gap-3 rounded-2xl border border-ink-200/70 p-4">
        <p class="text-sm font-semibold text-ink-900">Conceder plano pago</p>
        <label class="flex flex-col gap-1.5">
          <span class="text-xs font-medium text-ink-600">Plano</span>
          <select
            v-model="selectedPlanCode"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option v-for="plan in paidPlans" :key="plan.code" :value="plan.code">
              {{ plan.name }}
            </option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-xs font-medium text-ink-600">Ciclo</span>
          <select
            v-model="selectedCycle"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option value="monthly">Mensal</option>
            <option value="yearly">Anual</option>
          </select>
        </label>
        <button
          type="button"
          :disabled="submitting || !selectedPlanCode"
          class="rounded-full bg-ink-900 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
          @click="handleGrant"
        >
          {{ submitting ? "Salvando..." : "Conceder plano" }}
        </button>
      </div>

      <button
        v-if="info?.planCode !== PlanFree"
        type="button"
        :disabled="submitting"
        class="rounded-full border border-coral-200 px-5 py-2.5 text-sm font-semibold text-coral-700 transition hover:bg-coral-50 disabled:opacity-50"
        @click="handleRemove"
      >
        Remover plano (voltar ao Gratuito)
      </button>

      <p v-if="error" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ error }}
      </p>

      <div class="flex justify-end border-t border-ink-100 pt-4">
        <button
          type="button"
          class="rounded-full border border-ink-200 px-5 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-50"
          @click="emit('close')"
        >
          Fechar
        </button>
      </div>
    </div>
  </BaseModal>
</template>
