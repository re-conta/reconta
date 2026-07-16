<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import {
  ApiError,
  createCategory,
  deleteCategory,
  listCategories,
  updateCategory,
} from "../api/categories";
import { createTag, deleteTag, listTags, updateTag } from "../api/tags";
import type { Category, CategoryInput } from "../types/category";
import type { Tag, TagInput } from "../types/tag";

type Tab = "categories" | "tags";

const activeTab = ref<Tab>("categories");

const categories = ref<Category[]>([]);
const tags = ref<Tag[]>([]);
const errorMessage = ref("");
const loading = ref(true);
const submitting = ref(false);

const editingId = ref<number | null>(null);
const showForm = ref(false);
const showPatternsHelp = ref(false);
const form = reactive<CategoryInput & TagInput>({
  name: "",
  color: "#6366f1",
  icon: "circle",
  type: "both",
  patterns: "",
});

const categoryTypes = [
  { value: "expense", label: "Despesa" },
  { value: "income", label: "Receita" },
  { value: "both", label: "Ambos" },
];

const tabs: { id: Tab; label: string }[] = [
  { id: "categories", label: "Categorias" },
  { id: "tags", label: "Tags" },
];

const title = computed(() => (activeTab.value === "categories" ? "Categorias" : "Tags"));
const subtitle = computed(() =>
  activeTab.value === "categories"
    ? "Organize receitas e despesas"
    : "Etiquetas livres para organizar transações",
);
const addLabel = computed(() =>
  activeTab.value === "categories" ? "+ Nova categoria" : "+ Nova tag",
);
const emptyTitle = computed(() =>
  activeTab.value === "categories"
    ? "Nenhuma categoria cadastrada ainda"
    : "Nenhuma tag cadastrada ainda",
);
const emptySubtitle = computed(() =>
  activeTab.value === "categories"
    ? "Crie a primeira categoria para começar."
    : "Crie a primeira tag para começar.",
);

async function loadAll() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const [categoriesResult, tagsResult] = await Promise.all([listCategories(), listTags()]);
    categories.value = categoriesResult;
    tags.value = tagsResult;
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar dados";
  } finally {
    loading.value = false;
  }
}

function resetForm() {
  form.name = "";
  form.color = "#6366f1";
  form.icon = "circle";
  form.type = "both";
  form.patterns = "";
  editingId.value = null;
  showForm.value = false;
  showPatternsHelp.value = false;
}

function switchTab(tab: Tab) {
  activeTab.value = tab;
  resetForm();
}

function startCreate() {
  resetForm();
  showForm.value = true;
}

function startEditCategory(category: Category) {
  editingId.value = category.id;
  form.name = category.name;
  form.color = category.color;
  form.icon = category.icon;
  form.type = category.type;
  form.patterns = category.patterns;
  showForm.value = true;
}

function startEditTag(tag: Tag) {
  editingId.value = tag.id;
  form.name = tag.name;
  form.color = tag.color;
  showForm.value = true;
}

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    if (activeTab.value === "categories") {
      const input: CategoryInput = {
        name: form.name,
        color: form.color,
        icon: form.icon,
        type: form.type,
        patterns: form.patterns,
      };
      if (editingId.value) {
        await updateCategory(editingId.value, input);
      } else {
        await createCategory(input);
      }
    } else {
      const input: TagInput = { name: form.name, color: form.color };
      if (editingId.value) {
        await updateTag(editingId.value, input);
      } else {
        await createTag(input);
      }
    }
    resetForm();
    await loadAll();
  } catch (err) {
    const label = activeTab.value === "categories" ? "categoria" : "tag";
    errorMessage.value = err instanceof ApiError ? err.message : `Falha ao salvar ${label}`;
  } finally {
    submitting.value = false;
  }
}

async function handleDeleteCategory(id: number) {
  if (!confirm("Excluir esta categoria?")) return;
  try {
    await deleteCategory(id);
    await loadAll();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir categoria";
  }
}

async function handleDeleteTag(id: number) {
  if (!confirm("Excluir esta tag?")) return;
  try {
    await deleteTag(id);
    await loadAll();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir tag";
  }
}

