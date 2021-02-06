<template>
  <div class="contact-list">

    <div v-if="error!=null" class="alert alert-danger" v-translate>Can't change contact list: {{error}}</div>
    <div v-if="importing" class="alert alert-warning" v-translate>Importing contacts, head back later</div>
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
    <div v-if="contacts.length ==0" class="empty" v-translate>
      Contact list is empty...
    </div>
    <div v-if="contactsFilterActive">
      <div v-for="(contact) in contactsFilterd"
        v-bind:key="contact.Tel"
          :class="contact.Tel==editContactId?'selected btn col-12 chat':'btn col-12 chat'">
        <div class="row chat-entry">
          <div :class="'avatar col-3 '+contact.UUID&&contact.UUID[0]==0 && contact.UUID[contact.UUID.length-1]==0?'not-registered':''" @click="startChatModalOpen(contact,i)">
            <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
          </div>
          <div class="meta col-9" @click="startChatModalOpen(contact)"  v-longclick="()=>{showContactAction(contact)}">
            <p class="name">{{contact.Name}}</p>
            <p class="number">{{contact.Tel}}</p>
          </div>
        </div>
      </div>
    </div>
    <div v-else v-for="(contact) in contacts"
        v-bind:key="contact.Tel"
        :class="contact.Tel==editContactId?'selected btn col-12 chat':'btn col-12 chat'">
      <div class="row chat-entry">
        <!-- <div class="avatar col-3" @click="contactClick(contact)"> -->
        <div :class="'avatar col-3 '+contact.UUID&&contact.UUID[0]==0 && contact.UUID[contact.UUID.length-1]==0?' avatar col-3 not-registered':'avatar col-3'" @click="contactClick(contact,i)">
          <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
        </div>
        <!-- <div class="meta col-9" @click="contactClick(contact)"  v-longclick="()=>{showContactAction(contact)}"> -->
        <div class="meta col-9" @click="startChatModalOpen(contact)"  v-longclick="()=>{showContactAction(contact)}">
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
        <div v-if="startChatModal" class="startChatModal">
      <start-chat-modal
        @close="startChatModal=false"
      />
    </div>
    <button class="btn add-contact" @click="addContactModal=true"><font-awesome-icon icon="plus" /></button>
  </div>
</template>

<script>
import AddContactModal from "@/components/AddContactModal.vue"
import EditContactModal from "@/components/EditContactModal.vue"
import StartChatModal from "@/components/StartChatModal.vue"
export default {
  name: 'Contacts',
  props: {
    msg: String
  },
  components: {
    AddContactModal,
    EditContactModal,
    StartChatModal
  },
  data() {
    return {
      addContactModal: false,
      showActions: false,
      editContactModal: false,
      contact:null,
      contactId:null,
      startChatModal: false,
      editContactId:"",
      i:null
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
        if(contact.UUID!="" && (contact.UUID[0]!="0"||contact.UUID[contact.UUID.length-1]!="0"))
        this.$store.dispatch("createChat", contact.UUID)
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
    },
    startChatModalOpen(){
      this.startChatModal=true;
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
    justify-content: flex-end;
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
.empty{
  width:100%;
  height:70vh;
  display:flex;
  justify-content:center;
  align-items:center;
}
.not-registered .badge-name{
  background: linear-gradient(0deg,rgb(191, 191, 191) 8%, rgb(100, 100, 100) 42%, rgb(134, 134, 134) 100%);
}
</style>
