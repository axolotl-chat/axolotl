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
    newGroupName:null,
    newGroupMembers:[],
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
      if(!error){
        state.error = error;
      }
    },
        SET_CHATLIST(state, chatList){
              state.chatList = chatList;
        },
        SEND_MESSAGE(){

        },
        CREATE_CHAT(state, tel){
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
            else if(Object.keys(messageData)[0]=="Type"){
              this.commit("SET_REQUEST",messageData );
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
    getMessageList:function(context, chatId){
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
    leaveChat:function(){
      if(this.state.socket.isConnected){
        var message = {
          "request":"leaveChat",
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("LEAVE_CHAT");
    },
    createChat:function(context, tel){
      if(this.state.socket.isConnected){
        var message = {
          "request":"createChat",
          "tel": tel
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
      this.commit("CREATE_CHAT", tel);
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
    uploadVcf:function(context, vcf) {
      if(this.state.socket.isConnected){
        var message = {
          "request":"uploadVcf",
          "vcf": vcf
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    uploadAttachment:function(context, attachment) {
      if(this.state.socket.isConnected){
        var message = {
          "request":"uploadAttachment",
          "attachment": attachment
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    refreshContacts:function(context, chUrl){
      if(this.state.socket.isConnected){
        var message = {
          "request":"refreshContacts",
          "url": chUrl
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    delContact:function(state,tel){
      if(this.state.socket.isConnected){
        var message = {
          "request":"delContact",
          "phone":tel,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    editContact:function(state,data){
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
    sendPassword:function(context, password){
      if(this.state.socket.isConnected){
        var message = {
          "request":"sendPassword",
          "pw":  password,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))
      }
    },
    setPassword:function(context, password){
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
    createNewGroup:function(context, data){
      if(this.state.socket.isConnected){
        var message = {
          "request":"createGroup",
          "name": data.name,
          "members": data.members,
        }
        Vue.prototype.$socket.send(JSON.stringify(message))

      }
    },
    sendAttachment:function(context, data){
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
