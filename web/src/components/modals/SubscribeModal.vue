<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from "vue";
import {
  Barcode,
  CheckCircle2,
  Copy,
  CreditCard,
  ExternalLink,
  Landmark,
  QrCode,
  Sparkles,
} from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";
import { ApiError, getPaymentStatus, subscribe } from "../../api/billing";
import { useMercadoPago } from "../../composables/useMercadoPago";
import { formatPrice } from "../../types/billing";
import type {
  BillingCycle,
  PaymentMethod,
  Plan,
  SubscriptionPayment,
} from "../../types/billing";

const props = defineProps<{
  plan: Plan;
  cycle: BillingCycle;
}>();

const emit = defineEmits<{ close: []; subscribed: [] }>();

const { detectPaymentMethod, createCardToken } = useMercadoPago();

const amount = computed(() =>
  props.cycle === "yearly" ? props.plan.priceYearly : props.plan.priceMonthly,
);
const cycleLabel = computed(() => (props.cycle === "yearly" ? "por ano" : "por mês"));

const methods: { id: PaymentMethod; label: string; icon: typeof QrCode; hint: string }[] = [
  { id: "pix", label: "PIX", icon: QrCode, hint: "Aprovação em segundos" },
  { id: "boleto", label: "Boleto", icon: Barcode, hint: "Compensa em até 3 dias" },
  { id: "debit_card", label: "Débito", icon: Landmark, hint: "Débito direto na conta" },
  { id: "credit_card", label: "Crédito", icon: CreditCard, hint: "Aprovação na hora" },
];

const method = ref<PaymentMethod>("pix");
const submitting = ref(false);
const error = ref("");
const payment = ref<SubscriptionPayment | null>(null);
const approved = ref(false);
const copied = ref(false);

const form = reactive({
  docType: "CPF" as "CPF" | "CNPJ",
  docNumber: "",
  zipCode: "",
  streetName: "",
  streetNumber: "",
  neighborhood: "",
  city: "",
  federalUnit: "",
  cardNumber: "",
  cardholderName: "",
  expiry: "",
  securityCode: "",
});

const isCard = computed(() => method.value === "debit_card" || method.value === "credit_card");
const detectedBrand = ref("");

function selectMethod(id: PaymentMethod) {
  if (payment.value) return;
  method.value = id;
  error.value = "";
}

