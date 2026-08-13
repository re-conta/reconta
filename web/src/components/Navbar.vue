<script setup lang="ts">
import { ref, watch } from "vue";
import { PiggyBank } from "lucide-vue-next";
import { useAuth } from "../composables/useAuth";
import { useNotifications } from "../composables/useNotifications";
import { navLinks } from "../config";

const links = ref(navLinks);

const { currentUser } = useAuth();
const { start: startNotifications, disconnect: disconnectNotifications } = useNotifications();
const isPoupadorHost = window.location.hostname === "poupa.reconta.app";

function linkClasses(path: string) {
  return path === "/poupa"
    ? "inline-flex items-center gap-1.5 rounded-full bg-brand-200 px-3 py-1.5 font-bold text-ink-900 shadow-sm ring-1 ring-brand-300 transition hover:bg-brand-300"
    : "transition hover:text-ink-900";
}

function linkTarget(path: string) {
  return path === "/poupa" && isPoupadorHost ? "/" : path;
}

watch(
  currentUser,
  (user) => {
    if (user) startNotifications();
    else disconnectNotifications();
  },
  { immediate: true },
);
</script>

<template>
  <nav
    class="no-scrollbar flex w-full flex-row flex-nowrap items-center gap-5 overflow-x-auto text-sm font-medium text-ink-500 sm:w-auto sm:gap-6 sm:overflow-visible"
  >
    <template v-for="link in links">
      <RouterLink
        v-if="!link.authRequired || currentUser"
        :key="link.path"
        :to="linkTarget(link.path)"
        :class="linkClasses(link.path)"
        class="shrink-0 whitespace-nowrap"
        :active-class="link.path === '/poupa' ? 'bg-brand-300 text-ink-900' : 'text-ink-900'"
      >
        <PiggyBank v-if="link.path === '/poupa'" class="h-3.5 w-3.5" />
        {{ link.name }}
      </RouterLink>
    </template>
  </nav>
</template>
