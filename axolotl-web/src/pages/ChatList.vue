
<template>
  <div class="chatList">
    <div v-if="editActive" class="actions-header">
      <button class="btn hide-actions" @click="editDeactivate">
        <font-awesome-icon icon="times"/>
      </button>
    </div>
    <div v-if="chats.length>0" class="row">
      <button id="chat.id"  v-for="(chat) in chats" class="btn col-12 chat"
      @click="enterChat(chat )"
          >
        <div class="row chat-entry">
          <div class="avatar col-2">
            <div v-if="chat.IsGroup" class="badge-name"><img class="avatar-img" :src="'http://localhost:9080/avatars?file='+chat.Tel" @error="onImageError($event)"/><font-awesome-icon icon="user-friends" /></div>
            <div v-else class="badge-name">{{chat.Name[0]}}</div>
          </div>
  				<div class="meta col-10 row" v-longclick="editChat">
            <div class="col-9">
  					       <div class="name">
                     <font-awesome-icon class="mute" v-if="!chat.Notification" icon="volume-mute" />
                     <div v-if="chat.IsGroup&&chat.Name==chat.Tel">Unknown group</div>
                     <div v-else>{{chat.Name}}</div>
                     <div v-if="Number(chat.Unread)>0" class="counter badge badge-primary">{{chat.Unread}}</div>
                   </div>

            </div>
            <div v-if="!editActive" class="col-3 date-c">
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
import moment from 'moment';

export default {
  name: 'ChatList',
  props: {
    msg: String
  },
  created(){
    this.$store.dispatch("getChatList");
    this.$store.dispatch("clearMessageList");
  },
  mounted(){
    this.$store.dispatch("getContacts")
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
  },
  data() {
    return {
      editActive: false,
      editWasActive: false
    }
  },
  methods:{
    humanifyDate(inputDate){
      var date = new moment(inputDate);
      var min = moment().diff(date, 'minutes')
      if(min<60){
        if(min == 0) return "now"
        return moment().diff(date, 'minutes') +" min"
      }
      var hours = moment().diff(date, 'hours')
      if(hours <24) return hours + " h"
      return date.format("DD. MMM");
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
    onImageError(event){
      event.target.style.display = "none";
    },
    enterChat(chat){
      if(!this.editActive){
        this.$store.dispatch("setCurrentChat", chat);
        router.push ('/chat/'+chat.Tel)
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
  background: rgb(14,123,210);
  background: linear-gradient(0deg,rgba(14,123,210,1) 8%, rgba(32,144,234,1) 42%, rgba(107,180,238,1) 100%);
  /* padding: 14px; */
  width:44px;
  height:44px;
  border-radius: 50%;
  color: #FFF;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 14px;
  display:flex;
  justify-content: center;
  align-items:center;
  overflow: hidden;
  flex-wrap: wrap;
}
.avatar-img{
  max-width: 100%;
  max-height: 100%;
}
.chat{
      padding: 0px 10px;
}
.date-c{
  display: flex;
  justify-content: flex-end;
  align-items: center;
  padding-top: 10px;
}
.meta{
  text-align:left;
}
.meta p{
  margin:0px;
}
.chat-entry{
  padding: 0px 0px 10px 0px;
}
.meta>div,
.chat-entry>div{
  padding:0px 8px;
}
.meta .name{
  font-weight:bold;
  font-size: 18px;
  display:flex;
  align-items: center;
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
.chatList .mute{
  color: #999;
  margin-right: 10px;
}
.chatList .counter {
  border: 1px solid #FFF;
  border-radius: 50%;
  background-color: #2090ea;
  display:flex;
  justify-content:center;
  align-items: center;
  margin-left:10px;
  width: 28px;
  height: 28px;
}
</style>
