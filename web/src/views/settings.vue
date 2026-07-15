<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  getNotificationSettings,
  updateNotificationSettings,
  ApiError as NotificationApiError,
} from "../api/notificationSettings";
import { ApiError, updatePassword, updateProfile } from "../api/users";
import PasswordInput from "../components/PasswordInput.vue";
import { useAuth } from "../composables/useAuth";
import { OFFSET_OPTIONS } from "../types/notification";
import type { NotificationSettings } from "../types/notification";

const { currentUser, setCurrentUser } = useAuth();

const profileForm = reactive({ name: "", email: "" });
const profileError = ref("");
const profileSuccess = ref("");
const savingProfile = ref(false);

const passwordForm = reactive({ currentPassword: "", newPassword: "", confirmPassword: "" });
const passwordError = ref("");
const passwordSuccess = ref("");
const savingPassword = ref(false);

const avatarError = ref(false);
const avatarUrl = computed(() => (avatarError.value ? "" : currentUser.value?.avatarUrl || ""));

function handleAvatarError() {
  avatarError.value = true;
}

watch(
  currentUser,
  (user) => {
    if (!user) return;
    profileForm.name = user.name;
    profileForm.email = user.email;
  },
  { immediate: true },
);

const roleLabels: Record<string, string> = {
  user: "Usuário",
  pessoa_fisica: "Pessoa Física",
  pessoa_juridica: "Pessoa Jurídica",
  contador: "Contador / Técnico Contábil",
  admin: "Administrador",
  super_admin: "Super Administrador",
};

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString("pt-BR", { day: "2-digit", month: "long", year: "numeric" });
}

async function handleProfileSubmit() {
  profileError.value = "";
  profileSuccess.value = "";
  savingProfile.value = true;
  try {
    const updated = await updateProfile({ ...profileForm });
    setCurrentUser(updated);
    profileSuccess.value = "Dados atualizados com sucesso.";
  } catch (err) {
    profileError.value = err instanceof ApiError ? err.message : "Falha ao salvar os dados";
  } finally {
    savingProfile.value = false;
  }
}

const notificationSettings = reactive<NotificationSettings>({
  siteEnabled: true,
  emailEnabled: false,
  offsets: [],
});
const notificationsLoading = ref(true);
const notificationsError = ref("");
const notificationsSuccess = ref("");
const savingNotifications = ref(false);

async function loadNotificationSettings() {
  notificationsLoading.value = true;
  try {
    const settings = await getNotificationSettings();
    notificationSettings.siteEnabled = settings.siteEnabled;
    notificationSettings.emailEnabled = settings.emailEnabled;
    notificationSettings.offsets = settings.offsets;
  } catch (err) {
    notificationsError.value =
      err instanceof NotificationApiError ? err.message : "Falha ao carregar preferências";
  } finally {
    notificationsLoading.value = false;
  }
}

function toggleOffset(value: number) {
  const idx = notificationSettings.offsets.indexOf(value);
  if (idx >= 0) notificationSettings.offsets.splice(idx, 1);
  else notificationSettings.offsets.push(value);
}

async function handleNotificationSubmit() {
  notificationsError.value = "";
  notificationsSuccess.value = "";
  savingNotifications.value = true;
  try {
    const saved = await updateNotificationSettings({ ...notificationSettings });
    notificationSettings.siteEnabled = saved.siteEnabled;
    notificationSettings.emailEnabled = saved.emailEnabled;
    notificationSettings.offsets = saved.offsets;
    notificationsSuccess.value = "Preferências de notificação salvas.";
  } catch (err) {
    notificationsError.value =
      err instanceof NotificationApiError ? err.message : "Falha ao salvar preferências";
  } finally {
    savingNotifications.value = false;
  }
}

onMounted(loadNotificationSettings);

