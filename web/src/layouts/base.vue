<template>
  <div class="flex flex-col min-h-screen">
    <header class="sticky flex flex-row md:flex-col z-50 bg-gray-300 top-0 p-4">
      <div>Header</div>
      <nav>
        <RouterLink to="/">Início</RouterLink>
      </nav>
    </header>
    <main class="grow">
      <slot />
    </main>
    <footer class="sticky z-50 bg-gray-300 bottom-0 p-4">footer</footer>
  </div>
</template>
