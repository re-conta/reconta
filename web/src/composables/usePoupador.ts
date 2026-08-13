import { computed, shallowRef, watch } from "vue";
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

export function usePoupador() {
  const saved = loadEntries();
  const incomes = shallowRef<PoupadorEntry[]>(saved.incomes);
  const expenses = shallowRef<PoupadorEntry[]>(saved.expenses);

  watch([incomes, expenses], ([nextIncomes, nextExpenses]) => {
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ incomes: nextIncomes, expenses: nextExpenses }),
    );
  });

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
    saveEntry,
    removeEntry,
  };
}
