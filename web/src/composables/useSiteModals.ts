import { ref } from "vue";

export type SiteModal = "help" | "privacy" | "terms" | null;

const activeModal = ref<SiteModal>(null);

export function useSiteModals() {
  function open(modal: Exclude<SiteModal, null>) {
    activeModal.value = modal;
  }

  function close() {
    activeModal.value = null;
  }

  return { activeModal, open, close };
}
