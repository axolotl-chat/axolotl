<template>
  <component :is="$route.meta.layout || 'div'">
    <template #menu>
      <router-link v-translate class="dropdown-item" :to="'/contacts/'">
        Contacts
      </router-link>
      <router-link v-translate class="dropdown-item" :to="'/settings/'">
        Settings
      </router-link>
    </template>
    <div v-if="chatList" class="chatList">
      <div v-if="editActive" class="actions-header">
        <button class="btn hide-actions" @click="delChat($event)">
          <font-awesome-icon icon="trash" />
        </button>
        <button class="btn hide-actions" @click="editDeactivate">
          <font-awesome-icon icon="times" />
        </button>
      </div>
      <div v-for="chat in chatList" :key="chat.id.Contact ? chat.id.Contact : chat.id.Group" class="row">
        <!-- chat entry -->
        <div
          :id="chat.id.Contact ? chat.id.Contact : chat.id.Group" :class="editActive && selectedChat.indexOf(chat.id.Contact ? chat.id.Contact : chat.id.Group) >= 0
            ? 'selected col-12 chat-container'
            : 'col-12 chat-container '
          " data-long-press-delay="500" @click="enterChat(chat)"
          @long-press="editChat(chat.id.Contact ? chat.id.Contact : chat.id.Group)"
        >
          <div class="row chat-entry">
            <div class="avatar col-2">
              <div v-if="!isGroupCheck(chat)" class="badge-name">
                <!-- <img
                  class="avatar-img"
                  :src="'http://localhost:9080/avatars?session=' + chat.id"
                  alt="Avatar image"
                  @error="onImageError($event)"
                /> -->
                {{ `${chat.title[0] ? chat.title[0] : "?"}${chat.title[1] ? chat.title[1] : ""}` }}
              </div>
              <div v-else class="badge-name">
                <font-awesome-icon icon="user-friends" />
              </div>
            </div>
            <div class="meta col-10">
              <div class="row">
                <div class="col-9">
                  <div class="name">
                    <font-awesome-icon v-if="chat.muted" class="mute" icon="volume-mute" />
                    <font-awesome-icon v-if="chat.is_group" class="is_group me-1" icon="user-friends" />
                    <div
                      v-if="chat.is_group &&
                        chat.title === ''
                      " v-translate class="title"
                    >
                      Unknown group
                    </div>
                    <div v-else class="title">
                      {{ chat.title }}
                    </div>
                  </div>
                </div>
                <div v-if="!editActive" class="col-3 date-c">
                  <p
                    v-if="chat.last_message !== '' &&
                      chat.last_message_timestamp !== 0
                    " class="time"
                  >
                    {{ humanifyDate(chat.last_message_timestamp) }}
                  </p>
                </div>
              </div>
              <div class="row">
                <div class="col-9">
                  <p v-if="chat.last_message !== ''" class="preview">
                    {{ chat.last_message }}
                  </p>
                </div>
                <div class="col-3 unread-counter">
                  <div v-if="Number(chat.unread_messages_count) > 0" class="counter badge badge-primary">
                    {{ chat.unread_messages_count }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="chatList.length === 0" v-translate class="no-entries">
        No chats available
      </div>
      <!-- {{chats}} -->
      <router-link :to="'/contacts/'" class="btn start-chat">
        <font-awesome-icon icon="pencil-alt" />
      </router-link>
    </div>
  </component>
</template>

<script>
import moment from "moment";
import { mapState } from "vuex";
import { router } from "@/router/router";

export default {
  name: "ChatList",

  data() {
    return {
      editActive: false,
      editWasActive: false,
      selectedChat: [],
    };
  },
  computed: {
    ...mapState(["chatList", "lastMessages", "sessionNames"]),
    chats() {
      return this.chatList;
    },
  },
  created() {},
  mounted() {
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
    this.$language.current = navigator.language || navigator.userLanguage;
    this.$store.dispatch("leaveChat");
    this.$store.dispatch("clearMessageList");
    this.$store.dispatch("clearFilterContacts");
    this.$store.dispatch("getConfig");
    if (this.$store.state.contacts.length === 0)
    this.$store.dispatch("getContacts");
    this.$store.dispatch("getChatList");
  },
  methods: {
    humanifyDate(inputDate) {
      moment.locale(this.$language.current);
      const date = new moment(inputDate);
      const min = moment().diff(date, "minutes");
      if (min < 60) {
        if (min === 0) return "now";
        return moment().diff(date, "minutes") + " min";
      }
      const hours = moment().diff(date, "hours");
      if (hours < 24) return hours + " h";
      return date.format("DD. MMM");
    },
    editChat(e) {
      this.selectedChat.push(e);
      this.editActive = true;
    },
    editDeactivate(e) {
      this.editActive = false;
      e.preventDefault();
      e.stopPropagation();
      this.editWasActive = true;
      this.selectedChat = [];
    },
    delChat(e) {
      this.editActive = false;
      e.preventDefault();
      e.stopPropagation();
      if (this.selectedChat.length > 0) {
        this.selectedChat.forEach((c) => {
          this.$store.dispatch("delChat", c);
        });
      }
      this.editWasActive = true;
      this.selectedChat = [];
    },
    onImageError(event) {
      event.target.style.display = "none";
    },
    isGroupCheck(e) {
      if (e.DirectMessageRecipientID === -1) {
        return true;
      } else {
        return false;
      }
    },
    enterChat(chat) {
      if (!this.editActive) {
        // this.$store.dispatch("setCurrentChat", chat);
        router.push(`/chat/${JSON.stringify(chat.id)}`);
      } else {
        this.selectedChat.push(chat.Tel);
      }
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss">
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

.actions-header .btn {
  color: #fff;
}

.hide-actions {
  padding-right: 40px;
}

.chat-entry {
  padding: 10px 0;
  cursor: pointer;
}

.badge-name {
  background: rgb(14, 123, 210);
  background: linear-gradient(0deg,
      rgba(14, 123, 210, 1) 8%,
      rgba(32, 144, 234, 1) 42%,
      rgba(107, 180, 238, 1) 100%);
  /* padding: 14px; */
  width: 44px;
  height: 44px;
  border-radius: 50%;
  color: #fff;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 14px;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  flex-wrap: wrap;
}

.avatar-img {
  max-width: 100%;
  max-height: 100%;
  height: 100%;
}

.date-c {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.meta {
  text-align: left;
}

.meta p {
  margin: 0px;
}

.meta .name {
  font-weight: bold;
  font-size: 15px;
  display: flex;
  align-items: center;
}

.meta .preview {
  font-size: 13px;
}

a.chat-container {
  color: #000;
}

a:hover.chat-container {
  text-decoration: none;
}

.btn.start-chat {
  position: fixed;
  bottom: 16px;
  right: 10px;
  background-color: #2090ea;
  color: #fff;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.chatList {

  .preview {
    overflow: hidden;
    height: 20px;
  }

  .time {
    font-size: 12px;
  }

  .mute {
    color: #999;
    margin-right: 10px;
  }

  .counter {
    border-radius: 50%;
    background-color: #2090ea;
    display: flex;
    justify-content: center;
    align-items: center;
    margin-left: 10px;
    width: 20px;
    height: 20px;
  }

  .selected {
    background-color: #c5e4f0;
  }

  .unread-counter {
    text-align: right;
    display: flex;
    justify-content: flex-end;
  }

  .title {
    max-height: 2ch;
    overflow: hidden;
  }
}
</style>
