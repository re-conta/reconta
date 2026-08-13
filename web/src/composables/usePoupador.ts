import { computed, nextTick, onScopeDispose, shallowRef, watch } from "vue";
import { createPoupadorSnapshot, getPoupadorSnapshot } from "../api/poupador";
import type {
  PoupadorEntry,
  PoupadorEntryDraft,
  PoupadorFrequency,
  PoupadorKind,
} from "../types/poupador";

const STORAGE_KEY = "reconta-poupador-v1";
const MONTHS = ["Jan", "Fev", "Mar", "Abr", "Mai", "Jun", "Jul", "Ago", "Set", "Out", "Nov", "Dez"];

interface PoupadorStorage {
  incomes: PoupadorEntry[];
  expenses: PoupadorEntry[];
}

const SAVE_DEBOUNCE_MS = 700;
const SAVE_THROTTLE_MS = 2_000;

function loadEntries(): PoupadorStorage {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (!saved) return { incomes: [], expenses: [] };
    const parsed = JSON.parse(saved) as PoupadorStorage;
    return {
      incomes: Array.isArray(parsed.incomes) ? parsed.incomes : [],
      expenses: Array.isArray(parsed.expenses) ? parsed.expenses : [],
    };
  } catch {
    return { incomes: [], expenses: [] };
  }
}

function monthlyAmount(entry: PoupadorEntry) {
  const multipliers: Record<Exclude<PoupadorFrequency, "one-time">, number> = {
    monthly: 1,
    weekly: 52 / 12,
    biweekly: 26 / 12,
    yearly: 1 / 12,
  };
  return entry.frequency === "one-time" ? 0 : entry.amount * multipliers[entry.frequency];
}

function amountForMonth(entry: PoupadorEntry, month: number) {
  return entry.frequency === "one-time"
    ? entry.month === month
      ? entry.amount
      : 0
    : monthlyAmount(entry);
}

export function usePoupador(snapshotID?: string) {
  const saved = loadEntries();
  const incomes = shallowRef<PoupadorEntry[]>(saved.incomes);
  const expenses = shallowRef<PoupadorEntry[]>(saved.expenses);
  const loadedSnapshotID = shallowRef<string | null>(snapshotID ?? null);
  const shareURL = shallowRef<string | null>(snapshotID ? buildShareURL(snapshotID) : null);
  const saveStatus = shallowRef<"idle" | "saving" | "saved" | "error">("idle");
  const saveError = shallowRef<string | null>(null);
  const isLoadingSnapshot = shallowRef(Boolean(snapshotID));
  const snapshotLoadError = shallowRef<string | null>(null);
  let saveTimer: ReturnType<typeof setTimeout> | undefined;
  let lastRemoteSaveAt = 0;
  let changeVersion = 0;
  let savedVersion = 0;
  let saving = false;
  let loadingSnapshot = false;

  watch([incomes, expenses], ([nextIncomes, nextExpenses]) => {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ incomes: nextIncomes, expenses: nextExpenses }),
    );
    if (loadingSnapshot) return;
    changeVersion += 1;
    scheduleRemoteSave();
  });

  if (snapshotID) void loadSnapshot(snapshotID);

  const incomeMonthlyTotal = computed(() =>
    incomes.value.reduce((sum, entry) => sum + monthlyAmount(entry), 0),
  );
  const expenseMonthlyTotal = computed(() =>
    expenses.value.reduce((sum, entry) => sum + monthlyAmount(entry), 0),
  );
  const yearlyIncome = computed(() =>
    incomes.value.reduce(
      (sum, entry) =>
        sum + (entry.frequency === "one-time" ? entry.amount : monthlyAmount(entry) * 12),
      0,
    ),
  );
  const yearlyExpense = computed(() =>
    expenses.value.reduce(
      (sum, entry) =>
        sum + (entry.frequency === "one-time" ? entry.amount : monthlyAmount(entry) * 12),
      0,
    ),
  );
  const monthlyBalance = computed(() => incomeMonthlyTotal.value - expenseMonthlyTotal.value);
  const yearlyBalance = computed(() => yearlyIncome.value - yearlyExpense.value);
  const monthlySeries = computed(() =>
    MONTHS.map((label, index) => {
      const month = index + 1;
      const income = incomes.value.reduce((sum, entry) => sum + amountForMonth(entry, month), 0);
      const expense = expenses.value.reduce((sum, entry) => sum + amountForMonth(entry, month), 0);
      return { label, income, expense, balance: income - expense };
    }),
  );

  function saveEntry(kind: PoupadorKind, draft: PoupadorEntryDraft, id?: string) {
    const entry = { ...draft, id: id ?? crypto.randomUUID() };
    const target = kind === "income" ? incomes : expenses;
    target.value = id
      ? target.value.map((item) => (item.id === id ? entry : item))
      : [...target.value, entry];
  }

  function removeEntry(kind: PoupadorKind, id: string) {
    const target = kind === "income" ? incomes : expenses;
    target.value = target.value.filter((entry) => entry.id !== id);
  }

  async function loadSnapshot(id: string) {
    loadingSnapshot = true;
    isLoadingSnapshot.value = true;
    snapshotLoadError.value = null;
    try {
      const snapshot = await getPoupadorSnapshot(id);
      incomes.value = snapshot.incomes;
      expenses.value = snapshot.expenses;
      await nextTick();
      loadedSnapshotID.value = snapshot.id;
      shareURL.value = buildShareURL(snapshot.id);
      saveStatus.value = "saved";
    } catch {
      snapshotLoadError.value = "Não foi possível encontrar este resultado compartilhado.";
      loadedSnapshotID.value = null;
      shareURL.value = null;
    } finally {
      loadingSnapshot = false;
      isLoadingSnapshot.value = false;
    }
  }

  function scheduleRemoteSave() {
    if (saveTimer) clearTimeout(saveTimer);
    const elapsed = Date.now() - lastRemoteSaveAt;
    const delay = Math.max(SAVE_DEBOUNCE_MS, SAVE_THROTTLE_MS - elapsed);
    saveTimer = setTimeout(() => {
      saveTimer = undefined;
      void saveSnapshot();
    }, delay);
  }

  async function saveSnapshot() {
    if (saving) return;
    const version = changeVersion;
    saving = true;
    saveStatus.value = "saving";
    saveError.value = null;
    try {
      const snapshot = await createPoupadorSnapshot({
        incomes: incomes.value,
        expenses: expenses.value,
      });
      loadedSnapshotID.value = snapshot.id;
      shareURL.value = buildShareURL(snapshot.id);
      savedVersion = version;
      lastRemoteSaveAt = Date.now();
      saveStatus.value = "saved";
    } catch {
      saveError.value =
        "Não foi possível salvar o resultado. Tente novamente ao fazer outra alteração.";
      saveStatus.value = "error";
    } finally {
      saving = false;
      if (changeVersion > savedVersion) scheduleRemoteSave();
    }
  }

  onScopeDispose(() => {
    if (saveTimer) clearTimeout(saveTimer);
  });

  return {
    months: MONTHS,
    incomes,
    expenses,
    incomeMonthlyTotal,
    expenseMonthlyTotal,
    yearlyIncome,
    yearlyExpense,
    monthlyBalance,
    yearlyBalance,
    monthlySeries,
    loadedSnapshotID,
    shareURL,
    saveStatus,
    saveError,
    isLoadingSnapshot,
    snapshotLoadError,
    saveEntry,
    removeEntry,
  };
}

function buildShareURL(id: string) {
  return import.meta.env.PROD ? `https://poupa.reconta.app/${id}` : `/poupa/${id}`;
}
