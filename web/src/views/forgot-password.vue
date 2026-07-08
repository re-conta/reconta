<script setup lang="ts">
import { ref } from "vue";
import { forgotPassword } from "../api/auth";
import { ApiError } from "../api/users";

const email = ref("");
const errorMessage = ref("");
const submitting = ref(false);
const sent = ref(false);

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    await forgotPassword(email.value.trim());
    sent.value = true;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao enviar o e-mail";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-[calc(100vh-140px)] items-center justify-center px-6 py-4 md:py-12">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex flex-col items-center text-center">
        <h1 class="mt-4 font-display text-2xl font-bold text-ink-900">Esqueceu sua senha?</h1>
        <p class="mt-1 text-sm text-ink-500">
          Informe seu e-mail para receber um link de redefinição
        </p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-8 shadow-xl shadow-ink-900/5">
        <template v-if="sent">
          <p class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
            Se este e-mail estiver cadastrado, você receberá um link para redefinir sua senha em
            instantes.
          </p>
        </template>
        <form v-else class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">E-mail</span>
            <input
              v-model="email"
              type="email"
              placeholder="voce@exemplo.com"
              required
              autocomplete="username"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <button
            type="submit"
            :disabled="submitting"
            class="mt-2 rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Enviando..." : "Enviar link" }}
          </button>
        </form>

        <p v-if="errorMessage" class="mt-4 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>

        <p class="mt-6 text-center text-sm text-ink-500">
          Lembrou a senha?
          <RouterLink to="/login" class="font-semibold text-brand-700 hover:text-brand-800">
            Entrar
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>
