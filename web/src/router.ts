import { createWebHistory, createRouter } from "vue-router";

import AccountsView from "./views/accounts.vue";
import CategoriesView from "./views/categories.vue";
import HomeView from "./views/home.vue";
import LoginView from "./views/login.vue";
import RegisterView from "./views/register.vue";
import TagsView from "./views/tags.vue";
import TransactionsView from "./views/transactions.vue";
import UsersView from "./views/users.vue";
import NotFound from "./views/notfound.vue";

const routes = [
  { path: "/", component: HomeView },
  { path: "/login", name: "Login", component: LoginView },
  { path: "/register", name: "Register", component: RegisterView },
  { path: "/users", name: "Users", component: UsersView },
  { path: "/contas-bancarias", name: "Accounts", component: AccountsView },
  { path: "/categorias", name: "Categories", component: CategoriesView },
  { path: "/tags", name: "Tags", component: TagsView },
  { path: "/transacoes", name: "Transactions", component: TransactionsView },
  { path: "/:pathMatch(.*)*", name: "NotFound", component: NotFound },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
