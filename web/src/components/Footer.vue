<script setup lang="ts">
import SiteModals from "../components/modals/SiteModals.vue";
import { useSiteModals } from "../composables/useSiteModals";

const { open: openSiteModal } = useSiteModals();
const appName = import.meta.env.VITE_APP_NAME;
</script>
<template>
  <footer
    class="flex flex-col items-center gap-2.5 border-t border-ink-200/70 p-2 md:p-4 text-center text-sm text-ink-400"
  >
    <nav class="flex flex-wrap items-center justify-center gap-x-2 md:gap-x-5 gap-y-1.5 text-xs font-medium">
      <button type="button" class="transition hover:text-ink-700" @click="openSiteModal('help')">
        Ajuda
      </button>
      <button type="button" class="transition hover:text-ink-700" @click="openSiteModal('privacy')">
        Privacidade
      </button>
      <button type="button" class="transition hover:text-ink-700" @click="openSiteModal('terms')">
        Termos
      </button>
    </nav>
    <p class="text-[10px]">&copy; {{ new Date().getFullYear() }} {{ appName }}. Todos os direitos reservados.</p>
  </footer>
  <SiteModals />
</template>
