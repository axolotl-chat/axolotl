<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="chat" v-if="currentChat">
      <div id="messageList-container" class="messageList-container" @scroll="handleScroll($event)">
        <div v-if="messages && messages.length > 0" id="messageList" class="messageList row">
          <div v-if="showFullscreenImgSrc !== ''" class="fullscreenImage">
            <img :src="'http://localhost:9080/attachments/' + showFullscreenImgSrc
              " alt="Fullscreen image" />
            <button class="btn btn-secondary save" @click="saveImg($event)">
              <font-awesome-icon icon="arrow-down" />
            </button>
            <button class="btn btn-secondary close" @click="showFullscreenImgSrc = ''">
              <font-awesome-icon icon="times" />
            </button>
          </div>
          <div v-if="showFullscreenVideoSrc !== ''" class="fullscreenImage">
            <video controls>
              <source :src="'http://localhost:9080/attachments/' +
                showFullscreenVideoSrc
                " />
              <span v-translate>Your browser does not support the audio element.</span>
            </video>
            <button class="btn btn-secondary close" @click="showFullscreenVideoSrc = ''">
              <font-awesome-icon icon="times" />
            </button>
            <button class="btn btn-secondary save" @click="saveVideo($event)">
              <font-awesome-icon icon="arrow-down" />
            </button>
          </div>
          <message v-for="message in messageList" :key="message.ID" :message="message" :is-group="isGroup"
            @show-fullscreen-img="showFullscreenImg($event)" @show-fullscreen-video="showFullscreenVideo($event)"
            @click="handleClick($event)" />
        </div>
        <div v-else class="no-entries">
          <span v-translate>No messages available.</span>
        </div>
        <div id="chat-bottom" />
      </div>
      <div v-if="isGroup && currentGroup && currentGroup.JoinStatus == 1" class="messageInputBoxDisabled w-100">
        <p v-translate>
          Join this group? They won’t know you’ve seen their messages until you
          accept.
        </p>
        <div v-translate class="btn btn-primary" @click="joinGroupAccept">
          Join
        </div>
      </div>
      <div v-else-if="isGroup && currentGroup && currentGroup.JoinStatus == 2" class="messageInputBoxDisabled w-100">
        <p v-translate>You have been removed from this group.</p>
      </div>
      <div v-else class="bottom-wrapper">
        <div v-if="!voiceNote.recorded" class="messageInputBox">
          <div v-if="!voiceNote.recording" class="messageInput-container">
            <textarea id="messageInput" v-model="messageInput" type="textarea" contenteditable="true"
              data-long-press-delay="500" @long-press="paste" />
          </div>
          <div v-if="messageInput !== ''" class="messageInput-btn-container">
            <button class="btn send" @click="sendMessage">
              <font-awesome-icon icon="paper-plane" />
            </button>
          </div>
          <div v-else-if="voiceNote.recording" class="messageInput-btn-container d-flex justify-content-center w-100">
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
            <div>
              <span>{{ Math.floor(voiceNote.duration) }}</span><span v-translate class="me-2">s</span>
            </div>
            <button v-if="!voiceNote.playing" class="btn send play me-1" @click="playAudio">
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
        <audio v-if="voiceNote.blobUrl != ''" id="voiceNote" controls :src="voiceNote.blobUrl">
          Your browser does not support the
          <code>audio</code> element.
        </audio>
        <attachment-bar v-if="showAttachmentsBar" @close="showAttachmentsBar = false" @send="callContentHub($event)" />
        <input id="attachment" type="file" @change="sendDesktopAttachment" style="position: fixed; top: -100em" />
        <!--  TODO: Fix attachment sending
        <audio
          id="voiceNote"
          :src="voiceNote.blobUrl"
          style="position: fixed; top: -100em"
        />
        -->
      </div>
    </div>
  </component>
</template>

