<template>
  <div :class="route() + ' header '">
    <div class="container-fluid">
      <div v-if="showSettingsMenu" class="overlay" @click="showSettingsMenu = false" />
      <div class="header-row">
        <!-- chat page start -->
        <div v-if="route() === 'chat'" class="message-list-container row chat-page">
          <div v-if="errorConnection !== null" class="connection-error" />
          <div class="col-10 chat-header">
            <button class="btn" @click="$router.push('/')">
              <font-awesome-icon icon="arrow-left" />
            </button>
            <div v-if="currentChat !== null && currentChat && currentChat.thread" class="row w-100">
              <div class="col-2 badge-container">
                <div
                  v-if="!isGroup"
                  class="badge-name"
                  @click="openProfileForRecipient(currentChat?.thread?.Contact)"
                >
                  <img
                    class="avatar-img"
                    :src="
                      'http://localhost:9080/attachments/avatars/' + currentChat?.thread?.Contact
                    "
                    @error="onImageError($event)"
                  />
                  {{ currentChat.title ? currentChat.title[0] : '?' }}
                  {{ currentChat.title ? currentChat.title[1] : '' }}
                </div>
                <div v-else class="group-badge">
                  <font-awesome-icon icon="user-friends" />
                </div>
              </div>
              <div class="col-10 center">
                <div class="row">
                  <div class="col-12">
                    <div
                      v-if="isGroup && currentChat.title === currentChat.Tel"
                      class="header-text-chat"
                    >
                      <div v-if="currentChat.muted" class="mute-badge">
                        <font-awesome-icon class="mute" icon="volume-mute" />
                      </div>
                      <div v-translate>Unknown group</div>
                    </div>
                    <div v-else class="header-text-chat">
                      <div v-if="currentChat.muted" class="mute-badge">
                        <font-awesome-icon class="mute" icon="volume-mute" />
                      </div>
                      <div
                        v-if="currentChat.title !== currentChat.Tel"
                        @click="openProfileForRecipient(currentChat.thread.Contact)"
                      >
                        {{ currentChat.title }}
                      </div>
                    </div>
                  </div>
                  <div class="col-12">
                    <div
                      v-if="isGroup && currentGroup !== null && typeof currentGroup !== 'undefined'"
                      class="number-text"
                    >
                      <div v-for="e in currentGroup.Members" :key="e">
                        {{ getNameForTel(e) }}
                      </div>
                    </div>
                    <div
                      v-if="isGroup && currentGroup !== null && typeof currentGroup !== 'undefined'"
                      class="number-text"
                    >
                      <div v-for="(n, i) in names" :key="i" class="name">
                        {{ n }}
                        <span v-if="i < names.length - 1">,</span>
                      </div>
                    </div>
                    <div
                      v-if="!isGroup && currentChat.title === currentChat.Tel"
                      class="number-text"
                    >
                      {{ currentChat.Tel }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="col-2 text-right">
            <div class="dropdown">
              <button
                id="dropdownMenuButton"
                class="btn"
                type="button"
                data-toggle="dropdown"
                aria-haspopup="true"
                aria-expanded="false"
                @click="toggleSettings"
              >
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <div
                v-if="showSettingsMenu"
                id="settings-dropdown"
                class="dropdown-menu"
                aria-labelledby="dropdownMenuButton"
              >
                <button
                  v-if="currentChat !== null && !isGroup && currentChat.title !== currentChat.Tel"
                  class="dropdown-item"
                  @click="callNumber(currentChat.Tel)"
                >
                  {{ currentChat.Tel }}
                </button>
                <button
                  v-if="currentChat !== null && !currentChat.muted"
                  v-translate
                  class="dropdown-item"
                  @click="toggleNotifications"
                >
                  Mute
                </button>
                <button v-else v-translate class="dropdown-item" @click="toggleNotifications">
                  Unmute
                </button>
                <button
                  v-if="currentChat !== null && !isGroup && currentChat.title === currentChat.Tel"
                  v-translate
                  class="dropdown-item"
                  @click="addContactModal = true"
                >
                  Add contact
                </button>
                <button
                  v-if="currentChat !== null && !isGroup && currentChat.title !== currentChat.Tel"
                  v-translate
                  class="dropdown-item"
                  @click="openEditContactModal()"
                >
                  Edit contact
                </button>
                <!-- <button
                  v-if="currentChat !== null && !isGroup"
                  v-translate
                  class="dropdown-item"
                  @click="verifyIdentity"
                >
                  Show identity
                </button>
                <button
                  v-if="currentChat !== null && !isGroup"
                  v-translate
                  class="dropdown-item"
                  @click="resetEncryptionModal"
                >
                  Reset encryption
                </button> -->
                <button
                  v-if="currentChat !== null && !isGroup"
                  v-translate
                  class="dropdown-item"
                  @click="delChatModal"
                >
                  Delete chat
                </button>
                <!-- <router-link
                  v-if="currentChat !== null && currentChat.Type === 1"
                  v-translate
                  :to="'/editGroup/' + currentChat.Tel"
                  class="dropdown-item"
                >
                  Edit group
                </router-link> -->
              </div>
              <identity-modal
                v-if="showIdentityModal"
                @close="showIdentityModal = false"
                @confirm="showIdentityModal = false"
              />
              <confirmation-modal
                v-if="showConfirmationModal"
                :title="cMTitle"
                :text="cMText"
                @close="showConfirmationModal = false"
                @confirm="confirm"
              />
            </div>
          </div>
        </div>
        <!-- chat page end -->
        <!-- register page start -->
        <div v-else-if="route() === 'register'">
          <div class="header-text">
            <span v-translate>Connect with Signal</span>
          </div>
        </div>
        <!-- register page end -->
        <div v-else-if="route() === 'password'">
          <div class="header-text"><span v-translate>Enter password</span></div>
        </div>
        <!-- set password page -->
        <div v-else-if="route() === 'setPassword'" class="list-header-container">
          <router-link class="btn" :to="'/settings'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text">
            <span v-translate>Set encryption password</span>
          </div>
        </div>
        <!-- about page -->
        <div v-else-if="route() === 'about'" class="list-header-container">
          <router-link class="btn" :to="'/settings'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
        </div>
        <!-- settings page -->
        <div v-else-if="route() === 'settings'" class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>Settings</span></div>
        </div>
        <!-- new group page -->
        <!-- <div v-else-if="route() === 'newGroup'" class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>New group</span></div>
        </div> -->
        <!-- edit group page -->
        <!-- <div v-else-if="route() === 'editGroup'" class="list-header-container">
          <router-link class="btn" :to="'/chatList'">
            <font-awesome-icon icon="arrow-left" />
          </router-link>
          <div class="header-text"><span v-translate>Edit group</span></div>
        </div> -->
        <!-- linking devices page -->
        <div v-else-if="route() === 'devices'">
          <button class="back btn" @click="back()">
            <font-awesome-icon icon="arrow-left" />
          </button>
        </div>
        <!-- contacts page -->
        <div v-else-if="route() === 'contacts'" class="row">
          <div class="col-2">
            <button class="back btn" @click="back()">
              <font-awesome-icon icon="arrow-left" />
            </button>
          </div>
          <div class="col-10 text-right">
            <div class="input-container">
              <input
                v-if="toggleSearch"
                v-model="contactsFilter"
                type="text"
                class="form-control"
                @change="filterContacts()"
                @keyup="filterContacts()"
              />
            </div>
            <button
              v-if="toggleSearch"
              id="dropdownMenuButton"
              class="btn"
              type="button"
              data-toggle="dropdown"
              aria-haspopup="true"
              aria-expanded="false"
              @click="showSearch()"
            >
              <font-awesome-icon icon="times" />
            </button>
            <button
              v-if="!toggleSearch"
              id="dropdownMenuButton"
              class="btn"
              type="button"
              data-toggle="dropdown"
              aria-haspopup="true"
              aria-expanded="false"
              @click="showSearch()"
            >
              <font-awesome-icon icon="search" />
            </button>
            <!-- <div class="dropdown">
              <button
                id="dropdownMenuButton"
                class="btn"
                type="button"
                data-toggle="dropdown"
                aria-haspopup="true"
                aria-expanded="false"
                @click="toggleSettings"
              >
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <div
                v-if="showSettingsMenu"
                id="settings-dropdown"
                class="dropdown-menu"
                aria-labelledby="dropdownMenuButton"
              >
                <button
                  v-translate
                  class="dropdown-item"
                  @click="
                    showSettingsMenu = false;
                    showImportVcfModal = true;
                  "
                >
                  Import contacts
                </button>
              </div>
            </div> -->
          </div>
        </div>
        <div v-else>
          <!-- <button class="back btn" @click="back()">
            <font-awesome-icon icon="arrow-left" /></button> -->
        </div>
      </div>
    </div>
    <div v-if="addContactModal" class="addContactModal">
      <add-contact-modal
        :number="currentChat.Tel"
        :uuid="currentChat.UUID"
        @close="addContactModal = false"
        @add="addContact($event)"
      />
    </div>
    <div v-if="showImportVcfModal" class="addContactModal">
      <import-vcf-modal @close="showImportVcfModal = false" />
    </div>
    <div v-if="editContactModal && editContactId !== -1" class="editContactModal">
      <edit-contact-modal
        :id="editContactId.toString()"
        :contact="contacts[editContactId]"
        @close="editContactModal = false"
        @save="saveContact($event)"
      />
    </div>
  </div>
</template>

<script>
import IdentityModal from '@/components/IdentityModal.vue';
import ConfirmationModal from '@/components/ConfirmationModal.vue';
import ImportVcfModal from '@/components/ImportVcfModal.vue';
import AddContactModal from '@/components/AddContactModal.vue';
import EditContactModal from '@/components/EditContactModal.vue';

import { mapState } from 'vuex';
export default {
  name: 'HeaderComponent',
  components: {
    ConfirmationModal,
    IdentityModal,
    ImportVcfModal,
    AddContactModal,
    EditContactModal,
  },
  data() {
    return {
      showSettingsMenu: false,
      showConfirmationModal: false,
      showIdentityModal: false,
      showImportVcfModal: false,
      cMTitle: '',
      cMText: '',
      cMType: '',
      names: [],
      toggleSearch: false,
      contactsFilter: '',
      addContactModal: false,
      editContactModal: false,
      editContactId: -1,
    };
  },
  computed: {
    ...mapState([
      'messageList',
      'currentChat',
      'currentGroup',
      'contacts',
      'errorConnection',
      'currentContact',
      'sessionNames',
      'gui',
      'identity',
    ]),
    isGroup() {
      return this.currentChat?.thread.Group !== undefined;
    },
  },
  watch: {
    $route() {
      this.names = [];
      this.showSettingsMenu = false;
    },
    currentChat: {
      handler() {
        this.names = [];
        this.showSettingsMenu = false;
      },
      deep: true,
    },
  },
  mounted() {
    this.names = [];
  },
  methods: {
    route() {
      return this.$route.name;
    },
    back() {
      this.$router.go(-1);
      this.showSettingsMenu = false;
      this.contactsFilter = '';
      this.toggleSearch = false;
      this.names = [];
    },
    toggleSettings() {
      this.showSettingsMenu = !this.showSettingsMenu;
    },
    toggleNotifications() {
      this.showSettingsMenu = false;
      this.$store.dispatch('toggleNotifications');
    },
    resetEncryptionModal() {
      this.showSettingsMenu = false;
      this.showConfirmationModal = true;
      this.cMType = 'resetEncryption';
      this.cMTitle = this.$gettext('Reset secure session?');
      this.cMText = this.$gettext(
        "This may help if you're having encryption problems in this conversation. Your messages will be kept.",
      );
    },
    verifyIdentity() {
      this.$store.dispatch('verifyIdentity');
      this.showSettingsMenu = false;
      this.showIdentityModal = true;
    },
    delChatModal() {
      this.showSettingsMenu = false;
      this.showConfirmationModal = true;
      this.cMType = 'delChat';
      this.cMTitle = this.$gettext('Delete this chat?');
      this.cMText = this.$gettext(
        'This chat will be permanently deleted - but only from your device.',
      );
    },
    confirm() {
      if (this.cMType === 'resetEncryption') this.$store.dispatch('resetEncryption');
      else if (this.cMType === 'delChat') this.$store.dispatch('delChat', this.currentChat.thread);
      this.$router.push('/chatList');
      this.showConfirmationModal = false;
      this.showIdentityModal = false;
    },
    showSearch() {
      if (this.toggleSearch) {
        this.toggleSearch = false;
        this.$store.dispatch('clearFilterContacts');
      } else this.toggleSearch = true;
    },
    onImageError(event) {
      event.target.style.display = 'none';
    },
    filterContacts() {
      if (this.contactsFilter !== '') this.$store.dispatch('filterContacts', this.contactsFilter);
      else this.$store.dispatch('clearFilterContacts');
    },
    getNameForTel(tel) {
      this.contacts.forEach((c) => {
        if (c.Tel === tel) {
          if (this.names.length <= 3 && this.names.indexOf(c.Name) === -1) this.names.push(c.Name);
          return c;
        } else return tel;
      });
    },
    callNumber(n) {
      if (this.gui === 'ut') {
        window.prompt('call' + n);
        this.showSettingsMenu = false;
      } else {
        this.showSettingsMenu = false;
      }
    },
    createGroup() {},
    openEditContactModal() {
      const id = this.contacts.findIndex(
        (c) =>
          c.Tel === this.currentChat.Tel ||
          c.UUID === this.sessionNames[this.currentChat.thread].Name,
      );
      this.editContactId = id;
      if (id !== -1) {
        this.editContactModal = true;
      }
    },
    addContact(data) {
      this.$store.dispatch('addContact', data);
      this.addContactModal = false;
    },
    saveContact(data) {
      this.editContactModal = false;
      this.showActions = false;
      this.editContactId = '';
      this.$store.dispatch('editContact', data);
    },
    openProfileForRecipient(recipient) {
      if (recipient !== -1) {
        this.$router.push(`/profile/${recipient}`);
      }
    },
  },
};
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
  color: #fff;
  width: 300%;
}

