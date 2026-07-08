<script setup lang="ts">
import UserMenu from "../components/UserMenu.vue";
import { useAuth } from "../composables/useAuth";

const { currentUser } = useAuth();
const appName = import.meta.env.VITE_APP_NAME;
</script>

<template>
  <div class="flex min-h-screen flex-col bg-ink-50">
    <div
      class="pointer-events-none fixed inset-x-0 top-0 -z-10 h-120 bg-[radial-gradient(60%_60%_at_15%_0%,--theme(--color-brand-200/60%),transparent),radial-gradient(50%_50%_at_85%_10%,--theme(--color-coral-200/50%),transparent)]"
    ></div>

    <header
      class="sticky top-0 z-50 flex flex-col gap-3 border-b border-ink-200/70 bg-ink-50/80 px-4 py-3 backdrop-blur-md sm:flex-row sm:items-center sm:justify-between sm:px-8"
    >
      <div class="flex items-center justify-between gap-4">
        <RouterLink to="/" class="flex items-center gap-1.5">
          <img src="/images/favicon.svg" alt="" class="h-8 w-8" />
          <span class="font-display text-xl md:text-2xl font-bold tracking-tight text-ink-900">{{
            appName
          }}</span>
        </RouterLink>
        <div class="flex items-center gap-4 sm:hidden">
          <UserMenu v-if="currentUser" />
          <RouterLink
            v-else
            to="/login"
            class="rounded-full bg-ink-900 px-4 py-1.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          >
            Entrar
          </RouterLink>
        </div>
      </div>
      <div class="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:gap-8">
        <nav
          class="no-scrollbar flex w-full flex-row flex-nowrap items-center gap-5 overflow-x-auto text-sm font-medium text-ink-500 sm:w-auto sm:gap-6 sm:overflow-visible"
        >
          <RouterLink
            to="/"
            class="shrink-0 whitespace-nowrap transition hover:text-ink-900"
            active-class="text-ink-900"
          >
            Início
          </RouterLink>
          <template v-if="currentUser">
            <RouterLink
              to="/transacoes"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Transações
            </RouterLink>
            <RouterLink
              to="/importar-extrato"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Importar extrato
            </RouterLink>
            <RouterLink
              to="/contas-bancarias"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Contas
            </RouterLink>
            <RouterLink
              to="/categorias"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Categorias
            </RouterLink>
            <RouterLink
              to="/tags"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Tags
            </RouterLink>
            <RouterLink
              to="/relatorios"
              class="transition hover:text-ink-900 shrink-0 whitespace-nowrap"
              active-class="text-ink-900"
            >
              Relatórios
            </RouterLink>
          </template>
        </nav>
        <div class="hidden sm:block">
          <UserMenu v-if="currentUser" />
          <RouterLink
            v-else
            to="/login"
            class="rounded-full bg-ink-900 px-4 py-1.5 text-sm font-semibold text-white shadow-sm transition hover:bg-ink-800"
          >
            Entrar
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="grow">
      <slot />
    </main>

    <footer class="border-t border-ink-200/70 px-4 py-6 text-center text-sm text-ink-400 sm:px-8">
      <p>&copy; {{ new Date().getFullYear() }} {{ appName }}. Todos os direitos reservados.</p>
    </footer>
  </div>
</template>
