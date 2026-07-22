<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
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
import { ApiError as BillingApiError, listPlans, updatePlan } from "../api/billing";
import type { Plan } from "../types/billing";
import {
  ApiError as AnalyticsApiError,
  getActiveNow,
  getDeviceBreakdown,
  getOverview,
  getRecentVisits,
  getTopLocations,
  getTopPages,
  getTopReferrers,
} from "../api/analytics";
import type {
  AnalyticsOverview,
  DeviceBreakdown,
  LocationCount,
  PathCount,
  ReferrerCount,
  RecentVisit,
} from "../types/analytics";
import VisitsOverTimeChart from "../components/charts/VisitsOverTimeChart.vue";
import DeviceBreakdownChart from "../components/charts/DeviceBreakdownChart.vue";
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

const canManagePlans = computed(
  () => isSuperAdmin.value || currentUser.value?.permissions?.includes("manage_plans"),
);

const activeTab = ref<"users" | "permissions" | "health" | "plans" | "stats">("users");

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

// --- Aba de planos ---

// Formulário de edição por plano: benefícios são editados como texto, um por
// linha, e convertidos para lista no envio.
interface PlanForm {
  name: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  benefitsText: string;
  highlight: boolean;
}

const plans = ref<Plan[]>([]);
const planForms = ref<Record<number, PlanForm>>({});
const plansError = ref("");
const plansSuccess = ref("");
const loadingPlans = ref(true);
const savingPlanId = ref<number | null>(null);

async function loadPlans() {
  loadingPlans.value = true;
  plansError.value = "";
  try {
    plans.value = await listPlans();
    const forms: Record<number, PlanForm> = {};
    for (const plan of plans.value) {
      forms[plan.id] = {
        name: plan.name,
        description: plan.description,
        priceMonthly: plan.priceMonthly,
        priceYearly: plan.priceYearly,
        benefitsText: plan.benefits.join("\n"),
        highlight: plan.highlight,
      };
    }
    planForms.value = forms;
  } catch (err) {
    plansError.value = err instanceof BillingApiError ? err.message : "Falha ao carregar planos";
  } finally {
    loadingPlans.value = false;
  }
}

async function savePlan(plan: Plan) {
  const form = planForms.value[plan.id];
  if (!form || !canManagePlans.value) return;

  savingPlanId.value = plan.id;
  plansError.value = "";
  plansSuccess.value = "";
  try {
    const updated = await updatePlan(plan.id, {
      name: form.name,
      description: form.description,
      priceMonthly: Number(form.priceMonthly),
      priceYearly: Number(form.priceYearly),
      benefits: form.benefitsText
        .split("\n")
        .map((line) => line.trim())
        .filter(Boolean),
      highlight: form.highlight,
    });
    const idx = plans.value.findIndex((p) => p.id === plan.id);
    if (idx !== -1) plans.value[idx] = updated;
    plansSuccess.value = `Plano ${updated.name} atualizado.`;
  } catch (err) {
    plansError.value = err instanceof BillingApiError ? err.message : "Falha ao salvar o plano";
  } finally {
    savingPlanId.value = null;
  }
}

// --- Aba de estatísticas ---

type StatsPreset = "7d" | "30d" | "90d" | "custom";

function toInputDate(d: Date) {
  return d.toISOString().slice(0, 10);
}

const statsPreset = ref<StatsPreset>("30d");
const statsFrom = ref(toInputDate(new Date(Date.now() - 29 * 24 * 60 * 60 * 1000)));
const statsTo = ref(toInputDate(new Date()));

const statsOverview = ref<AnalyticsOverview | null>(null);
const statsPages = ref<PathCount[]>([]);
const statsReferrers = ref<ReferrerCount[]>([]);
const statsLocations = ref<LocationCount[]>([]);
const statsDevices = ref<DeviceBreakdown | null>(null);
const statsRecentVisits = ref<RecentVisit[]>([]);
const statsActiveNow = ref<number | null>(null);
const statsError = ref("");
const loadingStats = ref(true);

let activeNowTimer: ReturnType<typeof setInterval> | undefined;

function setStatsPreset(preset: StatsPreset) {
  statsPreset.value = preset;
  const days = { "7d": 7, "30d": 30, "90d": 90, custom: 30 }[preset];
  if (preset !== "custom") {
    statsFrom.value = toInputDate(new Date(Date.now() - (days - 1) * 24 * 60 * 60 * 1000));
    statsTo.value = toInputDate(new Date());
  }
  loadStats();
}

