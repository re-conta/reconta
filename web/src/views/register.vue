<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ApiError, createUser } from "../api/users";

const router = useRouter();

const form = reactive({ name: "", email: "", password: "" });
const errorMessage = ref("");
const submitting = ref(false);

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    await createUser({ ...form });
    router.push("/users");
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao cadastrar usuário";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-[calc(100vh-140px)] items-center justify-center py-4 md:py-12">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex flex-col items-center text-center">
        <h1 class="mt-4 font-display text-2xl font-bold text-ink-900">Crie sua conta</h1>
        <p class="mt-1 text-sm text-ink-500">Comece a organizar suas finanças</p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-8 shadow-xl shadow-ink-900/5">
        <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Nome</span>
            <input
              v-model="form.name"
              type="text"
              placeholder="Seu nome completo"
              required
              autocomplete="name"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
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
              placeholder="Mínimo 8 caracteres"
              minlength="8"
              required
              autocomplete="new-password"
              class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
            />
          </label>
          <button
            type="submit"
            :disabled="submitting"
            class="mt-2 rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Cadastrando..." : "Cadastrar" }}
          </button>
        </form>

        <p v-if="errorMessage" class="mt-4 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>

        <p class="mt-6 text-center text-sm text-ink-500">
          Já tem conta?
          <RouterLink to="/login" class="font-semibold text-brand-700 hover:text-brand-800">
            Entrar
          </RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>
