<script setup lang="ts">
import {
  BadgeDollarSign,
  Bell,
  CircleHelp,
  FolderKanban,
  LayoutDashboard,
  Repeat,
  Settings,
  Tags,
  Upload,
  Users,
  Wallet,
} from "lucide-vue-next";
import BaseModal from "./BaseModal.vue";

defineEmits<{ close: [] }>();

const appName = import.meta.env.VITE_APP_NAME;

const sections = [
  {
    icon: Wallet,
    title: "Contas",
    text: "Cadastre suas contas bancárias, carteiras e cartões. Elas servem para organizar de onde o dinheiro sai e para onde entra em cada transação.",
  },
  {
    icon: LayoutDashboard,
    title: "Transações",
    text: "Registre receitas e despesas manualmente ou importe um extrato. Cada transação pode ser vinculada a uma conta, categoria e tags para facilitar a análise depois.",
  },
  {
    icon: Upload,
    title: "Importar extratos",
    text: "Envie um arquivo de extrato (CSV/OFX) e o sistema identifica as transações automaticamente, evitando digitação manual e duplicidades.",
  },
  {
    icon: Repeat,
    title: "Contas recorrentes",
    text: "Cadastre contas fixas (aluguel, assinaturas, financiamentos) com periodicidade. O sistema avisa quando estão perto do vencimento e permite marcar como paga com um clique.",
  },
  {
    icon: FolderKanban,
    title: "Categorias",
    text: "Organize suas transações por categoria (ex.: Alimentação, Transporte, Lazer) para entender para onde o seu dinheiro está indo.",
  },
  {
    icon: Tags,
    title: "Tags",
    text: "Use tags livres para marcar transações com contextos específicos (ex.: viagem, reembolsável) que cruzam várias categorias.",
  },
  {
    icon: BadgeDollarSign,
    title: "Relatórios",
    text: "Visualize gráficos de fluxo de caixa e gastos por categoria em um período, para acompanhar sua evolução financeira ao longo do tempo.",
  },
  {
    icon: Bell,
    title: "Notificações",
    text: "Receba avisos no site e, se preferir, por e-mail, sobre contas fixas a vencer. É possível escolher com quantos dias de antecedência quer ser avisado, em Configurações.",
  },
  {
    icon: Settings,
    title: "Configurações",
    text: "Atualize seus dados pessoais, defina ou troque sua senha e ajuste suas preferências de notificação.",
  },
  {
    icon: Users,
    title: "Usuários (administradores)",
    text: "Contas com perfil de administrador podem gerenciar os demais usuários da conta a partir do menu de usuário.",
  },
];
</script>

<template>
  <BaseModal
    title="Central de ajuda"
    :subtitle="`Como o ${appName} funciona`"
    :icon="CircleHelp"
    @close="$emit('close')"
  >
    <div class="flex flex-col gap-6">
      <p class="text-sm leading-relaxed text-ink-600">
        O {{ appName }} centraliza suas contas, transações e contas fixas em um só lugar, para você
        organizar sua vida financeira sem depender de planilhas. Veja abaixo o que cada área do
        sistema faz.
      </p>

      <div class="grid gap-3 sm:grid-cols-2">
        <div
          v-for="section in sections"
          :key="section.title"
          class="flex gap-3 rounded-2xl border border-ink-100 bg-ink-50/50 p-4"
        >
          <span
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-white text-brand-600 shadow-sm ring-1 ring-ink-100"
          >
            <component :is="section.icon" class="h-4.5 w-4.5" />
          </span>
          <div class="min-w-0">
            <h3 class="text-sm font-semibold text-ink-900">{{ section.title }}</h3>
            <p class="mt-0.5 text-sm leading-relaxed text-ink-500">{{ section.text }}</p>
          </div>
        </div>
      </div>

      <div class="rounded-2xl bg-brand-50 p-4 text-sm text-brand-800">
        <p class="font-semibold">Dica</p>
        <p class="mt-1 leading-relaxed">
          Comece cadastrando suas contas em <strong>Contas</strong>, depois importe um extrato ou
          lance suas transações manualmente. Com algumas transações categorizadas, os
          <strong>Relatórios</strong> já mostram uma visão clara de como o seu dinheiro se move.
        </p>
      </div>

      <p class="text-xs text-ink-400">
        Ainda com dúvidas ou encontrou um problema? Entre em contato com o suporte pelo e-mail
        indicado no rodapé do site.
      </p>
    </div>
  </BaseModal>
</template>
