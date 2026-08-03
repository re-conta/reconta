<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { UserPlus } from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";
import { adminCreateUser, adminUpdateUser, ApiError } from "../../api/users";
import { roleLabels, type User, type UserRole } from "../../types/user";
import { formatCnpj, normalizeCnpj } from "../../utils/cnpj";

const props = defineProps<{
  user: User | null;
  assignableRoles: UserRole[];
}>();

const emit = defineEmits<{ close: []; saved: [user: User] }>();

const isEdit = computed(() => props.user !== null);

const form = reactive({
  name: props.user?.name ?? "",
  email: props.user?.email ?? "",
  password: "",
  role: (props.user?.role ?? "pessoa_fisica") as UserRole,
  cnpj: props.user?.cnpj ? formatCnpj(props.user.cnpj) : "",
});

const submitting = ref(false);
const error = ref("");

async function handleSubmit() {
  submitting.value = true;
  error.value = "";
  try {
    const cnpj = normalizeCnpj(form.cnpj);
    const user = isEdit.value
      ? await adminUpdateUser(props.user!.id, { name: form.name, email: form.email, cnpj })
      : await adminCreateUser({
          name: form.name,
          email: form.email,
          password: form.password,
          role: form.role,
          cnpj,
        });
    emit("saved", user);
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : "Falha ao salvar usuário";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <BaseModal
    :title="isEdit ? 'Editar usuário' : 'Novo usuário'"
    :icon="UserPlus"
    @close="emit('close')"
  >
    <form class="flex flex-col gap-4" @submit.prevent="handleSubmit">
      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-ink-700">Nome</span>
        <input
          v-model="form.name"
          type="text"
          required
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        />
      </label>

      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-ink-700">E-mail</span>
        <input
          v-model="form.email"
          type="email"
          required
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        />
      </label>

      <label v-if="!isEdit" class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-ink-700">Senha</span>
        <input
          v-model="form.password"
          type="password"
          minlength="8"
          required
          autocomplete="new-password"
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        />
        <span class="text-xs text-ink-400">Mínimo de 8 caracteres.</span>
      </label>

      <label v-if="!isEdit" class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-ink-700">Cargo</span>
        <select
          v-model="form.role"
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        >
          <option v-for="role in assignableRoles" :key="role" :value="role">
            {{ roleLabels[role] }}
          </option>
        </select>
      </label>

      <label
        v-if="!isEdit ? form.role === 'pessoa_juridica' : Boolean(user?.cnpj)"
        class="flex flex-col gap-1.5"
      >
        <span class="text-sm font-medium text-ink-700">CNPJ</span>
        <input
          v-model="form.cnpj"
          type="text"
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        />
      </label>

      <p v-if="error" class="rounded-xl bg-coral-50 px-3 py-2 text-sm text-coral-700">
        {{ error }}
      </p>

      <div class="flex justify-end gap-3 border-t border-ink-100 pt-4">
        <button
          type="button"
          class="rounded-full border border-ink-200 px-5 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-50"
          @click="emit('close')"
        >
          Cancelar
        </button>
        <button
          type="submit"
          :disabled="submitting"
          class="rounded-full bg-ink-900 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ submitting ? "Salvando..." : "Salvar" }}
        </button>
      </div>
    </form>
  </BaseModal>
</template>
