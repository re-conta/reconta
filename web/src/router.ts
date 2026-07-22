import { createWebHistory, createRouter } from "vue-router";

import AccountsView from "./views/accounts.vue";
import CategoriesView from "./views/categories.vue";
import SharesView from "./views/compartilhamentos.vue";
import FixedBillsView from "./views/contas-fixas.vue";
import ExportView from "./views/export.vue";
import ForgotPasswordView from "./views/forgot-password.vue";
import HomeView from "./views/home.vue";
import ImportView from "./views/import.vue";
import LoginView from "./views/login.vue";
import PlansView from "./views/planos.vue";
import NotificationsView from "./views/notificacoes.vue";
import RegisterView from "./views/register.vue";
import ReportsView from "./views/reports.vue";
import ResetPasswordView from "./views/reset-password.vue";
import SettingsView from "./views/settings.vue";
import TransactionsView from "./views/transactions.vue";
import AdminView from "./views/admin.vue";
import NotFound from "./views/notfound.vue";
import { useAuth } from "./composables/useAuth";
import { canAccessAdmin } from "./types/user";
import { trackPageView } from "./api/analytics";

const routes = [
  { path: "/", name: "Home", component: HomeView },
  { path: "/login", name: "Login", component: LoginView },
  { path: "/register", name: "Register", component: RegisterView },
  { path: "/planos", name: "Plans", component: PlansView },
  { path: "/esqueci-senha", name: "ForgotPassword", component: ForgotPasswordView },
  { path: "/redefinir-senha", name: "ResetPassword", component: ResetPasswordView },
  {
    path: "/admin",
    name: "Admin",
    component: AdminView,
    meta: { requiresAuth: true, requiresAdmin: true },
  },
  { path: "/users", redirect: "/admin" },
  {
    path: "/contas",
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
  { path: "/tags", redirect: "/categorias" },
  {
    path: "/compartilhamentos",
    name: "Shares",
    component: SharesView,
    meta: { requiresAuth: true },
  },
  { path: "/compartilhado", redirect: "/compartilhamentos" },
  {
    path: "/configuracoes",
    name: "Settings",
    component: SettingsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/transacoes",
    name: "Transactions",
    component: TransactionsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/importar",
    name: "Import",
    component: ImportView,
    meta: { requiresAuth: true },
  },
  {
    path: "/relatorios",
    name: "Reports",
    component: ReportsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/exportar",
    name: "Export",
    component: ExportView,
    meta: { requiresAuth: true },
  },
  {
    path: "/recorrentes",
    name: "FixedBills",
    component: FixedBillsView,
    meta: { requiresAuth: true },
  },
  {
    path: "/notificacoes",
    name: "Notifications",
    component: NotificationsView,
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

  if (to.meta.requiresAdmin && !canAccessAdmin(currentUser.value)) {
    return { name: "Home" };
  }

  return true;
});

// Como o Nginx serve a SPA via fallback para index.html, não há log de
// navegação por rota no servidor — cada troca de rota é reportada aqui.
// Ignora o próprio /admin para não poluir as estatísticas com o uso do painel.
router.afterEach((to) => {
  if (to.path.startsWith("/admin")) return;
  trackPageView(to.fullPath, document.referrer);
});
