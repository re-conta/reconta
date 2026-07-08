<script setup lang="ts">
import { useRouter } from "vue-router";
import { Bell } from "lucide-vue-next";
import { useNotifications } from "../composables/useNotifications";

const router = useRouter();
const { unreadCount } = useNotifications();
</script>

<template>
  <button
    type="button"
    class="relative rounded-full border border-ink-200 bg-white p-2 text-ink-700 shadow-sm transition hover:bg-ink-50"
    aria-label="Ver notificações"
    @click="router.push('/notificacoes')"
  >
    <Bell class="h-4.5 w-4.5" />
    <span
      v-if="unreadCount > 0"
      class="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-coral-500 px-1 text-[11px] font-bold text-white"
    >
      {{ unreadCount > 99 ? "99+" : unreadCount }}
    </span>
  </button>
</template>
