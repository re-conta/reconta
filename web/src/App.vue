<script setup lang="ts">
import { onMounted } from "vue";
import BaseLayout from "./layouts/base.vue";
import { useAuth } from "./composables/useAuth";

const { init } = useAuth();
onMounted(init);
</script>
<template>
  <BaseLayout>
    <RouterView />
  </BaseLayout>
</template>
