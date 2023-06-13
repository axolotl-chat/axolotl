<template>
  <div v-if="message.message && ((message.message_type == 'DataMessage' && message.message !== '') ||
    (message.message_type == 'SyncMessage' && message.message && message.message !== 'SyncMessage')) ||
    (message.attachments.length > 0)" :key="message.ID" :class="{
    'col-12': true,
    'message-container': true,
    outgoing: is_outgoing,
    sent: message.is_sent && is_outgoing,
    read: message.IsRead && is_outgoing,
    delivered: message.Receipt && is_outgoing,
    incoming: !is_outgoing,
    status:
      (message.Flags > 0 &&
        message.Flags !== 11 &&
        message.Flags !== 13 &&
        message.Flags !== 14) ||
      message.StatusMessage ||
      (message.Attachment &&
        message.Attachment.includes('null') &&
        message.Message === ''),
    hidden: message.Flags === 18,
    error: message.timestamp === 0 || message.SendingError,
    'group-message':
      isGroup,
  }">
    <div v-if="!is_outgoing &&
      isGroup
      " class="avatar">
      <div class="badge-name" @click="openProfileForRecipient(message.sender)">
        <div v-if="message.sender !== -1">
          <img class="avatar-img" :src="'http://localhost:9080/avatars?recipient=' + message.sender" @error="onImageError($event)" />
        </div>
        <div v-if="name && name !== ''">
          {{ name && name[0] }}
        </div>
      </div>
    </div>
    <div v-if="verifySelfDestruction(message)" :class="{
      message: true,
      'col-7':
        isGroup &&
        (message.Flags === 0 || message.Flags === 12 || message.Flags === 13) &&
        !is_outgoing
    }">
      <div v-if="isSenderNameDisplayed" class="sender">
        <div v-if="name !== ''">
          {{ name }}
        </div>
      </div>
      <blockquote v-if="message.QuotedMessage">
        <cite v-if="message.QuotedMessage && is_outgoing" v-translate>You</cite>
        <cite v-else>{{ name ? name : getName(message.QuotedMessage.SourceUUID) }}</cite>
        <p>{{ message.QuotedMessage.Message }}</p>
      </blockquote>
      <div v-if="message.attachments.length > 0" class="attachment">
        <div :class="`gallery-${message.attachments.length>3?'big':'small'}`">
          <div v-for=" m in message.attachments" :key="m.File" class="item">
            <div v-if="m.ctype === 'image'" class="attachment-img">
              <img :src="'http://localhost:9080/attachments/' + m.filename" alt="image"
                @click="$emit('show-fullscreen-img', m.filename)" />
            </div>
            <div v-else-if="m.ctype === 'audio'" class="attachment-audio">
              <div class="audio-player-container d-flex">
                <button id="play-icon">
                  <font-awesome-icon v-if="!isPlaying" class="play" icon="play" @click="play" />
                  <font-awesome-icon v-if="isPlaying" class="pause" icon="pause" @click="pause" />
                </button>
                <span v-if="!isPlaying" id="duration" class="time">{{
                  humanifyTimePeriod(duration)
                }}</span>
                <span v-if="isPlaying" id="currentTime" class="time">{{ humanifyTimePeriod(currentTime) }} /
                  {{ humanifyTimePeriod(duration) }}</span>
              </div>
            </div>
            <div v-else-if="m.filename !== '' && m.ctype === 'file'" class="attachment-file">
              <a :href="'http://localhost:9080/attachments/' + m.filename" @click="shareAttachment(m.filename, $event)">{{
                m.fileName
              }}</a>
            </div>
            <div v-else-if="m.ctype === 'video'" class="attachment-video"
              @click="$emit('show-fullscreen-video', m.filename)">
              <video>
                <source :src="'http://localhost:9080/attachments/' + m.filename" />
                <span v-translate>Your browser does not support the audio element.</span>
              </video>
              <img class="play-button" src="../assets/images/play.svg" alt="Play image" />
            </div>
            <!-- <div v-else-if="m.File !== ''" class="attachment">
              <span v-translate>Not supported mime type:</span> {{ m.CType }}
            </div> -->
          </div>
        </div>
        <!-- <div v-else-if="message.Attachment !== 'null'" class="attachment">
          <span v-translate>Not supported mime type:</span> {{ message.CType }}
        </div> -->
      </div>
      <div v-if="message.message" class="message-text">
        {{ message.message }}
        <!-- eslint-disable-next-line vue/no-v-html -->
        <!--<div
          v-if="message.Flags !== 17" class="message-text-content" data-test="message-text"
          v-html="linkify(sanitize(message.Message))"
        />-->
        <div v-if="message.Flags === 17" v-translate>Group changed.</div>
        <div v-if="message.attachments.length === 0 &&
          !message.message &&
          !isGroup
          " class="status-message">
          <span v-translate>Set timer for self-destructing messages </span>
          <div>{{ humanifyTimePeriod(message.ExpireTimer) }}</div>
        </div>
        <div v-if="message.Flags === 10" v-translate>
          Unsupported message type: sticker
        </div>
      </div>
      <div v-if="message.timestamp !== 0" class="meta">
        <div class="time">
          <span @click="showDate = !showDate">{{
            humanifyDateFromNow(message.timestamp)
          }}</span>
          <span v-if="showDate" class="fullDate">{{ humanifyDate(message.timestamp) }}</span>
        </div>
        <div v-if="message.ExpireTimer > 0">
          <div class="circle-wrap">
            <div class="circle">
              <div class="mask full" :style="'transform: rotate(' + timerPercentage(message) + 'deg)'">
                <div class="fill" :style="'transform: rotate(' + timerPercentage(message) + 'deg)'" />
              </div>
              <div class="mask half">
                <div class="fill" :style="'transform: rotate(' + timerPercentage(message) + 'deg)'" />
              </div>
              <div class="inside-circle" />
            </div>
          </div>
        </div>
        <!-- <div v-if="is_outgoing" class="transfer-indicator" /> -->
      </div>
      <div v-else class="col-12 meta">Error</div>
    </div>
  </div>
