import { createStore } from 'vuex'
import { router } from '../router/router';
import { validateUUID } from '@/helpers/uuidCheck'
import app from "../main";

function socketSend(message) {
  app.config.globalProperties.$socket.send(JSON.stringify(message))
}

export default createStore({
  state: {
    chatList: [],
    lastMessages: {},
    sessionNames: {},
    messageList: [],
    profile: {},
    request: '',
    contacts: [],
    contactsFiltered: [],
    contactsFilterActive: false,
    devices: [],
    gui: null,
    darkMode: false,
    error: null,
    errorConnection: null,
    fingerprint: {
      numbers: null,
      qrCode: null,
    },
    config: {},
    loginError: null,
    rateLimitError: null,
    newGroupName: null,
    currentChat: null,
    currentGroup: null,
    currentContact: null,
    newGroupMembers: [],
    importingContacts: false,
    verificationInProgress: false,
    verificationError: null,
    requestPin: false,
    registrationStatus: "registered",
    captchaToken: null,
    captchaTokenSent: false,
    deviceLinkCode: null,
    socket: {
      isConnected: false,
      message: '',
      reconnectError: false,
      heartBeatInterval: 50000,
      heartBeatTimer: 0
    }
  },

  getters: {
    // Here we will create a getter
    getMessages: state => {
      return state.messageList;
    }
  },

  mutations: {
    SET_ERROR(state, error) {
      if (error === "") {
        state.loginError = null;
        // state.rateLimitError = null;
        state.errorConnection = null;
      }
      else if (error === "wrong password") {
        state.loginError = error;
      } else if (error.includes("Rate")) {
        state.rateLimitError = `${error}. Try again later!`;
      } else if (error.includes("no such host") || error.includes("timeout")) {
        state.errorConnection = error;
      } else if (error.includes("Your registration is faulty")) {
        state.error = error;
      } else if (error.includes(400)) {
        //when pin is missing
        if (state.verificationInProgress)
          state.verificationError = 400
      } else if (error.includes(404)) {
        //when code was wrong
        if (state.verificationInProgress)
          state.verificationError = 404
      } else if (error.includes("RegistrationLockFailure")) {
        state.verificationError = "RegistrationLockFailure"
      }
    },
    SET_CHATLIST(state, chatList) {
      state.chatList = chatList;
    },
    SET_DEVICE_LINK_CODE(state, code) {
      state.deviceLinkCode = code;
    },
    SET_LASTMESSAGES(state, lastMessages) {
      state.lastMessages = lastMessages;
    },
    SET_SESSIONNAMES(state, sessionNames) {
      const sorted = {};
      for (let i = 0; i < sessionNames.length; i++) {
        sorted[sessionNames[i].ID] = sessionNames[i];
      }
      state.sessionNames = sorted;
    },
    SET_CURRENT_CHAT(state, chat) {
      state.currentChat = chat;
    },
    SET_CURRENT_CHAT_NAME(state, chatName) {
      state.currentChat.Name = chatName;
    },
    OPEN_CHAT(state, data) {
      state.currentChat = data.CurrentChat;
      state.currentGroup = data.Group;
      state.currentContact = data.Contact;
    },
    UPDATE_CURRENT_CHAT(state, data) {
      state.currentChat = data.CurrentChat;
      if (typeof state.messageList === "undefined") {
        state.messageList = []
      }
      const prepare = state.messageList.map(e => e.ID)
      if (data.CurrentChat.Messages !== null) {
        data.CurrentChat.Messages.forEach(m => {
          state.messageList[prepare.indexOf(m.ID)] = m;
        });
      }
    },
    SET_CONFIG(state, config) {
      state.config = config;
    },
    SEND_MESSAGE() {

    },
    CREATE_CHAT(state) {
      state.currentChat = null;
    },
    SET_DEVICELIST(state, devices) {
      state.devices = devices
    },
    SET_FINGERPRINT(state, data) {
      state.fingerprint = {
        numbers: data.FingerprintNumbers,
        qrCode: data.FingerprintQRCode,
      }
    },
    SET_REGISTRATION_STATUS(state, status) {
      state.registrationStatus = status;
    },
    SET_REQUEST(state, request) {
      const type = request.Type;
      state.request = request;
      if (type === "getPhoneNumber") {
        this.commit("SET_REGISTRATION_STATUS", "phoneNumber");
      } else if (type === "getVerificationCode") {
        this.commit("SET_REGISTRATION_STATUS", "verificationCode");
        state.verificationInProgress = true;
        router.push("/verify")
      } else if (type === "getPin") {
        this.commit("SET_REGISTRATION_STATUS", "pin");
        router.push("/pin")
        state.requestPin = true;
      } else if (type === "getCaptchaToken") {
        window.location = "https://signalcaptchas.org/registration/generate.html";
      } else if (type === "getEncryptionPw") {
        this.commit("SET_REGISTRATION_STATUS", "password");
        router.push("/password")
      } else if (type === "registrationDone") {
        this.commit("SET_REGISTRATION_STATUS", "registered");
        router.push("/")
      } else if (type === "requestEnterChat") {
        router.push(`/chat/${request.Chat}`)
        this.dispatch("getChatList")
      } else if (type === "config") {
        this.commit("SET_CONFIG", request)
      } else if (type === "getUsername") {
        this.commit("SET_REGISTRATION_STATUS", "getUsername");
        router.push("/setUsername")
      } else if (type === "Error") {
        this.commit("SET_ERROR", request.Error)
      }
      // this.dispatch("requestCode", "+123456")
    },
    SET_MESSAGELIST(state, messageList) {
      state.messageList = messageList.reverse();
    },
    SET_MORE_MESSAGELIST(state, messageList) {
      if (messageList.Messages !== null) {
        messageList.Messages.reverse()
        state.messageList = messageList.Messages.concat(state.messageList);
      }
    },
    SET_MESSAGE_RECEIVED(state, message) {
      if (state.currentChat !== null && state.currentChat.ID === message.SID) {
        const tmpList = state.messageList;
        tmpList.push(message);
        tmpList.sort(function (a, b) {
          return a.ID - b.ID
        })
        state.messageList = tmpList;
      }
      state.chatList.forEach((chat, i) => {
        if (chat.ID === message.SID) {
          state.chatList[i].Messages = [message]
        }
      })
    },
    SET_MESSAGE_UPDATE(state, message) {
      if (state.currentChat.ID === message.SID) {
        const index = state.messageList.findIndex(m => {
          return m.ID === message.ID;
        });
        // check if message exists
        if (index !== -1) {
          const tmpList = JSON.parse(JSON.stringify(state.messageList));
          tmpList[index] = message;
          tmpList.sort(function (a, b) {
            return a.ID - b.ID
          })
          // mark all as read if it's a is read update
          if (message.IsRead) {
            tmpList.forEach((m, i) => {
              if (m.Outgoing && !m.IsRead) {
                m.IsRead = true;
                tmpList[i] = m;
              }
            })
          }
          state.messageList = tmpList;
        } else {
          // add message to message list
          state.messageList.unshift(message)
        }
      }
    },
    CLEAR_MESSAGELIST(state) {
      state.messageList = [];
    },
    SET_PROFILE(state, profile) {
      state.profile = profile;
    },
    CLEAR_PROFILE(state) {
      state.profile = {};
    },
    SET_CONTACTS(state, contacts) {
      if (contacts !== null) {
        contacts = contacts.sort((a, b) => a.name.localeCompare(b.name));
        state.importingContacts = false;
        state.contacts = contacts;
      }
    },
    SET_CONTACTS_FILTER(state, filter) {
      filter = filter.toLowerCase()
      const f = state.contacts.filter(c => c.Name.toLowerCase().includes(filter));
      state.contactsFiltered = f;
      state.contactsFilterActive = true;
    },
    SET_CONTACTS_FOR_GROUP_FILTER(state, filter) {
      filter = filter.toLowerCase()
      let f = state.contacts.filter(c => {
        if (!validateUUID(c.UUID)) return false
        if (c.Name.toLowerCase().includes(filter))
          return true;
      });
      if (state.currentGroup !== null)
        f = f.filter(c => state.currentGroup.Members.indexOf(c.UUID) === -1);
      state.contactsFiltered = f;
      state.contactsFilterActive = true;
    },
    SET_CLEAR_CONTACTS_FILTER(state) {
      state.contactsFiltered = [];
      state.contactsFilterActive = false;
    },
    LEAVE_CHAT(state) {
      state.currentGroup = null;
      state.currentContact = null;
      state.currentChat = null;
      this.commit("CLEAR_MESSAGELIST");
    },
    SET_CAPTCHA_TOKEN(state, token) {
      state.captchaToken = token;
    },
    SET_CAPTCHA_TOKEN_SENT(state) {
      state.captchaTokenSent = true;
    },
    SOCKET_ONOPEN(state, event) {
      app.config.globalProperties.$socket = event.currentTarget;
      state.socket.isConnected = true;
      state.socket.heartBeatTimer = setInterval(() => {
        const message = "ping";
        state.socket.isConnected &&
          app.config.globalProperties.$socket.send(JSON.stringify({
            request: message,
            code: 200,
          }));
        this.dispatch("getChatList")
      }, state.socket.heartBeatInterval);
    },
    SOCKET_ONCLOSE(state) {
      state.socket.isConnected = false;
      clearInterval(state.socket.heartBeatTimer);
      state.socket.heartBeatTimer = 0;
    },
    SOCKET_ONERROR() {
      // console.error(state, event)
    },
    SOCKET_RECONNECT() {
      // do nothing
    },
    SOCKET_RECONNECT_ERROR(state) {
      state.socket.reconnectError = true;
    },
    // default handler called for all methods
    SOCKET_ONMESSAGE(state, message) {
      state.socket.message = message;
      console.log(message)
      const messageData = JSON.parse(message.data);
      if (messageData.Error) {
        this.commit("SET_ERROR", messageData.Error);
      }
      switch (Object.keys(messageData)[0]) {
        case "ChatList":
          const chats = messageData.ChatList
          if (messageData.LastMessages) {
            const lastMessages = {};
            for (let i = 0; i < messageData.LastMessages.length; i++) {
              lastMessages[messageData.LastMessages[i].SID] = messageData.LastMessages[i];
            }
            chats.sort((a, b) => {
              if (typeof lastMessages[a.ID] === "undefined" || typeof lastMessages[b.ID] === "undefined") {
                return 0;
              }
              if (lastMessages[a.ID].SentAt < lastMessages[b.ID].SentAt) {
                return 1;
              }
              if (lastMessages[a.ID].SentAt > lastMessages[b.ID].SentAt) {
                return -1;
              }
              return 0;
            }
            );
            this.commit("SET_CHATLIST", chats);
            this.commit("SET_LASTMESSAGES", lastMessages);
          } else {
            this.commit("SET_CHATLIST", chats);
            this.commit("SET_LASTMESSAGES", messageData.LastMessages);
          }
          this.commit("SET_SESSIONNAMES", messageData.SessionNames);
          break;
        case "MessageList":
          this.commit("SET_MESSAGELIST", messageData.MessageList);
          break;
        case "ContactList":
          this.commit("SET_CONTACTS", messageData.ContactList);
          break;
        case "MoreMessageList":
          this.commit("SET_MORE_MESSAGELIST", messageData.MoreMessageList);
          break;
        case "DeviceList":
          this.commit("SET_DEVICELIST", messageData.DeviceList);
          break;
        case "MessageReceived":
          this.commit("SET_MESSAGE_RECEIVED", messageData.MessageReceived);
          break;
        case "UpdateMessage":
          this.commit("SET_MESSAGE_UPDATE", messageData.UpdateMessage);
          break;
        case "Gui":
          this.commit("SET_CONFIG_GUI", messageData.Gui);
          break;
        case "DarkMode":
          this.commit("SET_CONFIG_DARK_MODE", messageData.DarkMode);
          break;
        case "CurrentChat":
          this.commit("SET_CURRENT_CHAT", messageData.CurrentChat);
          break;
        case "Type":
          this.commit("SET_REQUEST", messageData);
          break;
        case "FingerprintNumbers":
          this.commit("SET_FINGERPRINT", messageData);
          break;
        case "Error":
          this.commit("SET_ERROR", messageData.Errorx);
          break;
        case "OpenChat":
          this.commit("OPEN_CHAT", messageData.OpenChat);
          break;
        case "UpdateCurrentChat":
          this.commit("UPDATE_CURRENT_CHAT", messageData.UpdateCurrentChat);
          break;
        case "ProfileMessage":
          this.commit("SET_PROFILE", messageData.ProfileMessage);
          break;
        case "response_type":
          if (messageData.response_type === "contact_list") {
            this.commit("SET_CONTACTS", JSON.parse(messageData.data));
          } else if (messageData.response_type === "chat_list") {
            this.commit("SET_CHATLIST", JSON.parse(messageData.data));

          } else if (messageData.response_type === "message_list") {
            this.commit("SET_MESSAGELIST", JSON.parse(messageData.data));
          } else if (messageData.response_type === "qr_code") {
            this.commit("SET_DEVICE_LINK_CODE", messageData.data);
            router.push("/qr");
          }
          break;

        default:
          console.log("unkown message ", messageData, Object.keys(messageData)[0]);
      }
      this.commit("SET_SOCKET_MESSAGE_DATA", message.data)

    },
    SET_SOCKET_MESSAGE_DATA(state, data) {
      state.socket.message = data
    },
    SET_CONFIG_GUI(state, gui) {
      state.gui = gui;
    },
    SET_CONFIG_DARK_MODE(state, darkMode) {
      if (window.getCookie("darkMode") !== String(darkMode)) {
        const d = new Date();
        d.setTime(d.getTime() + (300 * 24 * 60 * 60 * 1000));
        const expires = `expires=${d.toUTCString()}`;
        document.cookie = `darkMode=${darkMode};${expires};path=/`;
        state.darkMode = darkMode;
        window.location.replace("/")
      }
    }
  },

  actions: {
    addDevice(state, url) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "addDevice",
          "data": url,
        }
        socketSend(message);
      }
    },
    delDevice(state, id) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "delDevice",
          "data": id,
        }
        socketSend(message);
      }
    },
    getDevices() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "getDevices",
        }
        socketSend(message);
      }
    },
    getChatList() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "getChatList",
        }
        socketSend(message);
      }
    },
    delChat(id) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "delChat",
          "data": this.state.currentChat.ID
        }
        socketSend(message);
      }
    },
    getMessageList(state, chatId) {

      this.commit("CLEAR_MESSAGELIST");
      if (this.state.socket.isConnected) {
        const data = {
          "id": chatId
        };
        if (chatId) {
          console.log(chatId)
          const message = {
            "request": "getMessageList",
            "data": JSON.stringify(data)
          }
          socketSend(message);
        }

      }
    },
    getProfile(state, id) {
      this.commit("CLEAR_PROFILE");
      if (this.state.socket.isConnected) {
        const message = {
          "request": "getProfile",
          id
        }
        socketSend(message);
      }
    },
    createRecipient(state, recipient) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "createRecipient",
          recipient
        }
        socketSend(message);
      }
    },
    createRecipientAndAddToGroup(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "createRecipientAndAddToGroup",
          "data":{
            "recipient": data.id,
            "group": data.group
          }
        }
        socketSend(message);
      }
    },
    openChat({ dispatch }, chatId) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "openChat",
          "data": chatId
        }
        socketSend(message);
        dispatch("getMessageList", chatId);
      }
    },
    getMoreMessages() {
      if (this.state.socket.isConnected && typeof this.state.messageList !== "undefined"
        && this.state.messageList !== null
        && this.state.messageList.length > 0 && this.state.messageList[0].SentAt > 0) {
        const firstMessage = this.state.messageList[0]
        const message = {
          "request": "getMoreMessages",
          "data": firstMessage.SentAt
        }
        socketSend(message);
      }
    },
    clearMessageList() {
      this.commit("CLEAR_MESSAGELIST");
    },
    setCurrentChat(state, chat) {
      this.commit("SET_CURRENT_CHAT", chat);
    },
    leaveChat() {
      this.commit("LEAVE_CHAT");
      if (this.state.socket.isConnected) {
        const message = {
          "request": "leaveChat",
        }
        socketSend(message);
      }
    },
    createChat(state, uuid) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "createChat",
          "data": uuid
        }
        socketSend(message);
      }
      this.commit("CREATE_CHAT", uuid);
    },
    sendMessage(state, messageContainer) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendMessage",
          "data":{
            "to": messageContainer.to,
            "message": messageContainer.message
          }
        }
        socketSend(message);
      }
    },
    toggleNotifications() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "toggleNotifications",
          "data": this.state.currentChat.ID
        }
        socketSend(message);
      }
    },
    resetEncryption() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "resetEncryption",
          "data": this.state.currentChat.ID
        }
        socketSend(message);
      }
    },
    verifyIdentity() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "verifyIdentity",
          "data": this.state.currentChat.ID
        }
        socketSend(message);
      }
    },
    getContacts(state) {
      if (this.state.socket.isConnected) {
        state.importingContacts = false;
        const message = {
          "request": "getContacts",
          "data": ""
        }
        socketSend(message);
      }
    },
    addContact(state, contact) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected
        && contact.name !== "" && contact.phone !== "") {
        if (this.state.currentChat !== null
          && this.state.currentChat.Tel === contact.phone) {
          this.commit("SET_CURRENT_CHAT_NAME", contact.name);
        }
        const message = {
          "request": "addContact",
          "data": {
            "name": contact.name,
            "phone": contact.phone,
            "uuid": contact.uuid
          }
        }
        socketSend(message);
      }
    },
    updateProfileName(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "updateProfileName",
          "data": {
            "name": data.name,
            "id": data.id
          }
        }
        socketSend(message);
      }
    },
    createChatForRecipient(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "createChatForRecipient",
          "data": data.id,
        }
        socketSend(message);
      }
    },
    filterContacts(state, filter) {
      this.commit("SET_CONTACTS_FILTER", filter);
    },
    filterContactsForGroup(state, filter) {
      this.commit("SET_CONTACTS_FOR_GROUP_FILTER", filter);
    },
    clearFilterContacts() {
      this.commit("SET_CLEAR_CONTACTS_FILTER");
    },
    uploadVcf(state, vcf) {
      state.rateLimitError = null;
      state.importingContacts = true;
      if (this.state.socket.isConnected) {
        const message = {
          "request": "uploadVcf",
          "data": vcf
        }
        socketSend(message);
      }
    },
    uploadAttachment(state, attachment) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "uploadAttachment",
          "data": {
            "attachment": attachment.attachment,
            "to": attachment.to,
          }
        }
        socketSend(message);
      }
    },
    refreshContacts(state, chUrl) {
      state.importingContacts = true;
      if (this.state.socket.isConnected) {
        const message = {
          "request": "refreshContacts",
          "data": chUrl
        }
        socketSend(message);
      }
    },
    delContact(state, id) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected) {
        const message = {
          "request": "delContact",
          "data": id,
        }
        socketSend(message);
      }
    },
    deleteSelfDestructingMessage(state, m) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "delMessage",
          "data": m.ID
        }
        socketSend(message);
      }
    },
    editContact(state, data) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected) {
        if (this.state.currentChat !== null
          && this.state.currentChat.Tel === data.contact.Tel) {
          this.commit("SET_CURRENT_CHAT_NAME", data.contact.Name);
        }
        const message = {
          "request": "editContact",
          "data": {
            "phone": data.contact.Tel,
            "name": data.contact.Name,
            "uuid": data.contact.UUID,
            "id": data.id
          }
        }
        socketSend(message);
      }
    },
    // registration functions
    requestCode(state, tel) {
      this.state.verificationError = null
      if (this.state.socket.isConnected) {
        const message = {
          "request": "requestCode",
          "data": tel,
        }
        socketSend(message);
      }
    },
    sendCode(state, code) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendCode",
          "data": code,
        }
        socketSend(message);
      }
    },
    setUsername(state, username) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendUsername",
          "data": username,
        }
        socketSend(message);
      }
    },
    sendPin(state, pin) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendPin",
          "data": pin,
        }
        socketSend(message);
      }
    },
    sendPassword(state, password) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendPassword",
          "data": password,
        }
        socketSend(message);
      }
    },
    setPassword(state, password) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "setPassword",
          "data": {
            "pw": password.pw,
            "currentPw": password.cPw
          }
        }
        socketSend(message);
        router.push("/chatList")
      }
    },
    getRegistrationStatus() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "getRegistrationStatus",
        }
        socketSend(message);
      }
    },
    unregister() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "unregister",
        }
        socketSend(message);

      }
    },
    getConfig() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "getConfig",
        }
        socketSend(message);
      }
    },
    createNewGroup(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "createGroup",
          "data": {
            "name": data.name,
            "members": data.members,
          }
        }
        socketSend(message);

      }
    },
    updateGroup(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "updateGroup",
          "data": {
            "name": data.name,
            "id": data.id,
            "members": data.members,
          }
        }
        socketSend(message);
      }
    },
    joinGroup(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "joinGroup",
          "data": data,
        }
        socketSend(message);
      }
    },
    sendAttachment(state, data) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendAttachment",
          "data": {
            "type": data.type,
            "path": data.path,
            "to": data.to,
            "message": data.message,
          }
        }
        socketSend(message);

      }
    },
    sendVoiceNote(state, voiceNote) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendVoiceNote",
          "data": {
            "voiceNote": voiceNote.note,
            "to": voiceNote.to,
          }
        }
        socketSend(message);
      }
    },
    setDarkMode(state, darkMode) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "setDarkMode",
          "data": darkMode,
        }
        socketSend(message);
        this.state.DarkMode = darkMode;
      }
    },
    sendCaptchaToken() {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "sendCaptchaToken",
          "data": this.state.captchaToken,
        }
        socketSend(message);
        this.commit("SET_CAPTCHA_TOKEN_SENT");

      }
    },
    setLogLevel(state, level) {
      if (this.state.socket.isConnected) {
        const message = {
          "request": "setLogLevel",
          level
        }
        socketSend(message);
      }
    },
    setCaptchaToken(state, token) {
      this.commit("SET_CAPTCHA_TOKEN", token);
    }
  }
});
