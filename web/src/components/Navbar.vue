<script setup lang="ts">
import { ref, watch } from "vue";
import { useAuth } from "../composables/useAuth";
import { useNotifications } from "../composables/useNotifications";
import { navLinks } from "../config";

const links = ref(navLinks);

const { currentUser } = useAuth();
const { start: startNotifications, disconnect: disconnectNotifications } = useNotifications();

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
        :to="link.path"
        class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
        active-class="text-ink-900"
      >
        {{ link.name }}
      </RouterLink>
    </template>
  </nav>
</template>
