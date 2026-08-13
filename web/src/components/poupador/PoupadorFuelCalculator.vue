<script setup lang="ts">
import { computed } from "vue";
import { Fuel, Gauge, Route } from "lucide-vue-next";
import type {
  PoupadorDistancePeriod,
  PoupadorFuelInput,
  PoupadorFuelType,
} from "../../types/poupador";

const fuel = defineModel<PoupadorFuelInput>("fuel", { required: true });
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const decimal = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 1 });

const monthlyDistance = computed(() =>
  fuel.value.distancePeriod === "daily" ? fuel.value.distance * 30 : fuel.value.distance,
);
const litersPerMonth = computed(() => monthlyDistance.value / fuel.value.consumption);
const monthlyCost = computed(() => litersPerMonth.value * fuel.value.fuelPrice);
const costPerKilometer = computed(() => fuel.value.fuelPrice / fuel.value.consumption);
const hasValidInput = computed(
  () => fuel.value.fuelPrice > 0 && fuel.value.distance > 0 && fuel.value.consumption > 0,
);

function updateNumber(field: "fuelPrice" | "distance" | "consumption", event: Event) {
  const value = Number((event.target as HTMLInputElement).value);
  fuel.value = { ...fuel.value, [field]: Number.isFinite(value) && value >= 0 ? value : 0 };
}

function updateFuelType(event: Event) {
  fuel.value = {
    ...fuel.value,
    fuelType: (event.target as HTMLSelectElement).value as PoupadorFuelType,
  };
}

function updateDistancePeriod(event: Event) {
  fuel.value = {
    ...fuel.value,
    distancePeriod: (event.target as HTMLSelectElement).value as PoupadorDistancePeriod,
  };
}
</script>

<template>
  <section class="mt-7 rounded-3xl border border-brand-200 bg-brand-50 p-4 shadow-sm sm:p-5">
    <div class="mb-4 flex items-start gap-3">
      <span
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-brand-200 text-brand-800"
      >
        <Fuel class="h-5 w-5" />
      </span>
      <div>
        <h2 class="font-display text-base font-bold text-ink-900">Consumo de combustível</h2>
        <p class="text-xs text-ink-600">Estime o consumo e o custo mensal do seu deslocamento.</p>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <label class="min-w-0 text-xs font-bold text-ink-600">
        Combustível
        <select
          :value="fuel.fuelType"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-brand-200 bg-white px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
          @change="updateFuelType"
        >
          <option value="gasoline">Gasolina</option>
          <option value="diesel">Diesel</option>
        </select>
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600">
        Preço por litro (R$)
        <input
          :value="fuel.fuelPrice"
          min="0"
          step="0.01"
          type="number"
          inputmode="decimal"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-brand-200 bg-white px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
          @input="updateNumber('fuelPrice', $event)"
        />
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600">
        Distância média
        <input
          :value="fuel.distance"
          min="0"
          step="0.1"
          type="number"
          inputmode="decimal"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-brand-200 bg-white px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
          @input="updateNumber('distance', $event)"
        />
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600">
        Período da distância
        <select
          :value="fuel.distancePeriod"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-brand-200 bg-white px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
          @change="updateDistancePeriod"
        >
          <option value="daily">Por dia</option>
          <option value="monthly">Por mês</option>
        </select>
      </label>
      <label class="min-w-0 text-xs font-bold text-ink-600 sm:col-span-2 lg:col-span-1">
        Consumo (km/L)
        <input
          :value="fuel.consumption"
          min="0"
          step="0.1"
          type="number"
          inputmode="decimal"
          class="mt-1.5 w-full min-w-0 max-w-full rounded-xl border border-brand-200 bg-white px-3 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
          @input="updateNumber('consumption', $event)"
        />
      </label>
    </div>

    <div v-if="hasValidInput" class="mt-5 grid gap-3 sm:grid-cols-3">
      <article class="rounded-2xl bg-white p-4">
        <div class="mb-3 flex items-center gap-2 text-xs text-ink-600">
          <Route class="h-4 w-4 text-brand-700" />
          <div class="text-xs font-bold text-ink-600">Distância mensal</div>
        </div>
        <p class="mt-1 font-display text-xl font-bold text-ink-900">
          {{ decimal.format(monthlyDistance) }} km
        </p>
      </article>
      <article class="rounded-2xl bg-white p-4">
        <div class="mb-3 flex items-center gap-2 text-xs text-ink-600">
          <Fuel class="h-4 w-4 text-brand-700" />
          <div class="text-xs font-bold text-ink-600">Consumo mensal</div>
        </div>
        <p class="mt-1 font-display text-xl font-bold text-ink-900">
          {{ decimal.format(litersPerMonth) }} L
        </p>
      </article>
      <article class="rounded-2xl bg-ink-900 p-4 text-white">
        <div class="mb-3 flex items-center gap-2 text-xs text-ink-600">
          <Gauge class="h-4 w-4 text-brand-300" />
          <div class="text-xs font-bold text-ink-300">Custo mensal estimado</div>
        </div>
        <p class="mt-1 font-display text-xl font-bold">{{ currency.format(monthlyCost) }}</p>
        <p class="mt-1 text-xs text-ink-300">{{ currency.format(costPerKilometer) }} por km</p>
      </article>
    </div>
    <p v-else class="mt-4 text-xs text-ink-600">
      Informe o preço, a distância e o consumo para calcular a estimativa mensal.
    </p>
  </section>
</template>
