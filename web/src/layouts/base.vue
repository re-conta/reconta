<script setup lang="ts">
import { watch } from "vue";
import NotificationBell from "../components/NotificationBell.vue";
import UserMenu from "../components/UserMenu.vue";
import Navbar from "../components/Navbar.vue";
import SiteModals from "../components/modals/SiteModals.vue";
import { useAuth } from "../composables/useAuth";
import { useNotifications } from "../composables/useNotifications";
import { useSiteModals } from "../composables/useSiteModals";

const { currentUser } = useAuth();
const { start: startNotifications, disconnect: disconnectNotifications } = useNotifications();
const { open: openSiteModal } = useSiteModals();
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
  <div class="flex min-h-screen flex-col bg-ink-50">
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
      <div class="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:gap-8">
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

    <main class="grow">
      <slot />
    </main>

    <footer
      class="flex flex-col items-center gap-3 border-t border-ink-200/70 px-4 py-6 text-center text-sm text-ink-400 sm:px-8"
    >
      <nav class="flex flex-wrap items-center justify-center gap-x-5 gap-y-1.5 text-xs font-medium">
        <button type="button" class="transition hover:text-ink-700" @click="openSiteModal('help')">
          Central de ajuda
        </button>
        <button
          type="button"
          class="transition hover:text-ink-700"
          @click="openSiteModal('privacy')"
        >
          Política de Privacidade
        </button>
        <button type="button" class="transition hover:text-ink-700" @click="openSiteModal('terms')">
          Termos de Uso
        </button>
      </nav>
      <p>&copy; {{ new Date().getFullYear() }} {{ appName }}. Todos os direitos reservados.</p>
    </footer>

    <SiteModals />
  </div>
</template>
