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
              <router-link class="dropdown-item" :to="'/devices/'">
                Linked devices
              </router-link>
              <button class="dropdown-item" href="#"></button>
              <button class="dropdown-item" @click="unregister">Unregister</button>
            </div>
          </div>
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
        this.showSettingsMenu =false;
        this.$store.dispatch("clearMessageList");
      },
      toggleSettings() {
        this.showSettingsMenu = !this.showSettingsMenu;
      },
      unregister(){
        this.$store.dispatch("unregister");

      },
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
    height:50px;
    z-index: 2;
    display: flex;
    justify-content: center;
    align-items: center;
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
    font-size: 20px;
  }
</style>
