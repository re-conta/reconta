<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import {
  ApiError,
  createCategory,
  deleteCategory,
  listCategories,
  updateCategory,
} from "../api/categories";
import type { Category, CategoryInput } from "../types/category";

const categories = ref<Category[]>([]);
const errorMessage = ref("");
const loading = ref(true);
const submitting = ref(false);

const editingId = ref<number | null>(null);
const showForm = ref(false);
const showPatternsHelp = ref(false);
const form = reactive<CategoryInput>({
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

async function loadCategories() {
  loading.value = true;
  errorMessage.value = "";
  try {
    categories.value = await listCategories();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao carregar categorias";
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
}

function startCreate() {
  resetForm();
  showForm.value = true;
}

function startEdit(category: Category) {
  editingId.value = category.id;
  form.name = category.name;
  form.color = category.color;
  form.icon = category.icon;
  form.type = category.type;
  form.patterns = category.patterns;
  showForm.value = true;
}

async function handleSubmit() {
  errorMessage.value = "";
  submitting.value = true;
  try {
    if (editingId.value) {
      await updateCategory(editingId.value, { ...form });
    } else {
      await createCategory({ ...form });
    }
    resetForm();
    await loadCategories();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao salvar categoria";
  } finally {
    submitting.value = false;
  }
}

async function handleDelete(id: number) {
  if (!confirm("Excluir esta categoria?")) return;
  try {
    await deleteCategory(id);
    await loadCategories();
  } catch (err) {
    errorMessage.value = err instanceof ApiError ? err.message : "Falha ao excluir categoria";
  }
}

onMounted(loadCategories);
</script>

<template>
  <div class="mx-auto flex w-full max-w-4xl flex-col gap-6 px-2 py-4 md:px-6 md:py-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-display text-2xl font-bold text-ink-900">Categorias</h1>
        <p class="mt-0.5 text-sm text-ink-500">Organize receitas e despesas</p>
      </div>
      <button
        type="button"
        class="rounded-full bg-ink-900 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
        @click="startCreate"
      >
        + Nova categoria
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
        <label class="flex flex-col gap-1.5">
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
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-ink-700">Ícone (nome lucide)</span>
          <input
            v-model="form.icon"
            type="text"
            placeholder="circle"
            class="rounded-xl border border-ink-200 bg-ink-50/50 px-3.5 py-2.5 text-sm text-ink-900 outline-none transition focus:border-brand-400 focus:bg-white focus:ring-4 focus:ring-brand-100"
          />
        </label>
      </div>
      <label class="flex flex-col gap-1.5">
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
        v-else-if="categories.length === 0"
        class="flex flex-col items-center gap-1 p-12 text-center"
      >
        <p class="text-sm font-medium text-ink-600">Nenhuma categoria cadastrada ainda</p>
        <p class="text-sm text-ink-400">Crie a primeira categoria para começar.</p>
      </div>
      <ul v-else class="divide-y divide-ink-100">
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
              @click="startEdit(category)"
            >
              Editar
            </button>
            <button
              type="button"
              class="text-xs font-semibold text-coral-600 hover:text-coral-700"
              @click="handleDelete(category.id)"
            >
              Excluir
            </button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>
