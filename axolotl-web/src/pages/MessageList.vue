
<template>
  <div class="chat">
    <div class="messageList-container" id="messageList-container" @scroll="handleScroll($event)">
      <div id="messageList" class="messageList row" v-if="messages && messages.length>0" >
        <div v-if="showFullscreenImgSrc!=''" class="fullscreenImage">
          <img :src="'http://localhost:9080/attachments?file='+showFullscreenImgSrc" >
          <button class="btn btn-secondary save" @click="saveImg($event)">
            <font-awesome-icon icon="arrow-down" />
          </button>
          <button class="btn btn-secondary close" @click="showFullscreenImgSrc=''">X</button>
        </div>
        <div v-if="showFullscreenVideoSrc!=''" class="fullscreenImage">
          <video controls>
            <source :src="'http://localhost:9080/attachments?file='+showFullscreenVideoSrc">
              <span v-translate>Your browser does not support the audio element.</span>
          </video>
          <button class="btn btn-secondary close" @click="showFullscreenVideoSrc=''">X</button>
          <button class="btn btn-secondary save" @click="saveVideo($event)">
            <font-awesome-icon icon="arrow-down" />
          </button>
        </div>
        <message v-for="(message) in messageList.Messages.slice().reverse()"
          :key="message.ID" :message="message" :isGroup="isGroup" :names="names"
          @showFullscreenImg="showFullscreenImg($event)"
          @showFullscreenVideo="showFullscreenVideo($event)">
        </message>
      </div>
      <div v-else class="no-entries">
        <span v-translate>No Messages available.</span>
      </div>
      <div id="chat-bottom"></div>
    </div>
    <div class="messageInputBox">
      <!-- <div v-if="chat&&chat.IsGroup&&chat.Name==chat.Tel" class="alert alert-warning">Group has to be updated by a member.</div>
      <div v-else class=""> -->
      <div class="messageInput-container">
        <textarea id="messageInput" type="textarea" v-model="messageInput"
        contenteditable="true" v-longclick="paste"/>
      </div>
      <div class="messageInput-btn-container" v-if="messageInput!=''">
        <button class="btn send" @click="sendMessage">
          <font-awesome-icon icon="paper-plane" />
        </button>
      </div>
      <div class="messageInput-btn-container" v-else>
        <button class="btn send" @click="loadAttachmentDialog">
          <font-awesome-icon icon="plus" />
        </button>
      </div>
    </div>
    <attachment-bar v-if="showAttachmentsBar"
    @close="showAttachmentsBar=false"
    @send="callContentHub($event)" />
    <input id="attachment" type="file" @change="sendDesktopAttachment" style="position: fixed; top: -100em">
  </div>
</template>

