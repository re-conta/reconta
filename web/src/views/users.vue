<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useAuth } from "../composables/useAuth";
import { ApiError, listUsers, updateUserRole } from "../api/users";
import type { User, UserRole } from "../types/user";

const { currentUser } = useAuth();

const users = ref<User[]>([]);
const errorMessage = ref("");
const loading = ref(true);
const roleUpdatingId = ref<number | null>(null);

const isSuperAdmin = computed(() => currentUser.value?.role === "super_admin");

const roleLabels: Record<UserRole, string> = {
  user: "Usuário",
  admin: "Admin",
  super_admin: "Super Admin",
};

async function loadUsers() {
  loading.value = true;
  errorMessage.value = "";
  try {
    users.value = await listUsers();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar usuários";
  } finally {
    loading.value = false;
  }
}

async function handleRoleChange(target: User, role: UserRole) {
  roleUpdatingId.value = target.id;
  errorMessage.value = "";
  try {
    const updated = await updateUserRole(target.id, role);
    const idx = users.value.findIndex((u) => u.id === target.id);
    if (idx !== -1) users.value[idx] = updated;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao atualizar role";
  } finally {
    roleUpdatingId.value = null;
  }
}

const gradients = [
  "from-brand-400 to-coral-500",
  "from-coral-400 to-brand-500",
  "from-brand-500 to-brand-300",
  "from-coral-500 to-coral-300",
];

function gradientFor(id: string | number) {
  const idx = String(id)
    .split("")
    .reduce((sum, ch) => sum + ch.charCodeAt(0), 0);
  return gradients[idx % gradients.length];
}

function initialsFor(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]!.toUpperCase())
    .join("");
}

const userCountLabel = computed(() => {
  const count = users.value.length;
  return count === 1 ? "1 usuário" : `${count} usuários`;
});

onMounted(loadUsers);
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-10 sm:py-14">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Usuários</h1>
        <p class="mt-0.5 text-sm text-ink-500">{{ userCountLabel }}</p>
      </div>
      <RouterLink
        to="/register"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
      >
        + Cadastrar
      </RouterLink>
    </div>

    <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span
          class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">
        {{ errorMessage }}
      </p>
      <div v-else-if="users.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">Nenhum usuário cadastrado ainda</p>
        <p class="text-sm text-ink-400">Cadastre o primeiro usuário para começar.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li
          v-for="user in users"
          :key="user.id"
          class="flex items-center gap-3 px-5 py-4 transition hover:bg-ink-50/60"
        >
          <span
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-linear-to-br text-sm font-semibold text-white shadow-sm"
            :class="gradientFor(user.id)"
          >
            {{ initialsFor(user.name) }}
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold text-ink-900">{{ user.name }}</p>
            <p class="truncate text-xs text-ink-500">{{ user.email }}</p>
          </div>
          <select
            v-if="isSuperAdmin && user.role !== 'super_admin'"
            :value="user.role"
            :disabled="roleUpdatingId === user.id"
            class="shrink-0 rounded-full border border-ink-200 bg-white px-3 py-1.5 text-xs font-semibold text-ink-700 outline-none transition focus:border-brand-400 focus:ring-4 focus:ring-brand-100 disabled:opacity-50"
            @change="handleRoleChange(user, ($event.target as HTMLSelectElement).value as UserRole)"
          >
            <option value="user">Usuário</option>
            <option value="admin">Admin</option>
          </select>
          <span
            v-else
            class="shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold"
            :class="
              user.role === 'super_admin'
                ? 'bg-ink-900 text-white'
                : user.role === 'admin'
                  ? 'bg-brand-100 text-brand-700'
                  : 'bg-ink-100 text-ink-600'
            "
          >
            {{ roleLabels[user.role] }}
          </span>
        </li>
      </ul>
    </div>
  </div>
</template>
