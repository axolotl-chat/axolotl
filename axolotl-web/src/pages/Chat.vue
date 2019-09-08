
<template>
  <div class="chat">
    <div class="chatList-container">
      <div class="chatList row" v-if="messages && messages.length>0" v-chat-scroll="{always: false, smooth: true}" v-on:scroll="handleScroll">

          <div v-for="message in messages.slice().reverse()" :class="{'col-12':true, 'sent':message.Outgoing, 'reply':!message.Outgoing}" >
            <div class="row w-100">
              <div class="col-12 data">
                <div class="avatar">
                </div>
                <div class="message">
                  <div v-if="message.Attachment!=''" class="attachment">
                    <div v-if="message.CType==2" class="attachment-img">
                      <img :src="'http://localhost:9080/attachments?file='+message.Attachment" />
                    </div>
                    <div v-else-if="message.CType==3" class="attachment-audio">
                      <audio controls>
                        <source :src="'http://localhost:9080/attachments?file='+message.Attachment" type="audio/mpeg">
                          Your browser does not support the audio element.
                      </audio>
                    </div>
                    <div v-else-if="message.CType==0" class="attachment-file">
                      <a :href="'http://localhost:9080/attachments?file='+message.Attachment">File</a>
                    </div>
                    <div v-else-if="message.CType==5" class="attachment-video">
                      <video controls>
                        <source :src="'http://localhost:9080/attachments?file='+message.Attachment">
                          Your browser does not support the audio element.
                      </video>
                    </div>
                    <div v-else class="attachment">
                      Not supported mime type: {{message.CType}}
                    </div>
                  </div>
                  <div class="message-text">
                    {{message.Message}}
                  </div>
                </div>
              </div>
              <div class="col-12 meta">
                {{humanifyDate(message.SentAt)}}

              </div>

            </div>
          </div>
        </div>

        <div v-else >
          No Messages aviable
        </div>
      </div>
    <div class="messageInputBox">
      <div class="container">
        <div class="row">

          <textarea id="messageInput" class="col-9" type="textarea" v-model="messageInput" />
          <div class="col-3 text-center">
            <button class="btn send" @click="sendMessage">send</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Chat',
  props: {
    chatId: String
  },
  data() {
    return {
      messageInput: ""
    }
  },
  methods: {
    getId: function() {
      return(this.chatId)
    },
    sendMessage(){
      if(this.messageInput!=""){
        this.$store.dispatch("sendMessage", {to:this.chatId, message:this.messageInput});
        this.messageInput=""
      }
    },
    handleScroll (event) {

      if(event.target.scrollTop==0){
        this.$store.dispatch("getMoreMessages");
      }
      // Any code to be executed when the window is scrolled
    },
    humanifyDate(date){
      var now = new Date();
      var date = new Date(date);
      var diff=(now-date)/1000;
      var seconds = diff;
      if(seconds<60)return "now";
      var minutes = seconds/60;
      if(minutes<60)return Math.floor(minutes)+" minutes ago";
      var hours = minutes/60
      if(hours<24)return Math.floor(hours)+" hours ago";
      return date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate() + " " + date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds()
    },
    back(){
      this.$router.go(-1)
    }

  },
  created(){
    this.$store.dispatch("getMessageList", this.getId());
  },
  mounted(){
    window.scrollTo(0,document.body.scrollHeight);

  },
  computed: {
    messages () {
      return this.$store.state.messageList.Messages
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.header {
  text-align: left;
}
.chatList{
  height:calc(100vh - 200px);
  overflow: hidden auto;
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.chat-list-container::-webkit-scrollbar {
    display: none;
}
.chat{
  position:relative;
  padding-top:30px;
}
.chat-list-container{
  padding-bottom:70px;
  overflow: hidden;
  height:90vh;
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.chatList > div:last-child {
    padding-bottom: 99px;
}
.avatar {
    justify-content: center;
    display: flex;
    align-items: center;
}
.reply{
  text-align:left;
  margin-bottom:10px;
}
.sent{
  display:flex;
  justify-content:flex-end;
}
.data{
    display:flex;
}
.sent .data,
.sent .meta{
  display:flex;
  justify-content:flex-end;
}
.meta {
    font-size: 13px;
    padding: 5px 20px;
}

.message{
  padding:10px;
  border-radius:10px;
  max-width:70%;
  background-color:#dfdfdf;
  text-align:left;
  min-width:100px;
}
video,
.attachment-img img {
    max-width: 100%;
    max-height: 80vh;
}
.sent .message{
  background-color:#d3f2d7;
}
.messageInputBox {
    position: fixed;
    bottom: 0px;
    width: 100%;
    left: 0px;
    padding: 10px;
    max-width:100vw;
    height:80px;
    z-index:2;
    background-color:#FFF;
}
#messageInput{
  padding-right:10px;
  border-radius:10px;
  border:none;
  resize: none;
  border:1px solid #2090ea;

}
.send{
  background-color:#2090ea;
  color:#FFF;
}
</style>
