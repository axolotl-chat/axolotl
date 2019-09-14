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
    error: null,
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
          state.error = error;
    },
        SET_CHATLIST(state, chatList){
              state.chatList = chatList;
        },
        SEND_MESSAGE(){

        },
        SET_DEVICELIST(state, devices){
          state.devices = devices
        },
        SET_REQUEST(state, request){
          state.request = request;
          if(request=="getPhoneNumber"){
            window.router.push("/")
          }
          else if (request == "getVerificationCode") {
            window.router.push("/verify")
          }
          else if (request =="registrationDone") {
            window.router.push("/chatList")
            this.dispatch("getChatList")
          }
          // this.dispatch("requestCode", "+123456")
        },
        SET_MESSAGELIST(state, messageList){
              state.messageList = messageList;
        },
        SET_MORE_MESSAGELIST(state, messageList){
              if(messageList.Messages!= null){
                state.messageList.Messages = state.messageList.Messages.concat(messageList.Messages);
              }
        },
        SET_MESSAGE_RECIEVED(state, message){
          if(state.messageList.ID==message.Source){
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
              state.contacts = contacts;
        },
        LEAVE_CHAT(){
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
        SOCKET_ONERROR (state, event)  {
          console.error(state, event)
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
            else if(Object.keys(messageData)[0]=="Type"){
              this.commit("SET_REQUEST",messageData["Type"] );
            }
            state.socket.message = message.data
          }


        },
        // mutations for reconnect methods
        SOCKET_RECONNECT(state, count) {
          console.info(state, count)
        },
        SOCKET_RECONNECT_ERROR(state) {
          state.socket.reconnectError = true;
        },
  },

  actions: {
    addDevice:function(state,url){
      if(this.state.socket.isConnected){
        var message = {
          "request":"addDevice",
          "url":url,
        }
        console.log("ad",url);
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delDevice:function(state,id){
      if(this.state.socket.isConnected){
        var message = {
          "request":"delDevice",
          "id":id,
        }
        console.log(message);
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
    getMessageList:function(context, chatId){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getMessageList",
          "id":  chatId
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getMoreMessages:function(){
      if(this.state.socket.isConnected){
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
    leaveChat:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"leaveChat",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("LEAVE_CHAT");
    },
    sendMessage:function(context, messageContainer){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendMessage",
          "to":  messageContainer.to,
          "message":  messageContainer.message
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    getContacts:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getContacts",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    addContact:function(context, contact){
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
    // registration functions
      requestCode:function(context, tel){
        if(this.state.socket.isConnected){
          var message = {
            "request":"requestCode",
            "tel":  tel,
          }
          Vue.prototype.$socket.send(JSON.stringify(message))
        }
    },
    sendCode:function(context, code){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendCode",
          "code":  code,
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
  }
});
