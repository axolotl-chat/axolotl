<template>
  <div class="password">

    <div v-if="error" class="alert alert-danger">Password is wrong</div>
    <input v-model="pw"  type="password" class="codeInput form-control" id="passwordInput"/>
    <button class="btn btn-primary" @click="sendPassword"> Decrypt</button>
    <button v-if="error" class="btn btn-danger" @click="unregister"> Unregister</button>
  </div>
</template>

<script>
export default {
  name: 'password',
  methods:{
    sendPassword(){
      this.$store.dispatch("sendPassword",this.pw);
      this.pw = null;
    },
    unregister(){
      console.log("vlub")
    }
  },
  data() {
    return {
      pw:""
    };
  },
  mounted(){
    document.getElementById("passwordInput").focus();
  },
  computed: {
    error () {
      return this.$store.state.loginError;
    }
  }
}
</script>
<style scoped>
  .password{
    display:flex;
    flex-direction: column;
  }
  .codeInput{
    margin-top:30px;
  }
  .btn{
    max-width: 300px;
    margin: auto;
    margin-top: auto;
    margin-top: 50px;
  }
  .alert{
    border:none;
    border-radius:0px;
    margin-top:20px;

  }
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
