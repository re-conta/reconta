<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useAuth } from "../composables/useAuth";
import {
  ApiError,
  fetchRolePermissions,
  listUsers,
  updateRolePermissions,
  updateUserRole,
} from "../api/users";
import {
  ApiError as HealthApiError,
  getHealthSettings,
  updateHealthSettings,
} from "../api/health";
import type { HealthSettings } from "../types/health";
import { formatCnpj } from "../utils/cnpj";
import {
  permissionLabels,
  roleLabels,
  type Permission,
  type User,
  type UserRole,
} from "../types/user";

const { currentUser } = useAuth();

const isSuperAdmin = computed(() => currentUser.value?.role === "super_admin");
const canManageUsers = computed(
  () => isSuperAdmin.value || currentUser.value?.permissions?.includes("manage_users"),
);
const canManagePermissions = computed(
  () => isSuperAdmin.value || currentUser.value?.permissions?.includes("manage_permissions"),
);

const activeTab = ref<"users" | "permissions" | "health">("users");

// --- Aba de usuários ---

const users = ref<User[]>([]);
const usersError = ref("");
const loadingUsers = ref(true);
const roleUpdatingId = ref<number | null>(null);

const assignableRoles = computed<UserRole[]>(() =>
  isSuperAdmin.value
    ? ["pessoa_fisica", "pessoa_juridica", "contador", "admin"]
    : ["pessoa_fisica", "pessoa_juridica", "contador"],
);

async function loadUsers() {
  loadingUsers.value = true;
  usersError.value = "";
  try {
    users.value = await listUsers();
  } catch (err) {
    usersError.value = err instanceof ApiError ? err.message : "Falha ao carregar usuários";
  } finally {
    loadingUsers.value = false;
  }
}

function canEditRole(target: User) {
  if (target.role === "super_admin") return false;
  if (!canManageUsers.value) return false;
  // Rebaixar um admin é reservado ao Super Admin.
  if (target.role === "admin" && !isSuperAdmin.value) return false;
  return true;
}

async function handleRoleChange(target: User, role: UserRole) {
  roleUpdatingId.value = target.id;
  usersError.value = "";
  try {
    const updated = await updateUserRole(target.id, role);
    const idx = users.value.findIndex((u) => u.id === target.id);
    if (idx !== -1) users.value[idx] = updated;
  } catch (err) {
    usersError.value = err instanceof ApiError ? err.message : "Falha ao atualizar cargo";
    await loadUsers();
  } finally {
    roleUpdatingId.value = null;
  }
}

// --- Aba de permissões ---

const permissionRoles = ref<UserRole[]>([]);
const availablePermissions = ref<Permission[]>([]);
const rolePermissions = ref<Record<string, Permission[]>>({});
const permissionsError = ref("");
const permissionsSuccess = ref("");
const loadingPermissions = ref(true);
const savingRole = ref<string | null>(null);

async function loadPermissions() {
  loadingPermissions.value = true;
  permissionsError.value = "";
  try {
    const data = await fetchRolePermissions();
    permissionRoles.value = data.roles;
    availablePermissions.value = data.available;
    rolePermissions.value = data.permissions;
  } catch (err) {
    permissionsError.value = err instanceof ApiError ? err.message : "Falha ao carregar permissões";
  } finally {
    loadingPermissions.value = false;
  }
}

function hasPermission(role: UserRole, perm: Permission) {
  return rolePermissions.value[role]?.includes(perm) ?? false;
}

async function togglePermission(role: UserRole, perm: Permission) {
  if (!canManagePermissions.value || savingRole.value) return;

  const current = rolePermissions.value[role] ?? [];
  const next = current.includes(perm) ? current.filter((p) => p !== perm) : [...current, perm];

  savingRole.value = role;
  permissionsError.value = "";
  permissionsSuccess.value = "";
  try {
    const result = await updateRolePermissions(role, next);
    rolePermissions.value = { ...rolePermissions.value, [role]: result.permissions };
    permissionsSuccess.value = `Permissões de ${roleLabels[role]} atualizadas.`;
  } catch (err) {
    permissionsError.value =
      err instanceof ApiError ? err.message : "Falha ao atualizar permissões";
  } finally {
    savingRole.value = null;
  }
}

