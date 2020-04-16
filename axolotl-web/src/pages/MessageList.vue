
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
              <span v-translate>Your browser does not support the audio element.</span>
          </video>
          <button class="btn btn-secondary close" @click="showFullscreenVideoSrc=''">X</button>
        </div>
          <div v-for="(message) in messageList.Messages.slice().reverse()"
              :class="{'col-12':true,
                      'sent':message.Outgoing,
                      'reply':!message.Outgoing,
                      'status':message.Flags>0||message.StatusMessage||message.Attachment.includes('null')&&message.Message=='',
                      'error':message.SentAt==0||message.SendingError,
                      /* 'sending':!message.IsSent&&message.Outgoing, */
                      /* 'receipt':message.Receipt||message.Outgoing&&message.SentAt<1586984922935 */
                      'receipt':true
                      }"
              v-bind:key="message.ID"
               >
            <div class="row w-100" v-if="verifySelfDestruction(message)">
              <div class="col-12 data">
                <div class="avatar" v-if="message.Flags==0">
                </div>
                <div class="message">
                  <div class="sender" v-if="!message.Outgoing&&isGroup&&message.Flags==0">
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
                              <span v-translate>Your browser does not support the audio element.</span>
                          </audio>
                        </div>
                        <div v-else-if="m.File!='' &&m.CType==0" class="attachment-file">
                          <a @click="shareAttachment(m.File,$event)" :href="'http://localhost:9080/attachments?file='+m.File">{{m.FileName?m.FileName:m.File}}</a>
                        </div>
                        <div v-else-if="m.CType==5" class="attachment-video" @click="showFullscreenVideo(m.File)">
                          <video @click="showFullscreenVideo(m.File)">
                            <source :src="'http://localhost:9080/attachments?file='+m.File">
                              <span v-translate>Your browser does not support the audio element.</span>
                          </video>
                        </div>
                        <div v-else-if="m.File!=''" class="attachment">
                          <span v-translate>Not supported mime type:</span> {{m.CType}}
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
                          <span v-translate>Your browser does not support the audio element.</span>
                      </audio>
                    </div>
                    <div v-else-if="message.Attachment!='null'&&message.CType==0" class="attachment-file">
                      {{message.Attachment}}
                      <a :href="'http://localhost:9080/attachments?file='+message.Attachment">File</a>
                    </div>
                    <div v-else-if="message.CType==5" class="attachment-video" @click="showFullscreenVideo(message.Attachment)">
                      <video @click="showFullscreenVideo(message.Attachment)">
                        <source :src="'http://localhost:9080/attachments?file='+message.Attachment">
                          <span v-translate>Your browser does not support the video element.</span>
                      </video>
                    </div>

                    <div v-else-if="message.Attachment!='null'" class="attachment">
                      <span v-translate>Not supported mime type:</span> {{message.CType}}
                    </div>
                  </div>
                  <div class="message-text">
                    <span v-html="message.Message" v-linkified ></span>
                    <div class="status-message" v-if="message.Attachment.includes('null')&&message.Message==''&&message.Flags==0">
                      <span v-translate>Set timer for self-destructing messages </span>
                      <div> {{humanifyTimePeriod(message.ExpireTimer)}}</div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-12 meta" v-if="message.SentAt!=0">
                <div class="time">{{humanifyDate(message.SentAt)}}</div>
                  <div v-if="message.ExpireTimer>0&&message.Message!=''">
                    <div class="circle-wrap">
                      <div class="circle">
                        <div class="mask full" :style="'transform: rotate('+timerPercentage(message)+'deg)'">
                          <div class="fill" :style="'transform: rotate('+timerPercentage(message)+'deg)'"></div>
                        </div>
                        <div class="mask half">
                          <div class="fill" :style="'transform: rotate('+timerPercentage(message)+'deg)'"></div>
                        </div>
                        <div class="inside-circle">
                        </div>
                      </div>
                  </div>
                </div>
              </div>
              <div v-else class="col-12 meta">
                Error
              </div>

            </div>
          </div>
        </div>

        <div v-else class="no-entries">
          <span v-translate>No Messages available.</span>
        </div>
    </div>
    <div class="messageInputBox">
      <!-- <div v-if="chat&&chat.IsGroup&&chat.Name==chat.Tel" class="alert alert-warning">Group has to be updated by a member.</div>
      <div v-else class=""> -->
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
    isAttachmentArray(input){
      try{
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
    verifySelfDestruction(m){
      if(m.ExpireTimer!=0){
        // console.log(m.ExpireTimer,m.SentAt, m.ReceivedAt, Date.now())
        if(m.ReceivedAt!=0){
          // hide destructed messages but not timer settings
          var r = moment(m.ReceivedAt)
          var duration = moment.duration(r.diff(moment.now()));
          if((duration.asSeconds()+m.ExpireTimer)<0&&m.Message!=""){
            return false;
          }
        }
        else if (m.SentAt!=0){
          var rS = moment(m.SentAt)
          var durationS = moment.duration(rS.diff(moment.now()));
          if((durationS.asSeconds()+m.ExpireTimer)<0&&m.Message!=""){
            return false;
          }
        }
      }
      return true
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
      if(document.getElementById("messageList-container")!=null)
      document.getElementById("messageList-container").scrollTo(0,document.getElementById("messageList-container").scrollHeight);
    },
    resetHeights(){
      document.getElementById("messageInput").style.height="33px";
      document.getElementById('messageList-container').style.height=window.innerHeight-135+'px';
    },
    keyupHandler(e){
      this.calcHeightsForInput(e);
    },
    humanifyTimePeriod(time){
      if(time<60)
      return time +" s";
      else if(time<60*60)
      return time/60+" m"
      else if(time<60*60*24)
      return time/60/60+" h"
      else if(time<60*60*24)
      return time/60/60/24+" d"

      return time
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
    , 600)
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
    ... mapState(['contacts','config','messageList']),
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
    padding: 0px 20px;
    display: flex;
}

.message {
  padding: 5px 8px;
  border-radius:10px;
  max-width:85%;
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
.status .message{
  background-color:transparent;
  width:100%;
  display:flex;
  justify-content:center;
  font-weight:600;
  text-align: center;
}
.status .data{
  justify-content:center;
}
.status{
  justify-content:center;

}
.status .status-message{
  width:100%;
  display:flex;
  justify-content:center;
  font-weight:600;
  text-align: center;
  flex-direction:column;
}
.status .status-message span{
  padding-right:4px;
}
.status .meta{
  text-align: center;
  justify-content:center;
}
.error .message{
  background-color:#f7663a;
}
.error .meta{
  color:#f7663a;
}
.messageInputBox {
  bottom: 0px;
  width: 100vw;
  left: 0px;
  padding: 4px;
  height: -3px;
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
  width: 75vw;
}
#messageInput{
  padding-right:10px;
  border-radius:0px;
  resize: none;
  bottom: 0px;
  width: 100%;
  max-height: 150px;
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
.circle-wrap {
  margin: 2px auto;
  width: 15px;
  height: 15px;
  background: #e6e2e7;
  border-radius: 50%;
}
.circle-wrap .circle .mask,
.circle-wrap .circle .fill {
  width: 16px;
  height: 16px;
  position: absolute;
  border-radius: 50%;
}
.circle-wrap .circle .mask {
  clip: rect(0px, 16px, 16px, 8px);
}
.circle-wrap .circle .mask .fill {
  clip: rect(0px, 8px, 16px, 0px);
  background-color: #9e00b1;
}

.circle-wrap .circle.p0 .mask.full,
.circle-wrap .circle.p0 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(00deg);
}
.circle-wrap .circle.p50 .mask.full,
.circle-wrap .circle.p50 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(180deg);
}
.circle-wrap .circle.p100 .mask.full,
.circle-wrap .circle.p100 .fill {
  /* animation: fill ease-in-out 3s; */
  transform: rotate(360deg);
}
</style>
