<script setup lang="ts">
import { ShieldCheck } from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";

defineEmits<{ close: [] }>();

const appName = import.meta.env.VITE_APP_NAME;

const sections = [
  {
    title: "1. Quais dados coletamos",
    body: [
      "Dados de cadastro: nome, e-mail, senha (armazenada de forma criptografada) e, quando você usa login com Google, sua foto de perfil.",
      "Dados financeiros que você mesmo informa: contas, transações, categorias, tags, contas recorrentes e extratos importados.",
      "Dados técnicos básicos de uso, como data de acesso e preferências de notificação, para o funcionamento do serviço.",
    ],
  },
  {
    title: "2. Como usamos seus dados",
    body: [
      `Para operar o ${appName}: exibir seus lançamentos, gerar relatórios, calcular vencimentos de contas fixas e enviar notificações que você configurar.`,
      "Para comunicação essencial, como e-mails de verificação de conta, redefinição de senha e lembretes de contas a vencer, quando ativados por você.",
      "Nunca vendemos ou compartilhamos seus dados financeiros com terceiros para fins de publicidade.",
    ],
  },
  {
    title: "3. Onde os dados ficam armazenados",
    body: [
      "Seus dados são armazenados em servidor próprio, protegido por HTTPS, e o acesso à aplicação exige autenticação.",
      "Senhas nunca são armazenadas em texto puro — apenas seu hash criptográfico.",
    ],
  },
  {
    title: "4. Compartilhamento com terceiros",
    body: [
      "Utilizamos provedores de infraestrutura (hospedagem e envio de e-mail transacional) estritamente para operar o serviço, sob obrigação de confidencialidade.",
      "Se você usa login com Google, compartilhamos com o Google apenas o necessário para autenticação — nunca enviamos seus dados financeiros a ele.",
    ],
  },
  {
    title: "5. Seus direitos",
    body: [
      "Você pode acessar, corrigir ou excluir seus dados a qualquer momento em Configurações, ou solicitando ao suporte.",
      "Você pode exportar suas transações e relatórios quando quiser, e desativar notificações por e-mail ou no site.",
      "Ao excluir sua conta, seus dados pessoais e financeiros são removidos permanentemente dos nossos sistemas.",
    ],
  },
  {
    title: "6. Cookies e sessão",
    body: [
      "Usamos apenas cookies/armazenamento local essenciais para manter você autenticado. Não usamos cookies de rastreamento publicitário.",
    ],
  },
  {
    title: "7. Alterações desta política",
    body: [
      "Podemos atualizar esta política para refletir mudanças no serviço. Alterações relevantes serão comunicadas por e-mail ou aviso no site.",
    ],
  },
];
</script>

<template>
  <BaseModal
    title="Política de Privacidade"
    :subtitle="`Como o ${appName} trata os seus dados`"
    :icon="ShieldCheck"
    @close="$emit('close')"
  >
    <div class="flex flex-col gap-6">
      <p class="text-sm leading-relaxed text-ink-600">
        Esta política descreve, de forma simples e transparente, quais dados o {{ appName }}
        coleta, como eles são usados e quais direitos você tem sobre eles.
      </p>

      <div v-for="section in sections" :key="section.title" class="flex flex-col gap-2">
        <h3 class="text-sm font-semibold text-ink-900">{{ section.title }}</h3>
        <ul class="flex flex-col gap-1.5">
          <li
            v-for="(item, idx) in section.body"
            :key="idx"
            class="flex gap-2 text-sm leading-relaxed text-ink-500"
          >
            <span class="mt-2 h-1 w-1 shrink-0 rounded-full bg-ink-300"></span>
            <span>{{ item }}</span>
          </li>
        </ul>
      </div>

      <p class="text-xs text-ink-400">
        Dúvidas sobre privacidade ou solicitações relacionadas aos seus dados podem ser enviadas
        para <a class="font-medium text-brand-600 hover:underline" href="mailto:contato@reconta.app">contato@reconta.app</a>.
      </p>
    </div>
  </BaseModal>
</template>
