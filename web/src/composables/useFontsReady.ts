import { onMounted, ref } from "vue";

/**
 * Reativo ao carregamento das fontes via Font Loading API.
 * Usado para não exibir texto na fonte "errada" (fallback) por muito tempo -
 * enquanto as fontes não estiverem prontas, a UI pode mostrar um skeleton.
 */
export function useFontsReady() {
  const fontsReady = ref(false);

  onMounted(async () => {
    if (typeof document === "undefined" || !("fonts" in document)) {
      fontsReady.value = true;
      return;
    }

    // `document.fonts.ready` só reflete fontes que o layout já decidiu
    // carregar - nesse ponto do onMounted isso ainda pode não ter
    // acontecido, fazendo a promise resolver "vazia" e cedo demais.
    // `fonts.load()` dispara o carregamento explicitamente pelos pesos
    // usados na página, garantindo que o estado reflita o download real.
    // .catch() garante que uma falha de rede (ex: bloqueador de anúncios,
    // offline) resolva em vez de rejeitar - senão o Promise.race abaixo
    // rejeitaria e o skeleton ficaria preso na tela para sempre.
    const load = Promise.all([
      document.fonts.load('400 1em "Nunito"').catch(() => null),
      document.fonts.load('700 1em "Nunito"').catch(() => null),
    ]);

    // Fallback de segurança caso o carregamento nunca resolva.
    const timeout = new Promise((resolve) => setTimeout(resolve, 2000));

    await Promise.race([load, timeout]);
    fontsReady.value = true;
  });

  return fontsReady;
}
