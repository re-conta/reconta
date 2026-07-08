<script setup lang="ts">
import { ScrollText } from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";

defineEmits<{ close: [] }>();

const appName = import.meta.env.VITE_APP_NAME;

const sections = [
  {
    title: "1. Aceitação dos termos",
    body: [
      `Ao criar uma conta e utilizar o ${appName}, você concorda com estes Termos de Uso e com a nossa Política de Privacidade. Se não concordar, não utilize o serviço.`,
    ],
  },
  {
    title: "2. Descrição do serviço",
    body: [
      `O ${appName} é uma ferramenta de organização financeira pessoal que permite cadastrar contas, registrar e importar transações, controlar contas fixas e visualizar relatórios sobre seus gastos e receitas.`,
      "O serviço não presta consultoria financeira, contábil ou de investimentos — as informações exibidas são baseadas exclusivamente nos dados que você insere.",
    ],
  },
  {
    title: "3. Cadastro e responsabilidade da conta",
    body: [
      "Você é responsável por manter a confidencialidade da sua senha e por todas as atividades realizadas na sua conta.",
      "As informações fornecidas no cadastro devem ser verdadeiras e mantidas atualizadas.",
      "Contas administradas por um administrador de organização podem ser gerenciadas por ele, incluindo acesso a usuários vinculados.",
    ],
  },
  {
    title: "4. Uso adequado",
    body: [
      "É proibido usar o serviço para fins ilícitos, tentar acessar dados de outros usuários sem autorização, ou comprometer a segurança da plataforma.",
      "Reservamo-nos o direito de suspender contas que violem estes termos.",
    ],
  },
  {
    title: "5. Dados importados",
    body: [
      "Ao importar extratos, você declara ter o direito de utilizar essas informações e é responsável por conferir a exatidão dos dados importados.",
    ],
  },
  {
    title: "6. Disponibilidade do serviço",
    body: [
      "Fazemos o possível para manter o serviço disponível, mas não garantimos operação ininterrupta. Manutenções podem causar indisponibilidade temporária.",
      "Recomendamos manter cópias dos seus dados importantes, como extratos originais e relatórios exportados.",
    ],
  },
  {
    title: "7. Cancelamento",
    body: [
      "Você pode encerrar sua conta a qualquer momento em Configurações. Ao encerrar, seus dados são removidos conforme descrito na Política de Privacidade.",
    ],
  },
  {
    title: "8. Alterações destes termos",
    body: [
      "Podemos atualizar estes termos periodicamente. Mudanças relevantes serão comunicadas por e-mail ou aviso no site antes de entrarem em vigor.",
    ],
  },
];
</script>

<template>
  <BaseModal
    title="Termos de Uso"
    :subtitle="`Regras de uso do ${appName}`"
    :icon="ScrollText"
    @close="$emit('close')"
  >
    <div class="flex flex-col gap-6">
      <p class="text-sm leading-relaxed text-ink-600">
        Estes termos regulam o uso do {{ appName }}. Ao utilizar o serviço, você concorda com as
        condições descritas abaixo.
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
        Dúvidas sobre estes termos podem ser enviadas para
        <a class="font-medium text-brand-600 hover:underline" href="mailto:contato@reconta.app">contato@reconta.app</a>.
      </p>
    </div>
  </BaseModal>
</template>
