<script setup>
import { ref } from "vue";

defineProps({
  src: { type: String, required: true },
  alt: { type: String, default: "" },
  class: { type: String, default: "" },
});

const isLoaded = ref(false);

function handleLoad() {
  isLoaded.value = true;
}
</script>

<template>
  <div class="lazy-image-container" :class="class">
    <!-- Skeleton Overlay -->
    <div v-if="!isLoaded" class="skeleton-loader"></div>

    <!-- Native Lazy Loaded Image -->
    <img
      :src="src"
      :alt="alt"
      loading="lazy"
      :class="{ 'is-hidden': !isLoaded }"
      @load="handleLoad"
    />
  </div>
</template>

<style scoped>
.lazy-image-container {
  position: relative;
  overflow: hidden;
  display: block;
}

img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity 0.3s ease-in-out;
}

.is-hidden {
  opacity: 0;
  position: absolute;
}

/* Skeleton Animation */
.skeleton-loader {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite linear;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
