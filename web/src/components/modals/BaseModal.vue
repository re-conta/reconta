<script setup lang="ts">
import { onBeforeUnmount, onMounted } from "vue";
import { X } from "lucide-vue-next";
import type { Component } from "vue";

const props = defineProps<{
  title: string;
  subtitle?: string;
  icon?: Component;
}>();

const emit = defineEmits<{ close: [] }>();

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") emit("close");
}

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
  document.body.style.overflow = "hidden";
});

onBeforeUnmount(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        class="fixed inset-0 z-[100] flex items-end justify-center bg-ink-900/50 backdrop-blur-sm p-0 sm:items-center sm:p-4"
        @mousedown.self="emit('close')"
      >
        <Transition
          appear
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 translate-y-4 sm:scale-95"
          enter-to-class="opacity-100 translate-y-0 sm:scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0 sm:scale-100"
          leave-to-class="opacity-0 translate-y-4 sm:scale-95"
        >
          <div
            class="relative flex max-h-[90vh] w-full max-w-2xl flex-col overflow-hidden rounded-t-3xl bg-white shadow-2xl ring-1 ring-ink-900/5 sm:max-h-[85vh] sm:rounded-3xl"
            role="dialog"
            aria-modal="true"
            :aria-label="props.title"
          >
            <div
              class="relative shrink-0 overflow-hidden border-b border-ink-100 bg-linear-to-br from-brand-50 via-white to-coral-50 px-6 py-6 sm:px-8 sm:py-7"
            >
              <div
                class="pointer-events-none absolute -right-10 -top-16 h-40 w-40 rounded-full bg-linear-to-br from-brand-200/60 to-coral-200/50 blur-2xl"
              ></div>
              <div class="relative flex items-start justify-between gap-4">
                <div class="flex items-center gap-3">
                  <span
                    v-if="props.icon"
                    class="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-linear-to-br from-brand-400 to-coral-500 text-white shadow-sm"
                  >
                    <component :is="props.icon" class="h-5.5 w-5.5" />
                  </span>
                  <div>
                    <h2 class="font-display text-xl font-bold text-ink-900 sm:text-2xl">
                      {{ props.title }}
                    </h2>
                    <p v-if="props.subtitle" class="mt-0.5 text-sm text-ink-500">
                      {{ props.subtitle }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="rounded-full p-2 text-ink-400 transition hover:bg-ink-900/5 hover:text-ink-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-400"
                  aria-label="Fechar"
                  @click="emit('close')"
                >
                  <X class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div class="grow overflow-y-auto px-6 py-6 sm:px-8 sm:py-7">
              <slot />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