async function loadStats() {
  loadingStats.value = true;
  statsError.value = "";
  const range = { from: statsFrom.value, to: statsTo.value };
  try {
    const [overview, pages, referrers, locations, devices, recentVisits] = await Promise.all([
      getOverview(range),
      getTopPages(range),
      getTopReferrers(range),
      getTopLocations(range),
      getDeviceBreakdown(range),
      getRecentVisits(range),
    ]);
    statsOverview.value = overview;
    statsPages.value = pages;
    statsReferrers.value = referrers;
    statsLocations.value = locations;
    statsDevices.value = devices;
    statsRecentVisits.value = recentVisits;
  } catch (err) {
    statsError.value =
      err instanceof AnalyticsApiError ? err.message : "Falha ao carregar estatísticas";
  } finally {
    loadingStats.value = false;
  }
}

async function pollActiveNow() {
  try {
    statsActiveNow.value = await getActiveNow();
  } catch {
    // silencioso: é um indicador auxiliar, não deve gerar erro visível
  }
}

function formatVisitDate(iso: string) {
  return new Date(iso).toLocaleString("pt-BR", { dateStyle: "short", timeStyle: "short" });
}

pollActiveNow();
activeNowTimer = setInterval(pollActiveNow, 30_000);

onUnmounted(() => {
  if (activeNowTimer) clearInterval(activeNowTimer);
});

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
  if (canManagePlans.value) loadPlans();
  loadStats();
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
      <button
        v-if="canManagePlans"
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="activeTab === 'plans' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'"
        @click="activeTab = 'plans'"
      >
        Planos
      </button>
      <button
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="activeTab === 'stats' ? 'bg-ink-900 text-white' : 'text-ink-500 hover:text-ink-900'"
        @click="activeTab = 'stats'"
      >
        Estatísticas
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

    <!-- Aba: Planos -->
    <div v-else-if="activeTab === 'plans'" class="flex flex-col gap-3">
      <p class="text-sm text-ink-500">
        Configure nome, preços e benefícios dos planos exibidos em /planos. O plano gratuito não
        tem preço; benefícios são um por linha.
      </p>

      <div v-if="loadingPlans" class="flex flex-col items-center gap-2 rounded-3xl border border-ink-200/70 bg-white p-12 text-sm text-ink-400 shadow-sm">
        <span
          class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Carregando...
      </div>
      <div v-else class="flex flex-col gap-4">
        <form
          v-for="plan in plans"
          :key="plan.id"
          class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
          @submit.prevent="savePlan(plan)"
        >
          <div class="flex items-center justify-between gap-3">
            <h2 class="text-sm font-semibold text-ink-900">
              {{ planForms[plan.id]?.name || plan.name }}
              <span class="ml-1.5 rounded-full bg-ink-100 px-2 py-0.5 text-[11px] font-semibold text-ink-500">
                {{ plan.code }}
              </span>
            </h2>
            <label
              v-if="plan.code !== 'gratuito'"
              class="flex cursor-pointer items-center gap-2 text-xs font-medium text-ink-600"
            >
              <input
                v-model="planForms[plan.id]!.highlight"
                type="checkbox"
                class="h-4 w-4 rounded border-ink-300 accent-brand-600 focus:ring-brand-400"
              />
              Destacar como "Mais popular"
            </label>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Nome</span>
              <input
                v-model="planForms[plan.id]!.name"
                type="text"
                required
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <label class="flex flex-col gap-1.5">
              <span class="text-sm font-medium text-ink-700">Descrição</span>
              <input
                v-model="planForms[plan.id]!.description"
                type="text"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              />
            </label>
            <template v-if="plan.code !== 'gratuito'">
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Preço mensal (R$)</span>
                <input
                  v-model.number="planForms[plan.id]!.priceMonthly"
                  type="number"
                  min="0.01"
                  step="0.01"
                  required
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                />
              </label>
              <label class="flex flex-col gap-1.5">
                <span class="text-sm font-medium text-ink-700">Preço anual (R$)</span>
                <input
                  v-model.number="planForms[plan.id]!.priceYearly"
                  type="number"
                  min="0.01"
                  step="0.01"
                  required
                  class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
                />
              </label>
            </template>
            <label class="flex flex-col gap-1.5 sm:col-span-2">
              <span class="text-sm font-medium text-ink-700">Benefícios (um por linha)</span>
              <textarea
                v-model="planForms[plan.id]!.benefitsText"
                rows="5"
                class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
              ></textarea>
            </label>
          </div>

          <div class="flex justify-end">
            <button
              type="submit"
              :disabled="savingPlanId === plan.id"
              class="rounded-full bg-ink-900 px-5 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
            >
              {{ savingPlanId === plan.id ? "Salvando..." : "Salvar" }}
            </button>
          </div>
        </form>
      </div>

      <p v-if="plansError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ plansError }}
      </p>
      <p v-else-if="plansSuccess" class="rounded-xl bg-brand-50 px-3 py-2 text-sm text-brand-700">
        {{ plansSuccess }}
      </p>
    </div>

    <!-- Aba: Estatísticas -->
    <div v-else-if="activeTab === 'stats'" class="flex flex-col gap-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <button
            v-for="preset in (['7d', '30d', '90d'] as const)"
            :key="preset"
            type="button"
            class="rounded-full px-3.5 py-1.5 text-xs font-semibold transition"
            :class="
              statsPreset === preset
                ? 'bg-ink-900 text-white'
                : 'border border-ink-200 text-ink-600 hover:border-ink-300'
            "
            @click="setStatsPreset(preset)"
          >
            {{ preset === "7d" ? "7 dias" : preset === "30d" ? "30 dias" : "90 dias" }}
          </button>
          <div class="flex items-center gap-1.5">
            <input
              v-model="statsFrom"
              type="date"
              class="rounded-full border border-ink-200 px-3 py-1.5 text-xs outline-none focus:border-brand-400"
              @change="
                statsPreset = 'custom';
                loadStats();
              "
            />
            <span class="text-xs text-ink-400">até</span>
            <input
              v-model="statsTo"
              type="date"
              class="rounded-full border border-ink-200 px-3 py-1.5 text-xs outline-none focus:border-brand-400"
              @change="
                statsPreset = 'custom';
                loadStats();
              "
            />
          </div>
        </div>
        <div
          v-if="statsActiveNow !== null"
          class="flex items-center gap-1.5 rounded-full bg-brand-50 px-3 py-1.5 text-xs font-semibold text-brand-700"
        >
          <span class="h-1.5 w-1.5 animate-pulse rounded-full bg-brand-500"></span>
          {{ statsActiveNow }} ativo{{ statsActiveNow === 1 ? "" : "s" }} agora
        </div>
      </div>

      <div
        v-if="loadingStats"
        class="flex flex-col items-center gap-2 rounded-3xl border border-ink-200/70 bg-white p-12 text-sm text-ink-400 shadow-sm"
      >
        <span
          class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Carregando...
      </div>
      <p v-else-if="statsError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ statsError }}
      </p>
      <template v-else>
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
            <p class="text-xs text-ink-500">Visitantes únicos</p>
            <p class="mt-1 font-display text-2xl font-bold text-ink-900">
              {{ statsOverview?.uniqueVisitors ?? 0 }}
            </p>
          </div>
          <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
            <p class="text-xs text-ink-500">Visitas totais</p>
            <p class="mt-1 font-display text-2xl font-bold text-ink-900">
              {{ statsOverview?.totalVisits ?? 0 }}
            </p>
          </div>
          <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
            <p class="text-xs text-ink-500">Novos visitantes</p>
            <p class="mt-1 font-display text-2xl font-bold text-ink-900">
              {{ statsOverview?.newVisitors ?? 0 }}
            </p>
          </div>
          <div class="rounded-3xl border border-ink-200/70 bg-white p-4 shadow-sm">
            <p class="text-xs text-ink-500">Recorrentes</p>
            <p class="mt-1 font-display text-2xl font-bold text-ink-900">
              {{ statsOverview?.returningVisitors ?? 0 }}
            </p>
          </div>
        </div>

        <VisitsOverTimeChart :by-day="statsOverview?.byDay ?? []" />

        <div class="grid gap-4 lg:grid-cols-2">
          <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
            <h2 class="px-5 pt-4 font-display text-sm font-bold text-ink-900">
              Páginas mais visitadas
            </h2>
            <ul class="divide-y divide-ink-100 px-5 py-2">
              <li
                v-for="p in statsPages"
                :key="p.path"
                class="flex items-center justify-between gap-3 py-2 text-sm"
              >
                <span class="truncate text-ink-700">{{ p.path }}</span>
                <span class="shrink-0 font-semibold text-ink-900">{{ p.visits }}</span>
              </li>
              <li v-if="statsPages.length === 0" class="py-4 text-center text-sm text-ink-400">
                Sem dados
              </li>
            </ul>
          </div>

          <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
            <h2 class="px-5 pt-4 font-display text-sm font-bold text-ink-900">Referrers</h2>
            <ul class="divide-y divide-ink-100 px-5 py-2">
              <li
                v-for="r in statsReferrers"
                :key="r.referrer"
                class="flex items-center justify-between gap-3 py-2 text-sm"
              >
                <span class="truncate text-ink-700">{{ r.referrer }}</span>
                <span class="shrink-0 font-semibold text-ink-900">{{ r.visits }}</span>
              </li>
              <li
                v-if="statsReferrers.length === 0"
                class="py-4 text-center text-sm text-ink-400"
              >
                Sem dados
              </li>
            </ul>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-3">
          <DeviceBreakdownChart title="Navegadores" :items="statsDevices?.browsers ?? []" />
          <DeviceBreakdownChart title="Sistemas operacionais" :items="statsDevices?.os ?? []" />
          <DeviceBreakdownChart title="Dispositivos" :items="statsDevices?.devices ?? []" />
        </div>

        <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
          <h2 class="px-5 pt-4 font-display text-sm font-bold text-ink-900">Localizações</h2>
          <ul class="divide-y divide-ink-100 px-5 py-2">
            <li
              v-for="(l, idx) in statsLocations"
              :key="`${l.country}-${l.city}-${idx}`"
              class="flex items-center justify-between gap-3 py-2 text-sm"
            >
              <span class="truncate text-ink-700">
                {{ l.city ? `${l.city}, ${l.country}` : l.country }}
              </span>
              <span class="shrink-0 font-semibold text-ink-900">{{ l.visits }}</span>
            </li>
            <li v-if="statsLocations.length === 0" class="py-4 text-center text-sm text-ink-400">
              Sem dados
            </li>
          </ul>
        </div>

        <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
          <h2 class="px-5 pt-4 font-display text-sm font-bold text-ink-900">Visitas recentes</h2>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr
                  class="border-b border-ink-100 text-left text-xs uppercase tracking-wide text-ink-400"
                >
                  <th class="px-5 py-3 font-semibold">Quando</th>
                  <th class="px-4 py-3 font-semibold">Página</th>
                  <th class="px-4 py-3 font-semibold">IP</th>
                  <th class="px-4 py-3 font-semibold">Local</th>
                  <th class="px-4 py-3 font-semibold">Navegador</th>
                  <th class="px-4 py-3 font-semibold">SO</th>
                  <th class="px-4 py-3 font-semibold">Referrer</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-ink-100">
                <tr v-for="v in statsRecentVisits" :key="v.id" class="hover:bg-ink-50/60">
                  <td class="whitespace-nowrap px-5 py-2.5 text-ink-500">
                    {{ formatVisitDate(v.createdAt) }}
                  </td>
                  <td class="max-w-[16rem] truncate px-4 py-2.5 text-ink-700">{{ v.path }}</td>
                  <td class="whitespace-nowrap px-4 py-2.5 font-mono text-xs text-ink-500">
                    {{ v.ip }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-2.5 text-ink-500">
                    {{ v.city ? `${v.city}, ${v.country}` : v.country || "—" }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-2.5 text-ink-500">
                    {{ v.browser || "—" }}
                  </td>
                  <td class="whitespace-nowrap px-4 py-2.5 text-ink-500">{{ v.os || "—" }}</td>
                  <td class="max-w-[12rem] truncate px-4 py-2.5 text-ink-500">
                    {{ v.referrer || "(direto)" }}
                  </td>
                </tr>
                <tr v-if="statsRecentVisits.length === 0">
                  <td colspan="7" class="px-5 py-6 text-center text-sm text-ink-400">
                    Sem visitas registradas neste período
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- Aba: Permissões -->
    <div v-else-if="activeTab === 'permissions'" class="flex flex-col gap-3">
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
