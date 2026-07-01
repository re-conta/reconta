import { createWebHistory, createRouter } from "vue-router";

import HomeView from "./views/home.vue";
import LoginView from "./views/login.vue";
import RegisterView from "./views/register.vue";
import UsersView from "./views/users.vue";
import NotFound from "./views/notfound.vue";

const routes = [
  { path: "/", component: HomeView },
  { path: "/login", name: "Login", component: LoginView },
  { path: "/register", name: "Register", component: RegisterView },
  { path: "/users", name: "Users", component: UsersView },
  { path: "/:pathMatch(.*)*", name: "NotFound", component: NotFound },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