<script>
import { mapState } from "vuex";
import Message from "@/components/Message";
import AttachmentBar from "@/components/AttachmentBar";
import DefaultLayout from "@/layouts/Default";
import { saveAs } from "file-saver";
import MicRecorder from "@jmd01/mic-recorder-to-mp3";
export default {
  name: "MessageList",
  components: {
    AttachmentBar,
    Message,
    DefaultLayout,
  },
  props: {
    chatId: { type: String, default: "" },
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
      voiceNote: {
        recorder: null,
        recording: false,
        recorded: false,
        playing: false,
        blobUrl: "",
        voiceNoteElem: null,
        duration: 0,
        firstScroll: true,
      },
    };
  },
  computed: {
    messages() {
      return this.$store.state.messageList;
    },
    isGroup() {
      if (!this.$store.state.currentChat)return false;
      return this.$store.state.currentChat?.thread?.Group !=undefined;
    },
    ...mapState(["contacts", "config", "messageList", "currentGroup", "currentChat"]),
  },
  watch: {
    messageInput() {
      // Adapt height of the textarea when its content changed
      let textarea = document.getElementById("messageInput");
      if (!textarea) return;
      if (this.messageInput === "") {
        textarea.style.height = "35px";
      } else {
        textarea.style.height = 0; // Set height to 0 to reset scrollHeight to its minimum
        textarea.style.height = textarea.scrollHeight + 5 + "px";
      }
      localStorage.setItem(JSON.stringify(this.currentChat.thread),this.messageInput);
    },
    currentChat: {
      handler() {
        const savedInput = localStorage.getItem(JSON.stringify(this.currentChat.thread));
        if(savedInput && savedInput !== null) this.messageInput = savedInput;
        this.scrollDown();
      },
      deep: true,
    },
    messages: {
      // This will let Vue know to look inside the array
      deep: true,
      handler() {
        if (!this.$data.scrollLocked) setTimeout(this.scrollDown, 600);
        if(!this.$data.firstScroll){
        setTimeout(this.$data.scrollLocked = false, 1000);

        }else this.$data.scrollLocked = false;
        
        this.$data.firstScroll = false;
        // this.scrollDown();
      },
    },
  },
  mounted() {
    this.$store.dispatch("openChat", this.getId());
    const mi = document.getElementById("messageInput");
    if (mi) mi.focus();
    setTimeout(this.scrollDown, 600);
  },
  methods: {
    getId: function () {
      return this.chatId;
    },
    callContentHub(type) {
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
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
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
        this.showAttachmentsBar = true;
      } else {
        document.getElementById("attachment").click();
      }
    },
    sendDesktopAttachment(evt) {
      const file = evt.target.files[0];
      if (file) {
        console.log(file);
        const reader = new FileReader();
        const that = this;
        reader.onload = function (e) {
          const attachment = e.target.result;
          that.$store.dispatch("uploadAttachment", {
            attachment: attachment,
            to: that.chatId,
            message: this.messageInput,
          });
        };
        reader.readAsDataURL(file);
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
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
        e.preventDefault();
        alert("[oP]" + this.showFullscreenImgSrc);
      } else
        saveAs(
          "http://localhost:9080/attachments/" + this.showFullscreenImgSrc
        );
    },
    saveVideo(e) {
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
        e.preventDefault();
        alert("[oV]" + this.showFullscreenVideoSrc);
      } else
        saveAs(
          "http://localhost:9080/attachments/" +
            this.showFullscreenVideoSrc
        );
    },
    sendMessage() {
      if (this.messageInput !== "") {
        console.log("send message")
        this.$store.dispatch("sendMessage", {
          to: this.chatId,
          message: this.messageInput,
        });
        if (this.$store.state.messageList === null) {
          this.$store.dispatch("getMessageList", this.getId());
        }
        this.messageInput = "";
        console.log(this.currentChat.thread)
      }
      this.scrollDown();
    },
    joinGroupAccept() {
      this.$store.dispatch("joinGroup", this.chat.UUID);
    },
    handleScroll(event) {
      if (
        !this.$data.scrollLocked &&
        event.target.scrollTop < 80 &&
        this.$store.state.messageList &&
        this.$store.state.messageList.length > 19
      ) {
        this.$data.scrollLocked = true;
        this.$store.dispatch("getMoreMessages");
      }
    },
    back() {
      this.$router.go(-1);
    },
    paste() {
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
        // Don't follow the link
        const result = window.prompt("paste");
        this.messageInput = this.messageInput + result;
      }
    },
    scrollDown() {
      if (this.messages && this.messages.length !== 0 && document.getElementById("chat-bottom"))
        document.getElementById("chat-bottom").scrollIntoView();
    },
    recordAudio() {
      var that = this;
      this.voiceNote.playing = false;
      navigator.mediaDevices
        .getUserMedia({
          video: false,
          audio: true,
        })
        .then(async function () {
          that.voiceNote.recorder = new MicRecorder({
            bitRate: 128,
          });
          that.voiceNote.recorder
            .start()
            .then(() => {
              // something else
              that.voiceNote.recording = true;
            })
            .catch((e) => {
              //skipqc: JS-0002
              /* eslint-disable no-console */
              console.error(e);
            });
        });
    },
    deleteAudio() {
      this.stopPlayAudio();
      this.voiceNote.playing = false;
      this.voiceNote.recorded = false;
      this.voiceNote.recorder = null;
    },
    playAudio() {
      this.voiceNote.playing = true;
      this.voiceNote.voiceNoteElem.play();
    },
    stopPlayAudio() {
      this.voiceNote.playing = false;
      this.voiceNote.voiceNoteElem.pause();
    },
    stopRecording() {
      this.voiceNote.recording = false;
      this.voiceNote.recorded = true;
      this.voiceNote.recorder.stop();
      this.voiceNote.recorder.getMp3().then(([, blob]) => {
        this.voiceNote.blobUrl = URL.createObjectURL(blob);
        this.voiceNote.blobObj = blob;
        let that = this;
        setTimeout(function () {
          that.voiceNote.voiceNoteElem = document.getElementById("voiceNote");
          that.voiceNote.duration = that.voiceNote.voiceNoteElem.duration;
        }, 200);
      });
    },
    sendVoiceNote() {
      this.voiceNote.recorded = false;
      let reader = new FileReader();
      this.voiceNote.voiceNoteElem.pause();
      this.voiceNote.playing = false;
      /* eslint-disable no-unused-vars */
      this.voiceNote.recorder.getMp3().then(() => {
        const file = new File([this.voiceNote.blobObj], "voice.mp3", {
          type: "audio/mpeg",
          lastModified: Date.now(),
        });
        reader.readAsDataURL(file); // converts the blob to base64 and calls onload
        var that = this;
        reader.onload = function () {
          var result = reader.result;
          const message = {
            note: result,
            to: that.chatId,
          };
          that.$store.dispatch("sendVoiceNote", message);
        };
      });
    },
    handleClick(event) {
      // If the clicked element doesn't have the right selector, bail
      if (!event.target.matches(".linkified")) return;
      if (typeof this.config.feature !== "undefined" && this.config.feature === "ut") {
        // Don't follow the link
        event.preventDefault();
        alert(event.target.href);
      }
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
.messageList>div:last-child {
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
  color: black;
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

#voiceNote {
  position: fixed;
  top: -100em;
}
</style>
