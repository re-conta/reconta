import { createWebHistory, createRouter } from "vue-router";

import AccountsView from "./views/accounts.vue";
import CategoriesView from "./views/categories.vue";
import HomeView from "./views/home.vue";
import ImportView from "./views/import.vue";
import LoginView from "./views/login.vue";
import RegisterView from "./views/register.vue";
import TagsView from "./views/tags.vue";
import TransactionsView from "./views/transactions.vue";
import UsersView from "./views/users.vue";
import NotFound from "./views/notfound.vue";
import { useAuth } from "./composables/useAuth";

const routes = [
  { path: "/", component: HomeView },
  { path: "/login", name: "Login", component: LoginView },
  { path: "/register", name: "Register", component: RegisterView },
  { path: "/users", name: "Users", component: UsersView },
  {
    path: "/contas-bancarias",
    name: "Accounts",
    component: AccountsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/categorias",
    name: "Categories",
    component: CategoriesView,
    meta: { requiresAuth: true },
  },
  { path: "/tags", name: "Tags", component: TagsView, meta: { requiresAuth: true } },
  {
    path: "/transacoes",
    name: "Transactions",
    component: TransactionsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/importar-extrato",
    name: "Import",
    component: ImportView,
    meta: { requiresAuth: true },
  },
  { path: "/:pathMatch(.*)*", name: "NotFound", component: NotFound },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return true;

  const { currentUser, initialized, init } = useAuth();
  if (!initialized.value) await init();

  if (!currentUser.value) {
    return { name: "Login", query: { redirect: to.fullPath } };
  }

  return true;
});
