<script setup lang="ts">
import AccountsManager from "../components/AccountsManager.vue";
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Contas</h1>
      <p class="mt-0.5 text-sm text-ink-500">Contas bancárias, cartões e investimentos</p>
    </div>

    <div class="rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm">
      <AccountsManager />
    </div>
  </div>
</template>
