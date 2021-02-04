<template>
  <div class="warning-box" v-if="mainWarningMessage">
      <span class="close-warning-box" @click="mainWarningMessage = false;">x</span>
      Due to upstream changes in Signal, some features are currently broken.
      We're working as fast as we can to bring them back.
      Follow the progress or join us to help with development <a href="https://t.me/axolotl_dev" target="_blank" @click="openExtern($event, 'https://t.me/axolotl_dev')">telegram</a>.
  </div>
</template>

<script>
export default {
  name: 'WarningMessage',
  data() {
    return {
      mainWarningMessage: true
    }
  },
  methods: {
    openExtern(e, url){
      if(this.gui=="ut"){
        e.preventDefault();
        alert(url)
      }
    }
  }
}
</script>