// --- Aba de saúde financeira ---

const healthSettings = ref<HealthSettings>({
  enabled: true,
  thresholdOtima: 20,
  thresholdBoa: 10,
  thresholdEstavel: 0,
  thresholdRuim: -10,
});
const healthError = ref("");
const healthSuccess = ref("");
const loadingHealth = ref(true);
const savingHealth = ref(false);

const healthLevels = [
  {
    key: "thresholdOtima",
    label: "Ótima",
    stars: 5,
    hint: "Taxa de poupança igual ou acima deste valor",
  },
  { key: "thresholdBoa", label: "Boa", stars: 4, hint: "Igual ou acima deste valor" },
  {
    key: "thresholdEstavel",
    label: "Normal / Estável",
    stars: 3,
    hint: "Igual ou acima deste valor",
  },
  { key: "thresholdRuim", label: "Ruim", stars: 2, hint: "Igual ou acima deste valor" },
] as const;

async function loadHealthSettings() {
  loadingHealth.value = true;
  healthError.value = "";
  try {
    healthSettings.value = await getHealthSettings();
  } catch (err) {
    healthError.value =
      err instanceof HealthApiError ? err.message : "Falha ao carregar configuração";
  } finally {
    loadingHealth.value = false;
  }
}

async function saveHealthSettings() {
  savingHealth.value = true;
  healthError.value = "";
  healthSuccess.value = "";
  try {
    healthSettings.value = await updateHealthSettings(healthSettings.value);
    healthSuccess.value = "Configuração de saúde financeira atualizada.";
  } catch (err) {
    healthError.value =
      err instanceof HealthApiError ? err.message : "Falha ao salvar configuração";
  } finally {
    savingHealth.value = false;
  }
}

// --- Visual ---

const roleBadgeClasses: Record<UserRole, string> = {
  pessoa_fisica: "bg-ink-100 text-ink-600",
  pessoa_juridica: "bg-brand-50 text-brand-700",
  contador: "bg-coral-50 text-coral-700",
  admin: "bg-brand-100 text-brand-700",
  super_admin: "bg-ink-900 text-white",
};

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

const avatarErrors = ref(new Set<number>());

function handleAvatarError(userId: number) {
  avatarErrors.value.add(userId);
}

const userCountLabel = computed(() => {
  const count = users.value.length;
  return count === 1 ? "1 usuário" : `${count} usuários`;
});

