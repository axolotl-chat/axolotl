<template>
  <div
    :key="message.ID"
    :class="{
      'col-12': true,
      outgoing: message.Outgoing,
      sent: message.IsSent && message.Outgoing,
      read: message.IsRead && message.Outgoing,
      delivered: message.Receipt && message.Outgoing,
      incoming: !message.Outgoing,
      status:
        (message.Flags > 0 &&
          message.Flags !== 11 &&
          message.Flags !== 13 &&
          message.Flags !== 14) ||
        message.StatusMessage ||
        (message.Attachment.includes('null') && message.Message === ''),
      hidden: message.Flags === 18,
      error: message.SentAt === 0 || message.SendingError,
    }"
  >
    <div v-if="verifySelfDestruction(message)" class="message">
      <div v-if="isSenderNameDisplayed" class="sender">
        <div v-if="names[message.Source]">
          {{ names[message.Source] }}
        </div>
        <div v-else>{{ getName(message.Source) }}</div>
      </div>
      <blockquote v-if="message.QuotedMessage !== null">
        <cite v-if="message.QuotedMessage.Outgoing" v-translate>You</cite>
        <cite v-else>{{ getName(message.QuotedMessage.Source) }}</cite>
        <p>{{ message.QuotedMessage.Message }}</p>
      </blockquote>
      <div v-if="message.Attachment !== ''" class="attachment">
        <div v-if="isAttachmentArray(message.Attachment)" class="gallery">
          <div
            v-for="m in isAttachmentArray(message.Attachment)"
            :key="m.File"
          >
            <div v-if="m.CType === 2" class="attachment-img">
              <img
                :src="'http://localhost:9080/attachments?file=' + m.File"
                alt="Fullscreen image"
                @click="$emit('showFullscreenImg', m.File)"
              >
            </div>
            <div v-else-if="m.CType === 3" class="attachment-audio">
              <audio controls>
                <source
                  :src="'http://localhost:9080/attachments?file=' + m.File"
                  type="audio/mpeg"
                >
                <span v-translate>Your browser does not support the audio element.</span>
              </audio>
            </div>
            <div
              v-else-if="m.File !== '' && m.CType === 0"
              class="attachment-file"
            >
              <a
                :href="'http://localhost:9080/attachments?file=' + m.File"
                @click="shareAttachment(m.File, $event)"
              >{{ m.FileName ? m.FileName : m.File }}</a>
            </div>
            <div
              v-else-if="m.CType === 5"
              class="attachment-video"
              @click="$emit('showFullscreenVideo', m.File)"
            >
              <video>
                <source
                  :src="'http://localhost:9080/attachments?file=' + m.File"
                >
                <span v-translate>Your browser does not support the audio element.</span>
              </video>
              <img class="play-button" src="../assets/images/play.svg" alt="Play image">
            </div>
            <div v-else-if="m.File !== ''" class="attachment">
              <span v-translate>Not supported mime type:</span> {{ m.CType }}
            </div>
          </div>
        </div>
        <!-- this is legacy code -->
        <div v-else-if="message.CType === 2" class="attachment-img">
          <img
            :src="
              'http://localhost:9080/attachments?file=' + message.Attachment
            "
            alt="Fullscreen image"
            @click="$emit('showFullscreenImg', message.Attachment)"
          >
        </div>
        <div v-else-if="message.CType === 3" class="attachment-audio">
          <audio controls>
            <source
              :src="
                'http://localhost:9080/attachments?file=' + message.Attachment
              "
              type="audio/mpeg"
            >
            <span v-translate>Your browser does not support the audio element.</span>
          </audio>
        </div>
        <div
          v-else-if="message.Attachment !== 'null' && message.CType === 0"
          class="attachment-file"
        >
          {{ message.Attachment }}
          <a
            :href="
              'http://localhost:9080/attachments?file=' + message.Attachment
            "
          >File</a>
        </div>
        <div
          v-else-if="message.CType === 5"
          class="attachment-video"
          @click="$emit('showFullscreenVideo', message.Attachment)"
        >
          <video>
            <source
              :src="
                'http://localhost:9080/attachments?file=' + message.Attachment
              "
            >
            <span v-translate>Your browser does not support the video element.</span>
          </video>
        </div>

        <div v-else-if="message.Attachment !== 'null'" class="attachment">
          <span v-translate>Not supported mime type:</span> {{ message.CType }}
        </div>
      </div>
      <div class="message-text">
        <!-- eslint-disable-next-line vue/no-v-html -->
        <div class="message-text-content" data-test="message-text" v-html="linkify(sanitize(message.Message))" />
        <div v-if="message.Flags===17" v-translate>Group changed.</div>
        <div
          v-if="
            message.Attachment.includes('null') &&
              message.Message === '' &&
              message.Flags === 0
          "
          class="status-message"
        >
          <span v-translate>Set timer for self-destructing messages </span>
          <div>{{ humanifyTimePeriod(message.ExpireTimer) }}</div>
        </div>
        <div v-if="message.Flags === 10" v-translate>
          Unsupported message type: sticker
        </div>
      </div>
      <div v-if="message.SentAt !== 0" class="meta">
        <div class="time">
          <span @click="showDate = !showDate">{{
            humanifyDateFromNow(message.SentAt)
          }}</span>
          <span v-if="showDate" class="fullDate">{{
            humanifyDate(message.SentAt)
          }}</span>
        </div>
        <div v-if="message.ExpireTimer > 0">
          <div class="circle-wrap">
            <div class="circle">
              <div
                class="mask full"
                :style="
                  'transform: rotate(' + timerPercentage(message) + 'deg)'
                "
              >
                <div
                  class="fill"
                  :style="
                    'transform: rotate(' + timerPercentage(message) + 'deg)'
                  "
                />
              </div>
              <div class="mask half">
                <div
                  class="fill"
                  :style="
                    'transform: rotate(' + timerPercentage(message) + 'deg)'
                  "
                />
              </div>
              <div class="inside-circle" />
            </div>
          </div>
        </div>
        <div v-if="message.Outgoing" class="transfer-indicator" />
      </div>
      <div v-else class="col-12 meta">Error</div>
    </div>
  </div>
