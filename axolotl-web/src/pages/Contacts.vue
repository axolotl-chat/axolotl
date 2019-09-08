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
    <add-contact-modal />
    <button class="btn add-contact" @click="addContact"><font-awesome-icon icon="plus" /></button>

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
  mounted(){
    this.$store.dispatch("getContacts")
  },
  methods: {
    addContact(){
      this.$store.dispatch("addContact")
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
  width: 70px;
  height: 70px;
  font-size: 30px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
