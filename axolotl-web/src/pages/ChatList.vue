
<template>
  <div class="chatList">
    <div v-if="editActive" class="actions-header">
      <button class="btn hide-actions" @click="editDeactivate">
        <font-awesome-icon icon="times"/>
      </button>
    </div>
    <div v-if="chats.length>0" class="row">
      <button id="chat.id"  v-for="chat in chats" class="btn col-12 chat"
      @click="enterChat('/chat/'+chat.Tel)"
          >
        <div class="row chat-entry">
          <div class="avatar col-3">
            <div v-if="chat.IsGroup" class="badge-name"><font-awesome-icon icon="user-friends" /></div>
            <div v-else class="badge-name">{{chat.Name[0]}}</div>
          </div>
  				<div class="meta col-9 row" v-longclick="editChat">
            <div class="col-9">
  					       <p class="name">{{chat.Name}}</p>
            </div>
            <div v-if="!editActive" class="col-3">
                <p v-if="chat.Messages&&chat.Messages!=null" class="time">{{humanifyDate(chat.Messages[chat.Messages.length-1].SentAt)}}</p>
            </div>
            <div v-else class="col-3 text-right">
              <font-awesome-icon icon="trash"  @click="delChat($event, chat)"/>
            </div>
            <div class="col-12">
              <p class="preview" v-if="chat.Messages&&chat.Messages!=null">{{chat.Messages[chat.Messages.length-1].Message}}</p>
            </div>
          </div>
				</div>
      </button>
    </div>
    <div v-else class="no-entries">
      No chats available
    </div>
    <!-- {{chats}} -->
    <router-link :to="'/contacts/'" class="btn start-chat"><font-awesome-icon icon="pencil-alt" /></router-link>

  </div>
</template>

<script>
export default {
  name: 'ChatList',
  props: {
    msg: String
  },
  created(){
    this.$store.dispatch("getChatList");
    this.$store.dispatch("clearMessageList");
    // this.$store.dispatch('addResponses', "1");
    // Vue.prototype.$store.dispatch("getChatList")
  },
  mounted(){
    this.$store.dispatch("getContacts")
  },
  data() {
    return {
      editActive: false,
      editWasActive: false
    }
  },
  methods:{
    humanifyDate(inputDate){
      var now = new Date();
      var date = new Date(inputDate);
      var diff=(now-date)/1000;
      var seconds = diff;
      if(seconds<60)return "now";
      var minutes = seconds/60;
      if(minutes<60)return Math.floor(minutes)+" MIN";
      var hours = minutes/60
      if(hours<24)return Math.floor(hours)+" H";
      return date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate()
    },
    editChat(e){
      this.editActive=true;
    },
    editDeactivate(e){
      this.editActive=false;
      e.preventDefault();
      e.stopPropagation();
      this.editWasActive = true;

    },
    delChat(e, chat){
      this.editActive=false;
      e.preventDefault();
      e.stopPropagation();
      this.$store.dispatch("delChat", chat.Tel);
      this.editWasActive = true;
    },
    enterChat(e){
      if(!this.editActive){
        router.push (e)
      }

    }
  },
  computed: {
    chats () {
      return this.$store.state.chatList.filter(e=>e.Messages!=null).sort(function(a, b) {
        return b.Messages[b.Messages.length-1].SentAt - a.Messages[a.Messages.length-1].SentAt;
      });
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
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
.avatar {
    justify-content: center;
    display: flex;
    align-items: center;
}
.badge-name{
  background-color: #2090ea;
  /* padding: 14px; */
  width:50px;
  height:50px;
  border-radius: 50%;
  color: #FFF;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 16px;
  display:flex;
  justify-content: center;
  align-items:center;
}
.meta{
  text-align:left;
}
.meta p{
  margin:0px;
}
.meta .name{
  font-weight:bold;
  font-size: 18px;
}
.meta .preview{
  font-size:15px;
}
.row.chat-entry{
  border-bottom:1px solid grey;
  border-bottom: 1px solid #c2c2c2;
  padding: 10px;
}
a.chat{
  color:#000;
}
a:hover.chat{
  text-decoration:none;
}
.btn.start-chat {
  position: fixed;
  bottom: 16px;
  right: 10px;
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
.chatList .preview{
  overflow: hidden;
  height: 20px;
}.chatList .time{
  font-size:12px;
}
</style>
