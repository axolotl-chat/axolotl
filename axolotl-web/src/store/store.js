import Vuex from 'vuex'
import Vue from 'vue'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    chatList: [],
    messageList: [],
    request: '',
    contacts:[],
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
        SET_CHATLIST(state, chatList){
              state.chatList = chatList;
        },
        SEND_MESSAGE(){

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
        CLEAR_MESSAGELIST(state){
          state.messageList = {};
        },
        SET_CONTACTS(state, contacts){
              state.contacts = contacts;
        },
        LEAVE_CHAT(state){
          this.commit("CLEAR_MESSAGELIST");
        },
        SOCKET_ONOPEN (state, event)  {
          Vue.prototype.$socket = event.currentTarget
          state.socket.isConnected = true
          this.dispatch("getRegistrationStatus")
          // Vue.prototype.$socket.send("getChatList")

        },
        SOCKET_ONCLOSE (state, event)  {
          state.socket.isConnected = false
        },
        SOCKET_ONERROR (state, event)  {
          console.error(state, event)
        },
        // default handler called for all methods
        SOCKET_ONMESSAGE (state, message)  {
          if(message.data!="Hi Client!"){
            var messageData =JSON.parse(message.data)
            console.log(messageData);
            if(Object.keys(messageData)[0]=="ChatList"){
              this.commit("SET_CHATLIST",messageData["ChatList"] );
            }
            else if(Object.keys(messageData)[0]=="MessageList"){
              this.commit("SET_MESSAGELIST",messageData["MessageList"] );
            }
            else if(Object.keys(messageData)[0]=="Type"){
              this.commit("SET_REQUEST",messageData["Type"] );
            }
            else if(Object.keys(messageData)[0]=="ContactList"){
              this.commit("SET_CONTACTS",messageData["ContactList"] );
            }
            else if(Object.keys(messageData)[0]=="MoreMessageList"){
              this.commit("SET_MORE_MESSAGELIST",messageData["MoreMessageList"] );
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
    addContact:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"addContact",
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
          console.log(JSON.stringify(message))
          Vue.prototype.$socket.send(JSON.stringify(message))
        }
    },
    sendCode:function(context, code){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendCode",
          "code":  code,
        }
        console.log(JSON.stringify(message))
        Vue.prototype.$socket.send(JSON.stringify(message))
        window.router.push("/chatList")

      }
    },
    getRegistrationStatus:function(context, code){
      if(this.state.socket.isConnected){
        var message = {
          "request":"getRegistrationStatus",
        }
        console.log(JSON.stringify(message))
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
  }
});
