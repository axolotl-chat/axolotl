
<template>
  <div class="chatList" v-if="chatList">
    <div v-if="editActive" class="actions-header">
      <button class="btn hide-actions" @click="delChat($event)">
        <font-awesome-icon icon="trash"/>
      </button>
      <button class="btn hide-actions" @click="editDeactivate">
        <font-awesome-icon icon="times"/>
      </button>
    </div>
    <div class="row" v-for="(chat) in chats" v-bind:key="chat.id">
      <div id="chat.id"
      :class="editActive&&selectedChat.indexOf(chat.Tel)>=0?'selected col-12 chat-container':'col-12 chat-container '"
      @click="enterChat(chat)">
        <div class="row chat-entry">
          <div class="avatar col-2">
            <div v-if="chat.IsGroup" class="badge-name">
              <img class="avatar-img" :src="'http://localhost:9080/avatars?file='+chat.Tel" @error="onImageError($event)"/>
              <font-awesome-icon icon="user-friends" />
            </div>
            <div v-else class="badge-name">{{chat.Name[0]}}</div>
          </div>
          <div class="meta col-10" v-longclick="()=>{editChat(chat.Tel)}">
            <div class="row">
              <div class="col-9">
                <div class="name">
                   <font-awesome-icon class="mute" v-if="!chat.Notification" icon="volume-mute" />
                   <div v-if="chat.IsGroup&&chat.Name==chat.Tel" v-translate>Unknown group</div>
                   <div v-else>{{chat.Name}}</div>
                   <div v-if="Number(chat.Unread)>0" class="counter badge badge-primary">{{chat.Unread}}</div>
                 </div>
              </div>
              <div v-if="!editActive" class="col-3 date-c">
                  <p v-if="chat.Messages&&chat.Messages!=null&&chat.Messages[chat.Messages.length-1].SentAt!=0" class="time">
                    {{humanifyDate(chat.Messages[chat.Messages.length-1].SentAt)}}
                  </p>
              </div>
            </div>
            <div class="row">
              <div class="col-12">
                <p class="preview" v-if="chat.Messages&&chat.Messages!=null">{{chat.Messages[chat.Messages.length-1].Message}}</p>
              </div>
            </div>
          </div>
				</div>
      </div>
    </div>
    <div v-if="chats.length==0" class="no-entries" v-translate>
      No chats available
    </div>
    <!-- {{chats}} -->
    <router-link :to="'/contacts/'" class="btn start-chat"><font-awesome-icon icon="pencil-alt" /></router-link>

  </div>
</template>

<script>
import moment from 'moment';
import { mapState } from 'vuex';
import { router } from '../router/router';

export default {
  name: 'ChatList',
  props: {
    msg: String
  },
  created(){
  },
  mounted(){
    this.$store.dispatch("getChatList");
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
    this.sortChats();
    var userLang = navigator.language || navigator.userLanguage;
    this.$language.current = userLang;
    this.$store.dispatch("leaveChat");
    this.$store.dispatch("clearMessageList");
    this.$store.dispatch("clearFilterContacts");
    this.$store.dispatch("getConfig");
    this.$store.dispatch("getContacts")
  },
  data() {
    return {
      editActive: false,
      editWasActive: false,
      selectedChat:[],
      chats:[]
    }
  },
  methods:{
    humanifyDate(inputDate){
      moment.locale(this.$language.current)
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
      this.selectedChat.push(e);
      this.editActive=true;
    },
    editDeactivate(e){
      this.editActive=false;
      e.preventDefault();
      e.stopPropagation();
      this.editWasActive = true;
      this.selectedChat = [];

    },
    delChat(e){
      this.editActive=false;
      e.preventDefault();
      e.stopPropagation();
      if(this.selectedChat.length>0){
        this.selectedChat.forEach(c=>{
          this.$store.dispatch("delChat", c);
        })
      }
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
      else{
          this.selectedChat.push(chat.Tel);
      }

    },
    sortChats(){
      this.chats = this.$store.state.chatList.filter(e=>e.Messages!=null).sort(function(a, b) {
              return b.Messages[b.Messages.length-1].SentAt - a.Messages[a.Messages.length-1].SentAt;
        });
    }
  },
  computed: mapState(['chatList']),
  watch:{
    chatList(){
      this.sortChats();
    }
  }

}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style>
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
.actions-header .btn{
  color:#FFF;
}

.hide-actions{
  padding-right:40px;
}

.chat-entry {
  padding: 10px 0;
  cursor: pointer;
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
  height: 100%;
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
.meta .name{
  font-weight:bold;
  font-size: 18px;
  display:flex;
  align-items: center;
}
.meta .preview{
  font-size:15px;
}
a.chat-container{
  color:#000;
}
a:hover.chat-container{
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
  border-radius: 50%;
  background-color: #2090ea;
  display:flex;
  justify-content:center;
  align-items: center;
  margin-left:10px;
  width: 28px;
  height: 28px;
}
.chatList .selected{
  background-color:#c5e4f0;
}
</style>
