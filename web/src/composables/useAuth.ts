import { ref } from "vue";
import { fetchCurrentUser, login as apiLogin, logout as apiLogout } from "../api/auth";
import type { User } from "../types/user";

const currentUser = ref<User | null>(null);
const initialized = ref(false);
const loading = ref(false);

let initPromise: Promise<void> | null = null;

async function init() {
  if (initPromise) return initPromise;
  loading.value = true;
  initPromise = fetchCurrentUser()
    .then((user) => {
      currentUser.value = user;
    })
    .finally(() => {
      loading.value = false;
      initialized.value = true;
    });
  return initPromise;
}

async function login(email: string, password: string) {
  const user = await apiLogin({ email, password });
  currentUser.value = user;
}

async function logout() {
  await apiLogout();
  currentUser.value = null;
}

function setCurrentUser(user: User) {
  currentUser.value = user;
}

export function useAuth() {
  return { currentUser, initialized, loading, init, login, logout, setCurrentUser };
}
