<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuth } from "../composables/useAuth";
import { useSiteModals } from "../composables/useSiteModals";

const { currentUser, logout } = useAuth();
const { open: openSiteModal } = useSiteModals();
const router = useRouter();

const open = ref(false);
const rootEl = ref<HTMLElement | null>(null);

const initials = computed(() => {
  if (!currentUser.value) return "";
  return currentUser.value.name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]!.toUpperCase())
    .join("");
});

const isAdmin = computed(
  () => currentUser.value?.role === "admin" || currentUser.value?.role === "super_admin",
);

const avatarError = ref(false);
const avatarUrl = computed(() => (avatarError.value ? "" : currentUser.value?.avatarUrl || ""));

function handleAvatarError() {
  avatarError.value = true;
}

function toggle() {
  open.value = !open.value;
}

function handleHelp() {
  close();
  openSiteModal("help");
}

function close() {
  open.value = false;
}

function handleClickOutside(event: MouseEvent) {
  if (rootEl.value && !rootEl.value.contains(event.target as Node)) {
    close();
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") close();
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
  document.addEventListener("keydown", handleKeydown);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
  document.removeEventListener("keydown", handleKeydown);
});

async function handleLogout() {
  close();
  await logout();
  router.push({ name: "Home" });
}
</script>

<template>
  <div v-if="currentUser" ref="rootEl" class="relative">
    <button
      type="button"
      class="flex items-center gap-2 rounded-full py-1 pl-1 pr-2 transition hover:bg-ink-900/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-brand-400 sm:pr-3"
      :aria-expanded="open"
      aria-haspopup="true"
      @click="toggle"
    >
      <img
        v-if="avatarUrl"
        :src="avatarUrl"
        alt=""
        referrerpolicy="no-referrer"
        class="h-8 w-8 shrink-0 rounded-full object-cover shadow-sm"
        @error="handleAvatarError"
      />
      <span
        v-else
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-linear-to-br from-brand-400 to-coral-500 text-sm font-semibold text-white shadow-sm"
      >
        {{ initials }}
      </span>
      <span class="hidden max-w-32 truncate text-sm font-medium text-ink-800 sm:inline">
        {{ currentUser.name }}
      </span>
      <svg
        class="hidden h-4 w-4 shrink-0 text-ink-500 transition-transform sm:block"
        :class="{ 'rotate-180': open }"
        viewBox="0 0 20 20"
        fill="currentColor"
      >
        <path
          fill-rule="evenodd"
          d="M5.23 7.21a.75.75 0 0 1 1.06.02L10 10.94l3.71-3.71a.75.75 0 1 1 1.06 1.06l-4.24 4.24a.75.75 0 0 1-1.06 0L5.21 8.29a.75.75 0 0 1 .02-1.08Z"
          clip-rule="evenodd"
        />
      </svg>
    </button>

    <transition
      enter-active-class="transition ease-out duration-150"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition ease-in duration-100"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 -translate-y-1"
    >
      <div
        v-if="open"
        class="absolute right-0 z-50 mt-2 w-64 max-w-[calc(100vw-2rem)] origin-top-right overflow-hidden rounded-xl border border-ink-200 bg-white shadow-lg"
        role="menu"
      >
        <div class="flex items-center gap-3 border-b border-ink-100 px-4 py-3">
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            alt=""
            referrerpolicy="no-referrer"
            class="h-10 w-10 shrink-0 rounded-full object-cover"
            @error="handleAvatarError"
          />
          <span
            v-else
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-linear-to-br from-brand-400 to-coral-500 text-sm font-semibold text-white"
          >
            {{ initials }}
          </span>
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-ink-900">{{ currentUser.name }}</p>
            <p class="truncate text-xs text-ink-500">{{ currentUser.email }}</p>
          </div>
        </div>

        <nav class="py-1">
          <RouterLink
            v-if="isAdmin"
            to="/users"
            class="flex items-center gap-2 px-4 py-2 text-sm text-ink-700 transition hover:bg-ink-50"
            role="menuitem"
            @click="close"
          >
            <svg class="h-4 w-4 text-ink-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                d="M10 9a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-6 9a6 6 0 1 1 12 0 1 1 0 0 1-1 1H5a1 1 0 0 1-1-1Z"
              />
            </svg>
            Usuários
          </RouterLink>
          <RouterLink
            to="/configuracoes"
            class="flex items-center gap-2 px-4 py-2 text-sm text-ink-700 transition hover:bg-ink-50"
            role="menuitem"
            @click="close"
          >
            <svg class="h-4 w-4 text-ink-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.53 1.53 0 0 1-2.28.95c-1.37-.83-2.94.74-2.11 2.11a1.53 1.53 0 0 1-.94 2.28c-1.56.38-1.56 2.6 0 2.98a1.53 1.53 0 0 1 .94 2.28c-.83 1.37.74 2.94 2.11 2.11a1.53 1.53 0 0 1 2.28.94c.38 1.56 2.6 1.56 2.98 0a1.53 1.53 0 0 1 2.28-.94c1.37.83 2.94-.74 2.11-2.11a1.53 1.53 0 0 1 .95-2.28c1.56-.38 1.56-2.6 0-2.98a1.53 1.53 0 0 1-.95-2.28c.83-1.37-.74-2.94-2.11-2.11a1.53 1.53 0 0 1-2.28-.95ZM10 13a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z"
                clip-rule="evenodd"
              />
            </svg>
            Configurações
          </RouterLink>
          <RouterLink
            to="/exportar"
            class="flex items-center gap-2 px-4 py-2 text-sm text-ink-700 transition hover:bg-ink-50"
            role="menuitem"
            @click="close"
          >
            <svg class="h-4 w-4 text-ink-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M10 3a.75.75 0 0 1 .75.75v7.19l2.72-2.72a.75.75 0 1 1 1.06 1.06l-4 4a.75.75 0 0 1-1.06 0l-4-4a.75.75 0 0 1 1.06-1.06l2.72 2.72V3.75A.75.75 0 0 1 10 3ZM4 15.25A.75.75 0 0 1 4.75 14.5h10.5a.75.75 0 0 1 0 1.5H4.75a.75.75 0 0 1-.75-.75Z"
                clip-rule="evenodd"
              />
            </svg>
            Exportar
          </RouterLink>
          <RouterLink
            to="/importar"
            class="flex items-center gap-2 px-4 py-2 text-sm text-ink-700 transition hover:bg-ink-50"
            role="menuitem"
            @click="close"
          >
            <svg class="h-4 w-4 text-ink-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M10 17a.75.75 0 0 1-.75-.75V9.06l-2.72 2.72a.75.75 0 1 1-1.06-1.06l4-4a.75.75 0 0 1 1.06 0l4 4a.75.75 0 1 1-1.06 1.06L10.75 9.06v7.19A.75.75 0 0 1 10 17ZM4 4.75A.75.75 0 0 1 4.75 4h10.5a.75.75 0 0 1 0 1.5H4.75A.75.75 0 0 1 4 4.75Z"
                clip-rule="evenodd"
              />
            </svg>
            Importar
          </RouterLink>
        </nav>

        <div class="border-t border-ink-100 py-1">
          <button
            type="button"
            class="flex w-full items-center gap-2 px-4 py-2 text-sm text-ink-700 transition hover:bg-ink-50"
            role="menuitem"
            @click="handleHelp"
          >
            <svg class="h-4 w-4 text-ink-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0ZM8.94 6.94a.75.75 0 1 1-1.061-1.061 3 3 0 1 1 2.871 5.026v.345a.75.75 0 0 1-1.5 0v-.5c0-.72.57-1.172 1.081-1.287a1.5 1.5 0 1 0-1.868-1.523ZM10 15a1 1 0 1 0 0-2 1 1 0 0 0 0 2Z"
                clip-rule="evenodd"
              />
            </svg>
            Central de ajuda
          </button>
        </div>

        <div class="border-t border-ink-100 py-1">
          <button
            type="button"
            class="flex w-full items-center gap-2 px-4 py-2 text-sm text-coral-600 transition hover:bg-coral-50"
            role="menuitem"
            @click="handleLogout"
          >
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
              <path
                fill-rule="evenodd"
                d="M3 4.25A2.25 2.25 0 0 1 5.25 2h5.5A2.25 2.25 0 0 1 13 4.25v2a.75.75 0 0 1-1.5 0v-2a.75.75 0 0 0-.75-.75h-5.5a.75.75 0 0 0-.75.75v11.5c0 .414.336.75.75.75h5.5a.75.75 0 0 0 .75-.75v-2a.75.75 0 0 1 1.5 0v2A2.25 2.25 0 0 1 10.75 18h-5.5A2.25 2.25 0 0 1 3 15.75V4.25Z"
                clip-rule="evenodd"
              />
              <path
                fill-rule="evenodd"
                d="M6 10a.75.75 0 0 1 .75-.75h9.69l-2.72-2.72a.75.75 0 1 1 1.06-1.06l4 4a.75.75 0 0 1 0 1.06l-4 4a.75.75 0 1 1-1.06-1.06l2.72-2.72H6.75A.75.75 0 0 1 6 10Z"
                clip-rule="evenodd"
              />
            </svg>
            Sair
          </button>
        </div>
      </div>
    </transition>
  </div>
</template>
