
<template>
  <div class="chat">
    <div class="messageList-container" id="messageList-container" @scroll="handleScroll($event)">
      <div id="messageList" class="messageList row" v-if="messages && messages.length>0" >
        <div v-if="showFullscreenImgSrc!=''" class="fullscreenImage">
          <img :src="'http://localhost:9080/attachments?file='+showFullscreenImgSrc" >
          <button class="btn btn-secondary close" @click="showFullscreenImgSrc=''">X</button>
        </div>
        <div v-if="showFullscreenVideoSrc!=''" class="fullscreenImage">
          <video controls>
            <source :src="'http://localhost:9080/attachments?file='+showFullscreenVideoSrc">
              Your browser does not support the audio element.
          </video>
          <button class="btn btn-secondary close" @click="showFullscreenVideoSrc=''">X</button>
        </div>
          <div v-for="(message,i) in messages.slice().reverse()"
              :class="{'col-12':true, 'sent':message.Outgoing, 'reply':!message.Outgoing, 'error':message.SentAt==0}"
              v-bind:key="i"
               >
            <div class="row w-100">
              <div class="col-12 data">
                <div class="avatar">
                </div>
                <div class="message">
                  <div class="sender" v-if="!message.Outgoing&&isGroup">
                    <div v-if="names[message.Source]">
                      {{names[message.Source]}}
                    </div>
                    <div v-else>{{getName(message.Source)}}</div>
                  </div>
                  <div v-if="message.Attachment!=''" class="attachment">
                    <div class="gallery" v-if="isAttachmentArray(message.Attachment)">
                      <div  v-for="m in isAttachmentArray(message.Attachment)"
                        v-bind:key="m.File">
                        <div v-if="m.CType==2" class="attachment-img">
                          <img  :src="'http://localhost:9080/attachments?file='+m.File" @click="showFullscreenImg(m.File)"/>
                        </div>
                        <div v-else-if="m.CType==3" class="attachment-audio">
                          <audio controls>
                            <source :src="'http://localhost:9080/attachments?file='+m.File" type="audio/mpeg">
                              Your browser does not support the audio element.
                          </audio>
                        </div>
                        <div v-else-if="m.CType==0" class="attachment-file">
                          <a :href="'http://localhost:9080/attachments?file='+m.File">File</a>
                        </div>
                        <div v-else-if="m.CType==5" class="attachment-video" @click="showFullscreenVideo(m.File)">
                          <video @click="showFullscreenVideo(m.File)">
                            <source :src="'http://localhost:9080/attachments?file='+m.File">
                              Your browser does not support the audio element.
                          </video>
                        </div>
                        <div v-else class="attachment">
                          Not supported mime type: {{m.CType}}
                        </div>
                      </div>
                    </div>
                    <!-- this is legacy code -->
                    <div v-else-if="message.CType==2" class="attachment-img">
                      <img  :src="'http://localhost:9080/attachments?file='+message.Attachment" @click="showFullscreenImg(message.Attachment)"/>
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
                    <div v-else-if="message.CType==5" class="attachment-video" @click="showFullscreenVideo(message.Attachment)">
                      <video @click="showFullscreenVideo(message.Attachment)">
                        <source :src="'http://localhost:9080/attachments?file='+message.Attachment">
                          Your browser does not support the audio element.
                      </video>
                    </div>
                    <div v-else class="attachment">
                      Not supported mime type: {{message.CType}}
                    </div>
                  </div>
                  <div class="message-text">
                    <span v-html="message.Message" v-linkified ></span>
                  </div>
                </div>
              </div>
              <div class="col-12 meta" v-if="message.SentAt!=0">
                {{humanifyDate(message.SentAt)}}
              </div>
              <div v-else class="col-12 meta">
                Error
              </div>

            </div>
          </div>
        </div>

        <div v-else class="no-entries">
          No Messages available.
        </div>
    </div>
    <div class="messageInputBox">
      <!-- <div v-if="chat&&chat.IsGroup&&chat.Name==chat.Tel" class="alert alert-warning">Group has to be updated by a member.</div>
      <div v-else class=""> -->
      <div>
        <div class="row">
          <div class="messageInput-container col-10">
            <textarea id="messageInput" type="textarea" v-model="messageInput"
            @keyup="keyupHandler($event)"
            @click="calcHeightsForInput($event)"
            @focus="calcHeightsForInput($event)"
            @focusout="resetHeights"
            contenteditable="true" v-longclick="paste"/>
          </div>
          <div v-if="messageInput!=''" class="col-2 text-right">
            <button class="btn send" @click="sendMessage"><font-awesome-icon icon="paper-plane" /></button>
          </div>
          <div v-else class="col-2 text-right">
            <button class="btn send" @click="loadAttachmentDialog"><font-awesome-icon icon="plus" /></button>
          </div>
        </div>
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
import AttachmentBar from "@/components/AttachmentBar"
export default {
  name: 'Chat',
  props: {
    chatId: String
  },
  components:{
    AttachmentBar
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
    getName(tel){
      if(this.contacts!=null){
        if(typeof this.names[tel]=="undefined"){
          var contact = this.contacts.find(function(element) {
            return element.Tel == tel;
          });
          if(typeof contact!="undefined"){
            this.names[tel]=contact.Name;
            return contact.Name
          }else{
            this.names[tel] = tel;
            return tel
          }
        }else return this.names[tel]
      }
      return tel;

    },
    loadAttachmentDialog(){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
      this.showAttachmentsBar=true
      }
      else{
        document.getElementById("attachment").click()

      }
    },
    sendDesktopAttachment(evt){
      var f = evt.target.files[0];
      if (f) {
        var r = new FileReader();
        var that = this;
        r.onload = function(e) {
            var attachment = e.target.result;
            that.$store.dispatch("uploadAttachment", attachment);
        }
        r.readAsText(f)
      } else {
        alert("Failed to load file");
      }
    },
    isAttachmentArray(input){
      try{
        console.log()
        var attachments = JSON.parse(input)
        return attachments;

      } catch(e){
        return false;
      }
      // JSON.parse(input)
    },
    showFullscreenImg(img){
      this.showFullscreenImgSrc = img;
    },
    showFullscreenVideo(video){
      this.showFullscreenVideoSrc = video;
    },
    sendMessage(){
      if(this.messageInput!=""){
        this.$store.dispatch("sendMessage", {to:this.chatId, message:this.messageInput});
        this.messageInput=""
        if(this.$store.state.messageList.Messages==null)
        this.$store.dispatch("getMessageList", this.getId());
        this.resetHeights()
      }

      this.scrollDown();
    },
    handleScroll (event) {
      if(event.target.scrollTop<50
        && this.$store.state.messageList.Messages!=null
        &&this.$store.state.messageList.Messages.length>19){
        this.$store.dispatch("getMoreMessages");
      }
      // Any code to be executed when the window is scrolled
    },
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
    back(){
      this.$router.go(-1)
    },
    paste(){
      if(typeof this.config.Gui!="undefined"&&this.config.Gui=="ut"){
        // Don't follow the link
        var result = window.prompt("paste");
        this.messageInput=result;
      }
    },
    scrollDown(){
      document.getElementById("messageList-container").scrollTo(0,document.getElementById("messageList-container").scrollHeight);
    },
    resetHeights(){
      document.getElementById("messageInput").style.height="33px";
      document.getElementById('messageList-container').style.height=window.innerHeight-135+'px';
    },
    keyupHandler(e){
      this.calcHeightsForInput(e);

    },
    calcHeightsForInput(){
      var el = document.getElementById("messageInput");
      var c = document.getElementById("messageList-container");
      if(window.innerHeight-c.clientHeight<200){
        var scroll = el.scrollHeight;
        if(scroll>150)scroll= 150;
        if(Math.abs(this.lastCHeight-c.style.height)>10){
          c.style.height = window.innerHeight-scroll-100+'px';
          this.lastCHeight = c.style.height;
        }
        if(Math.abs(this.lastMHeight-el.style.height)>10){
          el.style.height=el.scrollHeight+'px';
          this.lastMHeight = c.style.height;
        }
      }
      if(el.scrollHeight > el.clientHeight && el.scrollHeight<150){
          el.style.height=el.scrollHeight+5+'px';
          c.style.height = window.innerHeight-el.scrollHeight-100+'px';
          if(document.body.scrollTop+550<document.body.scrollHeight)
          window.scrollTo(0,document.body.scrollHeight);
        }
        // var scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    }


  },
  mounted(){
    this.$store.dispatch("openChat", this.getId());
    this.$store.dispatch("getMessageList", this.getId());
    window.addEventListener('resize', this.resetHeights);
    document.getElementById('messageInput').focus()
    setTimeout(this.scrollDown
    , 300)
      document.addEventListener("scroll", () => {
        var scrolled = document.scrollingElement.scrollTop;
        if(scrolled==0){
          this.$store.dispatch("getMoreMessages");
        }
      });
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
      this.calcHeightsForInput();
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
    }
  },
  computed: {
    chat() {
      return this.$store.state.currentChat
    },
    messages () {
      return this.$store.state.messageList.Messages
    },
    isGroup () {
      return this.$store.state.messageList.Session.IsGroup
    },
    ... mapState(['contacts','config']),
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
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
  height: calc(100vh - 140px);
  transition: width 0.5s, height 0.5s;
}