async function handlePasswordSubmit() {
  passwordError.value = "";
  passwordSuccess.value = "";

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordError.value = "As senhas não coincidem";
    return;
  }

  savingPassword.value = true;
  try {
    await updatePassword({
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword,
    });
    passwordForm.currentPassword = "";
    passwordForm.newPassword = "";
    passwordForm.confirmPassword = "";
    if (currentUser.value) setCurrentUser({ ...currentUser.value, hasPassword: true });
    passwordSuccess.value = "Senha atualizada com sucesso.";
  } catch (err) {
    passwordError.value = err instanceof ApiError ? err.message : "Falha ao atualizar a senha";
  } finally {
    savingPassword.value = false;
  }
}
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Configurações</h1>
      <p class="mt-0.5 text-sm text-ink-500">Gerencie seus dados de conta</p>
    </div>

    <div
      v-if="currentUser"
      class="flex items-center gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
    >
      <img
        v-if="avatarUrl"
        :src="avatarUrl"
        alt=""
        referrerpolicy="no-referrer"
        class="h-14 w-14 shrink-0 rounded-full object-cover shadow-sm"
        @error="handleAvatarError"
      />
      <span
        v-else
        class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-linear-to-br from-brand-400 to-coral-500 text-lg font-semibold text-white shadow-sm"
      >
        {{
          currentUser.name
            .split(" ")
            .filter(Boolean)
            .slice(0, 2)
            .map((part) => part[0]!.toUpperCase())
            .join("")
        }}
      </span>
      <div class="min-w-0">
        <p class="truncate text-sm font-semibold text-ink-900">{{ currentUser.name }}</p>
        <p class="truncate text-xs text-ink-500">{{ currentUser.email }}</p>
        <p class="mt-1 text-xs text-ink-400">
          {{ roleLabels[currentUser.role] ?? currentUser.role }} &middot; desde
          {{ formatDate(currentUser.createdAt) }}
        </p>
      </div>
    </div>

    <form
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleProfileSubmit"
    >
      <h2 class="text-sm font-semibold text-ink-900">Dados pessoais</h2>
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Nome</span>
          <input
            v-model="profileForm.name"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">E-mail</span>
          <input
            v-model="profileForm.email"
            type="email"
            required
            autocomplete="username"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>

      <p v-if="profileError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ profileError }}
      </p>
      <p v-if="profileSuccess" class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
        {{ profileSuccess }}
      </p>

      <div>
        <button
          type="submit"
          :disabled="savingProfile"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ savingProfile ? "Salvando..." : "Salvar" }}
        </button>
      </div>
    </form>

    <form
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handlePasswordSubmit"
    >
      <h2 class="text-sm font-semibold text-ink-900">
        {{ currentUser?.hasPassword ? "Alterar senha" : "Definir senha" }}
      </h2>
      <p v-if="!currentUser?.hasPassword" class="text-sm text-ink-500">
        Sua conta usa login via Google. Defina uma senha para também poder entrar com e-mail e
        senha.
      </p>

      <div class="grid gap-4 sm:grid-cols-2">
        <label v-if="currentUser?.hasPassword" class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-ink-700">Senha atual</span>
          <PasswordInput
            v-model="passwordForm.currentPassword"
            required
            autocomplete="current-password"
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Nova senha</span>
          <PasswordInput
            v-model="passwordForm.newPassword"
            required
            :minlength="8"
            autocomplete="new-password"
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Confirmar nova senha</span>
          <PasswordInput
            v-model="passwordForm.confirmPassword"
            required
            :minlength="8"
            autocomplete="new-password"
          />
        </label>
      </div>

      <p v-if="passwordError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ passwordError }}
      </p>
      <p v-if="passwordSuccess" class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
        {{ passwordSuccess }}
      </p>

      <div>
        <button
          type="submit"
          :disabled="savingPassword"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ savingPassword ? "Salvando..." : "Salvar" }}
        </button>
      </div>
    </form>

    <form
      id="notificacoes"
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleNotificationSubmit"
    >
      <div>
        <h2 class="text-sm font-semibold text-ink-900">Notificações</h2>
        <p class="mt-0.5 text-xs text-ink-500">Lembretes de contas fixas vencendo ou vencidas</p>
      </div>

      <div v-if="notificationsLoading" class="text-sm text-ink-400">Carregando...</div>
      <template v-else>
        <div class="flex flex-col gap-3 sm:flex-row sm:gap-6">
          <label class="flex items-center gap-2 text-sm font-medium text-ink-700">
            <input
              v-model="notificationSettings.siteEnabled"
              type="checkbox"
              class="h-4 w-4 rounded border-ink-300 text-brand-600 focus:ring-brand-400"
            />
            Notificações no site
          </label>
          <label class="flex items-center gap-2 text-sm font-medium text-ink-700">
            <input
              v-model="notificationSettings.emailEnabled"
              type="checkbox"
              class="h-4 w-4 rounded border-ink-300 text-brand-600 focus:ring-brand-400"
            />
            Notificações por e-mail
          </label>
        </div>

        <div>
          <span class="text-sm font-medium text-ink-700">Antecedência dos lembretes</span>
          <div class="mt-2 grid grid-cols-2 gap-2 sm:grid-cols-4">
            <label
              v-for="option in OFFSET_OPTIONS"
              :key="option.value"
              class="flex items-center gap-2 rounded-xl border border-ink-200 bg-ink-50/50 px-3 py-2 text-xs font-medium text-ink-700"
            >
              <input
                type="checkbox"
                :checked="notificationSettings.offsets.includes(option.value)"
                class="h-3.5 w-3.5 rounded border-ink-300 text-brand-600 focus:ring-brand-400"
                @change="toggleOffset(option.value)"
              />
              {{ option.label }}
            </label>
          </div>
        </div>
      </template>

      <p v-if="notificationsError" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ notificationsError }}
      </p>
      <p
        v-if="notificationsSuccess"
        class="rounded-xl bg-emerald-50 px-3 py-2 text-sm text-emerald-700"
      >
        {{ notificationsSuccess }}
      </p>

      <div>
        <button
          type="submit"
          :disabled="savingNotifications || notificationsLoading"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ savingNotifications ? "Salvando..." : "Salvar" }}
        </button>
      </div>
    </form>
  </div>
</template>