onMounted(loadAll);
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div class="flex items-start justify-between">
      <div>
        <h1 class="font-display text-xl md:text-2xl font-bold text-ink-900">{{ title }}</h1>
        <p class="mt-0.5 text-xs md:text-sm text-ink-500">{{ subtitle }}</p>
      </div>
      <button
        type="button"
        class="shrink-0 rounded-full bg-ink-900 px-2 md:px-4 py-2 text-xs md:text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="startCreate"
      >
        {{ addLabel }}
      </button>
    </div>

    <div class="flex gap-1 rounded-full border border-ink-200/70 bg-white p-1 shadow-sm w-fit">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="rounded-full px-4 py-1.5 text-sm font-semibold transition"
        :class="
          activeTab === tab.id
            ? 'bg-ink-900 text-white'
            : 'text-ink-600 hover:bg-ink-100'
        "
        @click="switchTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <form
      v-if="showForm"
      class="flex flex-col gap-4 rounded-3xl border border-ink-200/70 bg-white p-6 shadow-sm"
      @submit.prevent="handleSubmit"
    >
      <div class="grid gap-4 sm:grid-cols-2">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Nome</span>
          <input
            v-model="form.name"
            type="text"
            required
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
        <label v-if="activeTab === 'categories'" class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Tipo</span>
          <select
            v-model="form.type"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          >
            <option v-for="t in categoryTypes" :key="t.value" :value="t.value">
              {{ t.label }}
            </option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Cor</span>
          <input
            v-model="form.color"
            type="color"
            class="h-10.5 w-16 cursor-pointer rounded-xl border border-ink-200"
          />
        </label>
        <label v-if="activeTab === 'categories'" class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Ícone (nome lucide)</span>
          <input
            v-model="form.icon"
            type="text"
            placeholder="circle"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>
      <label v-if="activeTab === 'categories'" class="flex flex-col gap-1.5">
        <span class="flex items-center justify-between text-sm font-medium text-ink-700">
          Padrões de auto-categorização (opcional)
          <button
            type="button"
            class="text-xs font-semibold text-brand-700 hover:text-brand-800"
            @click="showPatternsHelp = !showPatternsHelp"
          >
            {{ showPatternsHelp ? "ocultar ajuda" : "como funciona?" }}
          </button>
        </span>
        <textarea
          v-model="form.patterns"
          rows="4"
          placeholder="uma expressão regular por linha, ex.: ifood&#10;uber"
          class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 font-mono text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
        ></textarea>
        <p v-if="showPatternsHelp" class="rounded-xl bg-ink-50 px-3 py-2 text-xs text-ink-500">
          Uma expressão regular por linha (sem diferenciar maiúsculas/minúsculas). Ao rodar
          "Auto-categorizar" na tela de Transações, cada lançamento sem categoria é comparado à
          descrição + beneficiário PIX. Ex.: <code>ifood</code> casa "iFood *Restaurante";
          <code>^uber</code> casa apenas descrições que começam com "uber".
        </p>
      </label>
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
        <span
          class="h-5 w-5 animate-spin rounded-full border-2 border-brand-300 border-t-transparent"
        ></span>
        Carregando...
      </div>
      <p v-else-if="errorMessage" class="p-8 text-center text-sm text-coral-600">
        {{ errorMessage }}
      </p>
      <div
        v-else-if="activeTab === 'categories' && categories.length === 0"
        class="flex flex-col items-center gap-1 p-12 text-center"
      >
        <p class="text-sm font-medium text-ink-600">{{ emptyTitle }}</p>
        <p class="text-sm text-ink-400">{{ emptySubtitle }}</p>
      </div>
      <ul v-else-if="activeTab === 'categories'" class="divide-y divide-ink-100">
        <li
          v-for="category in categories"
          :key="category.id"
          class="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-ink-50/60"
        >
          <div class="flex min-w-0 items-center gap-2.5">
            <span
              class="h-3 w-3 shrink-0 rounded-full"
              :style="{ backgroundColor: category.color }"
            ></span>
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-ink-900">{{ category.name }}</p>
              <p class="truncate text-xs text-ink-500">
                {{ categoryTypes.find((t) => t.value === category.type)?.label ?? category.type }}
                <span v-if="category.patterns"> &middot; auto-categorização ativa</span>
              </p>
            </div>
          </div>
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              class="text-xs font-semibold text-brand-700 hover:text-brand-800"
              @click="startEditCategory(category)"
            >
              Editar
            </button>
            <button
              type="button"
              class="text-xs font-semibold text-coral-600 hover:text-coral-700"
              @click="handleDeleteCategory(category.id)"
            >
              Excluir
            </button>
          </div>
        </li>
      </ul>
      <div v-else-if="tags.length === 0" class="flex flex-col items-center gap-1 p-12 text-center">
        <p class="text-sm font-medium text-ink-600">{{ emptyTitle }}</p>
        <p class="text-sm text-ink-400">{{ emptySubtitle }}</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
        <li
          v-for="tag in tags"
          :key="tag.id"
          class="flex items-center justify-between gap-3 px-5 py-4 transition hover:bg-ink-50/60"
        >
          <div class="flex items-center gap-2.5">
            <span
              class="h-3 w-3 shrink-0 rounded-full"
              :style="{ backgroundColor: tag.color }"
            ></span>
            <p class="truncate text-sm font-semibold text-ink-900">{{ tag.name }}</p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button
              type="button"
              class="text-xs font-semibold text-brand-700 hover:text-brand-800"
              @click="startEditTag(tag)"
            >
              Editar
            </button>
            <button
              type="button"
              class="text-xs font-semibold text-coral-600 hover:text-coral-700"
              @click="handleDeleteTag(tag.id)"
            >
              Excluir
            </button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
