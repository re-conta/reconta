import { ref } from "vue";
import { getUnreadCount } from "../api/notifications";
import type { Notification } from "../types/notification";

const unreadCount = ref(0);
const latest = ref<Notification | null>(null);

let eventSource: EventSource | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let shouldStayConnected = false;

async function refreshUnreadCount() {
  try {
    unreadCount.value = await getUnreadCount();
  } catch {
    // silencioso: o contador só é relevante enquanto autenticado
  }
}

function scheduleReconnect() {
  if (reconnectTimer || !shouldStayConnected) return;
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    if (shouldStayConnected) connect();
  }, 5000);
}

function connect() {
  if (eventSource || !shouldStayConnected) return;

  refreshUnreadCount();

  eventSource = new EventSource("/api/notifications/stream");
  eventSource.onmessage = (event) => {
    try {
      const notification = JSON.parse(event.data) as Notification;
      latest.value = notification;
      unreadCount.value += 1;
    } catch {
      // ignora eventos mal formados (ex.: comentários de keep-alive)
    }
  };
  eventSource.onerror = () => {
    eventSource?.close();
    eventSource = null;
    scheduleReconnect();
  };
}

function disconnect() {
  shouldStayConnected = false;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  eventSource?.close();
  eventSource = null;
  unreadCount.value = 0;
  latest.value = null;
}

function start() {
  shouldStayConnected = true;
  connect();
}

export function useNotifications() {
  return { unreadCount, latest, start, disconnect, refreshUnreadCount };
}
