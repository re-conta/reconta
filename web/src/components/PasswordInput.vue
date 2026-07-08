<script setup lang="ts">
import { Eye, EyeOff } from "lucide-vue-next";
import { ref } from "vue";

defineProps<{
  modelValue: string;
  placeholder?: string;
  required?: boolean;
  minlength?: number;
  autocomplete?: string;
}>();

defineEmits<{ (e: "update:modelValue", value: string): void }>();

const visible = ref(false);
</script>

<template>
  <div class="relative">
    <input
      :value="modelValue"
      :type="visible ? 'text' : 'password'"
      :placeholder="placeholder"
      :required="required"
      :minlength="minlength"
      :autocomplete="autocomplete"
      class="w-full rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 pr-10 text-sm text-ink-900 outline-none transition placeholder:text-ink-400 focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <button
      type="button"
      tabindex="-1"
      class="absolute inset-y-0 right-0 flex items-center px-3 text-ink-400 transition hover:text-ink-600"
      :aria-label="visible ? 'Esconder senha' : 'Mostrar senha'"
      @click="visible = !visible"
    >
      <EyeOff v-if="visible" class="h-4 w-4" />
      <Eye v-else class="h-4 w-4" />
    </button>
  </div>
</template>
