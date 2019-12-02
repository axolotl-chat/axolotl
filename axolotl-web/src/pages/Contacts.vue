<template>
  <div class="contact-list">
    <div v-if="error!=null" class="alert alert-danger">Can't change contact list: {{error}}</div>
    <div v-if="importing" class="alert alert-warning">Importing contacts, head back later</div>
    <div v-if="showActions" class="actions-header">
      <button class="btn" @click="delContact(i)">
        <font-awesome-icon icon="trash"  />
      </button>
      <button class="btn" @click="editContactModalOpen(contact,i)">
        <font-awesome-icon icon="pencil-alt"  />
      </button>
      <button class="btn hide-actions">
        <font-awesome-icon icon="times"  @click="showActions=false"/>
      </button>
    </div>
    <div v-if="contactsFilterActive">
      <div v-for="(contact, i) in contactsFilterd"
          :class="contact.Tel==editContactId?'selected btn col-12 chat':'btn col-12 chat'">
        <div class="row chat-entry">
          <div class="avatar col-3" @click="contactClick(contact)">
            <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
          </div>
          <div class="meta col-9" @click="contactClick(contact)"  v-longclick="()=>{showContactAction(contact)}">
            <p class="name">{{contact.Name}}</p>
            <p class="number">{{contact.Tel}}</p>
          </div>
        </div>
      </div>
    </div>
    <div v-else v-for="(contact, i) in contacts"
        :class="contact.Tel==editContactId?'selected btn col-12 chat':'btn col-12 chat'">
      <div class="row chat-entry">
        <div class="avatar col-3" @click="contactClick(contact)">
          <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
        </div>
        <div class="meta col-9" @click="contactClick(contact)"  v-longclick="()=>{showContactAction(contact)}">
          <p class="name">{{contact.Name}}</p>
          <p class="number">{{contact.Tel}}</p>
        </div>
      </div>
    </div>
    <div v-if="addContactModal" class="addContactModal">
      <add-contact-modal
      @close="addContactModal=false"
      @add="addContact($event)"
      />
    </div>
    <div v-if="editContactModal" class="editContactModal">
      <edit-contact-modal
      :contact="contact"
      :id="contactId"
      @close="editContactModal=false"
      @save="saveContact($event)"
      />
    </div>
    <button class="btn add-contact" @click="addContactModal=true"><font-awesome-icon icon="plus" /></button>
  </div>
</template>

<script>
import AddContactModal from "@/components/AddContactModal.vue"
import EditContactModal from "@/components/EditContactModal.vue"
export default {
  name: 'Contacts',
  props: {
    msg: String
  },
  components: {
    AddContactModal,
    EditContactModal
  },
  data() {
    return {
      addContactModal: false,
      showActions: false,
      editContactModal: false,
      contact:null,
      contactId:null,
      editContactId:""
    }
  },
  mounted(){
    this.$store.dispatch("getContacts")
  },
  methods: {
    addContact(data){
      this.$store.dispatch("addContact", data)
      this.addContactModal=false
    },
    delContact(){
      this.$store.dispatch("delContact", this.editContactId)
      this.showActions = false;
      this.editContactId ="";
    },
    saveContact(data){
      this.editContactModal=false
      this.showActions = false;
      this.editContactId ="";
      this.$store.dispatch("editContact", data)
    },
    showContactAction(contact){
      this.editContactId=contact.Tel;
      this.contact = contact;
      this.showActions = true;
    },
    contactClick(contact){
      if(!this.showActions){
        this.$store.dispatch("createChat", contact.Tel)
      }
      else{
        this.editContactId=contact.Tel;
      }
    },
    editContactModalOpen(){
      this.editContactModal=true;
      this.contact = this.contact;
      this.contactId = this.editContactId;
      this.showActions = false;
      this.editContactId ="";
    }
  },
  computed: {
    contacts () {
      return this.$store.state.contacts
    },
    contactsFilterd () {
      return this.$store.state.contactsFilterd
    },
    contactsFilterActive () {
      return this.$store.state.contactsFilterActive
    },
    error () {
      return this.$store.state.ratelimitError;
    },
    importing () {
      return this.$store.state.importingContacts;
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn.add-contact {
  position: fixed;
  bottom: 16px;
  right: 10px;
  background-color: #2090ea;
  color: #FFF;
  border-radius: 50%;
  width: 45px;
  height: 45px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.chat{
  padding: 0px;
}
.number{
  font-size:14px;
}
.actions-header {
    position: fixed;
    background-color: #173d5c;
    width: 100%;
    left: 0;
    display: flex;
    justify-content: end;
    z-index: 2;
    top: 0;
    height: 51px;
}
.hide-actions{
  padding-right:40px;
}
.col-2.actions {
    position: absolute;
    display: flex;
    right: 0px;
    justify-content:center;
    align-items:center;
}
.col-2.actions .btn {
    font-size: 15px;
    padding: 5px;
}
.selected{
  background-color:#c5e4f0;
}
</style>