<script>
import { mapState } from 'vuex';
import moment from 'moment';
import Message from "@/components/Message"
import AttachmentBar from "@/components/AttachmentBar"
import { saveAs } from 'file-saver';
export default {
  name: 'Chat',
  props: {
    chatId: String
  },
  components:{
    AttachmentBar,
    Message
  },
  data() {
    return {
      messageInput: "",
      scrolled:false,
      showAttachmentsBar:false,
      showFullscreenImgSrc:"",
      showFullscreenVideoSrc:"",
      names:{},
      lastCHeight:0,
      lastMHeight:0,
      scrollLocked: false
    }
  },
  methods: {
    getId: function() {
      return(this.chatId)
    },
    callContentHub(type) {
      this.showAttachmentsBar = false;
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        var result = window.prompt(type);
        this.showSettingsMenu = false;
        if(result!="canceld")
        this.$store.dispatch("sendAttachment", {type:type, path:result, to: this.chatId, message:this.messageInput});
      } else {
        // this.showSettingsMenu = false;
        // document.getElementById("addVcf").click()
      }
    },
    loadAttachmentDialog(){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
      this.showAttachmentsBar=true
      }
      else{
        document.getElementById("attachment").click()
      }
    },
    shareAttachment(file, e){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        e.preventDefault();
        alert("[oD]"+file)
      // this.showAttachmentsBar=true
      }
      else{
        // alert(file)
          // console.log(file)
      }
    },
    sendDesktopAttachment(evt){
      var f = evt.target.files[0];
      if (f) {
        var r = new FileReader();
        var that = this;
        r.onload = function(e) {
            var attachment = e.target.result;
            that.$store.dispatch("uploadAttachment", {attachment:attachment, to: that.chatId, message:this.messageInput});
        }
        r.readAsDataURL(f)
      } else {
        alert("Failed to load file");
      }
    },
    showFullscreenImg(img){
      this.showFullscreenImgSrc = img;
    },
    showFullscreenVideo(video){
      this.showFullscreenVideoSrc = video;
    },
    saveImg(e){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        e.preventDefault();
        alert("[oP]"+this.showFullscreenImgSrc)
      } else
      saveAs('http://localhost:9080/attachments?file='+this.showFullscreenImgSrc)
    },
    saveVideo(e){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        e.preventDefault();
        alert("[oV]"+this.showFullscreenVideoSrc)
      } else
      saveAs('http://localhost:9080/attachments?file='+this.showFullscreenVideoSrc)
    },
    sendMessage(){
      if(this.messageInput!=""){
        this.$store.dispatch("sendMessage", {to:this.chatId, message:this.messageInput});
        if(this.$store.state.messageList.Messages == null) {
          this.$store.dispatch("getMessageList", this.getId());
        }
        this.messageInput = "";
      }

      this.scrollDown();
    },
    timerPercentage(m){
      var r = moment(m.ReceivedAt)
      var duration = moment.duration(r.diff(moment.now()));
      var percentage = 1-((m.ExpireTimer+duration.asSeconds())/m.ExpireTimer)
      if(percentage<1)
      return 179*percentage
      else return 0
    },
    handleScroll (event) {
      if(!this.$data.scrollLocked
        && event.target.scrollTop < 80
        && this.$store.state.messageList.Messages != null
        && this.$store.state.messageList.Messages.length > 19) {
        this.$data.scrollLocked = true;
        this.$store.dispatch("getMoreMessages");
      }
    },
    back(){
      this.$router.go(-1)
    },
    paste(){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        // Don't follow the link
        var result = window.prompt("paste");
        this.messageInput=this.messageInput+result;
      }
    },
    scrollDown(){
      document.getElementById("chat-bottom").scrollIntoView();
    }
  },
  mounted(){
    this.$store.dispatch("openChat", this.getId());
    this.$store.dispatch("getMessageList", this.getId());
    document.getElementById('messageInput').focus();
    setTimeout(this.scrollDown, 600);
    var that = this;
    document.addEventListener('click', function (event) {

      // If the clicked element doesn't have the right selector, bail
      if (!event.target.matches('.linkified')) return;
      if(typeof that.config.Gui!="undefined"&&that.config.Gui=="ut"){
        // Don't follow the link
        event.preventDefault();
        alert(event.target.href)
      }
      // else
      // console.log(that.config.Gui)
    }, false);
  },
  watch:{
    messageInput(){
      // Adapt height of the textarea when its content changed
      let textarea = document.getElementById("messageInput");
      if (this.messageInput == "") {
        textarea.style.height = "35px";
      } else {
        textarea.style.height = 0; // Set height to 0 to reset scrollHeight to its minimum
        textarea.style.height = textarea.scrollHeight + 5 + "px";
      }
    },
    contacts(){
      if(this.contacts!=null){
        Object.keys(this.names).forEach((i)=>{
          var contact = this.contacts.find(function(element) {
            return element.Tel == i;
          });
          if(typeof contact!="undefined"){
            this.names[i]=contact.Name;
          }
        });
      }
    },
    messages: {
      // This will let Vue know to look inside the array
      deep:true,
      handler(){
        this.$data.scrollLocked = false;
      }
    }
  },
  computed: {
    chat() {
      return this.$store.state.currentChat
    },
    messages() {
      return this.$store.state.messageList.Messages
    },
    isGroup() {
      return this.$store.state.messageList.Session.IsGroup
    },
    ... mapState(['contacts','config','messageList']),
  }
}
</script>

<style scoped>
.chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  justify-content: space-between;
}

.header {
  text-align: left;
}
.messageList{
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.messageList-container{
  overflow-x: hidden;
  overflow-y: scroll;
  transition: width 0.5s, height 0.5s;
  padding-top: 5px;
}

.chat-list-container::-webkit-scrollbar {
    display: none;
}
/* .chat-list-container{
  padding-bottom:70px;
  overflow: hidden;
  height:90vh;
  -ms-overflow-style: none;
  scrollbar-width: none;
} */
.messageList > div:last-child {
    padding-bottom: 20px;
}
.messageInputBox {
  display: flex;
  margin: 5px 0px;
}
.messageInput-container {
  flex-grow: 1;
}
.messageInput-btn-container {
  flex-grow: 0;
  flex-shrink: 1;
  margin-left: 15px;
}
#messageInput{
  resize: none;
  width: 100%;
  height: 35px;
  max-height: 150px;
  padding: 3px 10px;
  border-radius: 4px;
  ::-webkit-scrollbar {
      display: block;
  }
}
textarea:focus, input:focus{
    outline: none;
}
.send{
  background-color: #2090ea;
  color: #FFF;
  border-radius: 50%;
  width: 35px;
  height: 35px;
  font-size: 15px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.fullscreenImage {
    position: fixed;
    z-index: 100;
    top: 0px;
    left: 0px;
    width: 100vw;
    height: 100vh;
    background-color: black;
    display: flex;
    justify-content: center;
    align-items: center;
}
.fullscreenImage img,
.fullscreenImage video {
  max-height: 100%;
  max-width: 100%;
  height:unset;
}
.fullscreenImage .close {
  position:absolute;
  right:10px;
  top:10px;
  padding:10px;
  background-color:#FFFFFF;
}
.fullscreenImage .save {
  position:absolute;
  right:50px;
  top:10px;
  padding:10px;
  background-color:#FFFFFF;
  color:black;
}
</style>
