import Vuex from 'vuex'
import Vue from 'vue'
import { router } from '../router/router';

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    chatList: [],
    messageList: {},
    request: '',
    contacts: [],
    contactsFilterd: [],
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
    ratelimitError: null,
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
    socket: {
      isConnected: false,
      message: '',
      reconnectError: false,
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
      if (error == "") {
        state.loginError = null;
        // state.ratelimitError = null;
        state.errorConnection = null;
      }
      else if (error == "wrong password") {
        state.loginError = error;
      } else if (error.includes("Rate")) {
        state.ratelimitError = error + ". Try again later!";
      } else if (error.includes("no such host") || error.includes("timeout")) {
        state.errorConnection = error;
      } else if (error.includes("Your registration is faulty")) {
        state.error = error + " .Please consider to register again";
      } else if (error.includes(400)) {
        //when pin is missing
        if (state.verificationInProgress)
          state.verificationError = 400
      }
      else if (error.includes(404)) {
        //when code was wrong
        if (state.verificationInProgress)
          state.verificationError = 404
      }
      else if (error.includes("RegistrationLockFailure")) {
        state.verificationError = "RegistrationLockFailure"
      }
    },
    SET_CHATLIST(state, chatList) {
      state.chatList = chatList;
    },
    SET_CURRENT_CHAT(state, chat) {
      state.currentChat = chat;
    },
    OPEN_CHAT(state, data) {
      state.currentChat = data.CurrentChat;
      // console.log(data);
      state.currentGroup = data.Group;
      state.currentContact = data.Contact;
    },
    UPDATE_CURRENT_CHAT(state, data) {
      state.currentChat = data.CurrentChat;
      if (state.messageList.Messages == undefined) {
        state.messageList.Messages = []
      }
      var prepare = state.messageList.Messages.map(function(e) { return e.ID; })
      if (data.CurrentChat.Messages != null) {
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
    CREATE_CHAT(state, tel) {
      state.currentChat = null;
      router.push('/chat/' + tel)
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
      if (type == "getPhoneNumber") {
        this.commit("SET_REGISTRATION_STATUS", "phoneNumber");
      } else if (type == "getVerificationCode") {
        this.commit("SET_REGISTRATION_STATUS", "verificationCode");
        state.verificationInProgress = true;
      } else if (type == "getPin") {
        this.commit("SET_REGISTRATION_STATUS", "pin");
        state.requestPin = true;
      } else if (type == "getEncryptionPw") {
        this.commit("SET_REGISTRATION_STATUS", "password");
      } else if (type == "registrationDone") {
        this.commit("SET_REGISTRATION_STATUS", "registered");
      } else if (type == "requestEnterChat") {
        router.push("/chat/" + request["Chat"])
        this.dispatch("getChatList")
      } else if (type == "config") {
        this.commit("SET_CONFIG", request)
      } else if (type == "Error") {
        this.commit("SET_ERROR", request.Error)
      }
      // this.dispatch("requestCode", "+123456")
    },
    SET_MESSAGELIST(state, messageList) {
      state.messageList = messageList;
    },
    SET_MORE_MESSAGELIST(state, messageList) {
      if (messageList.Messages != null) {
        state.messageList.Messages = state.messageList.Messages.concat(messageList.Messages);
      }
    },
    SET_MESSAGE_RECIEVED(state, message) {
      if (state.messageList.ID == message.ChatID) {
        var tmpList = state.messageList.Messages;
        tmpList.push(message);
        tmpList.sort(function(a, b) {
          return b.ID - a.ID
        })
        state.messageList.Messages = tmpList;
      }
      state.chatList.forEach((chat, i) => {
        if (chat.Tel == message.ChatID) {
          state.chatList[i].Messages = [message]
        }
      })
    },
    SET_MESSAGE_UPDATE(state, message) {
      if (state.messageList.Session.ID == message.SID) {
        var index = state.messageList.Messages.findIndex(m => {
          return m.ID === message.ID;
        });
        if (index != -1) {
          var tmpList = JSON.parse(JSON.stringify(state.messageList.Messages));
          tmpList[index] = message;
          tmpList.sort(function(a, b) {
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
        }
      }
    },
    CLEAR_MESSAGELIST(state) {
      state.messageList = {};
    },
    SET_CONTACTS(state, contacts) {
      if (contacts != null) {
        contacts = contacts.sort((a, b) => a.Name.localeCompare(b.Name));
        state.importingContacts = false;
        state.contacts = contacts;
      }
    },
    SET_CONTACTS_FILTER(state, filter) {
      filter = filter.toLowerCase()
      var f = state.contacts.filter(c => c.Name.toLowerCase().includes(filter));
      state.contactsFilterd = f;
      state.contactsFilterActive = true;
    },
    SET_CONTACTS_FOR_GROUP_FILTER(state, filter) {
      filter = filter.toLowerCase()
      var f = state.contacts.filter(c => c.Name.toLowerCase().includes(filter));
      f = f.filter(c => state.currentGroup.Members.indexOf(c.Tel) == -1);
      state.contactsFilterd = f;
      state.contactsFilterActive = true;
    },
    SET_CLEAR_CONTACTS_FILTER(state) {
      state.contactsFilterd = [];
      state.contactsFilterActive = false;
    },
    LEAVE_CHAT(state) {
      state.currentGroup = null;
      state.currentContact = null;
      state.currentChat = null;
      this.commit("CLEAR_MESSAGELIST");
    },
    SOCKET_ONOPEN(state, event) {
      Vue.prototype.$socket = event.currentTarget
      state.socket.isConnected = true
    },
    SOCKET_ONCLOSE(state) {
      state.socket.isConnected = false
    },
    SOCKET_ONERROR() {
      // console.error(state, event)
    },
    // default handler called for all methods
    SOCKET_ONMESSAGE(state, message) {
      if (message.data != "Hi Client!") {
        var messageData = JSON.parse(message.data);
        if (typeof messageData.Error != "undefined") {
          this.commit("SET_ERROR", messageData["Error"]);
        }
        if (Object.keys(messageData)[0] == "ChatList") {
          this.commit("SET_CHATLIST", messageData["ChatList"]);
        }
        else if (Object.keys(messageData)[0] == "MessageList") {
          this.commit("SET_MESSAGELIST", messageData["MessageList"]);
        }

        else if (Object.keys(messageData)[0] == "ContactList") {
          this.commit("SET_CONTACTS", messageData["ContactList"]);
        }
        else if (Object.keys(messageData)[0] == "MoreMessageList") {
          this.commit("SET_MORE_MESSAGELIST", messageData["MoreMessageList"]);
        }
        else if (Object.keys(messageData)[0] == "DeviceList") {
          this.commit("SET_DEVICELIST", messageData["DeviceList"]);
        }
        else if (Object.keys(messageData)[0] == "MessageRecieved") {
          this.commit("SET_MESSAGE_RECIEVED", messageData["MessageRecieved"]);
        }
        else if (Object.keys(messageData)[0] == "UpdateMessage") {
          this.commit("SET_MESSAGE_UPDATE", messageData["UpdateMessage"]);
        }
        else if (Object.keys(messageData)[0] == "Gui") {
          this.commit("SET_CONFIG_GUI", messageData["Gui"]);
        }
        else if (Object.keys(messageData)[0] == "DarkMode") {
          this.commit("SET_CONFIG_DARK_MODE", messageData["DarkMode"]);
        }
        else if (Object.keys(messageData)[0] == "CurrentChat") {
          this.commit("SET_CURRENT_CHAT", messageData["CurrentChat"]);
        }
        else if (Object.keys(messageData)[0] == "Type") {
          this.commit("SET_REQUEST", messageData);
        }
        else if (Object.keys(messageData)[0] == "FingerprintNumbers") {
          this.commit("SET_FINGERPRINT", messageData);
        }
        else if (Object.keys(messageData)[0] == "Error") {
          this.commit("SET_ERROR", messageData.Errorx);
        }
        else if (Object.keys(messageData)[0] == "OpenChat") {
          this.commit("OPEN_CHAT", messageData["OpenChat"]);
        }
        else if (Object.keys(messageData)[0] == "UpdateCurrentChat") {
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
    // mutations for reconnect methods
    SOCKET_RECONNECT() {
    },
    SOCKET_RECONNECT_ERROR(state) {
      state.socket.reconnectError = true;
    },
    SET_CONFIG_GUI(state, gui) {
      state.gui = gui;
    },
    SET_CONFIG_DARK_MODE(state, darkMode) {
      if (window.getCookie("darkMode") != String(darkMode)) {
        var d = new Date();
        d.setTime(d.getTime() + (300 * 24 * 60 * 60 * 1000));
        var expires = "expires=" + d.toUTCString();
        document.cookie = "darkMode" + "=" + darkMode + ";" + expires + ";path=/";
        state.darkMode = darkMode;
        window.location.replace("/")
      }
    },
  },

  actions: {
    addDevice: function(state, url) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "addDevice",
          "url": url,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delDevice: function(state, id) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delDevice",
          "id": id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getDevices: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getDevices",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getChatList: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getChatList",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delChat: function(state, id) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delChat",
          "id": id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getMessageList: function(state, chatId) {
      this.commit("CLEAR_MESSAGELIST");
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getMessageList",
          "id": chatId
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    openChat: function(state, chatId) {
      this.commit("CLEAR_MESSAGELIST");
      if (this.state.socket.isConnected) {
        var message = {
          "request": "openChat",
          "id": chatId
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getMoreMessages: function() {
      if (this.state.socket.isConnected && typeof this.state.messageList.Messages != "undefined"
        && this.state.messageList.Messages != null
        && this.state.messageList.Messages.length > 19 && this.state.messageList.Messages.slice(-1)[0].ID > 1) {
        var message = {
          "request": "getMoreMessages",
          "lastId": String(this.state.messageList.Messages.slice(-1)[0].ID)
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    clearMessageList: function() {
      this.commit("CLEAR_MESSAGELIST");
    },
    setCurrentChat: function(state, chat) {
      this.commit("SET_CURRENT_CHAT", chat);
    },
    leaveChat: function() {
      this.commit("LEAVE_CHAT");
      if (this.state.socket.isConnected) {
        var message = {
          "request": "leaveChat",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    createChat: function(state, tel) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "createChat",
          "tel": tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("CREATE_CHAT", tel);
    },
    sendMessage: function(state, messageContainer) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendMessage",
          "to": messageContainer.to,
          "message": messageContainer.message
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    toggleNotifcations: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "toggleNotifcations",
          "chat": this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    resetEncryption: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "resetEncryption",
          "chat": this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    verifyIdentity: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "verifyIdentity",
          "chat": this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getContacts: function(state) {
      if (this.state.socket.isConnected) {
        state.importingContacts = false;
        var message = {
          "request": "getContacts",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    addContact: function(state, contact) {
      state.ratelimitError = null;
      if (this.state.socket.isConnected
        && contact.name != "" && contact.phone != "") {
        var message = {
          "request": "addContact",
          "name": contact.name,
          "phone": contact.phone,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    filterContacts: function(state, filter) {
      this.commit("SET_CONTACTS_FILTER", filter);
    },
    filterContactsForGroup: function(state, filter) {
      this.commit("SET_CONTACTS_FOR_GROUP_FILTER", filter);
    },
    clearFilterContacts: function() {
      this.commit("SET_CLEAR_CONTACTS_FILTER");
    },
    uploadVcf: function(state, vcf) {
      state.ratelimitError = null;
      state.importingContacts = true;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "uploadVcf",
          "vcf": vcf
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    uploadAttachment: function(state, attachment) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "uploadAttachment",
          "attachment": attachment.attachment,
          "to": attachment.to,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    refreshContacts: function(state, chUrl) {
      state.importingContacts = true;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "refreshContacts",
          "url": chUrl
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delContact: function(state, id) {
      state.ratelimitError = null;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "delContact",
          "id": id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    editContact: function(state, data) {
      state.ratelimitError = null;
      if (this.state.socket.isConnected) {
        var message = {
          "request": "editContact",
          "phone": data.contact.Tel,
          "name": data.contact.Name,
          "id": data.id
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    // registration functions
    requestCode: function(state, tel) {
      this.state.verificationError = null
      if (this.state.socket.isConnected) {
        var message = {
          "request": "requestCode",
          "tel": tel,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    sendCode: function(state, code) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendCode",
          "code": code,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    sendPin: function(state, pin) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendPin",
          "pin": pin,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    sendPassword: function(state, password) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendPassword",
          "pw": password,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    setPassword: function(state, password) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "setPassword",
          "pw": password.pw,
          "currentPw": password.cPw
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
        router.push("/chatList")
      }
    },
    getRegistrationStatus: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getRegistrationStatus",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    unregister: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "unregister",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    getConfig: function() {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "getConfig",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    createNewGroup: function(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "createGroup",
          "name": data.name,
          "members": data.members,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    updateGroup: function(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "updateGroup",
          "name": data.name,
          "id": data.id,
          "members": data.members,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    sendAttachment: function(state, data) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "sendAttachment",
          "type": data.type,
          "path": data.path,
          "to": data.to,
          "message": data.message,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    setDarkMode: function(state, darkMode) {
      if (this.state.socket.isConnected) {
        var message = {
          "request": "setDarkMode",
          "darkMode": darkMode,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
        this.state.DarkMode = darkMode;
      }
    }
  }
});