</template>

<script>
import moment from "moment";
import { mapState } from "vuex";
let decoder;

export default {
  name: "MessageComponent",
  props: {
    message: {
      type: Object,
      default: () => {}
    },
    isGroup: {
      type: Boolean,
      default: false,
    },
    names: {
      type: Object,
      default: () => {}
    }
  },
  emits: ["showFullscreenImg", "showFullscreenVideo"],
  data() {
    return {
      showDate: false,
    };
  },
  computed: {
    ...mapState(["contacts", "config"]),
    isSenderNameDisplayed() {
      return (
        !this.message.Outgoing &&
        this.isGroup &&
        (this.message.Flags === 0 || this.message.Flags === 14)
      ); // #14 is the flag for quoting messages
	    // see this list for all message types: https://github.com/nanu-c/axolotl/blob/main/app/helpers/models.go#L15
    },
  },
  methods: {
    sanitize(msg) {
      decoder = decoder || document.createElement("div");
      decoder.textContent = msg;
      let result = decoder.innerHTML;
      decoder.textContent = result;//escapes twice in order to negate v-html's unescaping
      result = decoder.innerHTML;
      return result;
    },
    getName(tel) {
      if (this.contacts !== null) {
        if (typeof this.names[tel] === "undefined") {
          const contact = this.contacts.find(function (element) {
            return element.Tel === tel;
          });
          if (typeof contact !== "undefined") {
            this.names[tel] = contact.Name;
            return contact.Name;
          } else {
            this.names[tel] = tel;
            return tel;
          }
        } else return this.names[tel];
      }
      return tel;
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
      } else {
        // alert(file)
        // console.log(file)
      }
    },
    timerPercentage(m) {
      const r = moment(m.ReceivedAt);
      const duration = moment.duration(r.diff(moment.now()));
      const percentage =
        1 - (m.ExpireTimer + duration.asSeconds()) / m.ExpireTimer;
      if (percentage < 1) {
        const fullCircle = 179;
        return fullCircle * percentage;
      }
      else return 0;
    },
    verifySelfDestruction(m) {
      if (m.ExpireTimer !== 0) {
        // console.log(m.ExpireTimer,m.SentAt, m.ReceivedAt, Date.now())
        if (m.ReceivedAt !== 0) {
          // hide destructed messages but not timer settings
          const r = moment(m.ReceivedAt);
          const duration = moment.duration(r.diff(moment.now()));
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
      else if (time < 60 * 60 * 24) return time / 60 / 60 / 24 + " d";
      return time;
    },
  },
};
</script>

<style lang="scss" scoped>
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
  font-size: 0.9rem;
  font-weight: bold;
}
.gallery {
  display: flex;

  div + div {
    margin-left: 3px;
  }
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
.hidden{
  display: none;
}
</style>