async function detectBrand() {
  detectedBrand.value = "";
  const bin = form.cardNumber.replace(/\D/g, "").slice(0, 6);
  if (bin.length < 6 || !isCard.value) return;
  try {
    const found = await detectPaymentMethod(
      bin,
      method.value as "credit_card" | "debit_card",
    );
    if (found) detectedBrand.value = found.id;
  } catch {
    // Detecção é só conveniência visual; o submit tenta de novo.
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null;

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function startPolling() {
  stopPolling();
  pollTimer = setInterval(async () => {
    if (!payment.value) return;
    try {
      const updated = await getPaymentStatus(payment.value.id);
      payment.value = updated;
      if (updated.status === "approved") {
        approved.value = true;
        stopPolling();
        emit("subscribed");
      } else if (updated.status === "rejected" || updated.status === "cancelled") {
        stopPolling();
        error.value = "O pagamento não foi concluído. Tente novamente.";
        payment.value = null;
      }
    } catch {
      // Erro transitório de rede: a próxima rodada tenta de novo.
    }
  }, 5000);
}

onBeforeUnmount(stopPolling);

async function handleSubmit() {
  error.value = "";
  submitting.value = true;
  try {
    let token = "";
    let paymentMethodId = "";
    let issuerId = "";

    if (isCard.value) {
      const cardNumber = form.cardNumber.replace(/\D/g, "");
      const [month, year] = form.expiry.split("/").map((p) => p.trim());
      if (!month || !year) throw new ApiError("Informe a validade no formato MM/AA");

      const detected = await detectPaymentMethod(
        cardNumber.slice(0, 6),
        method.value as "credit_card" | "debit_card",
      );
      if (!detected) throw new ApiError("Não reconhecemos este cartão para o tipo escolhido");
      paymentMethodId = detected.id;
      if (detected.issuer) issuerId = String(detected.issuer.id);

      token = await createCardToken({
        cardNumber,
        cardholderName: form.cardholderName,
        cardExpirationMonth: month.padStart(2, "0"),
        cardExpirationYear: year.length === 2 ? `20${year}` : year,
        securityCode: form.securityCode,
        identificationType: form.docType,
        identificationNumber: form.docNumber.replace(/\D/g, ""),
      });
    }

    const result = await subscribe({
      planCode: props.plan.code,
      cycle: props.cycle,
      method: method.value,
      token,
      paymentMethodId,
      issuerId,
      installments: 1,
      docType: form.docType,
      docNumber: form.docNumber.replace(/\D/g, ""),
      zipCode: form.zipCode.replace(/\D/g, ""),
      streetName: form.streetName,
      streetNumber: form.streetNumber,
      neighborhood: form.neighborhood,
      city: form.city,
      federalUnit: form.federalUnit.toUpperCase(),
    });

    payment.value = result.payment;
    if (result.payment.status === "approved") {
      approved.value = true;
      emit("subscribed");
    } else if (result.payment.status === "rejected" || result.payment.status === "cancelled") {
      payment.value = null;
      error.value = "Pagamento recusado. Confira os dados do cartão ou tente outro método.";
    } else {
      startPolling();
    }
  } catch (err) {
    error.value =
      err instanceof ApiError || err instanceof Error
        ? err.message
        : "Falha ao processar o pagamento";
  } finally {
    submitting.value = false;
  }
}

async function copyPixCode() {
  if (!payment.value?.pixQr) return;
  try {
    await navigator.clipboard.writeText(payment.value.pixQr);
    copied.value = true;
    setTimeout(() => (copied.value = false), 2500);
  } catch {
    // Clipboard indisponível (ex.: contexto não seguro): o usuário ainda pode
    // escanear o QR Code.
  }
}
</script>

<template>
  <BaseModal
    :title="`Assinar plano ${plan.name}`"
    :subtitle="`${formatPrice(amount)} ${cycleLabel} · cancele quando quiser`"
    :icon="Sparkles"
    @close="emit('close')"
  >
    <!-- Sucesso -->
    <div v-if="approved" class="flex flex-col items-center gap-4 py-8 text-center">
      <span
        class="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-100 text-emerald-600"
      >
        <CheckCircle2 class="h-9 w-9" />
      </span>
      <div>
        <h3 class="font-display text-xl font-bold text-ink-900">Assinatura ativa!</h3>
        <p class="mt-1 text-sm text-ink-500">
          Pagamento confirmado. Bem-vindo(a) ao plano {{ plan.name }}!
        </p>
      </div>
      <button
        type="button"
        class="rounded-full bg-ink-900 px-6 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="emit('close')"
      >
        Começar a usar
      </button>
    </div>

    <!-- Aguardando pagamento PIX/Boleto -->
    <div v-else-if="payment" class="flex flex-col items-center gap-5 text-center">
      <template v-if="payment.method === 'pix'">
        <p class="text-sm text-ink-600">
          Escaneie o QR Code no app do seu banco ou copie o código PIX. A confirmação é
          automática.
        </p>
        <img
          v-if="payment.pixQrBase64"
          :src="`data:image/png;base64,${payment.pixQrBase64}`"
          alt="QR Code PIX"
          class="h-52 w-52 rounded-2xl border border-ink-200 bg-white p-2 shadow-sm"
        />
        <button
          type="button"
          class="flex items-center gap-2 rounded-full border border-ink-200 bg-white px-5 py-2.5 text-sm font-semibold text-ink-700 shadow-sm transition hover:border-brand-400 hover:text-brand-700"
          @click="copyPixCode"
        >
          <Copy class="h-4 w-4" />
          {{ copied ? "Código copiado!" : "Copiar código PIX" }}
        </button>
      </template>

      <template v-else-if="payment.method === 'boleto'">
        <p class="text-sm text-ink-600">
          Boleto gerado! Ele compensa em até 3 dias úteis — sua assinatura ativa assim que o
          pagamento for confirmado.
        </p>
        <a
          v-if="payment.ticketUrl"
          :href="payment.ticketUrl"
          target="_blank"
          rel="noopener"
          class="flex items-center gap-2 rounded-full bg-ink-900 px-6 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        >
          <ExternalLink class="h-4 w-4" />
          Abrir boleto
        </a>
      </template>

      <div class="flex items-center gap-2 text-xs text-ink-400">
        <span
          class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Aguardando confirmação do pagamento...
      </div>
    </div>

    <!-- Escolha de método + formulário -->
    <form v-else class="flex flex-col gap-5" @submit.prevent="handleSubmit">
      <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
        <button
          v-for="m in methods"
          :key="m.id"
          type="button"
          class="flex flex-col items-center gap-1.5 rounded-2xl border px-3 py-3.5 text-center transition"
          :class="
            method === m.id
              ? 'border-brand-400 bg-brand-50 text-brand-700 ring-4 ring-brand-100'
              : 'border-ink-200 bg-white text-ink-600 hover:border-ink-300'
          "
          @click="selectMethod(m.id)"
        >
          <component :is="m.icon" class="h-5 w-5" />
          <span class="text-sm font-semibold">{{ m.label }}</span>
          <span class="text-[11px] leading-tight text-ink-400">{{ m.hint }}</span>
        </button>
      </div>

      <!-- Documento (todos os métodos; obrigatório para boleto e cartão) -->
      <div class="grid gap-3 sm:grid-cols-3">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Documento</span>
          <select
            v-model="form.docType"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option value="CPF">CPF</option>
            <option value="CNPJ">CNPJ</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Número do {{ form.docType }}</span>
          <input
            v-model="form.docNumber"
            type="text"
            inputmode="numeric"
            :required="method !== 'pix'"
            :placeholder="form.docType === 'CPF' ? '000.000.000-00' : '00.000.000/0000-00'"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>

      <!-- Endereço (boleto) -->
      <div v-if="method === 'boleto'" class="grid gap-3 sm:grid-cols-6">
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">CEP</span>
          <input
            v-model="form.zipCode"
            type="text"
            inputmode="numeric"
            required
            placeholder="00000-000"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-3">
          <span class="text-sm font-medium text-ink-700">Rua</span>
          <input
            v-model="form.streetName"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-1">
          <span class="text-sm font-medium text-ink-700">Nº</span>
          <input
            v-model="form.streetNumber"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Bairro</span>
          <input
            v-model="form.neighborhood"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-3">
          <span class="text-sm font-medium text-ink-700">Cidade</span>
          <input
            v-model="form.city"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-1">
          <span class="text-sm font-medium text-ink-700">UF</span>
          <input
            v-model="form.federalUnit"
            type="text"
            required
            maxlength="2"
            placeholder="SP"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm uppercase outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>

      <!-- Cartão (débito/crédito) -->
      <div v-if="isCard" class="grid gap-3 sm:grid-cols-2">
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="flex items-center justify-between text-sm font-medium text-ink-700">
            Número do cartão
            <span v-if="detectedBrand" class="text-xs font-semibold uppercase text-brand-600">
              {{ detectedBrand }}
            </span>
          </span>
          <input
            v-model="form.cardNumber"
            type="text"
            inputmode="numeric"
            autocomplete="cc-number"
            required
            placeholder="0000 0000 0000 0000"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            @input="detectBrand"
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Nome impresso no cartão</span>
          <input
            v-model="form.cardholderName"
            type="text"
            autocomplete="cc-name"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm uppercase outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Validade</span>
          <input
            v-model="form.expiry"
            type="text"
            inputmode="numeric"
            autocomplete="cc-exp"
            required
            placeholder="MM/AA"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">CVV</span>
          <input
            v-model="form.securityCode"
            type="text"
            inputmode="numeric"
            autocomplete="cc-csc"
            required
            maxlength="4"
            placeholder="123"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>

      <p v-if="error" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ error }}
      </p>

      <div class="flex items-center justify-between gap-3 border-t border-ink-100 pt-4">
        <div>
          <p class="text-xs text-ink-400">Total</p>
          <p class="font-display text-lg font-bold text-ink-900">
            {{ formatPrice(amount) }}
            <span class="text-xs font-normal text-ink-400">{{ cycleLabel }}</span>
          </p>
        </div>
        <button
          type="submit"
          :disabled="submitting"
          class="rounded-full bg-ink-900 px-6 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{
            submitting
              ? "Processando..."
              : method === "pix"
                ? "Gerar QR Code"
                : method === "boleto"
                  ? "Gerar boleto"
                  : "Pagar agora"
          }}
        </button>
      </div>

      <p class="text-center text-xs text-ink-400">
        Pagamento processado com segurança pelo Mercado Pago. Cancele quando quiser, com
        reembolso proporcional ao tempo não usado.
      </p>
    </form>
  </BaseModal>
</template>
