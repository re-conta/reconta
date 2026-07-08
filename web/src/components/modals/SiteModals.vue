<script setup lang="ts">
import { useSiteModals } from "../../composables/useSiteModals";
import HelpModal from "./HelpModal.vue";
import PrivacyModal from "./PrivacyModal.vue";
import TermsModal from "./TermsModal.vue";

const { activeModal, close } = useSiteModals();
</script>

<template>
  <HelpModal v-if="activeModal === 'help'" @close="close" />
  <PrivacyModal v-else-if="activeModal === 'privacy'" @close="close" />
  <TermsModal v-else-if="activeModal === 'terms'" @close="close" />
</template>
