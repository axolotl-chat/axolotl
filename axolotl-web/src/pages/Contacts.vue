<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="contact-list">
      <div v-if="error !== null" v-translate class="alert alert-danger">
        Can't change contact list: {{ error }}
      </div>
      <div v-if="importing" v-translate class="alert alert-warning">
        Importing contacts, head back later
      </div>
      <div v-if="showActions" class="actions-header">
        <button class="btn" @click="delContact()">
          <font-awesome-icon icon="trash" />
        </button>
        <button class="btn" @click="openEditContactModal()">
          <font-awesome-icon icon="pencil-alt" />
        </button>
        <button class="btn hide-actions">
          <font-awesome-icon icon="times" @click="closeActionMode()" />
        </button>
      </div>
      <div v-if="contacts.length === 0" v-translate class="empty">
        Contact list is empty...
      </div>
      <div v-if="contactsFilterActive">
        <div
          v-for="c in contactsFiltered"
          :key="c.Tel"
          :class="
            c === selectedContact
              ? 'selected btn col-12 chat'
              : 'btn col-12 chat'
          "
        >
          <div class="row chat-entry">
            <div
              :class="'avatar col-3 ' + checkForUUIDClass(c)"
              @click="contactClick(c)"
            >
              <div class="badge-name">
                <img
                  class="avatar-img"
                  :src="'http://localhost:9080/avatars?e164=' + c.Tel"
                  alt="Avatar image"
                  @error="onImageError($event)"
                />
                {{ c.name[0] ? c.name[0] + c.name[1] : c.name[0] }}
              </div>
            </div>
            <div class="meta col-8" @click="contactClick(c)">
              <p class="name">{{ c.name }}</p>
              <p class="number">{{ c.Tel }}</p>
            </div>
            <div class="col-1" @click="showContactAction(c)">
              <font-awesome-icon icon="wrench" />
            </div>
          </div>
        </div>
      </div>
      <div
        v-for="c in contacts"
        v-else
        :key="c.uuid"
        :class="
          c === selectedContact ? 'selected btn col-12 chat' : 'btn col-12 chat'
        "
      >
        <div class="row chat-entry">
          <div
            :class="'avatar col-3 avatar ' + checkForUUIDClass(c)"
            @click="contactClick(c)"
          >
            <div class="badge-name">
              <img
                class="avatar-img"
                :src="'http://localhost:9080/avatars?e164=' + c.Tel"
                alt="Avatar image"
                @error="onImageError($event)"
              />
              {{ c.name[0] ? c.name[0] + c.name[1] : c.name[0] }}
            </div>
          </div>
          <div class="meta col-8" @click="contactClick(c)">
            <p class="name">{{ c.name }}</p>
            <p class="number" v-if="c.phonenumber">{{ `+${ c.phonenumber}` }}</p>
          </div>
          <!-- <div class="col-1" @click="showContactAction(c)">
            <font-awesome-icon icon="wrench" />
          </div> -->
        </div>
      </div>

      <div v-if="addContactModal" class="addContactModal">
        <add-contact-modal
          @close="closeAddContactModal()"
          @add="addContact($event)"
        />
      </div>
      <div v-if="editContactModal" class="editContactModal">
        <edit-contact-modal
          :contact="selectedContact"
          @close="closeEditContactModal()"
          @save="saveContact($event)"
        />
      </div>
      <!-- <button class="btn add-contact" @click="openAddContactModal()">
        <font-awesome-icon icon="plus" />
      </button> -->
    </div>
  </component>
</template>

<script>
import AddContactModal from "@/components/AddContactModal.vue";
import EditContactModal from "@/components/EditContactModal.vue";
import { validateUUID } from "@/helpers/uuidCheck";

export default {
  name: "ContactsPage",
  components: {
    AddContactModal,
    EditContactModal,
  },
  data() {
    return {
      showActions: false,
      addContactModal: false,
      editContactModal: false,
      startChatModal: false,
      selectedContact: null,
    };
  },
  computed: {
    contacts() {
      return this.$store.state.contacts;
    },
    contactsFiltered() {
      return this.$store.state.contactsFiltered;
    },
    contactsFilterActive() {
      return this.$store.state.contactsFilterActive;
    },
    error() {
      return this.$store.state.rateLimitError;
    },
    importing() {
      return this.$store.state.importingContacts;
    },
  },
  mounted() {
    this.$store.dispatch("getContacts");
  },
  methods: {
    validateUUID,
    openAddContactModal() {
      this.addContactModal = true;
    },
    closeAddContactModal() {
      this.addContactModal = false;
    },
    addContact(data) {
      this.$store.dispatch("addContact", data);
      this.addContactModal = false;
    },
    showContactAction(contact) {
      this.selectedContact = contact;
      this.showActions = true;
    },
    closeActionMode() {
      this.showActions = false;
      this.selectedContact = null;
    },
    delContact() {
      this.$store.dispatch("delContact", this.selectedContact.Tel);
      this.closeActionMode();
    },
    onImageError(event) {
      event.target.style.display = "none";
    },
    openEditContactModal() {
      this.editContactModal = true;
      this.showActions = false;
    },
    closeEditContactModal() {
      this.editContactModal = false;
      this.closeActionMode();
    },
    saveContact(data) {
      this.$store.dispatch("editContact", data);
      this.closeEditContactModal();
    },
    contactClick(contact) {
      if (!this.showActions) {
        if (this.validateUUID(contact.uuid)){
          this.$router.push(`/chat/${JSON.stringify({Contact:contact.uuid})}`);

        }
        
      } else {
        this.closeActionMode();
      }
    },
    checkForUUIDClass(contact) {
      var isValid = this.validateUUID(contact.uuid);
      return isValid ? "" : "not-registered";
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn.add-contact {
  position: fixed;
  bottom: 16px;
  right: 10px;
  background-color: #2090ea;
  color: #fff;
  border-radius: 50%;
  width: 45px;
  height: 45px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.chat {
  padding: 0px;
}
.number {
  font-size: 14px;
}
.actions-header {
  position: fixed;
  background-color: #173d5c;
  width: 100%;
  left: 0;
  display: flex;
  justify-content: flex-end;
  z-index: 2;
  top: 0;
  height: 51px;
}
.hide-actions {
  padding-right: 40px;
}
.col-2.actions {
  position: absolute;
  display: flex;
  right: 0px;
  justify-content: center;
  align-items: center;
}
.col-2.actions .btn {
  font-size: 15px;
  padding: 5px;
}
.selected {
  background-color: #c5e4f0;
}
.empty {
  width: 100%;
  height: 70vh;
  display: flex;
  justify-content: center;
  align-items: center;
}
.not-registered .badge-name {
  background: linear-gradient(
    0deg,
    rgb(191, 191, 191) 8%,
    rgb(100, 100, 100) 42%,
    rgb(134, 134, 134) 100%
  );
}
  /*
.avatar{
  border-radius: 50px;
background-color: #2090ea;
width: 50px;
height: 50px;
justify-content: center;
align-items: center;
display: flex;
color: #FFF;
margin:20px;
} */
.meta{
  text-align: left;
  display: flex;
  justify-content: center;
  flex-direction: column;
}
.meta p{
  margin: 0;
}
.meta .name{
  font-size: 16px;
  font-weight: 600;
}
</style>
