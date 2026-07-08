<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { resetPassword } from "../api/auth";
import { ApiError } from "../api/users";
import PasswordInput from "../components/PasswordInput.vue";

const route = useRoute();
const router = useRouter();

const token = typeof route.query.token === "string" ? route.query.token : "";

const form = reactive({ password: "", confirmPassword: "" });
const errorMessage = ref("");
const submitting = ref(false);
const done = ref(false);

async function handleSubmit() {
  errorMessage.value = "";

  if (!token) {
    errorMessage.value = "Link de redefinição inválido ou expirado";
    return;
  }
  if (form.password.trim() !== form.confirmPassword.trim()) {
    errorMessage.value = "As senhas não coincidem";
    return;
  }

  submitting.value = true;
  try {
    await resetPassword(token, form.password.trim());
    done.value = true;
    setTimeout(() => router.push("/login"), 2000);
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao redefinir a senha";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-[calc(100vh-140px)] items-center justify-center px-6 py-4 md:py-12">
    <div class="w-full max-w-sm">
      <div class="mb-8 flex flex-col items-center text-center">
        <h1 class="mt-4 font-display text-2xl font-bold text-ink-900">Redefinir senha</h1>
        <p class="mt-1 text-sm text-ink-500">Escolha uma nova senha para sua conta</p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-8 shadow-xl shadow-ink-900/5">
        <template v-if="done">
          <p class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
            Senha redefinida com sucesso. Redirecionando para o login...
          </p>
        </template>
        <template v-else-if="!token">
          <p class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
            Link de redefinição inválido ou expirado.
          </p>
        </template>
        <form v-else class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Nova senha</span>
            <PasswordInput
              v-model="form.password"
              placeholder="Mínimo 8 caracteres"
              :minlength="8"
              required
              autocomplete="new-password"
            />
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">Confirmar nova senha</span>
            <PasswordInput
              v-model="form.confirmPassword"
              :minlength="8"
              required
              autocomplete="new-password"
            />
          </label>
          <button
            type="submit"
            :disabled="submitting"
            class="mt-2 rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800 disabled:opacity-50"
          >
            {{ submitting ? "Salvando..." : "Redefinir senha" }}
          </button>
        </form>

        <p v-if="errorMessage" class="mt-4 rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
          {{ errorMessage }}
        </p>
      </div>
    </div>
  </div>
</template>