.number-text .name {
  margin-right: 10px;
}

.header {
  padding: 5px 0;
  background-color: #2090ea;
  z-index: 2;
  box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
  min-height: 49px;
}

.header .text-right {
  justify-content: flex-end;
  display: flex;
  align-items: center;
  padding: 0px;
}

.chat.header .btn {
  margin-right: 10px;
}

.header #dropdownMenuButton {
  margin-right: 0px;
}

.btn {
  color: #fff;
}

#dropdownMenuButton {
  color: #fff;
}

.settings-container {
  align-self: flex-end;
  display: flex;
  justify-content: flex-end;
}

.list-header-container {
  display: flex;
  align-items: center;
}

#settings-dropdown {
  display: block !important;
  border-radius: 3px;
  right: 5px;
  left: auto;
  border: 1px solid #2090ea;
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
  color: #ffffff;
}

.header-text-chat {
  font-weight: bold;
  font-size: 15px;
  color: #ffffff;
  display: flex;
}

.group-badge {
  background-color: #fff;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  justify-content: center;
  align-items: center;
  color: #2090ea;
}

.mute-badge {
  color: #fff;
  margin-right: 10px;
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
  background-color: blue;
  background: linear-gradient(-45deg, #ee7752, #e73c7e, #23a6d5, #23d5ab);
  background-size: 400% 400%;
  -webkit-animation: gradientBG 10s ease infinite;
  animation: gradientBG 10s ease infinite;
  z-index: 1000;
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
