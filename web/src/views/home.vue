<script setup lang="ts">
import { ref, watchEffect } from "vue";
import { useAuth } from "../composables/useAuth";
import { listTransactions } from "../api/transactions";
import LazyImage from "../components/LazyImage.vue";

const { currentUser } = useAuth();
const appName = import.meta.env.VITE_APP_NAME;

const hasTransactions = ref(false);

watchEffect(async () => {
  if (!currentUser.value) return;
  try {
    const result = await listTransactions({ limit: 1 });
    hasTransactions.value = result.pagination.total > 0;
  } catch {
    hasTransactions.value = false;
  }
});
</script>

<template>
  <div
    class="mx-auto flex max-w-6xl flex-col items-center gap-8 px-2 md:px-6 py-4 md:py-10 sm:gap-12 sm:py-12 lg:flex-row lg:items-center lg:justify-between lg:py-32"
  >
    <div class="max-w-xl text-center lg:text-left">
      <span
        class="inline-flex items-center gap-2 rounded-full border border-brand-300/60 bg-brand-100/70 px-3 py-1 text-xs font-semibold text-brand-700"
      >
        Suas finanças, organizadas
      </span>
      <h1
        class="mt-5 font-display text-3xl font-bold leading-tight tracking-tight text-ink-900 sm:text-5xl"
      >
        {{ appName }} seu dinheiro com
        <span class="bg-linear-to-r from-brand-500 to-coral-500 bg-clip-text text-transparent"
          >clareza</span
        >
      </h1>
      <p class="mt-4 text-base leading-relaxed text-ink-500 sm:mt-5 sm:text-lg">
        Centralize contas, transações e categorias em um só lugar &mdash; sem planilhas, sem
        complicação.
      </p>
      <div class="mt-6 flex flex-col gap-3 sm:mt-8 sm:flex-row sm:justify-center lg:justify-start">
        <RouterLink
          v-if="!currentUser"
          to="/login"
          class="rounded-full bg-ink-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
        >
          Entre
        </RouterLink>
        <RouterLink
          v-else-if="!hasTransactions"
          to="/importar-extrato"
          class="rounded-full bg-ink-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
        >
          Comece
        </RouterLink>
        <RouterLink
          v-else
          to="/transacoes"
          class="rounded-full bg-ink-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
        >
          Transações
        </RouterLink>
      </div>
    </div>

    <div class="relative order-first w-4/5 max-w-xs shrink-0 sm:max-w-sm lg:order-last lg:w-full lg:max-w-md">
      <div
        class="absolute -inset-6 -z-10 rounded-full bg-linear-to-br from-brand-200 via-coral-100 to-transparent blur-2xl"
      ></div>

      <LazyImage
        src="/images/moneybag.svg"
        alt="Ilustração de um cofre de dinheiro"
        class="w-full drop-shadow-xl"
      />
    </div>
  </div>
</template>
