<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Add members</h5>
          <div class="actions" v-if="!searchActive">
            <button type="button" class="btn search" @click="searchActive=true;">
              <font-awesome-icon icon="search"/>
            </button>
            <button type="button" class="btn" @click="$emit('close')">
              <font-awesome-icon icon="times"/>
            </button>
          </div>
          <div class="actions" v-if="searchActive">
            <div class="input-container">
              <input type="text" class="form-control"
                    v-model="contactsFilter"
                    @change="filterContacts()"
                    @keyup="filterContacts()"></input>
            </div>
            <button type="button" class="btn" @click="searchActive=false;">
              <font-awesome-icon icon="times"/>
            </button>
          </div>
        </div>
        <div class="modal-body">
          <div class="contact-list">
            <div v-for="(contact, i) in contacts"
                v-if="contacts.length>0&&contactsFilter==''"
                class="btn col-12 chat">
              <div class="row chat-entry">
                <div class="avatar col-3" @click="contactClick(contact)">
                  <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
                </div>
                <div class="meta col-7" @click="$emit('add', contact)" >
                  <p class="name">{{contact.Name}}</p>
                  <p class="number">{{contact.Tel}}</p>
                </div>
              </div>
            </div>
            <div v-for="(contact, i) in contactsFilterd"
                v-if="contactsFilter!=''"
                class="btn col-12 chat">
              <div class="row chat-entry">
                <div class="avatar col-3" @click="contactClick(contact)">
                  <div class="badge-name">{{contact.Name[0]+contact.Name[1]}}</div>
                </div>
                <div class="meta col-7" @click="$emit('add', contact)" >
                  <p class="name">{{contact.Name}}</p>
                  <p class="number">{{contact.Tel}}</p>
                </div>
              </div>
            </div>
            <div v-else>Enter Contacts in list first<div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
  </div>
</template>

<script>
export default {
  name: 'AddGroupMembersModal',
  props: {
    allreadyAdded:Array
  },
  components: {
  },
  data() {
    return {
      contacts:[],
      searchActive:false,
      contactsFilter:"",
    }
  },
  mounted(){
    this.contacts = this.$store.state.contacts
  },
  methods: {
    contactClick(contact){
      this.$store.dispatch("addNewGroupMember", contact)
    },
    filterContacts(){
      if(this.contactsFilter!="")
      this.$store.dispatch("filterContactsForGroup", this.contactsFilter);
      else  this.$store.dispatch("clearFilterContacts");
    },
  },
  computed: {
    contactsFilterd () {
      return this.$store.state.contactsFilterd
    },
  },
  watch:{
    allreadyAdded(newVal, oldVal){
      var that = this
      this.contacts=that.$store.state.contacts.filter(function (el) {
        var found = that.allreadyAdded.find(function(element) {
            return element.Tel == el.Tel;
        });
        if(typeof found =="undefined")return true;
        else return false;
      });
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
.modal {
    display: block;
    border:none;
}
.modal-content {
  border-radius:0px;
}
.modal-body {
  max-height:80vh;
  overflow:auto;
}
.modal-header {
  border-bottom: none;
  background-color: #2090ea;
  border-radius: 0px;
  color: #FFF;
}
.modal-title{
  display:flex;
}
.modal-title > div{
  margin-left:10px;
}
.modal-footer{
  border-top:0px;
}
.actions .btn{
  color:#FFF;
  opacity:1;
}
.actions{
  display: flex;
}
</style>
