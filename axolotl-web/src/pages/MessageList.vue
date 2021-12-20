<template>
  <div class="chat">
    <div
      id="messageList-container"
      class="messageList-container"
      @scroll="handleScroll($event)"
    >
      <div
        v-if="messages && messages.length > 0"
        id="messageList"
        class="messageList row"
      >
        <div v-if="showFullscreenImgSrc !== ''" class="fullscreenImage">
          <img
            :src="
              'http://localhost:9080/attachments?file=' + showFullscreenImgSrc
            "
            alt="Fullscreen image"
          />
          <button class="btn btn-secondary save" @click="saveImg($event)">
            <font-awesome-icon icon="arrow-down" />
          </button>
          <button
            class="btn btn-secondary close"
            @click="showFullscreenImgSrc = ''"
          >
            X
          </button>
        </div>
        <div v-if="showFullscreenVideoSrc !== ''" class="fullscreenImage">
          <video controls>
            <source
              :src="
                'http://localhost:9080/attachments?file=' +
                  showFullscreenVideoSrc
              "
            />
            <span v-translate>Your browser does not support the audio element.</span>
          </video>
          <button
            class="btn btn-secondary close"
            @click="showFullscreenVideoSrc = ''"
          >
            X
          </button>
          <button class="btn btn-secondary save" @click="saveVideo($event)">
            <font-awesome-icon icon="arrow-down" />
          </button>
        </div>
        <message
          v-for="message in messageList.Messages.slice().reverse()"
          :key="message.ID"
          :message="message"
          :is-group="isGroup"
          :names="names"
          @showFullscreenImg="showFullscreenImg($event)"
          @showFullscreenVideo="showFullscreenVideo($event)"
        />
      </div>
      <div v-else class="no-entries">
        <span v-translate>No messages available.</span>
      </div>
      <div id="chat-bottom" />
    </div>
    <div
      v-if="chat.IsGroup && chat.GroupJoinStatus !== 0"
      v-translate
      class="messageInputBoxDisabled w-100"
    >
      You have been invited to this group. Accepting invitations is not yet
      supported. In order to join please send private a message to the other
      side to exchange the profile key. After that he/she has to remove and add
      you again.
    </div>
    <div v-if="!recorded" class="messageInputBox">
      <div v-if="!recording" class="messageInput-container">
        <textarea
          id="messageInput"
          v-model="messageInput"
          type="textarea"
          contenteditable="true"
          data-long-press-delay="500"
          @long-press="paste"
        />
      </div>
      <div v-if="messageInput !== ''" class="messageInput-btn-container">
        <button class="btn send" @click="sendMessage">
          <font-awesome-icon icon="paper-plane" />
        </button>
      </div>
      <div v-else-if="recording" class="messageInput-btn-container d-flex justify-content-center w-100">
        <div v-translate class="me-5">Recording...</div>
        <button class="btn send record-stop" @click="stopRecording">
          <font-awesome-icon icon="stop-circle" />
        </button>
      </div>
      <div v-else class="messageInput-btn-container d-flex">
        <button class="btn send record me-2" @click="recordAudio">
          <font-awesome-icon icon="microphone" />
        </button>
        <button class="btn send" @click="loadAttachmentDialog">
          <font-awesome-icon icon="plus" />
        </button>
      </div>
    </div>
    <div v-else class="messageInputBox justify-content-center">
      <div class="messageInput-btn-container d-flex justify-content-center align-items-center">
        <div><span>{{ duration }}</span><span v-translate class="me-2">s</span></div>
        <button v-if="!playing" class="btn send play me-1" @click="playAudio">
          <font-awesome-icon icon="play" />
        </button>
        <button v-else class="btn send stop me-1" @click="stopPlayAudio">
          <font-awesome-icon icon="stop-circle" />
        </button>
        <button class="btn send delete me-1" @click="deleteAudio">
          <font-awesome-icon icon="times" />
        </button>
        <button class="btn send send-voice" @click="sendVoiceNote">
          <font-awesome-icon icon="paper-plane" />
        </button>
      </div>
    </div>
    <audio v-if="blobUrl!=''"
           id="voiceNote"
           controls
           :src="blobUrl"
    >
      Your browser does not support the
      <code>audio</code> element.
    </audio>
    <attachment-bar
      v-if="showAttachmentsBar"
      @close="showAttachmentsBar = false"
      @send="callContentHub($event)"
    />
    <input
      id="attachment"
      type="file"
      style="position: fixed; top: -100em"
      @change="sendDesktopAttachment"
    />
    <audio id="voiceNote" :src="blobUrl" style="position: fixed; top: -100em" />
  </div>
