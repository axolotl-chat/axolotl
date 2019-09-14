<template>
  <div class="contact-list">
    <router-link :to="'/chat/'+contact.Tel" v-for="contact in contacts" class="col-12 chat">
      <div class="row chat-entry">
        <div class="avatar col-3">
          <div class="badge-name">{{contact.Name[0]}}</div>
        </div>
        <div class="meta col-9">
          <p class="name">{{contact.Name}}</p>
        </div>
      </div>
    </router-link>
    <div v-if="addContactModal" class="addContactModal">
      <add-contact-modal
      @close="addContactModal=false"
      @add="addContact($event)"
      />
    </div>
    <button class="btn add-contact" @click="addContactModal=true"><font-awesome-icon icon="plus" /></button>
  </div>
</template>

<script>
import AddContactModal from "@/components/AddContactModal.vue"
export default {
  name: 'Contacts',
  props: {
    msg: String
  },
  components: {
    AddContactModal
  },
  data() {
    return {
      addContactModal: false
    }
  },
  mounted(){
    this.$store.dispatch("getContacts")
  },
  methods: {
    addContact(data){
      this.$store.dispatch("addContact", data)
      this.addContactModal=false

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
  right: 20%;
  background-color: #2090ea;
  color: #FFF;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
