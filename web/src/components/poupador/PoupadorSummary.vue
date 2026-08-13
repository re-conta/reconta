<script setup lang="ts">
import { computed } from "vue";
import { ArrowDownRight, ArrowUpRight, PiggyBank, Wallet } from "lucide-vue-next";

const props = defineProps<{
  income: number;
  expense: number;
  balance: number;
  yearlyBalance: number;
}>();
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const balanceDescription = computed(() =>
  props.balance >= 0 ? "Disponível para guardar por mês" : "Ajuste seus gastos ou receitas",
);
</script>

<template>
  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <article class="rounded-3xl border border-brand-200 bg-brand-50 p-5">
      <div class="flex items-center gap-2 text-sm text-brand-800">
        <ArrowUpRight class="mb-5 h-5 w-5 text-brand-700" />
        <div class="text-xs font-bold text-brand-800">Receitas mensais</div>
      </div>
      <p class="mt-1 font-display text-2xl font-bold text-ink-900">{{ currency.format(income) }}</p>
    </article>
    <article class="rounded-3xl border border-coral-200 bg-coral-50 p-5">
      <div class="flex items-center gap-2 text-sm text-brand-800">
        <ArrowDownRight class="mb-5 h-5 w-5 text-coral-700" />
        <div class="text-xs font-bold text-coral-800">Gastos mensais</div>
      </div>
      <p class="mt-1 font-display text-2xl font-bold text-ink-900">
        {{ currency.format(expense) }}
      </p>
    </article>
    <article class="rounded-3xl border p-5"
      :class="balance >= 0 ? 'border-brand-300 bg-white' : 'border-coral-300 bg-white'">
      <div class="flex items-center gap-2 text-sm text-brand-800">
        <PiggyBank class="mb-5 h-5 w-5" :class="balance >= 0 ? 'text-brand-700' : 'text-coral-700'" />
        <div class="text-xs font-bold text-ink-600">Saldo mensal</div>
      </div>
      <p class="mt-1 font-display text-2xl font-bold text-ink-900">
        {{ currency.format(balance) }}
      </p>
      <p class="mt-1 text-xs text-ink-500">{{ balanceDescription }}</p>
    </article>
    <article class="rounded-3xl bg-ink-900 p-5 text-white">
      <div class="flex items-center gap-2 text-sm text-brand-800">
        <Wallet class="mb-5 h-5 w-5 text-brand-300" />
        <div class="text-xs font-bold text-ink-300">Saldo projetado no ano</div>
      </div>
      <p class="mt-1 font-display text-2xl font-bold">{{ currency.format(yearlyBalance) }}</p>
      <p class="mt-1 text-xs text-ink-300">Inclui valores únicos no mês escolhido</p>
    </article>
  </section>
</template>
