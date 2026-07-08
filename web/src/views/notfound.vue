<template>
  <div
    class="mx-auto flex min-h-[calc(100vh-140px)] max-w-md flex-col items-center justify-center gap-6 px-6 text-center"
  >
    <img src="/images/not-found.svg" alt="" class="w-56 drop-shadow-lg" />
    <div>
      <h1 class="font-display text-2xl font-bold text-ink-900">Página não encontrada</h1>
      <p class="mt-2 text-sm text-ink-500">O endereço que você acessou não existe ou foi movido.</p>
    </div>
    <RouterLink
      to="/"
      class="rounded-full bg-ink-900 px-6 py-2.5 text-sm font-semibold text-white shadow-lg shadow-ink-900/10 transition hover:bg-ink-800"
    >
      Voltar ao início
    </RouterLink>
  </div>
</template>
