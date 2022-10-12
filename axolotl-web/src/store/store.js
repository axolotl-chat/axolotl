import { createStore } from 'vuex'
import { router } from '../router/router';
import { validateUUID } from '@/helpers/uuidCheck'
import app from "../main";

const socketSendPlugin = store => {
  store.socketSend = function(message) {
    if (store.state.socket.isConnected) {
      app.config.globalProperties.$socket.send(JSON.stringify(message))
      return true
    }
    return false
  }
}

export default createStore({
  plugins: [socketSendPlugin],
  state: {
    chatList: [],
    lastMessages: {},
    sessionNames: {},
    messageList: {},
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
    registrationStatus: null,
    captchaToken: null,
    captchaTokenSent: false,
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
    SET_LASTMESSAGES(state, lastMessages) {
      state.lastMessages = lastMessages;
    },
    SET_SESSIONNAMES(state, sessionNames) {
      const sessionNamesMap = {};
      for (let i = 0; i < sessionNames.length; i++) {
        sessionNamesMap[sessionNames[i].ID] = sessionNames[i];
      }
      state.sessionNames = sessionNamesMap;
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
        for (let i = state.messageList.length - 1; i >= 0; i--) {
          if(state.messageList[i].SentAt < message.SentAt) {
            state.messageList.splice(i+1, 0, message)
            break
          }
        }
      }
      const lastMessage = state.lastMessages[message.SID]
      if(lastMessage !== null && lastMessage.SentAt < message.SentAt) {
        state.lastMessages[message.SID] = message
      }
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
      state.messageList = {};
    },
    SET_PROFILE(state, profile) {
      state.profile = profile;
    },
    CLEAR_PROFILE(state) {
      state.profile = {};
    },
    SET_CONTACTS(state, contacts) {
      if (contacts !== null) {
        state.importingContacts = false;
        state.contacts = contacts.sort((a, b) => a.Name.localeCompare(b.Name));
      }
    },
    SET_CONTACTS_FILTER(state, filter) {
      const lowerFilter = filter.toLowerCase()
      const f = state.contacts.filter(c => c.Name.toLowerCase().includes(lowerFilter));
      state.contactsFiltered = f;
      state.contactsFilterActive = true;
    },
    SET_CONTACTS_FOR_GROUP_FILTER(state, filter) {
      const lowerFilter = filter.toLowerCase()
      let f = state.contacts.filter(
        c => validateUUID(c.UUID) &&
        c.Name.toLowerCase().includes(lowerFilter)
      )
      if (state.currentGroup !== null){
        f = f.filter(c => state.currentGroup.Members.indexOf(c.UUID) === -1);
      }
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
            code: 200,
            msg: message
          }));
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
      const messageData = JSON.parse(message.data);
      if (messageData.Error) {
        this.commit("SET_ERROR", messageData.Error);
      }
      switch (Object.keys(messageData)[0]) {
        case "ChatList": {
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
            this.commit("SET_LASTMESSAGES", {});
          }
          this.commit("SET_SESSIONNAMES", messageData.SessionNames);
          break;
        }
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
        default:
        // console.log("unkown message ", Object.keys(messageData)[0]);
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

        document.cookie = `darkMode=${darkMode};expires=${2**63-1};path=/`;
        state.darkMode = darkMode;
        window.location.replace("/")
      }
    }
  },

  actions: {
    addDevice(state, url) {
      const message = {
        "request": "addDevice",
        url,
      }
      this.socketSend(message)
    },
    delDevice(state, id) {
      const message = {
        "request": "delDevice",
        id,
      }
      this.socketSend(message)
    },
    getDevices() {
      const message = {
        "request": "getDevices",
      }
      this.socketSend(message)
    },
    getChatList() {
      const message = {
        "request": "getChatList",
      }
      this.socketSend(message)
    },
    delChat(id) {
      const message = {
        "request": "delChat",
        id
      }
      this.socketSend(message)
    },
    getMessageList(state, id) {
      this.commit("CLEAR_MESSAGELIST");
      const message = {
        "request": "getMessageList",
        id
      }
      this.socketSend(message)
    },
    getProfile(state, id) {
      this.commit("CLEAR_PROFILE");
      const message = {
        "request": "getProfile",
        id
      }
      this.socketSend(message)
    },
    createRecipient(state, recipient) {
      const message = {
        "request": "createRecipient",
        recipient
      }
      this.socketSend(message)
    },
    createRecipientAndAddToGroup(state, data) {
      const message = {
        "request": "createRecipientAndAddToGroup",
        "recipient": data.id,
        "group": data.group
      }
      this.socketSend(message)
    },
    openChat({ dispatch }, id) {
      const message = {
        "request": "openChat",
        id
      }
      if(this.socketSend(message)) {
        dispatch("getMessageList", id)
      }
    },
    getMoreMessages() {
      if (typeof this.state.messageList !== "undefined"
        && this.state.messageList !== null
        && this.state.messageList.length > 19 && this.state.messageList[0].ID > 1) {
        const message = {
          "request": "getMoreMessages",
          "lastId": String(this.state.messageList[0].ID)
        }
        this.socketSend(message);
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
      const message = {
        "request": "leaveChat",
      }
      this.socketSend(message)
    },
    createChat(state, uuid) {
      const message = {
        "request": "createChat",
        uuid
      }
      if(this.socketSend(message)) {
        this.commit("CREATE_CHAT", uuid);
      }
    },
    sendMessage(state, messageContainer) {
      const message = {
        "request": "sendMessage",
        "to": messageContainer.to,
        "message": messageContainer.message
      }
      this.socketSend(message)
    },
    toggleNotifications() {
      const message = {
        "request": "toggleNotifications",
        "chat": this.state.currentChat.ID
      }
      this.socketSend(message)
    },
    resetEncryption() {
      const message = {
        "request": "resetEncryption",
        "chat": this.state.currentChat.ID
      }
      this.socketSend(message)
    },
    verifyIdentity() {
      const message = {
        "request": "verifyIdentity",
        "chat": this.state.currentChat.ID
      }
      this.socketSend(message)
    },
    getContacts(state) {
      state.importingContacts = false;
      const message = {
        "request": "getContacts",
      }
      this.socketSend(message)
    },
    addContact(state, contact) {
      state.rateLimitError = null;
      if (contact.name !== "" && contact.phone !== "") {
        if (this.state.currentChat !== null
          && this.state.currentChat.Tel === contact.phone) {
          this.commit("SET_CURRENT_CHAT_NAME", contact.name);
        }
        const message = {
          "request": "addContact",
          "name": contact.name,
          "phone": contact.phone,
          "uuid": contact.uuid
        }
        this.socketSend(message);
      }
    },
    updateProfileName(state, data) {
      const message = {
        "request": "updateProfileName",
        "name": data.name,
        "id": data.id
      }
      this.socketSend(message)
    },
    createChatForRecipient(state, data) {
      const message = {
        "request": "createChatForRecipient",
        "id": data.id,
      }
      this.socketSend(message)
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
      const message = {
        "request": "uploadVcf",
        vcf
      }
      this.socketSend(message)
    },
    uploadAttachment(state, attachment) {
      const message = {
        "request": "uploadAttachment",
        "attachment": attachment.attachment,
        "to": attachment.to,
      }
      this.socketSend(message)
    },
    refreshContacts(state, url) {
      state.importingContacts = true;
      const message = {
        "request": "refreshContacts",
        url
      }
      this.socketSend(message)
    },
    delContact(state, id) {
      state.rateLimitError = null;
      const message = {
        "request": "delContact",
        id,
      }
      this.socketSend(message)
    },
    deleteSelfDestructingMessage(state, m) {
      const message = {
        "request": "delMessage",
        "id": m.ID
      }
      this.socketSend(message)
    },
    editContact(state, data) {
      state.rateLimitError = null;
      if (this.state.currentChat !== null
        && this.state.currentChat.Tel === data.contact.Tel) {
        this.commit("SET_CURRENT_CHAT_NAME", data.contact.Name);
      }
      const message = {
        "request": "editContact",
        "phone": data.contact.Tel,
        "name": data.contact.Name,
        "uuid": data.contact.UUID,
        "id": data.id
      }
      this.socketSend(message);
    },
    // registration functions
    requestCode(state, tel) {
      this.state.verificationError = null
      const message = {
        "request": "requestCode",
        tel,
      }
      this.socketSend(message)
    },
    sendCode(state, code) {
      const message = {
        "request": "sendCode",
        code,
      }
      this.socketSend(message)
    },
    setUsername(state, username) {
      const message = {
        "request": "sendUsername",
        username,
      }
      if(this.socketSend(message)) {
        this.commit("SET_REGISTRATION_STATUS", "");
        router.push("/")
      }
    },
    sendPin(state, pin) {
      const message = {
        "request": "sendPin",
        pin,
      }
      this.socketSend(message)
    },
    sendPassword(state, pw) {
      const message = {
        "request": "sendPassword",
        pw,
      }
      this.socketSend(message)
    },
    setPassword(state, password) {
      const message = {
        "request": "setPassword",
        "pw": password.pw,
        "currentPw": password.cPw
      }
      if(this.socketSend(message)) {
        router.push("/chatList")
      }
    },
    getRegistrationStatus() {
      const message = {
        "request": "getRegistrationStatus",
      }
      this.socketSend(message)
    },
    unregister() {
      const message = {
        "request": "unregister",
      }
      this.socketSend(message)
    },
    getConfig() {
      const message = {
        "request": "getConfig",
      }
      this.socketSend(message)
    },
    createNewGroup(state, data) {
      const message = {
        "request": "createGroup",
        "name": data.name,
        "members": data.members,
      }
      this.socketSend(message)
    },
    updateGroup(state, data) {
      const message = {
        "request": "updateGroup",
        "name": data.name,
        "id": data.id,
        "members": data.members,
      }
      this.socketSend(message)
    },
    joinGroup(state, id) {
      const message = {
        "request": "joinGroup",
        id,
      }
      this.socketSend(message)
    },
    sendAttachment(state, data) {
      const message = {
        "request": "sendAttachment",
        "type": data.type,
        "path": data.path,
        "to": data.to,
        "message": data.message,
      }
      this.socketSend(message)
    },
    sendVoiceNote(state, voiceNote) {
      const message = {
        "request": "sendVoiceNote",
        "voiceNote": voiceNote.note,
        "to": voiceNote.to,
      }
      this.socketSend(message)
    },
    setDarkMode(state, darkMode) {
      const message = {
        "request": "setDarkMode",
        darkMode,
      }
      if(this.socketSend(message)) {
        this.state.DarkMode = darkMode;
      }
    },
    sendCaptchaToken() {
      const message = {
        "request": "sendCaptchaToken",
        "token": this.state.captchaToken,
      }
      if(this.socketSend(message)) {
        this.commit("SET_CAPTCHA_TOKEN_SENT");
      }
    },
    setLogLevel(state, level) {
      const message = {
        "request": "setLogLevel",
        level
      }
      this.socketSend(message);
    },
    setCaptchaToken(state, token) {
      this.commit("SET_CAPTCHA_TOKEN", token);
    }
  }
});
