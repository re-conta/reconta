<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { AlertTriangle, CalendarClock, CheckCheck, Settings } from "lucide-vue-next";
import {
  ApiError,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
} from "../api/notifications";
import { useNotifications } from "../composables/useNotifications";
import type { Notification } from "../types/notification";

const { unreadCount, refreshUnreadCount, latest } = useNotifications();

const notifications = ref<Notification[]>([]);
const loading = ref(true);
const errorMessage = ref("");

async function loadNotifications() {
  loading.value = true;
  errorMessage.value = "";
  try {
    notifications.value = await listNotifications();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar notificações";
  } finally {
    loading.value = false;
  }
}

watch(latest, (notification) => {
  if (!notification) return;
  if (!notifications.value.some((n) => n.id === notification.id)) {
    notifications.value.unshift(notification);
  }
});

async function handleMarkRead(notification: Notification) {
  if (notification.readAt) return;
  try {
    await markNotificationRead(notification.id);
    notification.readAt = new Date().toISOString();
    await refreshUnreadCount();
  } catch {
    // ignora falha silenciosamente, o usuário pode tentar novamente
  }
}

async function handleMarkAllRead() {
  try {
    await markAllNotificationsRead();
    const now = new Date().toISOString();
    notifications.value.forEach((n) => {
      if (!n.readAt) n.readAt = now;
    });
    await refreshUnreadCount();
  } catch {
    // ignora
  }
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleString("pt-BR", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function dayLabel(value: string) {
  const date = new Date(value);
  const today = new Date();
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  const sameDay = (a: Date, b: Date) =>
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate();

  if (sameDay(date, today)) return "Hoje";
  if (sameDay(date, yesterday)) return "Ontem";
  return date.toLocaleDateString("pt-BR", { day: "2-digit", month: "long", year: "numeric" });
}

const groups = computed(() => {
  const map = new Map<string, Notification[]>();
  for (const n of notifications.value) {
    const key = dayLabel(n.createdAt);
    if (!map.has(key)) map.set(key, []);
    map.get(key)!.push(n);
  }
  return Array.from(map.entries());
});

onMounted(loadNotifications);
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-6 px-2 md:px-4 py-4 md:py-8">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Notificações</h1>
        <p class="mt-0.5 text-sm text-ink-500">Lembretes de contas fixas vencendo ou vencidas</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <button
          v-if="unreadCount > 0"
          type="button"
          class="flex items-center gap-1.5 rounded-full border border-ink-200 px-3.5 py-2 text-xs font-semibold text-ink-700 transition hover:bg-ink-100"
          @click="handleMarkAllRead"
        >
          <CheckCheck class="h-4 w-4" />
          Marcar tudo como lido
        </button>
        <RouterLink
          to="/configuracoes#notificacoes"
          class="flex items-center gap-1.5 rounded-full border border-ink-200 px-3.5 py-2 text-xs font-semibold text-ink-700 transition hover:bg-ink-100"
          aria-label="Configurar notificações"
        >
          <Settings class="h-4 w-4" />
          Configurar
        </RouterLink>
      </div>
    </div>

    <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
      <span
        class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
      ></span>
      Carregando...
    </div>
    <p
      v-else-if="errorMessage"
      class="rounded-3xl border border-ink-200/70 bg-white p-8 text-center text-sm text-coral-600"
    >
      {{ errorMessage }}
    </p>
    <div
      v-else-if="notifications.length === 0"
      class="flex flex-col items-center gap-1 rounded-3xl border border-ink-200/70 bg-white p-12 text-center shadow-sm"
    >
      <p class="text-sm font-medium text-ink-600">Nenhuma notificação por aqui</p>
      <p class="text-sm text-ink-400">Você será avisado quando uma conta fixa estiver vencendo.</p>
    </div>

    <div v-else class="flex flex-col gap-8">
      <div v-for="[label, items] in groups" :key="label" class="flex flex-col gap-3">
        <h2 class="text-xs font-semibold uppercase tracking-wide text-ink-400">{{ label }}</h2>
        <ol class="relative flex flex-col gap-4 border-l-2 border-ink-100 pl-6">
          <li
            v-for="notification in items"
            :key="notification.id"
            class="relative cursor-pointer rounded-2xl border border-ink-200/70 bg-white p-4 shadow-sm transition hover:border-brand-200"
            :class="{ 'bg-ink-50/40': notification.readAt }"
            @click="handleMarkRead(notification)"
          >
            <span
              class="absolute left-[-1.95rem] top-4 flex h-6 w-6 items-center justify-center rounded-full ring-4 ring-white"
              :class="
                notification.kind === 'bill_overdue'
                  ? 'bg-coral-100 text-coral-600'
                  : 'bg-brand-100 text-brand-600'
              "
            >
              <AlertTriangle v-if="notification.kind === 'bill_overdue'" class="h-3.5 w-3.5" />
              <CalendarClock v-else class="h-3.5 w-3.5" />
            </span>

            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-sm font-semibold text-ink-900">{{ notification.title }}</p>
                <p class="mt-0.5 text-sm text-ink-600">{{ notification.message }}</p>
                <p class="mt-1.5 text-xs text-ink-400">
                  {{ formatDateTime(notification.createdAt) }}
                </p>
              </div>
              <span
                v-if="!notification.readAt"
                class="mt-1 h-2 w-2 shrink-0 rounded-full bg-brand-500"
              ></span>
            </div>
          </li>
        </ol>
      </div>
    </div>
  </div>
</template>
