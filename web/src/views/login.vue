<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ApiError } from "../api/users";
import { useAuth } from "../composables/useAuth";

const router = useRouter();
const { login } = useAuth();

const form = reactive({ email: "", password: "" });
const errorMessage = ref("");
const submitting = ref(false);

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    await login(form.email, form.password);
    router.push("/");
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao entrar";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-[calc(100vh-140px)] items-center justify-center px-6 py-12">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex flex-col items-center text-center">
        <img src="/images/favicon.svg" alt="" class="h-12 w-12" />
        <h1 class="mt-4 font-display text-2xl font-bold text-ink-900">Bem-vindo de volta</h1>
        <p class="mt-1 text-sm text-ink-500">Acesse sua conta Reconta</p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-8 shadow-xl shadow-ink-900/5">
        <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">E-mail</span>
            <input
              v-model="form.email"
              type="email"
              placeholder="voce@exemplo.com"
              required
              autocomplete="username"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Senha</span>
            <input
              v-model="form.password"
              type="password"
              placeholder="••••••••"
              required
              autocomplete="current-password"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <button
            type="submit"
            :disabled="submitting"
            class="mt-2 rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Entrando..." : "Entrar" }}
          </button>
        </form>

        <p v-if="errorMessage" class="mt-4 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>

        <div class="mt-6 flex items-center gap-3 text-xs text-ink-400">
          <span class="h-px flex-1 bg-ink-200"></span>
          ou
          <span class="h-px flex-1 bg-ink-200"></span>
        </div>

        <a
          href="/api/auth/google/login"
          class="mt-4 flex items-center justify-center gap-2 rounded-full border border-ink-200 bg-white px-4 py-2.5 text-sm font-semibold text-ink-700 shadow-sm transition hover:bg-ink-100"
        >
          Continuar com Google
        </a>

        <p class="mt-6 text-center text-sm text-ink-500">
          Não tem conta?
          <RouterLink to="/register" class="font-semibold text-brand-700 hover:text-brand-800">
            Cadastre-se
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>
