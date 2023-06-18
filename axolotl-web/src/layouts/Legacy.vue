<template>
  <div class="legacy-layout">
    <header>
      <legacy-header />
    </header>
    <main class="container">
      <slot />
    </main>
    <footer>
      <slot name="footer" />
    </footer>
  </div>
</template>
<script>
import LegacyHeader from "@/components/LegacyHeader.vue";
export default {
  name: "LegacyLayout",
  components: {
    LegacyHeader,
  },
};
</script>
<style scoped>
.legacy-layout {
  height: 100vh;
}
</style>
