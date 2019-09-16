
<template>
  <div class="chatList">
    <div v-if="chats.length>0" class="row">
      <router-link :to="'/chat/'+chat.Tel" v-for="chat in chats" class="col-12 chat">
        <div class="row chat-entry">
          <div class="avatar col-3">
            <div v-if="chat.IsGroup" class="badge-name"><font-awesome-icon icon="user-friends" /></div>
            <div v-else class="badge-name">{{chat.Name[0]}}</div>
          </div>
  				<div class="meta col-9 row">
            <div class="col-9">
  					       <p class="name">{{chat.Name}}</p>
            </div>
            <div class="col-3">
                <p v-if="chat.Messages" class="time">{{humanifyDate(chat.Messages[chat.Messages.length-1].SentAt)}}</p>
            </div>
            <div class="col-12">
              <p class="preview" v-if="chat.Messages">{{chat.Messages[chat.Messages.length-1].Message}}</p>
            </div>
          </div>
				</div>
      </router-link>
    </div>
    <div v-else class="no-entries">
      No chats aviable
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
  },
  computed: {
    chats () {
      return this.$store.state.chatList
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
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
  font-size:20px;
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
