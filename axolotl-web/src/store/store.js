import Vuex from 'vuex'
import Vue from 'vue'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    chatList: [],
    messageList: [],
    request: '',
    contacts:[],
    devices: [],
    gui:null,
    error: null,
    identity: {
      me:null,
      their:null
    },
    loginError:null,
    ratelimitError:null,
    newGroupName:null,
    currentChat:null,
    newGroupMembers:[],
    importingContacts:false,
    socket: {
      isConnected: false,
      message: '',
      reconnectError: false,
    }
  },

  getters: {
    // Here we will create a getter
  },

  mutations: {

    SET_ERROR(state, error){
      if(error=="wrong password"){
        state.loginError = error;
      }else if(error=="Rate limit exceeded: 413"){
        state.ratelimitError = error+". Try again later!";
      }
    },
        SET_CHATLIST(state, chatList){
              state.chatList = chatList;
        },
        SET_CURRENT_CHAT(state, chat){
          state.currentChat = chat;
        },
        SEND_MESSAGE(){

        },
        CREATE_CHAT(state, tel){
          state.currentChat = null;
          window.router.push('/chat/'+tel)
        },
        SET_DEVICELIST(state, devices){
          state.devices = devices
        },
        SET_REQUEST(state, request){
          var type = request["Type"]
          state.request = request;
          if(type=="getPhoneNumber"){
            window.router.push("/")
          }
          else if (type == "getVerificationCode") {
            window.router.push("/verify")
          }
          else if (type == "getEncryptionPw") {
            if(window.router.currentRoute.name !="password")
            window.router.push("/password")
          }
          else if (type =="registrationDone") {
            window.router.push("/chatList")
            this.dispatch("getChatList")
          }
          else if (type =="requestEnterChat") {
            window.router.push("/chat/"+request["Chat"])
            this.dispatch("getChatList")
          }
          else if (type =="registrationDone") {
            window.router.push("/chatList")
            this.dispatch("getChatList")
          }
          else if (type =="Error") {
            this.commit("SET_ERROR", request.Error)
          }
          // this.dispatch("requestCode", "+123456")
        },
        SET_MESSAGELIST(state, messageList){
              state.messageList = messageList;
              // router.push('/chat/'+)

        },
        SET_MORE_MESSAGELIST(state, messageList){
              if(messageList.Messages!= null){
                state.messageList.Messages = state.messageList.Messages.concat(messageList.Messages);
              }
        },
        SET_MESSAGE_RECIEVED(state, message){
          if(state.messageList.ID==message.ChatID){
            var tmpList = state.messageList.Messages;
            tmpList.push(message);
            tmpList.sort(function(a, b){
                return b.ID-a.ID
            })
            state.messageList.Messages = tmpList;
          }
          state.chatList.forEach((chat, i)=>{
            if(chat.Tel == message.Source){
              state.chatList[i].Messages = [message]
            }
          })

        },
        CLEAR_MESSAGELIST(state){
          state.messageList = {};
        },
        SET_CONTACTS(state, contacts){
          state.importingContacts = false;
              state.contacts = contacts;
        },
        SET_IDENTITY(state, identity){
          state.identity.me = identity.Identity;
          state.identity.their = identity.TheirId;
        },
        LEAVE_CHAT(state){
          state.currentChat = null;
          this.commit("CLEAR_MESSAGELIST");
        },
        SOCKET_ONOPEN (state, event)  {
          Vue.prototype.$socket = event.currentTarget
          state.socket.isConnected = true
          this.dispatch("getRegistrationStatus")
          // Vue.prototype.$socket.send("getChatList")

        },
        SOCKET_ONCLOSE (state)  {
          state.socket.isConnected = false
        },
        SOCKET_ONERROR ()  {
          // console.error(state, event)
        },
        // default handler called for all methods
        SOCKET_ONMESSAGE (state, message)  {
          if(message.data!="Hi Client!"){
            var messageData =JSON.parse(message.data)
            if(typeof messageData.Error !="undefined"){
              this.commit("SET_ERROR", messageData["Error"] );
            }
            if(Object.keys(messageData)[0]=="ChatList"){
              this.commit("SET_CHATLIST",messageData["ChatList"] );
            }
            else if(Object.keys(messageData)[0]=="MessageList"){
              this.commit("SET_MESSAGELIST",messageData["MessageList"] );
            }

            else if(Object.keys(messageData)[0]=="ContactList"){
              this.commit("SET_CONTACTS",messageData["ContactList"] );
            }
            else if(Object.keys(messageData)[0]=="MoreMessageList"){
              this.commit("SET_MORE_MESSAGELIST",messageData["MoreMessageList"] );
            }
            else if(Object.keys(messageData)[0]=="DeviceList"){
              this.commit("SET_DEVICELIST",messageData["DeviceList"] );
            }
            else if(Object.keys(messageData)[0]=="MessageRecieved"){
              this.commit("SET_MESSAGE_RECIEVED",messageData["MessageRecieved"] );
            }
            else if(Object.keys(messageData)[0]=="Gui"){
              this.commit("SET_CONFIG_GUI", messageData["Gui"]);
            }
            else if(Object.keys(messageData)[0]=="CurrentChat"){
              this.commit("SET_CURRENT_CHAT", messageData["CurrentChat"]);
            }
            else if(Object.keys(messageData)[0]=="Identity"){
              this.commit("SET_IDENTITY", messageData);
            }
            else if(Object.keys(messageData)[0]=="Type"){
              this.commit("SET_REQUEST",messageData );
            }
            else if(Object.keys(messageData)[0]=="Error"){
              this.commit("SET_ERROR",messageData.Errorx );
            }
            state.socket.message = message.data
          }
        },
        // mutations for reconnect methods
        SOCKET_RECONNECT() {
        },
        SOCKET_RECONNECT_ERROR(state) {
          state.socket.reconnectError = true;
        },
        SET_CONFIG_GUI(state, gui) {
          state.gui =  gui;
        },
  },

  actions: {
    addDevice:function(state,url){
      if(this.state.socket.isConnected){
        var message = {
          "request":"addDevice",
          "url":url,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delDevice:function(state,id){
      if(this.state.socket.isConnected){
        var message = {
          "request":"delDevice",
          "id":id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getDevices:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getDevices",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getChatList:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getChatList",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delChat:function(state,id){
      if(this.state.socket.isConnected){
        var message = {
          "request":"delChat",
          "id":id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getMessageList:function(state, chatId){
      this.commit("CLEAR_MESSAGELIST");
      if(this.state.socket.isConnected){
        var message = {
          "request":"getMessageList",
          "id":  chatId
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getMoreMessages:function(){
      if(this.state.socket.isConnected && typeof this.state.messageList.Messages !="undefined"
        && this.state.messageList.Messages !=null
        &&this.state.messageList.Messages.length>20 && this.state.messageList.Messages.slice(-1)[0].ID>1){
        var message = {
          "request":"getMoreMessages",
          "lastId":  String(this.state.messageList.Messages.slice(-1)[0].ID)
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    clearMessageList:function(){
      this.commit("CLEAR_MESSAGELIST");
    },
    setCurrentChat:function(state, chat){
      this.commit("SET_CURRENT_CHAT", chat);
    },
    leaveChat:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"leaveChat",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("LEAVE_CHAT");
    },
    createChat:function(state, tel){
      if(this.state.socket.isConnected){
        var message = {
          "request":"createChat",
          "tel": tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("CREATE_CHAT", tel);
    },
    sendMessage:function(state, messageContainer){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendMessage",
          "to":  messageContainer.to,
          "message":  messageContainer.message
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    toggleNotifcations:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"toggleNotifcations",
          "chat":this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    resetEncryption:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"resetEncryption",
          "chat":this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    verifyIdentity:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"verifyIdentity",
          "chat":this.state.currentChat.Tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getContacts:function(state){
      if(this.state.socket.isConnected){
        state.importingContacts = false;
        var message = {
          "request":"getContacts",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    addContact:function(state, contact){
      state.ratelimitError = null;
      if(this.state.socket.isConnected
        &&contact.name!="" && contact.phone!=""){
        var message = {
          "request":"addContact",
          "name": contact.name,
          "phone": contact.phone,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    uploadVcf:function(state, vcf) {
      state.ratelimitError = null;
      state.importingContacts = true;
      if(this.state.socket.isConnected){
        var message = {
          "request":"uploadVcf",
          "vcf": vcf
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    uploadAttachment:function(state, attachment) {
      if(this.state.socket.isConnected){
        var message = {
          "request":"uploadAttachment",
          "attachment": attachment
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    refreshContacts:function(state, chUrl){
      state.importingContacts = true;
      if(this.state.socket.isConnected){
        var message = {
          "request":"refreshContacts",
          "url": chUrl
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delContact:function(state,id){
      state.ratelimitError = null;
      if(this.state.socket.isConnected){
        var message = {
          "request":"delContact",
          "id":id,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    editContact:function(state,data){
      state.ratelimitError = null;
      if(this.state.socket.isConnected){
        var message = {
          "request":"editContact",
          "phone":data.contact.Tel,
          "name":data.contact.Name,
          "id": data.id
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    // registration functions
      requestCode:function(state, tel){
        if(this.state.socket.isConnected){
          var message = {
            "request":"requestCode",
            "tel":  tel,
          }
          Vue.prototype.$socket.send(JSON.stringify(message))
        }
    },
    sendCode:function(state, code){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendCode",
          "code":  code,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
        window.router.push("/chatList")

      }
    },
    sendPassword:function(state, password){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendPassword",
          "pw":  password,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    setPassword:function(state, password){
      if(this.state.socket.isConnected){
        var message = {
          "request":"setPassword",
          "pw":  password.pw,
          "currentPw": password.cPw
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
        window.router.push("/chatList")
      }
    },
    getRegistrationStatus:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getRegistrationStatus",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    unregister:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"unregister",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    createNewGroup:function(state, data){
      if(this.state.socket.isConnected){
        var message = {
          "request":"createGroup",
          "name": data.name,
          "members": data.members,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    sendAttachment:function(state, data){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendAttachment",
          "type": data.type,
          "path": data.path,
          "to": data.to,
          "message": data.message,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    }
  }
});
