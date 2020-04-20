<template>
  <div :class="route()+' header '">
    <div class="container-fluid">
      <div class="overlay" v-if="showSettingsMenu" @click="showSettingsMenu=false" />
      <div class="header-row">
        <div v-if="route()=='chat'" class="message-list-container row">
          <div v-if="errorConnection!=null" class="connection-error"></div>
          <div class="col-10 chat-header">
            <button class="btn" @click="back()">
              <font-awesome-icon icon="arrow-left" />
            </button>
            <div class="row w-100" v-if="currentChat!=null">
              <div class="col-2 badge-container">
                <div v-if="currentChat!=null&&currentChat.IsGroup" class="badge-name">
                  <img class="avatar-img" :src="'http://localhost:9080/avatars?file='+currentChat.Tel" @error="onImageError($event)" />
                  <font-awesome-icon icon="user-friends" />
                </div>
                <div v-else class="group-badge">{{currentChat.Name[0]}}</div>
              </div>
              <div class="col-10 center">
                <div class="row">
                  <div class="col-12">
                    <div v-if="currentChat!=null&&!currentChat.Notification" class="mute-badge">
                      <font-awesome-icon class="mute" icon="volume-mute" />
                    </div>
                    <div class="header-text-chat" v-if="currentChat!=null&&currentChat.IsGroup&&currentChat.Name==currentChat.Tel">
                      <div v-translate>Unknown group</div>
                    </div>
                    <div class="header-text-chat" v-else>
                      <div v-if="currentChat!=null&&currentChat.Name!=currentChat.Tel" class="">{{currentChat.Name}}</div>
                    </div>
                  </div>
                  <div class="col-12">
                    <div class="number-text" v-if="currentChat!=null&&currentChat.IsGroup&&currentGroup!=null&&typeof currentGroup!='undefined'">
                      <div v-for="e in currentGroup.Members" v-bind:key="e">{{getNameForTel(e)}}</div>
                    </div>
                    <div class="number-text" v-if="currentChat!=null&&currentChat.IsGroup&&currentGroup!=null&&typeof currentGroup!='undefined'">
                      <div class="name" v-for="(n,i) in names" v-bind:key="i">{{n}}<span v-if="i<names.length-1">,</span></div>
                    </div>
                    <div class="number-text" v-if="currentChat!=null&&!currentChat.IsGroup&&currentChat.Name==currentChat.Tel"> {{messageList.ID}}</div>

                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="col-2 text-right">
            <div class="dropdown">
              <button class="btn" type="button" @click="toggleSettings" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
                <button v-if="currentChat!=null&&!currentChat.IsGroup&&currentChat.Name!=currentChat.Tel" class="dropdown-item" @click="callNumber(currentChat.Tel)">
                {{currentChat.Tel}}
                </button>
                <button v-if="currentChat!=null&&currentChat.Notification" class="dropdown-item" @click="toggleNotifications" v-translate>Mute</button>
                <button v-else class="dropdown-item" @click="toggleNotifications" v-translate>Unmute</button>
                <!-- <button v-if="currentChat!=null&&currentChat.IsGroup" class="dropdown-item" @click="editGroup">Edit group</button> -->
                <button v-if="currentChat!=null&&!currentChat.IsGroup" class="dropdown-item" @click="verifyIdentity" v-translate>Show identity</button>
                <button v-if="currentChat!=null&&!currentChat.IsGroup" class="dropdown-item" @click="resetEncryptionModal" v-translate>Reset encryption</button>
                <router-link v-if="currentChat!=null&&currentChat.IsGroup" :to="'/editGroup/'+currentChat.Tel" class="dropdown-item" v-translate>Edit Group</router-link>

              </div>
              <identity-modal v-if="showIdentityModal" @close="showIdentityModal=false" @confirm="showIdentityModal=false" />
              <confirmation-modal v-if="showConfirmationModal" @close="showConfirmationModal=false" @confirm="confirm" :title="cMTitle" :text="cMText" />
            </div>
          </div>
        </div>
        <div v-else-if="route()=='register' ">
          <div class="header-text"><span v-translate>Connect with Signal</span></div>
        </div>
        <div v-else-if="route()=='password' ">
          <div class="header-text"><span v-translate>Enter password</span></div>
        </div>
        <div v-else-if="route()=='setPassword' " class="list-header-container">
          <router-link class="btn" :to="'/settings'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>Set encryption password</span></div>
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
          <div class="header-text"><span v-translate>Settings</span></div>
        </div>
        <div v-else-if="route()=='newGroup' " class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>New group</span></div>
        </div>
        <div v-else-if="route()=='editGroup' " class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>Edit group</span></div>
        </div>
        <div v-else-if="route()=='devices' ">
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
            <div class="input-container">
              <input v-if="toggleSearch" type="text" class="form-control" v-model="contactsFilter" @change="filterContacts()" @keyup="filterContacts()" />
            </div>
            <button class="btn" type="button" v-if="toggleSearch" @click="showSearch()" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <font-awesome-icon icon="times" />
            </button>
            <button class="btn" type="button" v-if="!toggleSearch" @click="showSearch()" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <font-awesome-icon icon="search" />
            </button>
            <div class="dropdown">
              <button class="btn" type="button" @click="toggleSettings" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
                <button class="dropdown-item" @click="refreshContacts" v-translate>Add Contacts</button>
                <input id="addVcf" type="file" @change="readVcf" style="position: fixed; top: -100em">
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="route()=='chatList'" class="settings-container row">
          <div v-if="errorConnection!=null" class="connection-error"></div>
          <div class="dropdown">
            <button class="btn" type="button" @click="toggleSettings" id="dropdownMenuButton" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              <font-awesome-icon icon="ellipsis-v" />
            </button>
            <div v-if="showSettingsMenu" class="dropdown-menu" id="settings-dropdown" aria-labelledby="dropdownMenuButton">
              <router-link class="dropdown-item" :to="'/newGroup'" @click="showSettingsMenu=false" v-translate>
                New group
              </router-link>
              <router-link class="dropdown-item" :to="'/settings/'" @click="showSettingsMenu=false" v-translate>
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
  import {
    mapState
  } from 'vuex';
  export default {
    name: 'Header',
    components: {
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
        showIdentityModal: false,
        cMTitle: "",
        cMText: "",
        cMType: "",
        names: [],
        toggleSearch: false,
        contactsFilter: "",
      }
    },
    methods: {
      route() {
        return this.$route.name
      },
      back() {
        this.$router.go(-1)
        this.showSettingsMenu = false;
        this.$store.dispatch("leaveChat");
        this.$store.dispatch("clearMessageList");
        this.contactsFilter != "";
        this.toggleSearch = false;
        this.names = []
        this.$store.dispatch("clearFilterContacts");
      },
      toggleSettings() {
        this.showSettingsMenu = !this.showSettingsMenu;
      },
      toggleNotifications() {
        this.showSettingsMenu = false;
        this.$store.dispatch("toggleNotifcations");
      },
      resetEncryptionModal() {
        this.showSettingsMenu = false;
        this.showConfirmationModal = true;
        this.cMType = "resetEncryption";
        this.cMTitle = "Reset secure session?";
        this.cMText = "This may help if you're having encryption problems in this conversation. Your messages will be kept.";
      },
      verifyIdentity() {
        this.$store.dispatch("verifyIdentity");
        this.showSettingsMenu = false;
        this.showIdentityModal = true;
      },
      confirm() {
        if (this.cMType == "resetEncryption") this.$store.dispatch("resetEncryption");
        this.showConfirmationModal = false;
        this.showIdentityModal = false;
      },
      showSearch() {
        if (this.toggleSearch) {
          this.toggleSearch = false;
          this.$store.dispatch("clearFilterContacts");
        } else this.toggleSearch = true;

      },
      onImageError(event) {
        event.target.style.display = "none";
      },
      filterContacts() {
        if (this.contactsFilter != "")
          this.$store.dispatch("filterContacts", this.contactsFilter);
        else this.$store.dispatch("clearFilterContacts");
      },
      getNameForTel(tel) {

        this.contacts.forEach(c => {
          if (c.Tel == tel) {
            if (this.names.length <= 3 && this.names.indexOf(c.Name) == -1)
              this.names.push(c.Name);
            return c;
          } else return tel;
        })
      },
      refreshContacts() {
        this.$store.state.importingContacts = true;
        // console.log("Import contacts for gui " + this.gui)
        this.showSettingsMenu = false;
        if (this.gui == "ut") {
          var result = window.prompt("refreshContacts");
          if (result != "canceld")
            this.$store.dispatch("refreshContacts", result);
        } else {
          this.showSettingsMenu = false;
          document.getElementById("addVcf").click()
        }
      },
      callNumber(n){
        if (this.gui == "ut") {
          window.prompt("call"+n);
          this.showSettingsMenu = false;

        } else {
          this.showSettingsMenu = false;
        }
      },
      createGroup() {

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
    computed: mapState(['messageList',
      'currentChat',
      'currentGroup',
      'contacts',
      'errorConnection',
      'currentContact',
      'gui',
      'identity',
    ]),

    mounted() {
      window.router = this.$router;
      this.names = [];
    },
    watch: {
      $route() {
        this.names = [];
        this.showSettingsMenu = false;
      }
    }
  }
</script>

<style scoped>
  .overlay {
    position: fixed;
    width: 100vh;
    height: 100vh;
  }

  .badge-container {
    justify-content: center;
    align-items: center;
    display: flex;
    padding: 0px;
  }

  .number-text {
    display: flex;
    color: #FFF;
    width: 300%;
  }

  .number-text .name {
    margin-right: 10px;
  }

  .header {
    position: fixed;
    width: 100%;
    background-color: #2090ea;
    top: 0px;
    height: 50px;
    z-index: 2;
    display: flex;
    justify-content: center;
    align-items: center;
    -webkit-box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
    -moz-box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
    box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
  }

  .header .text-right {
    justify-content: flex-end;
    display: flex;
    align-items: center;
    padding: 0px;
  }

  .chat.header {
    height: 70px;
  }

  .chat.header .btn {
    margin-right: 10px;
  }

  .header #dropdownMenuButton {
    margin-right: 0px
  }

  .btn {
    color: #FFF;
  }

  .settings-container {
    align-self: flex-end;
    width: 100%;
    display: flex;
    justify-content: flex-end;
    margin: auto;
  }

  .list-header-container {
    display: flex;
    align-items: center;
  }

  #settings-dropdown {
    display: block !important;
    border-radius: 0px;
    right: 5px;
    left: auto;
  }

  .back {
    font-size: 20px;
  }

  .message-list-container {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .header-text {
    font-weight: bold;
    font-size: 20px;
    padding-left: 10px;
    color: #FFFFFF;
  }

  .header-text-chat {
    font-weight: bold;
    font-size: 18px;
    color: #FFFFFF;
  }

  .group-badge {
    background-color: #FFF;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    justify-content: center;
    align-items: center;
    color: #2090ea;
  }

  .mute-badge {
    color: #FFF;
    margin-left: 10px;
  }

  .chat-header {
    display: flex;
    align-items: center;
    padding: 0px;
  }

  .center {
    padding: 5px;
    overflow: hidden;
  }

  .input-container {
    display: flex;
    max-width: 100%;
    overflow: hidden;
  }
  </style>
<style>
  .connection-error {
    position: fixed;
    width: 100vw;
    height: 3px;
    left: 0;
    bottom: 0px;
    background-color:blue;
    background: linear-gradient(-45deg, #ee7752, #e73c7e, #23a6d5, #23d5ab);
    background-size: 400% 400%;
    -webkit-animation: gradientBG 10s ease infinite;
    animation: gradientBG 10s ease infinite;
    z-index:1000;
  }

  @-webkit-keyframes gradientBG {
    0% {
      background-position: 0% 50%;
    }

    50% {
      background-position: 100% 50%;
    }

    100% {
      background-position: 0% 50%;
    }
  }

  @keyframes gradientBG {
    0% {
      background-position: 0% 50%;
    }

    50% {
      background-position: 100% 50%;
    }

    100% {
      background-position: 0% 50%;
    }
  }
</style>
