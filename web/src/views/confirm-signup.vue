<script setup lang="ts">
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ApiError, resendSignupOtp, verifySignupOtp } from "../api/users";
import { useAuth } from "../composables/useAuth";

const route = useRoute();
const router = useRouter();
const { setCurrentUser } = useAuth();

const email = typeof route.query.email === "string" ? route.query.email : "";

const code = ref("");
const errorMessage = ref("");
const infoMessage = ref("");
const submitting = ref(false);
const resending = ref(false);

async function handleSubmit() {
  errorMessage.value = "";
  infoMessage.value = "";

  if (!email) {
    errorMessage.value = "E-mail não informado, refaça o cadastro.";
    return;
  }

  submitting.value = true;
  try {
    const user = await verifySignupOtp(email, code.value.trim());
    setCurrentUser(user);
    router.push("/");
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao confirmar o código";
  } finally {
    submitting.value = false;
  }
}

async function handleResend() {
  errorMessage.value = "";
  infoMessage.value = "";

  if (!email) {
    errorMessage.value = "E-mail não informado, refaça o cadastro.";
    return;
  }

  resending.value = true;
  try {
    await resendSignupOtp(email);
    infoMessage.value = "Enviamos um novo código para o seu e-mail.";
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao reenviar o código";
  } finally {
    resending.value = false;
  }
}
</script>

<template>
  <div class="flex items-center justify-center px-2 py-2 md:py-4">
    <div class="w-full max-w-sm">
      <div class="mb-4 md:mb-8 flex flex-col items-center text-center">
        <h1 class="mt-2 md:mt-4 font-display text-2xl font-bold text-ink-900">Confirme seu e-mail</h1>
        <p class="md:mt-1 text-sm text-ink-500">
          Enviamos um código de confirmação para
          <strong class="text-ink-700">{{ email || "seu e-mail" }}</strong>
        </p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 md:p-8 shadow-xl shadow-ink-900/5">
        <template v-if="!email">
          <p class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
            Não encontramos o e-mail do cadastro.
            <RouterLink to="/register" class="font-semibold underline">Refaça o cadastro</RouterLink>.
          </p>
        </template>
        <form v-else class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Código de confirmação</span>
            <input
              v-model="code"
              type="text"
              inputmode="numeric"
              autocomplete="one-time-code"
              placeholder="000000"
              maxlength="6"
              required
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-center text-lg tracking-[0.5em] text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>

          <button
            type="submit"
            :disabled="submitting"
            class="mt-2 rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Confirmando..." : "Confirmar" }}
          </button>
        </form>

        <p v-if="errorMessage" class="mt-4 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>
        <p v-if="infoMessage" class="mt-4 rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
          {{ infoMessage }}
        </p>

        <p v-if="email" class="mt-6 text-center text-sm text-ink-500">
          Não recebeu o código?
          <button
            type="button"
            :disabled="resending"
            class="font-semibold text-brand-700 hover:text-brand-800 disabled:opacity-50"
            @click="handleResend"
          >
            {{ resending ? "Enviando..." : "Reenviar código" }}
          </button>
        </p>
      </div>
    </div>
  </div>
</template>
