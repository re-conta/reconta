<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ApiError, createTag, deleteTag, listTags, updateTag } from "../api/tags";
import type { Tag, TagInput } from "../types/tag";

const tags = ref<Tag[]>([]);
const errorMessage = ref("");
const loading = ref(true);
const submitting = ref(false);

const editingId = ref<number | null>(null);
const showForm = ref(false);
const form = reactive<TagInput>({ name: "", color: "#6366f1" });

async function loadTags() {
  loading.value = true;
  errorMessage.value = "";
  try {
    tags.value = await listTags();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar tags";
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  form.name = "";
  form.color = "#6366f1";
  editingId.value = null;
  showForm.value = false;
}

function startCreate() {
  resetForm();
  showForm.value = true;
}

function startEdit(tag: Tag) {
  editingId.value = tag.id;
  form.name = tag.name;
  form.color = tag.color;
  showForm.value = true;
}

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    if (editingId.value) {
      await updateTag(editingId.value, { ...form });
    } else {
      await createTag({ ...form });
    }
    resetForm();
    await loadTags();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar tag";
  } finally {
    submitting.value = false;
  }
}

async function handleDelete(id: number) {
  if (!confirm("Excluir esta tag?")) return;
  try {
    await deleteTag(id);
    await loadTags();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir tag";
  }
}

onMounted(loadTags);
</script>

<template>
  <div class="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-10 sm:py-14">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Tags</h1>
        <p class="mt-0.5 text-sm text-ink-500">Etiquetas livres para organizar transações</p>
      </div>
      <button
        type="button"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="startCreate"
      >
        + Nova tag
      </button>
    </div>

    <form
      v-if="showForm"
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleSubmit"
    >
      <div class="grid gap-4 sm:grid-cols-[1fr_auto]">
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
          <span class="text-sm font-medium text-ink-700">Cor</span>
          <input v-model="form.color" type="color" class="h-[42px] w-16 cursor-pointer rounded-xl border border-ink-200" />
        </label>
      </div>
      <div class="flex gap-3">
        <button
          type="submit"
          :disabled="submitting"
          class="rounded-full bg-ink-900 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800 disabled:opacity-50"
        >
          {{ submitting ? "Salvando..." : "Salvar" }}
        </button>
        <button
          type="button"
          class="rounded-full border border-ink-200 px-4 py-2.5 text-sm font-semibold text-ink-700 transition hover:bg-ink-100"
          @click="resetForm"
        >
          Cancelar
        </button>
      </div>
    </form>

    <div class="overflow-hidden rounded-3xl border border-ink-200/70 bg-white shadow-sm">
      <div v-if="loading" class="flex flex-col items-center gap-2 p-12 text-sm text-ink-400">
        <span class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">{{ errorMessage }}</p>
      <div v-else-if="tags.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">Nenhuma tag cadastrada ainda</p>
        <p class="text-sm text-ink-400">Crie a primeira tag para começar.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li v-for="tag in tags" :key="tag.id" class="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-ink-50/60">
          <div class="flex items-center gap-2.5">
            <span class="h-3 w-3 shrink-0 rounded-full" :style="{ backgroundColor: tag.color }"></span>
            <p class="truncate text-sm font-semibold text-ink-900">{{ tag.name }}</p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button type="button" class="text-xs font-semibold text-brand-700 hover:text-brand-800" @click="startEdit(tag)">
              Editar
            </button>
            <button
              type="button"
              class="text-xs font-semibold text-coral-600 hover:text-coral-700"
              @click="handleDelete(tag.id)"
            >
              Excluir
            </button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