</template>

<script>
import { mapState } from "vuex";
import Message from "@/components/Message";
import AttachmentBar from "@/components/AttachmentBar";
import { saveAs } from "file-saver";
import * as MicRecorder from 'mic-recorder-to-mp3';
export default {
  name: "MessageList",
  components: {
    AttachmentBar,
    Message,
  },
  props: {
    chatId: { type: Number, default: -1 },
  },
  data() {
    return {
      messageInput: "",
      scrolled: false,
      showAttachmentsBar: false,
      showFullscreenImgSrc: "",
      showFullscreenVideoSrc: "",
      names: {},
      lastCHeight: 0,
      lastMHeight: 0,
      scrollLocked: false,
      recorder: null,
      recording: false,
      recorded: false,
      playing: false,
      blobUrl: "",
      voiceNoteElem: null,
      duration:0,
    };
  },
  computed: {
    chat() {
      return this.$store.state.currentChat;
    },
    messages() {
      return this.$store.state.messageList.Messages;
    },
    isGroup() {
      return this.$store.state.messageList.Session.IsGroup;
    },
    ...mapState(["contacts", "config", "messageList"]),
  },
  watch: {
    messageInput() {
      // Adapt height of the textarea when its content changed
      let textarea = document.getElementById("messageInput");
      if (this.messageInput === "") {
        textarea.style.height = "35px";
      } else {
        textarea.style.height = 0; // Set height to 0 to reset scrollHeight to its minimum
        textarea.style.height = textarea.scrollHeight + 5 + "px";
      }
    },
    contacts() {
      if (this.contacts !== null) {
        Object.keys(this.names).forEach((i) => {
          const contact = this.contacts.find(function (element) {
            return element.Tel === i;
          });
          if (typeof contact !== "undefined") {
            this.names[i] = contact.Name;
          }
        });
      }
    },
    messages: {
      // This will let Vue know to look inside the array
      deep: true,
      handler() {
        if (!this.$data.scrollLocked) setTimeout(this.scrollDown, 600);
        this.$data.scrollLocked = false;
        // this.scrollDown();
      },
    },
  },
  mounted() {
    this.$store.dispatch("openChat", this.getId());
    this.$store.dispatch("getMessageList", this.getId());
    document.getElementById("messageInput").focus();
    setTimeout(this.scrollDown, 600);
    const that = this;
    document.addEventListener(
      "click",
      function (event) {
        // If the clicked element doesn't have the right selector, bail
        if (!event.target.matches(".linkified")) return;
        if (
          typeof that.config.Gui !== "undefined" &&
          that.config.Gui === "ut"
        ) {
          // Don't follow the link
          event.preventDefault();
          alert(event.target.href);
        }
        // else
        // console.log(that.config.Gui)
      },
      false
    );
  },
  methods: {
    getId: function () {
      return this.chatId;
    },
    callContentHub(type) {
      this.showAttachmentsBar = false;
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        const result = window.prompt(type);
        this.showSettingsMenu = false;
        if (result !== "canceld")
          this.$store.dispatch("sendAttachment", {
            type: type,
            path: result,
            to: this.chatId,
            message: this.messageInput,
          });
      } else {
        // this.showSettingsMenu = false;
        // document.getElementById("addVcf").click()
      }
    },
    loadAttachmentDialog() {
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        this.showAttachmentsBar = true;
      } else {
        document.getElementById("attachment").click();
      }
    },
    sendDesktopAttachment(evt) {
      const f = evt.target.files[0];
      if (f) {
        const r = new FileReader();
        const that = this;
        r.onload = function (e) {
          const attachment = e.target.result;
          that.$store.dispatch("uploadAttachment", {
            attachment: attachment,
            to: that.chatId,
            message: this.messageInput,
          });
        };
        r.readAsDataURL(f);
      } else {
        alert("Failed to load file");
      }
    },
    showFullscreenImg(img) {
      this.showFullscreenImgSrc = img;
    },
    showFullscreenVideo(video) {
      this.showFullscreenVideoSrc = video;
    },
    saveImg(e) {
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        e.preventDefault();
        alert("[oP]" + this.showFullscreenImgSrc);
      } else
        saveAs(
          "http://localhost:9080/attachments?file=" + this.showFullscreenImgSrc
        );
    },
    saveVideo(e) {
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        e.preventDefault();
        alert("[oV]" + this.showFullscreenVideoSrc);
      } else
        saveAs(
          "http://localhost:9080/attachments?file=" +
            this.showFullscreenVideoSrc
        );
    },
    sendMessage() {
      if (this.messageInput !== "") {
        this.$store.dispatch("sendMessage", {
          to: this.chatId,
          message: this.messageInput,
        });
        if (this.$store.state.messageList.Messages === null) {
          this.$store.dispatch("getMessageList", this.getId());
        }
        this.messageInput = "";
      }
      this.scrollDown();
    },
    handleScroll(event) {
      if (
        !this.$data.scrollLocked &&
        event.target.scrollTop < 80 &&
        this.$store.state.messageList !== null &&
        this.$store.state.messageList.Messages !== null &&
        this.$store.state.messageList.Messages !== undefined &&
        this.$store.state.messageList.Messages.length > 19
      ) {
        this.$data.scrollLocked = true;
        this.$store.dispatch("getMoreMessages");
      }
    },
    back() {
      this.$router.go(-1);
    },
    paste() {
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        // Don't follow the link
        const result = window.prompt("paste");
        this.messageInput = this.messageInput + result;
      }
    },
    scrollDown() {
      document.getElementById("chat-bottom").scrollIntoView();
    },
    recordAudio(){
      var that = this;
      this.recording = true;
      this.playing =  false;
       navigator.mediaDevices.getUserMedia({
          video: false,
          audio: true
      }).then(async function() {
          that.recorder = new MicRecorder({
            bitRate: 128
          });
          that.recorder.start().then(() => {
            // something else
          }).catch((e) => {
            /* eslint-disable no-console */
            console.error(e);
          });
      });
    },
    deleteAudio(){
      this.playing = false;
      this.recorded = false;
      this.recorder = null;
    },
    playAudio(){
      this.playing = true;
      this.voiceNoteElem.play();

    },
    stopPlayAudio(){
      this.playing = false;
      this.voiceNoteElem = document.getElementById("voiceNote");
      this.voiceNoteElem.pause();
    },
    stopRecording(){
      this.recording = false;
      this.recorded = true;
      this.recorder.stop()
      this.recorder.getMp3().then(([_buffer, blob]) => {
        this.blobUrl = URL.createObjectURL(blob);
        this.blobObj = blob;
        let that = this;
        setTimeout(function(){
          that.voiceNoteElem = document.getElementById("voiceNote");
          that.duration = that.voiceNoteElem.duration;
        }, 200);
      })

    },
    sendVoiceNote(){
      this.recorded = false;
      let reader = new FileReader();
      let voiceNoteElem = document.getElementById("voiceNote");
      voiceNoteElem.pause()
      this.playing = false;
      /* eslint-disable no-unused-vars */
      this.recorder.getMp3().then(([buffer, _blob]) => {
        const file = new File([this.blobObj] , 'voice.mp3', {
          type: "audio/mpeg",
          lastModified: Date.now()
        });
        reader.readAsDataURL(file); // converts the blob to base64 and calls onload
        var that = this;
        reader.onload = function() {
          var r = reader.result
          const m = {
            note:r,
            to: that.chatId,
          }
          that.$store.dispatch("sendVoiceNote", m);
        };
      });
    },
  },
};
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
.messageList {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.messageList-container {
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
#messageInput {
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
textarea:focus,
input:focus {
  outline: none;
}
.send {
  background-color: #2090ea;
  color: #fff;
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
  height: unset;
}
.fullscreenImage .close {
  position: absolute;
  right: 10px;
  top: 10px;
  padding: 10px;
  background-color: #ffffff;
}
.fullscreenImage .save {
  position: absolute;
  right: 50px;
  top: 10px;
  padding: 10px;
  background-color: #ffffff;
  color: black;
}
.messageInputBoxDisabled {
  color: red;
}
#voiceNote{
  position:fixed;
  top:-100em;
}
</style>
