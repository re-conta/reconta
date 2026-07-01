<script setup lang="ts">
import { useAuth } from "../composables/useAuth";

const { currentUser } = useAuth();
</script>

<template>
  <div class="mx-auto flex max-w-6xl flex-col items-center gap-14 px-6 py-16 sm:py-24 lg:flex-row lg:items-center lg:justify-between lg:py-32">
    <div class="max-w-xl text-center lg:text-left">
      <span
        class="inline-flex items-center gap-2 rounded-full border border-brand-300/60 bg-brand-100/70 px-3 py-1 text-xs font-semibold text-brand-700"
      >
        Suas finanças, organizadas
      </span>
      <h1 class="mt-5 font-display text-4xl font-bold leading-tight tracking-tight text-ink-900 sm:text-5xl">
        Reconta seu dinheiro com
        <span class="bg-gradient-to-r from-brand-500 to-coral-500 bg-clip-text text-transparent">clareza</span>
      </h1>
      <p class="mt-5 text-lg leading-relaxed text-ink-500">
        Centralize contas, transações e categorias em um só lugar &mdash; sem planilhas, sem complicação.
      </p>
      <div class="mt-8 flex flex-col gap-3 sm:flex-row sm:justify-center lg:justify-start">
        <RouterLink
          v-if="!currentUser"
          to="/register"
          class="rounded-full bg-ink-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
        >
          Começar agora
        </RouterLink>
        <RouterLink
          v-else
          to="/users"
          class="rounded-full bg-ink-900 px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
        >
          Ver usuários
        </RouterLink>
        <RouterLink
          v-if="!currentUser"
          to="/login"
          class="rounded-full border border-ink-200 bg-white px-6 py-3 text-sm font-semibold text-ink-700 transition hover:border-ink-300 hover:bg-ink-100"
        >
          Já tenho conta
        </RouterLink>
      </div>
    </div>

    <div class="relative w-full max-w-sm shrink-0 lg:max-w-md">
      <div
        class="absolute -inset-6 -z-10 rounded-full bg-gradient-to-br from-brand-200 via-coral-100 to-transparent blur-2xl"
      ></div>
      <img src="/images/moneybag.svg" alt="Ilustração de um cofre de dinheiro" class="w-full drop-shadow-xl" />
    </div>
  </div>
</template>
