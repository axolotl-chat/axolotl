<template>
  <div :class="route()+' header '">
    <div class="container">
      <div class="header-row row">
        <div v-if="route()!='chatList' && route()!='register' ">
          <button class="back btn" @click="back()">
            <font-awesome-icon icon="arrow-left" /></button>
        </div>
        <div v-else-if="route()=='chatList'" class="settings-container">
          <div class="dropdown">
            <button class="btn"
                    type="button"
                    @click="toggleSettings"
                    id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <font-awesome-icon icon="ellipsis-v" />
            </button>
            <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
              <button class="dropdown-item" @click="linkDevice">
                Link Device
                </button>
              <a class="dropdown-item" href="#">Another action</a>
              <a class="dropdown-item" href="#">Something else here</a>
            </div>
          </div>
        </div>
        <div v-else="route()=='chatList'">
          <button class="btn" @click="openSettings">
            <font-awesome-icon icon="ellipsis-v" /></button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'Header',
    props: {
      msg: String
    },
    data() {
      return {
        showSettingsMenu: false
      }
    },
    methods: {
      route() {
        return this.$route.name
      },
      back() {
        this.$router.go(-1)
        that.$store.dispatch("clearMessageList");
      },
      toggleSettings() {
        this.showSettingsMenu = !this.showSettingsMenu;
      },
      linkDevice() {
        var result = window.prompt("desktopLink");
        this.showSettingsMenu = false;
        console.log("desktopSync", result, typeof result)
      }
    },
    mounted() {
      window.router = this.$router;
      console.info('App this router:', this.$router)
    }
  }
</script>

<style scoped>
  .header {
    position: fixed;
    width: 100%;
    background-color: #2090ea;
    top: 0px;
    z-index: 2;
  }
  .btn {
    color: #FFF;
  }
  .settings-container{
    align-self: flex-end;
    width: 100%;
    display: flex;
    justify-content: flex-end;
  }
  #settings-dropdown{
    display: block!important;
    border-radius: 0px;
  }
  .back {
    font-size: 30px;
  }
</style>
