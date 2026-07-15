<script setup lang="ts">
import { watch } from "vue";
import NotificationBell from "../components/NotificationBell.vue";
import UserMenu from "../components/UserMenu.vue";
import Navbar from "../components/Navbar.vue";
import Footer from "../components/Footer.vue";
import { useAuth } from "../composables/useAuth";
import { useNotifications } from "../composables/useNotifications";

const { currentUser } = useAuth();
const { start: startNotifications, disconnect: disconnectNotifications } = useNotifications();
const appName = import.meta.env.VITE_APP_NAME;

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
  <div class="flex h-svh flex-col bg-ink-50">
    <div
      class="pointer-events-none fixed inset-x-0 top-0 -z-10 h-120 bg-[radial-gradient(60%_60%_at_15%_0%,--theme(--color-brand-200/60%),transparent),radial-gradient(50%_50%_at_85%_10%,--theme(--color-coral-200/50%),transparent)]"
    ></div>

    <header
      class="sticky top-0 z-50 flex flex-col gap-3 border-b border-ink-200/70 bg-ink-50/80 px-4 py-3 backdrop-blur-md sm:flex-row sm:items-center sm:justify-between sm:px-8"
    >
      <div class="flex items-center justify-between gap-4">
        <RouterLink to="/" class="flex items-center gap-1.5">
          <img src="/images/favicon.svg" alt="" class="h-8 w-8" />
          <span class="font-display text-xl md:text-2xl font-bold tracking-tight text-ink-900">{{
            appName
          }}</span>
        </RouterLink>
        <div class="flex items-center gap-3 sm:hidden">
          <NotificationBell v-if="currentUser" />
          <UserMenu v-if="currentUser" />
          <RouterLink
            v-else
            to="/login"
            class="rounded-full bg-ink-900 px-4 py-1.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          >
            Entrar
          </RouterLink>
        </div>
      </div>
      <div
        class="flex-col items-start gap-3 sm:flex-row sm:items-center sm:gap-8"
        :class="currentUser ? 'flex' : 'hidden sm:flex'"
      >
        <Navbar />
        <div class="hidden items-center gap-3 sm:flex">
          <NotificationBell v-if="currentUser" />
          <UserMenu v-if="currentUser" />
          <RouterLink
            v-else
            to="/login"
            class="rounded-full bg-ink-900 px-4 py-1.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          >
            Entrar
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="flex grow flex-col">
      <slot />
    </main>

    <Footer />
  </div>
</template>