.chat-list-container::-webkit-scrollbar {
    display: none;
}
.chat{
  position:relative;
  padding-top: 26px;
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
.avatar {
    justify-content: center;
    display: flex;
    align-items: center;
}
.message-text{
  overflow-wrap: break-word;
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
.sender{
  font-size:0.9rem;
  font-weight:bold;
}
video,
.attachment-img img {
    max-width: 100%;
    max-height: 80vh;
}
.sent .message{
  background-color:#d3f2d7;
}
.error .message{
  background-color:#f7663a;
}
.messageInputBox {
  bottom: 0px;
  width: 100vw;
  left: 0px;
  padding: 4px;
  height: -3px;
  z-index: 2;
  background-color:
  #FFF;
  -webkit-transition: width 0.5s, height 0.5s;
  transition: width 0.5s, height 0.5s;
  position: fixed;
  display: flex;
  justify-content: center;
}
.messageInput-container{
  position: relative;
  transition: width 0.5s, height 0.5s;
  padding:0px;
}
#messageInput{
  padding-right:10px;
  border-radius:0px;
  border:none;
  resize: none;
  bottom: 0px;
  width: 100%;
  max-height: 150px;
  border:1px solid #2090ea;
  height: 35px;
  padding:3px 10px;
  transition: width 0.5s, height 0.5s;
  border-radius:4px;
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
.gallery{
  display:flex;

}
.gallery img{
  padding-right:3px;
  padding-bottom:3px;
}

</style>