onMounted(() => {
  loadUsers();
  loadPermissions();
  loadHealthSettings();
});
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Administração</h1>
        <p class="mt-0.5 text-sm text-ink-500">Gerencie usuários, cargos e permissões</p>
      </div>
      <RouterLink
        to="/register"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
      >
        + Cadastrar
      </RouterLink>
    </div>

    <div class="flex gap-1 rounded-full border border-ink-200/70 bg-white p-1 shadow-sm w-fit">
      <button
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="activeTab === 'users' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'"
        @click="activeTab = 'users'"
      >
        Usuários
      </button>
      <button
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="
          activeTab === 'permissions' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'
        "
        @click="activeTab = 'permissions'"
      >
        Permissões
      </button>
      <button
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="
          activeTab === 'health' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'
        "
        @click="activeTab = 'health'"
      >
        Saúde Financeira
      </button>
    </div>

    <!-- Aba: Usuários -->
    <div v-if="activeTab === 'users'" class="flex flex-col gap-3">
      <p class="text-sm text-ink-500">{{ userCountLabel }}</p>
      <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
        <div v-if="loadingUsers" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
          <span
            class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
          ></span>
          Carregando...
        </div>
        <p v-else-if="usersError" class="p-8 text-center text-sm text-coral-600">
          {{ usersError }}
        </p>
        <div
          v-else-if="users.length === 0"
          class="flex flex-col items-center gap-1 p-12 text-center"
        >
          <p class="text-sm font-medium text-ink-600">Nenhum usuário cadastrado ainda</p>
          <p class="text-sm text-ink-400">Cadastre o primeiro usuário para começar.</p>
        </div>
        <ul v-else class="divide-y divide-ink-100">
          <li
            v-for="user in users"
            :key="user.id"
            class="flex items-center gap-3 px-5 py-4 transition hover:bg-ink-50/60"
          >
            <img
              v-if="user.avatarUrl && !avatarErrors.has(user.id)"
              :src="user.avatarUrl"
              alt=""
              referrerpolicy="no-referrer"
              class="h-10 w-10 shrink-0 rounded-full object-cover shadow-sm"
              @error="handleAvatarError(user.id)"
            />
            <span
              v-else
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-linear-to-br text-sm font-semibold text-white shadow-sm"
              :class="gradientFor(user.id)"
            >
              {{ initialsFor(user.name) }}
            </span>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-semibold text-ink-900">{{ user.name }}</p>
              <p class="truncate text-xs text-ink-500">
                {{ user.email }}
                <span v-if="user.cnpj" class="text-ink-400"
                  >&middot; CNPJ {{ formatCnpj(user.cnpj) }}</span
                >
              </p>
            </div>
            <select
              v-if="canEditRole(user)"
              :value="user.role"
              :disabled="roleUpdatingId === user.id"
              class="shrink-0 rounded-full border border-ink-200 bg-white px-3 py-1.5 text-xs font-semibold text-ink-700 outline-none transition focus:border-brand-400 focus:ring-4 focus:ring-brand-100 disabled:opacity-50"
              @change="
                handleRoleChange(user, ($event.target as HTMLSelectElement).value as UserRole)
              "
            >
              <option v-for="role in assignableRoles" :key="role" :value="role">
                {{ roleLabels[role] }}
              </option>
            </select>
            <span
              v-else
              class="shrink-0 rounded-full px-3 py-1.5 text-xs font-semibold"
              :class="roleBadgeClasses[user.role]"
            >
              {{ roleLabels[user.role] }}
            </span>
          </li>
        </ul>
      </div>
    </div>

    <!-- Aba: Saúde Financeira -->
    <div v-else-if="activeTab === 'health'" class="flex flex-col gap-3">
      <p class="text-sm text-ink-500">
        O bloco de saúde financeira classifica o mês do usuário pela taxa de poupança &mdash;
        percentual das receitas que sobra após as despesas. Ajuste abaixo o limite mínimo de cada
        nível.
      </p>

      <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
        <div
          v-if="loadingHealth"
          class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400"
        >
          <span
            class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
          ></span>
          Carregando...
        </div>
        <div v-else class="flex flex-col divide-y divide-ink-100">
          <label class="flex cursor-pointer items-center justify-between gap-3 px-5 py-4">
            <div>
              <p class="text-sm font-semibold text-ink-900">Exibir bloco de saúde financeira</p>
              <p class="text-xs text-ink-500">
                Quando desativado, o bloco some da barra lateral de todos os usuários.
              </p>
            </div>
            <input
              v-model="healthSettings.enabled"
              type="checkbox"
              class="h-4 w-4 cursor-pointer rounded border-ink-300 accent-brand-600 focus:ring-brand-400"
            />
          </label>

          <div
            v-for="lvl in healthLevels"
            :key="lvl.key"
            class="flex items-center justify-between gap-3 px-5 py-3.5"
          >
            <div class="min-w-0">
              <p class="text-sm font-semibold text-ink-900">
                {{ lvl.label }}
                <span class="ml-1 text-xs font-normal text-brand-600">{{
                  "★".repeat(lvl.stars) + "☆".repeat(5 - lvl.stars)
                }}</span>
              </p>
              <p class="text-xs text-ink-500">{{ lvl.hint }}</p>
            </div>
            <div class="flex shrink-0 items-center gap-1.5">
              <input
                v-model.number="healthSettings[lvl.key]"
                type="number"
                step="1"
                class="w-20 rounded-lg border border-ink-200 px-2.5 py-1.5 text-right text-sm outline-none transition focus:border-brand-400 focus:ring-4 focus:ring-brand-100"
              />
              <span class="text-sm text-ink-500">%</span>
            </div>
          </div>

          <div class="flex items-center justify-between gap-3 px-5 py-3.5">
            <div>
              <p class="text-sm font-semibold text-ink-900">
                Péssima
                <span class="ml-1 text-xs font-normal text-brand-600">★☆☆☆☆</span>
              </p>
              <p class="text-xs text-ink-500">
                Qualquer taxa abaixo do limite de "Ruim" ({{ healthSettings.thresholdRuim }}%).
              </p>
            </div>
          </div>

          <div class="flex items-center justify-end gap-3 px-5 py-4">
            <button
              type="button"
              :disabled="savingHealth"
              class="rounded-full bg-ink-900 px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
              @click="saveHealthSettings"
            >
              {{ savingHealth ? "Salvando..." : "Salvar" }}
            </button>
          </div>
        </div>
      </div>

      <p v-if="healthError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ healthError }}
      </p>
      <p v-else-if="healthSuccess" class="rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700">
        {{ healthSuccess }}
      </p>
    </div>

    <!-- Aba: Permissões -->
    <div v-else class="flex flex-col gap-3">
      <p class="text-sm text-ink-500">
        O Super Administrador sempre possui todas as permissões e não pode ser alterado.
        <span v-if="!canManagePermissions">Você não tem permissão para editar — somente visualizar.</span>
      </p>

      <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
        <div
          v-if="loadingPermissions"
          class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400"
        >
          <span
            class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
          ></span>
          Carregando...
        </div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-ink-100 text-left text-xs uppercase tracking-wide text-ink-400">
                  <th class="px-5 py-3 font-semibold">Cargo</th>
                  <th
                    v-for="perm in availablePermissions"
                    :key="perm"
                    class="px-4 py-3 text-center font-semibold"
                  >
                    {{ permissionLabels[perm] }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-ink-100">
                <tr
                  v-for="role in permissionRoles"
                  :key="role"
                  class="transition hover:bg-ink-50/60"
                >
                  <td class="px-5 py-3.5 font-semibold text-ink-900">{{ roleLabels[role] }}</td>
                  <td
                    v-for="perm in availablePermissions"
                    :key="perm"
                    class="px-4 py-3.5 text-center"
                  >
                    <input
                      type="checkbox"
                      :checked="hasPermission(role, perm)"
                      :disabled="!canManagePermissions || savingRole === role"
                      class="h-4 w-4 cursor-pointer rounded border-ink-300 text-brand-600 accent-brand-600 focus:ring-brand-400 disabled:cursor-not-allowed disabled:opacity-50"
                      @change="togglePermission(role, perm)"
                    />
                  </td>
                </tr>
                <tr class="bg-ink-50/40">
                  <td class="px-5 py-3.5 font-semibold text-ink-900">
                    {{ roleLabels.super_admin }}
                  </td>
                  <td
                    v-for="perm in availablePermissions"
                    :key="perm"
                    class="px-4 py-3.5 text-center"
                  >
                    <input
                      type="checkbox"
                      checked
                      disabled
                      class="h-4 w-4 rounded border-ink-300 accent-ink-900 opacity-60"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>

      <p v-if="permissionsError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ permissionsError }}
      </p>
      <p
        v-else-if="permissionsSuccess"
        class="rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700"
      >
        {{ permissionsSuccess }}
      </p>
    </div>
  </div>
</template>
