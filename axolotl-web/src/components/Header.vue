<template>
  <div :class="route()+' header '">
    <div class="container" >
      <div class="overlay" v-if="showSettingsMenu" @click="showSettingsMenu=false"/>
      <div class="header-row row">
        <div v-if="route()=='chat'" class="message-list-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
            <div v-if="messageList.Session" class="header-text">{{messageList.Session.Name}}</div>
        </div>
        <div v-else-if="route()=='register' ">
          <div class="header-text">Connect with Signal</div>
        </div>
        <div v-else-if="route()=='devices' " >
          <button class="back btn" @click="back()">
            <font-awesome-icon icon="arrow-left" /></button>
        </div>
        <div v-else-if="route()=='contacts' " class="row w-100">
          <div class="col-2">
            <button class="back btn" @click="back()">
              <font-awesome-icon icon="arrow-left" />
            </button>
          </div>
          <div class="col-10 text-right">
            <div class="dropdown">
              <button class="btn"
                      type="button"
                      @click="toggleSettings"
                      id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
                <button class="dropdown-item" @click="refreshContacts">Add Contacts</button>
                <input id="addVcf" type="file" @change="readVcf" style="position: fixed; top: -100em">
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="route()=='chatList'" class="settings-container row col-12">
          <div class="dropdown">
            <button class="btn"
                    type="button"
                    @click="toggleSettings"
                    id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <font-awesome-icon icon="ellipsis-v" />
            </button>
            <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
              <router-link class="dropdown-item" :to="'/devices/'" @click="showSettingsMenu=false">
                Linked devices
              </router-link>
              <button class="dropdown-item" @click="unregister">Unregister</button>
            </div>
          </div>
        </div>
        <div v-else>
          <!-- <button class="back btn" @click="back()">
            <font-awesome-icon icon="arrow-left" /></button> -->
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
        this.showSettingsMenu = false;
        this.$store.dispatch("unregister");
      },
      refreshContacts(){
        if(typeof ut !="undefined"){
          var result = window.prompt("refreshContacts");
          console.log(result);
          this.$store.dispatch("refreshContacts", result);
        } else {
          this.showSettingsMenu = false;
          document.getElementById("addVcf").click()
        }
      },
      readVcf(evt) {
          var f = evt.target.files[0];
          if (f) {
            var r = new FileReader();
            var that = this;
            r.onload = function(e) {
                var vcf = e.target.result;
                that.$store.dispatch("uploadVcf", vcf);
            }
            r.readAsText(f)
          } else {
            alert("Failed to load file");
          }
      }
    },
    computed: {
      messageList() {
        return this.$store.state.messageList
      },

    },
    mounted() {
      window.router = this.$router;
      console.info('App this router:', this.$router)
    },
    watch:{
    $route (to, from){
        this.showSettingsMenu = false;
    }
}
  }
</script>

<style scoped>
  .overlay{
    position: fixed;
    width:100vh;
    height:100vh;
  }
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
  .header .text-right{
    justify-content: flex-end;
    display: flex;
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
    right: 5px;
    left: auto;
  }
  .back {
    font-size: 20px;
  }
  .message-list-container{
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .header-text{
    font-weight:bold;
    font-size:20px;
    padding-left:10px;
    color:#FFFFFF;
  }

</style>
