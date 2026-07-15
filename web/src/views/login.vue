<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ApiError } from "../api/users";
import PasswordInput from "../components/PasswordInput.vue";
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
    await login(form.email.trim(), form.password.trim());
    const redirect = router.currentRoute.value.query.redirect;
    router.push(typeof redirect === "string" ? redirect : "/");
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao entrar";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex items-center justify-center px-2 md:px-6 md:py-12">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex flex-col items-center text-center">
        <h1 class="mt-4 font-display text-2xl font-bold text-ink-900">Bem-vindo de volta</h1>
        <p class="mt-1 text-sm text-ink-500">Acesse sua conta</p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 md:p-8 shadow-xl shadow-ink-900/5">
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
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-ink-700">Senha</span>
              <RouterLink
                to="/esqueci-senha"
                class="text-xs font-semibold text-brand-700 hover:text-brand-800"
              >
                Esqueceu sua senha?
              </RouterLink>
            </div>
            <PasswordInput
              v-model="form.password"
              placeholder="••••••••"
              required
              autocomplete="current-password"
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
          <svg class="h-5 w-5" viewBox="0 0 48 48" aria-hidden="true">
            <path
              fill="#FFC107"
              d="M43.611 20.083H42V20H24v8h11.303c-1.649 4.657-6.08 8-11.303 8-6.627 0-12-5.373-12-12s5.373-12 12-12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 12.955 4 4 12.955 4 24s8.955 20 20 20 20-8.955 20-20c0-1.341-.138-2.65-.389-3.917z"
            />
            <path
              fill="#FF3D00"
              d="M6.306 14.691l6.571 4.819C14.655 15.108 18.961 12 24 12c3.059 0 5.842 1.154 7.961 3.039l5.657-5.657C34.046 6.053 29.268 4 24 4 16.318 4 9.656 8.337 6.306 14.691z"
            />
            <path
              fill="#4CAF50"
              d="M24 44c5.166 0 9.86-1.977 13.409-5.192l-6.19-5.238A11.91 11.91 0 0 1 24 36c-5.202 0-9.619-3.317-11.283-7.946l-6.522 5.025C9.505 39.556 16.227 44 24 44z"
            />
            <path
              fill="#1976D2"
              d="M43.611 20.083H42V20H24v8h11.303a12.04 12.04 0 0 1-4.087 5.571l.003-.002 6.19 5.238C36.971 39.205 44 34 44 24c0-1.341-.138-2.65-.389-3.917z"
            />
          </svg>
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
