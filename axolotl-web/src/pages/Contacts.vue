<template>
  <div class="contact-list">
    <div v-if="showActions" class="actions-header">
      <button class="btn hide-actions">
        <font-awesome-icon icon="times"  @click="showActions=false"/>
      </button>
    </div>
    <div v-for="(contact, i) in contacts"
        class="btn col-12 chat">
      <div class="row chat-entry">
        <div class="avatar col-3" @click="contactClick(contact)">
          <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
        </div>
        <div class="meta col-7" @click="contactClick(contact)"  v-longclick="showContactAction">
          <p class="name">{{contact.Name}}</p>
          <p class="number">{{contact.Tel}}</p>
        </div>
        <div class="col-2 actions" v-if="showActions" >

          <button class="btn" @click="delContact(i)">
            <font-awesome-icon icon="trash"  />
          </button>
          <button class="btn" @click="editContactModalOpen(contact,i)">
            <font-awesome-icon icon="pencil-alt"  />
          </button>
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
      contactId:null
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
    delContact(id){
      this.$store.dispatch("delContact", id)
      this.showActions = false;

    },
    saveContact(data){
      this.editContactModal=false
      this.showActions = false;
      this.$store.dispatch("editContact", data)
    },
    showContactAction(){
      this.showActions = true;
    },
    contactClick(contact){
      if(!this.showActions){
        router.push('/chat/'+contact.Tel)
      }
    },
    editContactModalOpen(contact,id){
      this.editContactModal=true;
      this.contact = contact;
      this.contactId = id;
    }
  },
  computed: {
    contacts () {
      return this.$store.state.contacts
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
    background-color: #cacaca;
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
</style>