</template>

<script>
import moment from "moment";
import { mapState } from "vuex";
import { router } from "@/router/router";
let decoder;

export default {
  name: "MessageComponent",
  props: {
    message: {
      type: Object,
      default: () => { },
    },
    isGroup: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["show-fullscreen-img", "show-fullscreen-video"],
  data() {
    return {
      showDate: false,
      audio: null,
      duration: 0.0,
      currentTime: 0.0,
      isPlaying: false,
      id: -1,
      senderName: "",
    };
  },
  computed: {
    ...mapState(["currentGroup", "config", "contacts"]),
    is_outgoing() {
      if (this.message?.sender == this.config.uuid)
        return true;
      else if (this.message && this.message.is_outgoing) {
        return true;
      } else {
        return false;
      }
    },
    isSenderNameDisplayed() {
      return (
        !this.is_outgoing &&
        this.isGroup
        //&&
        // (this.message.Flags === 0 || this.message.Flags === 14)
      ); // #14 is the flag for quoting messages
      // see this list for all message types: https://github.com/nanu-c/axolotl/blob/main/app/helpers/models.go#L15
    },
    name() {
      if (this.senderName == "") {
        const uuid = this.message.sender
        const contact = this.contacts.find(function (element) {
          return element.uuid === uuid;
        });
        if (typeof contact !== "undefined") {
          return contact.name;
        }

      }
        return this.senderName;
    }
  },
  mounted() {
    if (
      this.message.Attachment &&
      this.message.Attachment !== "" &&
      this.message.Attachment !== null
    ) {
      const attachment = JSON.parse(this.message.Attachment);
      if (attachment && attachment.length > 0 && attachment[0].CType === 3) {
        this.audio = new Audio(
          "http://localhost:9080/attachments/" + attachment[0].File
        );
        var that = this;
        this.audio.onloadedmetadata = function () {
          that.duration = that.audio.duration.toFixed(2);
        };
        this.audio.onended = function () {
          that.audio.currentTime = 0;
          that.isPlaying = false;
        };
        this.audio.ontimeupdate = function () {
          that.currentTime = that.audio.currentTime.toFixed(2);
        };
      }
    }
  },
  methods: {
    sanitize(msg) {
      decoder = decoder || document.createElement("div");
      decoder.textContent = msg;
      let result = decoder.innerHTML;
      decoder.textContent = result; //escapes twice in order to negate v-html's unescaping
      result = decoder.innerHTML;
      return result;
    },
    onImageError(event) {
      event.target.style.display = "none";
    },
    openProfileForRecipient(recipient) {
      if (recipient !== -1) {
        router.push("/profile/" + recipient);
      } else {
        // name == uuid of the recipient
        // this.$store.dispatch("createRecipient", this.name);
        this.$store.dispatch("createRecipientAndAddToGroup", {
          id: this.message.SourceUUID,
          group: this.currentGroup.HexId,
        });
      }
    },
    getName() {
      if(!this.isGroup) return "";
      if (this.contacts !== null) {
        const uuid = this.message.sender;
        const contact = this.contacts.find(function (element) {
          return element.uuid === uuid;
        });
        if (typeof contact !== "undefined") {
          this.name = contact.name;
          return this.name;
        }
      }
      if (this.currentGroup !== null && this.currentGroup.Members !== null) {
        const contact = this.currentGroup.Members.find(function (element) {
          return element.UUID === uuid;
        });
        if (typeof contact !== "undefined") {
          this.id = contact.Id;
          if (contact.ProfileGivenName !== "") this.name = contact.ProfileGivenName;
          else this.name = contact.Username;
          return this.name;
        }
      }
      this.name = this.message.sender;
      return this.message.sender;
    },
    isAttachmentArray(input) {
      try {
        return JSON.parse(input);
      } catch (e) {
        return false;
      }
      // JSON.parse(input)
    },
    shareAttachment(file, e) {
      if (typeof this.config.Gui !== "undefined" && this.config.Gui === "ut") {
        e.preventDefault();
        alert("[oD]" + file);
        // this.showAttachmentsBar=true
      }
    },
    timerPercentage(m) {
      const received = moment(m.ReceivedAt);
      const duration = moment.duration(received.diff(moment.now()));
      const percentage = 1 - (m.ExpireTimer + duration.asSeconds()) / m.ExpireTimer;
      if (percentage < 1) {
        const fullCircle = 179;
        return fullCircle * percentage;
      } else return 0;
    },
    verifySelfDestruction(m) {
      if (m.ExpireTimer !== 0) {
        if (m.ReceivedAt !== 0) {
          // hide destructed messages but not timer settings
          const received = moment(m.ReceivedAt);
          const duration = moment.duration(received.diff(moment.now()));
          if (duration.asSeconds() + m.ExpireTimer < 0) {
            this.$store.dispatch("deleteSelfDestructingMessage", m);
            return false;
          }
        } else if (m.SentAt !== 0) {
          const rS = moment(m.SentAt);
          const durationS = moment.duration(rS.diff(moment.now()));
          if (durationS.asSeconds() + m.ExpireTimer < 0 && m.Message !== "") {
            this.$store.dispatch("deleteSelfDestructingMessage", m);
            return false;
          }
        }
      }
      return true;
    },
    humanifyDate(inputDate) {
      return new moment(inputDate).format("lll");
    },
    humanifyDateFromNow(inputDate) {
      return new moment(inputDate).fromNow();
    },
    humanifyTimePeriod(time) {
      if (time < 60) return time + " s";
      else if (time < 60 * 60) return time / 60 + " m";
      else if (time < 60 * 60 * 24) return time / 60 / 60 + " h";
      else if (time > 60 * 60 * 24) return time / 60 / 60 / 24 + " d";
      return time;
    },
    play() {
      this.audio.play();
      this.isPlaying = true;
    },
    pause() {
      if (this.audio && this.audio.currentTime > 0) {
        this.audio.pause();
        this.isPlaying = false;
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.gallery-big {
  display: flex;
  flex-flow: wrap;
  align-content: space-between;
  margin: auto;

  .item {
    box-sizing: border-box;
    width: 32%;
    margin-bottom: 2%;
    padding: 5px;
  }

  .item:nth-child(3n+1) {
    order: 1;
  }

  .item:nth-child(3n+2) {
    order: 2;
  }

  .item:nth-child(3n) {
    order: 3;
  }
}

.message-text {
  overflow-wrap: break-word;
}

.incoming {
  text-align: left;
}

.outgoing {
  display: flex;
  justify-content: flex-end;
}

.meta {
  display: flex;
  align-items: center;
  font-size: 11px;
}

.outgoing .meta {
  justify-content: flex-end;
}

.message {
  margin-bottom: 10px;
  padding: 8px 12px;
  border-radius: 10px;
  max-width: 85%;
  text-align: left;
  min-width: 100px;
}

.error .message {
  border: solid #f7663a 2px;
}

.sender {
  font-size: 0.7rem;
}

.gallery {
  // display: flex;
  max-width: 100%;
}

video,
.attachment-img img {
  max-width: 100%;
  max-height: 80vh;
}

.outgoing .attachment-img {
  background: center center no-repeat;
  background-color: #000;
  background-image: url("../assets/images/loading.svg");

  img {
    opacity: 0.2;
  }
}

.sent .attachment-img {
  background-color: #eee; // To deal with images with a transparent background
  background-image: none;

  img {
    opacity: 1;
  }
}

.status .message {
  background-color: transparent;
  width: 100%;
  font-weight: 600;
  text-align: center;
}

.status .status-message {
  width: 100%;
  display: flex;
  justify-content: center;
  font-weight: 600;
  text-align: center;
  flex-direction: column;
}

.status .status-message span {
  padding-right: 4px;
}

.status .meta {
  text-align: center;
  justify-content: center;
}

.transfer-indicator {
  width: 18px;
  height: 12px;
  margin-left: 4px;
  background-repeat: no-repeat;
  background-position: left center;
}

.error .transfer-indicator {
  background-image: url("../assets/images/warning.svg");
}

.circle-wrap {
  margin-top: 3px;
  margin-left: 5px;
  width: 15px;
  height: 15px;
  background: #e6e2e7;
  border-radius: 50%;
  position: relative;
}

.circle-wrap .circle .mask,
.circle-wrap .circle .fill {
  width: 16px;
  height: 16px;
  top: 0;
  left: 0;
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

.message-text .message-text-content {
  white-space: pre-line;
}

.attachment-video {
  position: relative;

  .play-button {
    margin: auto;
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    bottom: 0;
  }
}

blockquote {
  padding: 0.5rem;
  margin-top: 3px;
  margin-bottom: 5px;
  background-color: #00000044;
  border-left: solid 4px #00000044;
  border-radius: 4px;

  cite {
    font-style: normal;
    font-weight: bold;
  }

  p {
    margin: 0;
  }
}

.fullDate {
  font-style: italic;
  margin-left: 2px;
}

.hidden {
  display: none;
}

button {
  padding: 0;
  border: 0;
  background: inherit;
  cursor: pointer;
  outline: none;
  width: 40px;
  height: 40px;
}

.audio-player-container {
  position: relative;
  width: 90%;
  max-width: 500px;
  height: 80px;

  p {
    position: absolute;
    top: -18px;
    right: 5%;
    padding: 0 5px;
    margin: 0;
    font-size: 28px;
    background: #fff;
  }

  #play-icon {
    margin: 20px 2.5% 20px 2.5%;
  }
}

.incoming.group-message {
  display: flex;

  .badge-name {
    margin-right: 10px;
  }
}</style>
