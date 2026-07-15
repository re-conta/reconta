<script setup lang="ts">
import { computed, ref } from "vue";
import { CircleDollarSign, Hourglass, XCircle } from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";
import { ApiError, cancelSubscription } from "../../api/billing";
import type { CancelResult, Subscription } from "../../types/billing";

const props = defineProps<{
  subscription: Subscription;
}>();

const emit = defineEmits<{ close: []; canceled: [result: CancelResult] }>();

const mode = ref<"refund" | "end_of_cycle">("end_of_cycle");
const submitting = ref(false);
const error = ref("");

const periodEndLabel = computed(() =>
  props.subscription.currentPeriodEnd
    ? new Date(props.subscription.currentPeriodEnd).toLocaleDateString("pt-BR")
    : "-",
);

// Estimativa exibida ao usuário; o valor final é calculado pelo servidor no
// momento do cancelamento.
const estimatedRefund = computed(() => {
  const end = props.subscription.currentPeriodEnd;
  if (!end) return 0;
  const endDate = new Date(end);
  const start = new Date(endDate);
  if (props.subscription.cycle === "yearly") start.setFullYear(start.getFullYear() - 1);
  else start.setMonth(start.getMonth() - 1);

  const now = Date.now();
  const total = endDate.getTime() - start.getTime();
  const remaining = endDate.getTime() - now;
  if (total <= 0 || remaining <= 0) return 0;
  return Math.max(0, (remaining / total) * 100);
});

async function handleConfirm() {
  submitting.value = true;
  error.value = "";
  try {
    const result = await cancelSubscription(mode.value);
    emit("canceled", result);
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao cancelar a assinatura";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <BaseModal
    title="Cancelar assinatura"
    :subtitle="`Plano ${subscription.planName}`"
    :icon="XCircle"
    @close="emit('close')"
  >
    <div class="flex flex-col gap-4">
      <p class="text-sm text-ink-600">Como você prefere encerrar sua assinatura?</p>

      <label
        class="flex cursor-pointer items-start gap-3 rounded-2xl border p-4 transition"
        :class="
          mode === 'end_of_cycle'
            ? 'border-brand-400 bg-brand-50/60 ring-4 ring-brand-100'
            : 'border-ink-200 hover:border-ink-300'
        "
      >
        <input v-model="mode" type="radio" value="end_of_cycle" class="mt-1 accent-brand-600" />
        <div class="flex items-start gap-3">
          <Hourglass class="mt-0.5 h-5 w-5 shrink-0 text-brand-600" />
          <div>
            <p class="text-sm font-semibold text-ink-900">Usar até o fim do ciclo</p>
            <p class="mt-0.5 text-xs text-ink-500">
              Sem reembolso. Você mantém todos os benefícios até
              <strong>{{ periodEndLabel }}</strong> e a assinatura não renova.
            </p>
          </div>
        </div>
      </label>

      <label
        class="flex cursor-pointer items-start gap-3 rounded-2xl border p-4 transition"
        :class="
          mode === 'refund'
            ? 'border-brand-400 bg-brand-50/60 ring-4 ring-brand-100'
            : 'border-ink-200 hover:border-ink-300'
        "
      >
        <input v-model="mode" type="radio" value="refund" class="mt-1 accent-brand-600" />
        <div class="flex items-start gap-3">
          <CircleDollarSign class="mt-0.5 h-5 w-5 shrink-0 text-brand-600" />
          <div>
            <p class="text-sm font-semibold text-ink-900">Cancelar agora com reembolso parcial</p>
            <p class="mt-0.5 text-xs text-ink-500">
              O acesso termina imediatamente e devolvemos o valor proporcional ao tempo não usado
              (~{{ estimatedRefund.toFixed(0) }}% do ciclo) pelo Mercado Pago.
            </p>
          </div>
        </div>
      </label>

      <p v-if="error" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ error }}
      </p>

      <div class="flex justify-end gap-3 border-t border-ink-100 pt-4">
        <button
          type="button"
          class="rounded-full border border-ink-200 px-5 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-50"
          @click="emit('close')"
        >
          Voltar
        </button>
        <button
          type="button"
          :disabled="submitting"
          class="rounded-full bg-coral-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-coral-700 disabled:opacity-50"
          @click="handleConfirm"
        >
          {{ submitting ? "Cancelando..." : "Confirmar cancelamento" }}
        </button>
      </div>
    </div>
  </BaseModal>
</template>
