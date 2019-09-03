
<template>
  <div class="chatList">
    <div v-if="chats.length>0" class="row">
      <router-link :to="'/chat/'+chat.Tel" v-for="chat in chats" class="col-12 chat">
        <div class="row">
          <div class="avatar col-3">
            <div class="badge-name">{{chat.Name[0]}}</div>
          </div>
  				<div class="meta col-9">
  					<p class="name">{{chat.Name}}</p>
  					<p class="preview">{{chat.Messages[0].Message}}</p>
          </div>
				</div>
      </router-link>
    </div>
    <div v-else >
      No chats aviable
    </div>
    <!-- {{chats}} -->
  </div>
</template>

<script>
export default {
  name: 'ChatList',
  props: {
    msg: String
  },
  created(){
    // console.log("ChatList", this);
    this.$store.dispatch("getChatList");
    // this.$store.dispatch('addResponses', "1");
    // Vue.prototype.$store.dispatch("getChatList")

    // this.$store.dispatch("getChatList")
  },
  computed: {
    chats () {
      return this.$store.state.chatList
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.avatar {
    justify-content: center;
    display: flex;
    align-items: center;
}
.badge-name{
  background-color: #2090ea;
  /* padding: 14px; */
  width:50px;
  height:50px;
  border-radius: 50%;
  color: #FFF;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 16px;
  display:flex;
  justify-content: center;
  align-items:center;
}
.meta{
  text-align:left;
}
.meta p{
  margin:0px;
}
.meta .name{
  font-weight:bold;
  font-size:20px;
}
.meta .preview{
  font-size:15px;
}
.chat .row{
  border-bottom:1px solid grey;
  border-bottom: 1px solid #c2c2c2;
  padding: 10px;
}
a.chat{
  color:#000;
}
a:hover.chat{
  text-decoration:none;
}
</style>
