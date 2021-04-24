<template>
  <div v-if="mainWarningMessage" class="warning-box mb-2">
    <span class="close-warning-box" @click="disableWarning">x</span>
    <p v-translate class="pb-0 mb-0">
      Due to upstream changes in Signal, some features are currently broken.
      We're working as fast as we can to bring them back.
    </p>
    <div class="d-flex">
      <span v-translate class="mr-1">
        Follow the progress or join us to help with development on
      </span>
      <a
        href="https://t.me/axolotl_dev"
        target="_blank"
        @click="openExtern($event, 'https://t.me/axolotl_dev')"
      >
        telegram
      </a>.
    </div>
  </div>
</template>

<script>
export default {
  name: "WarningMessage",
  data() {
    return {
      mainWarningMessage: true,
    };
  },
  mounted(){
    if(localStorage.getItem("upstreamWarning")){
      this.mainWarningMessage = false; 
    }
  },
  methods: {
    disableWarning(){
      localStorage.setItem("upstreamWarning", true)
    },
    openExtern(e, url) {
      if (this.gui === "ut") {
        e.preventDefault();
        alert(url);
      }
    },
  },
};
</script>
<style scoped>
  .close-warning-box{
    cursor: pointer;
  }
</style>