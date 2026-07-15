// Integração com o SDK JS v2 do Mercado Pago, usado apenas para pagamentos
// com cartão: o número do cartão nunca passa pelo nosso backend — o SDK envia
// os dados direto ao Mercado Pago e devolve um token de uso único.

export interface MpPaymentMethod {
  id: string;
  payment_type_id: "credit_card" | "debit_card" | string;
  issuer?: { id: number | string };
}

export interface CardTokenInput {
  cardNumber: string;
  cardholderName: string;
  cardExpirationMonth: string;
  cardExpirationYear: string;
  securityCode: string;
  identificationType: string;
  identificationNumber: string;
}

interface MercadoPagoInstance {
  getPaymentMethods(params: { bin: string }): Promise<{ results: MpPaymentMethod[] }>;
  createCardToken(input: CardTokenInput): Promise<{ id: string }>;
}

declare global {
  interface Window {
    MercadoPago?: new (publicKey: string, options?: { locale?: string }) => MercadoPagoInstance;
  }
}

const SDK_URL = "https://sdk.mercadopago.com/js/v2";

let sdkPromise: Promise<MercadoPagoInstance> | null = null;

function loadSdk(): Promise<MercadoPagoInstance> {
  if (sdkPromise) return sdkPromise;

  const publicKey = import.meta.env.VITE_MP_PUBLIC_KEY as string | undefined;
  if (!publicKey) {
    return Promise.reject(
      new Error("Pagamento com cartão indisponível: chave pública do Mercado Pago não configurada"),
    );
  }

  sdkPromise = new Promise<MercadoPagoInstance>((resolve, reject) => {
    const create = () => {
      if (!window.MercadoPago) {
        reject(new Error("Falha ao carregar o SDK do Mercado Pago"));
        return;
      }
      resolve(new window.MercadoPago(publicKey, { locale: "pt-BR" }));
    };

    if (window.MercadoPago) {
      create();
      return;
    }
    const script = document.createElement("script");
    script.src = SDK_URL;
    script.async = true;
    script.onload = create;
    script.onerror = () => {
      sdkPromise = null;
      reject(new Error("Falha ao carregar o SDK do Mercado Pago"));
    };
    document.head.appendChild(script);
  });
  return sdkPromise;
}

export function useMercadoPago() {
  // Detecta bandeira/emissor pelos 6 primeiros dígitos do cartão, filtrando
  // pelo tipo escolhido no modal (crédito ou débito).
  async function detectPaymentMethod(
    bin: string,
    type: "credit_card" | "debit_card",
  ): Promise<MpPaymentMethod | null> {
    const mp = await loadSdk();
    const { results } = await mp.getPaymentMethods({ bin });
    return results.find((m) => m.payment_type_id === type) ?? results[0] ?? null;
  }

  async function createCardToken(input: CardTokenInput): Promise<string> {
    const mp = await loadSdk();
    const { id } = await mp.createCardToken(input);
    return id;
  }

  return { detectPaymentMethod, createCardToken };
}
