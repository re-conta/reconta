<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  modelValue: number;
  required?: boolean;
  placeholder?: string;
  class?: string;
  allowNegative?: boolean;
}>();

const emit = defineEmits<{ (e: "update:modelValue", value: number): void }>();

const formatter = new Intl.NumberFormat("pt-BR", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

function format(cents: number) {
  return formatter.format(cents / 100);
}

const displayValue = computed(() => format(Math.round(props.modelValue * 100)));

function onInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const negative = props.allowNegative && target.value.includes("-");
  const digits = target.value.replace(/\D/g, "");
  const cents = (digits ? parseInt(digits, 10) : 0) * (negative ? -1 : 1);
  target.value = format(cents);
  emit("update:modelValue", cents / 100);
}
</script>

<template>
  <input
    :value="displayValue"
    type="text"
    inputmode="numeric"
    :required="required"
    :placeholder="placeholder"
    :class="props.class"
    @input="onInput"
  />
</template>
