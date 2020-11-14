<template>
  <div class="password">
    <div v-if="error" class="alert alert-danger" v-translate>Password is wrong</div>
    <input v-model="pw"  type="password" class="codeInput form-control" id="passwordInput" @keydown="checkEnter($event)"/>
    <button class="btn btn-primary" @click="sendPassword" v-translate> Decrypt</button>
    <button v-if="error" class="btn btn-danger" @click="unregister" v-translate> Unregister</button>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import checkRegistrationStatus from '@/helpers/registrationStatus'

export default {
  name: 'password',
  methods:{
    sendPassword(){
      this.$store.dispatch("sendPassword",this.pw);
      this.pw = null;
    },
    checkEnter(e){
      if(e.keyCode==13) this.sendPassword();
    },
    unregister(){
      this.$store.dispatch("unregister");

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
  watch:{
    registrationStatus() {
      checkRegistrationStatus(this.registrationStatus)
    }
  },
  computed: {
    ...mapState(['registrationStatus']),
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
