import { createWebHistory, createRouter } from "vue-router";

import HomeView from "./views/home.vue";
import NotFound from "./views/notfound.vue";

const routes = [
  { path: "/", component: HomeView },
  // { path: '/users/:id', component: () => import('./views/UserDetails.vue') },
  { path: "/:pathMatch(.*)*", name: "NotFound", component: NotFound },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
