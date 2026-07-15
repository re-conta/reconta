<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ApiError, createUser } from "../api/users";
import PasswordInput from "../components/PasswordInput.vue";
import { formatCnpj, isValidCnpj, normalizeCnpj } from "../utils/cnpj";
import type { UserRole } from "../types/user";

const router = useRouter();

const accountTypes: { value: UserRole; label: string; description: string }[] = [
  { value: "pessoa_fisica", label: "Pessoa Física", description: "Para uso pessoal" },
  { value: "pessoa_juridica", label: "Pessoa Jurídica", description: "Para empresas (CNPJ)" },
  { value: "contador", label: "Contador", description: "Contador ou Técnico Contábil" },
];

const form = reactive({
  name: "",
  email: "",
  password: "",
  role: "pessoa_fisica" as UserRole,
  cnpj: "",
});
const errorMessage = ref("");
const submitting = ref(false);

const isPessoaJuridica = computed(() => form.role === "pessoa_juridica");
const cnpjTouched = ref(false);
const cnpjInvalid = computed(
  () => cnpjTouched.value && form.cnpj.length > 0 && !isValidCnpj(form.cnpj),
);

function handleCnpjInput(event: Event) {
  const target = event.target as HTMLInputElement;
  form.cnpj = formatCnpj(target.value);
  target.value = form.cnpj;
}

async function handleSubmit() {
  errorMessage.value = "";

  if (isPessoaJuridica.value && !isValidCnpj(form.cnpj)) {
    cnpjTouched.value = true;
    errorMessage.value = "Informe um CNPJ válido para contas Pessoa Jurídica.";
    return;
  }

  submitting.value = true;
  try {
    await createUser({
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password.trim(),
      role: form.role,
      cnpj: isPessoaJuridica.value ? normalizeCnpj(form.cnpj) : undefined,
    });
    router.push("/login");
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao cadastrar usuário";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="flex items-center justify-center px-2 md:py-12">
    <div class="w-full max-w-sm">
      <div class="mb-4 md:mb-8 flex flex-col items-center text-center">
        <h1 class="mt-2 md:mt-4 font-display text-2xl font-bold text-ink-900">Crie sua conta</h1>
        <p class="md:mt-1 text-sm text-ink-500">Comece a organizar suas finanças</p>
      </div>

      <div class="rounded-3xl border border-ink-200/70 bg-white p-4 md:p-8 shadow-xl shadow-ink-900/5">
        <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
          <fieldset class="flex flex-col gap-1.5">
            <legend class="text-sm font-medium text-ink-700">Tipo de conta</legend>
            <div class="mt-1.5 grid grid-cols-3 gap-2">
              <label
                v-for="type in accountTypes"
                :key="type.value"
                class="flex cursor-pointer flex-col items-center gap-0.5 rounded-xl border px-2 py-2.5 text-center transition"
                :class="
                  form.role === type.value
                    ? 'border-brand-400 bg-brand-50 ring-2 ring-brand-100'
                    : 'border-ink-200 bg-ink-50/50 hover:border-ink-300'
                "
              >
                <input v-model="form.role" type="radio" :value="type.value" class="sr-only" />
                <span class="text-xs font-semibold text-ink-900">{{ type.label }}</span>
                <span class="text-[10px] leading-tight text-ink-500">{{ type.description }}</span>
              </label>
            </div>
          </fieldset>
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
          <label v-if="isPessoaJuridica" class="flex flex-col gap-1.5">
            <span class="text-sm font-medium text-ink-700">CNPJ</span>
            <input
              :value="form.cnpj"
              type="text"
              inputmode="numeric"
              placeholder="00.000.000/0000-00"
              required
              maxlength="18"
              class="rounded-xl border bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:bg-white focus:ring-4"
              :class="
                cnpjInvalid
                  ? 'border-coral-400 focus:border-coral-400 focus:ring-coral-100'
                  : 'border-ink-200 focus:border-brand-400 focus:ring-brand-100'
              "
              @input="handleCnpjInput"
              @blur="cnpjTouched = true"
            />
            <span v-if="cnpjInvalid" class="text-xs text-coral-600">CNPJ inválido</span>
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
            <PasswordInput
              v-model="form.password"
              placeholder="Mínimo 8 caracteres"
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
