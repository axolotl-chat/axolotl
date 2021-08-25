import { createStore } from 'vuex'
import { router } from '../router/router';
import { validateUUID } from '@/helpers/uuidCheck'
import app from "../main";

export default createStore({
  state: {
    chatList: [],
    messageList: {},
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
      return state.messageList.Messages;
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
        state.rateLimitError = error + ". Try again later!";
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
      if (typeof state.messageList.Messages === "undefined") {
        state.messageList.Messages = []
      }
      var prepare = state.messageList.Messages.map(function (e) { return e.ID; })
      if (data.CurrentChat.Messages !== null) {
        data.CurrentChat.Messages.forEach(m => {
          state.messageList.Messages[prepare.indexOf(m.ID)] = m;
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
      var type = request["Type"]
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
        router.push("/chat/" + request["Chat"])
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
      state.messageList = messageList;
    },
    SET_MORE_MESSAGELIST(state, messageList) {
      if (messageList.Messages !== null) {
        state.messageList.Messages = state.messageList.Messages.concat(messageList.Messages);
      }
    },
    SET_MESSAGE_RECIEVED(state, message) {
      if (state.messageList.ID === message.SID) {
        var tmpList = state.messageList.Messages;
        tmpList.push(message);
        tmpList.sort(function (a, b) {
          return b.ID - a.ID
        })
        state.messageList.Messages = tmpList;
      }
      state.chatList.forEach((chat, i) => {
        if (chat.Tel === message.ChatID) {
          state.chatList[i].Messages = [message]
        }
      })
    },
    SET_MESSAGE_UPDATE(state, message) {
      if (state.messageList.ID === message.SID) {
        var index = state.messageList.Messages.findIndex(m => {
          return m.ID === message.ID;
        });
        // check if message exists
        if (index !== -1) {
          var tmpList = JSON.parse(JSON.stringify(state.messageList.Messages));
          tmpList[index] = message;
          tmpList.sort(function (a, b) {
            return b.ID - a.ID
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
          state.messageList.Messages = tmpList;
        } else {
          // add message to message list
          state.messageList.Messages.unshift(message)
        }
      }
    },
    CLEAR_MESSAGELIST(state) {
      state.messageList = {};
    },
    SET_CONTACTS(state, contacts) {
      if (contacts !== null) {
        contacts = contacts.sort((a, b) => a.Name.localeCompare(b.Name));
        state.importingContacts = false;
        state.contacts = contacts;
      }
    },
    SET_CONTACTS_FILTER(state, filter) {
      filter = filter.toLowerCase()
      var f = state.contacts.filter(c => c.Name.toLowerCase().includes(filter));
      state.contactsFiltered = f;
      state.contactsFilterActive = true;
    },
    SET_CONTACTS_FOR_GROUP_FILTER(state, filter) {
      filter = filter.toLowerCase()
      var f = state.contacts.filter(c => {
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
    SOCKET_RECONNECT_ERROR(state) {
      state.socket.reconnectError = true;
    },
    // default handler called for all methods
    SOCKET_ONMESSAGE(state, message) {
      state.socket.message = message;
      if (message.data !== "Hi Client!") {
        var messageData = JSON.parse(message.data);
        if (typeof messageData.Error !== "undefined") {
          this.commit("SET_ERROR", messageData["Error"]);
        }
        if (Object.keys(messageData)[0] === "ChatList") {
          this.commit("SET_CHATLIST", messageData["ChatList"]);
        }
        else if (Object.keys(messageData)[0] === "MessageList") {
          this.commit("SET_MESSAGELIST", messageData["MessageList"]);
        }

        else if (Object.keys(messageData)[0] === "ContactList") {
          this.commit("SET_CONTACTS", messageData["ContactList"]);
        }
        else if (Object.keys(messageData)[0] === "MoreMessageList") {
          this.commit("SET_MORE_MESSAGELIST", messageData["MoreMessageList"]);
        }
        else if (Object.keys(messageData)[0] === "DeviceList") {
          this.commit("SET_DEVICELIST", messageData["DeviceList"]);
        }
        else if (Object.keys(messageData)[0] === "MessageRecieved") {
          this.commit("SET_MESSAGE_RECIEVED", messageData["MessageRecieved"]);
        }
        else if (Object.keys(messageData)[0] === "UpdateMessage") {
          this.commit("SET_MESSAGE_UPDATE", messageData["UpdateMessage"]);
        }
        else if (Object.keys(messageData)[0] === "Gui") {
          this.commit("SET_CONFIG_GUI", messageData["Gui"]);
        }
        else if (Object.keys(messageData)[0] === "DarkMode") {
          this.commit("SET_CONFIG_DARK_MODE", messageData["DarkMode"]);
        }
        else if (Object.keys(messageData)[0] === "CurrentChat") {
          this.commit("SET_CURRENT_CHAT", messageData["CurrentChat"]);
        }
        else if (Object.keys(messageData)[0] === "Type") {
          this.commit("SET_REQUEST", messageData);
        }
        else if (Object.keys(messageData)[0] === "FingerprintNumbers") {
          this.commit("SET_FINGERPRINT", messageData);
        }
        else if (Object.keys(messageData)[0] === "Error") {
          this.commit("SET_ERROR", messageData.Errorx);
        }
        else if (Object.keys(messageData)[0] === "OpenChat") {
          this.commit("OPEN_CHAT", messageData["OpenChat"]);
        }
        else if (Object.keys(messageData)[0] === "UpdateCurrentChat") {
          this.commit("UPDATE_CURRENT_CHAT", messageData["UpdateCurrentChat"]);
        }
        else {
          // console.log("unkown message ", Object.keys(messageData)[0]);
        }
        this.commit("SET_SOCKET_MESSAGE_DATA", message.data)

      }
    },
    SET_SOCKET_MESSAGE_DATA(state, data) {
      state.socket.message = data
    },
    SET_CONFIG_GUI(state, gui) {
      state.gui = gui;
    },
    SET_CONFIG_DARK_MODE(state, darkMode) {
      if (window.getCookie("darkMode") !== String(darkMode)) {
        var d = new Date();
        d.setTime(d.getTime() + (300 * 24 * 60 * 60 * 1000));
        var expires = "expires=" + d.toUTCString();
        document.cookie = "darkMode" + "=" + darkMode + ";" + expires + ";path=/";
        state.darkMode = darkMode;
        window.location.replace("/")
      }
    }
  },

  actions: {
    addDevice(state, url) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "addDevice",
          "url": url,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    delDevice(state, id) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delDevice",
          "id": id,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    getDevices() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getDevices",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    getChatList() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getChatList",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    delChat(id) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delChat",
          "id": id,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    getMessageList(state, chatId) {
      this.commit("CLEAR_MESSAGELIST");
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getMessageList",
          "id": chatId
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    openChat(state, chatId) {
      this.commit("CLEAR_MESSAGELIST");
      if (this.state.socket.isConnected) {
        var message = {
          "request": "openChat",
          "id": chatId
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    getMoreMessages() {
      if (this.state.socket.isConnected && typeof this.state.messageList.Messages !== "undefined"
        && this.state.messageList.Messages !== null
        && this.state.messageList.Messages.length > 19 && this.state.messageList.Messages.slice(-1)[0].ID > 1) {
        var message = {
          "request": "getMoreMessages",
          "lastId": String(this.state.messageList.Messages.slice(-1)[0].ID)
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
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
        var message = {
          "request": "leaveChat",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    createChat(state, uuid) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "createChat",
          "uuid": uuid
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
      this.commit("CREATE_CHAT", uuid);
    },
    sendMessage(state, messageContainer) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendMessage",
          "to": messageContainer.to,
          "message": messageContainer.message
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    toggleNotifications() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "toggleNotifications",
          "chat": this.state.currentChat.ID
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    resetEncryption() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "resetEncryption",
          "chat": this.state.currentChat.ID
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    verifyIdentity() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "verifyIdentity",
          "chat": this.state.currentChat.ID
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    getContacts(state) {
      if (this.state.socket.isConnected) {
        state.importingContacts = false;
        var message = {
          "request": "getContacts",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    addContact(state, contact) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected
        && contact.name !== "" && contact.phone !== "") {
          if(this.state.currentChat.Tel === contact.phone){
            this.commit("SET_CURRENT_CHAT_NAME", contact.name);
          }
        var message = {
          "request": "addContact",
          "name": contact.name,
          "phone": contact.phone,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
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
        var message = {
          "request": "uploadVcf",
          "vcf": vcf
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    uploadAttachment(state, attachment) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "uploadAttachment",
          "attachment": attachment.attachment,
          "to": attachment.to,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    refreshContacts(state, chUrl) {
      state.importingContacts = true;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "refreshContacts",
          "url": chUrl
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    delContact(state, id) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delContact",
          "id": id,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    deleteSelfDestructingMessage(state, m) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delMessage",
          "id": m.ID
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    editContact(state, data) {
      state.rateLimitError = null;
      if (this.state.socket.isConnected) {
        if(this.state.currentChat.Tel === data.contact.Tel){
          this.commit("SET_CURRENT_CHAT_NAME", data.contact.Name);
        }
        var message = {
          "request": "editContact",
          "phone": data.contact.Tel,
          "name": data.contact.Name,
          "id": data.id
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    // registration functions
    requestCode(state, tel) {
      this.state.verificationError = null
      if (this.state.socket.isConnected) {
        var message = {
          "request": "requestCode",
          "tel": tel,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    sendCode(state, code) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendCode",
          "code": code,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    setUsername(state, username) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendUsername",
          "username": username,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
      this.commit("SET_REGISTRATION_STATUS", "");
      router.push("/")
    },
    sendPin(state, pin) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendPin",
          "pin": pin,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    sendPassword(state, password) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendPassword",
          "pw": password,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    setPassword(state, password) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "setPassword",
          "pw": password.pw,
          "currentPw": password.cPw
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
        router.push("/chatList")
      }
    },
    getRegistrationStatus() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getRegistrationStatus",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    unregister() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "unregister",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))

      }
    },
    getConfig() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getConfig",
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
      }
    },
    createNewGroup(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "createGroup",
          "name": data.name,
          "members": data.members,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))

      }
    },
    updateGroup(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "updateGroup",
          "name": data.name,
          "id": data.id,
          "members": data.members,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))

      }
    },
    sendAttachment(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendAttachment",
          "type": data.type,
          "path": data.path,
          "to": data.to,
          "message": data.message,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))

      }
    },
    setDarkMode(state, darkMode) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "setDarkMode",
          "darkMode": darkMode,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
        this.state.DarkMode = darkMode;
      }
    },
    sendCaptchaToken() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendCaptchaToken",
          "token": this.state.captchaToken,
        }
        app.config.globalProperties.$socket.send(JSON.stringify(message))
        this.commit("SET_CAPTCHA_TOKEN_SENT");

      }
    },
    setCaptchaToken(state, token) {
      this.commit("SET_CAPTCHA_TOKEN", token);
    }
  }
});
