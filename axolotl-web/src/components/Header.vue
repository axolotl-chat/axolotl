<template>
  <div :class="route()+' header '">
    <div class="container" >
      <div class="overlay" v-if="showSettingsMenu" @click="showSettingsMenu=false"/>
      <div class="header-row row">
        <div v-if="route()=='chat'" class="message-list-container row w-100">
            <div class="col-10 chat-header">
              <router-link class="btn" :to="'/chatList'">
                <font-awesome-icon icon="arrow-left" />
              </router-link>
              <div v-if="currentChat!=null&&currentChat.IsGroup" class="group-badge"><font-awesome-icon icon="user-friends" /></div>
              <div v-else class="group-badge">{{currentChat.Name[0]}}</div>
              <div v-if="currentChat!=null&&!currentChat.Notification" class="mute-badge"> <font-awesome-icon class="mute" icon="volume-mute" /></div>
              <div class="header-text" v-if="currentChat!=null&&currentChat.IsGroup&&currentChat.Name==currentChat.Tel"><div>Unknown group</div></div>
              <div class="header-text" v-else><div v-if="currentChat!=null" class="header-text">{{currentChat.Name}}</div>
              </div>
            </div>
            <div class="col-2 text-right">
              <div class="dropdown">
                <button class="btn"
                        type="button"
                        @click="toggleSettings"
                        id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                  <font-awesome-icon icon="ellipsis-v" />
                </button>
                <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
                  <button v-if="currentChat!=null&&currentChat.Notification" class="dropdown-item" @click="toggleNotifications">Mute</button>
                  <button v-else class="dropdown-item" @click="toggleNotifications">Unmute</button>
                  <button v-if="currentChat!=null&&!currentChat.IsGroup" class="dropdown-item" @click="verifyIdentity">Show identity</button>
                  <button v-if="currentChat!=null&&!currentChat.IsGroup" class="dropdown-item" @click="resetEncryptionModal">Reset encryption</button>

                </div>
                <identity-modal
                v-if="showIdentityModal"
                @close="showIdentityModal=false"
                @confirm="showIdentityModal=false" />
                <confirmation-modal
                v-if="showConfirmationModal"
                @close="showConfirmationModal=false"
                @confirm="confirm"
                :title="cMTitle"
                :text="cMText" />
              </div>
            </div>
        </div>
        <div v-else-if="route()=='register' ">
          <div class="header-text">Connect with Signal</div>
        </div>
        <div v-else-if="route()=='password' ">
          <div class="header-text">Enter password</div>
        </div>
        <div v-else-if="route()=='setPassword' " class="list-header-container">
          <router-link class="btn" :to="'/settings'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text">Set encryption password</div>
        </div>
        <div v-else-if="route()=='about' " class="list-header-container">
          <router-link class="btn" :to="'/settings'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
        </div>
        <div v-else-if="route()=='settings'" class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text">Settings</div>
        </div>
        <div v-else-if="route()=='newGroup' " class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text">New group</div>
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
              <router-link class="dropdown-item" :to="'/newGroup'" @click="showSettingsMenu=false">
                New group
              </router-link>
              <router-link class="dropdown-item" :to="'/settings/'" @click="showSettingsMenu=false">
                Settings
              </router-link>
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
  import IdentityModal from "@/components/IdentityModal.vue"
  import ConfirmationModal from "@/components/ConfirmationModal.vue"
  export default {
    name: 'Header',
    components:{
      ConfirmationModal,
      IdentityModal
    },
    props: {
      msg: String
    },
    data() {
      return {
        showSettingsMenu: false,
        showConfirmationModal: false,
        showIdentityModal:false,
        cMTitle:"",
        cMText: "",
        cMType:""
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
      toggleNotifications(){
        this.showSettingsMenu = false;
        this.$store.dispatch("toggleNotifcations");
      },
      resetEncryptionModal(){
        this.showSettingsMenu = false;
        this.showConfirmationModal = true;
        this.cMType="resetEncryption";
        this.cMTitle="Reset secure session?";
        this.cMText="This may help if you're having encryption problems in this conversation. Your messages will be kept.";
      },
      verifyIdentity(){
        this.$store.dispatch("verifyIdentity");
        this.showSettingsMenu = false;
        this.showIdentityModal = true;
      },
      confirm(){
        if(this.cMType=="resetEncryption")this.$store.dispatch("resetEncryption");
        this.showConfirmationModal = false;
        this.showIdentityModal = flase;
      },
      refreshContacts(){
        this.$store.state.importingContacts = true;
        console.log("Import contacts for gui " + this.gui)
        this.showSettingsMenu = false;
        if(this.gui=="ut"){
          var result = window.prompt("refreshContacts");
          if(result!="canceld")
          this.$store.dispatch("refreshContacts", result);
        } else {
          this.showSettingsMenu = false;
          document.getElementById("addVcf").click()
        }
      },
      createGroup(){

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
      currentChat() {
        return this.$store.state.currentChat
      },
      gui() {
        return this.$store.state.gui
      },
      identity() {
        return this.$store.state.identity
      },
    },
    mounted() {
      window.router = this.$router;
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
    -webkit-box-shadow: 0px -11px 14px 7px rgba(0,0,0,0.75);
    -moz-box-shadow: 0px -11px 14px 7px rgba(0,0,0,0.75);
    box-shadow: 0px -11px 14px 7px rgba(0,0,0,0.75);
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
  .list-header-container{
    display:flex;
    align-items: center;
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
  .group-badge{
    background-color:#FFF;
    border-radius:50%;
    width:28px;
    height:28px;
    display:flex;
    justify-content:center;
    align-items:center;
    color: #2090ea;
  }
  .mute-badge {
    color: #FFF;
    margin-left: 10px;
  }
  .chat-header{
    display:flex;
    align-items:center;
  }
</style>
