<template>
  <div class="verify">
    <h3 v-translate>Enter your registration pin</h3>
    <div v-if="verificationError=='RegistrationLockFailure'||requestPin"
        class="verify">
      <p v-translate>or disable it on Android/IOs</p>
      <input v-model="pin" type="text"/>
      <button  class="btn btn-primary" @click="sendPin()" v-translate>Send pin</button>

    </div>
    <div  class="verify" v-if="!requestPin">
      <Sms v-model="code" class="codeInput"></Sms>
      <button :disabled="inProgress" class="btn btn-primary" @click="sendCode()" v-translate> Send code </button>
    </div>
    <div v-if="inProgress&&verificationError==null&&!requestPin" class="spinner">
      <div class="spinner-border" role="status">
          <span class="sr-only" v-translate>Loading...</span>
      </div>
    </div>
    <div v-if="verificationError==404">
      <div v-translate>Wrong code entered. Restart for another try.</div>
    </div>

  </div>
</template>

<script>
import Sms from 'ofcold-security-code';
import { mapState } from 'vuex';
import checkRegistrationStatus from '@/helpers/registrationStatus'

export default {
  name: 'verification',
  components: {
    Sms
  },
  props: {
    msg: String
  },
  mounted(){
    checkRegistrationStatus(this.registrationStatus)
    document.getElementsByClassName("form-control")[0].focus()
  },
  methods:{
    sendCode(){
      if(this.code.length==6){
        this.$store.dispatch("sendCode",this.code);
        this.inProgress = true;
      }
    },
    sendPin(){
      if(this.code.length==6){
        this.$store.dispatch("sendPin",this.pin);
        this.inProgress = true;
      }
    }
  },
  watch:{
    registrationStatus() {
      checkRegistrationStatus(this.registrationStatus)
    }
  },
  computed: mapState(['verificationError', 'requestPin', 'registrationStatus']),
  data() {
    return {
      code:"",
      pin:"",
      inProgress:false,
    };
  },
}
</script>
<style>
  .verify{
    display:flex;
    flex-direction: column;
    padding-top:30px;
  }
  .verify h3{
    text-align:center;
  }
  .verify .codeInput{
    margin-top:30px;
  }
  .verify .btn{
    max-width: 300px;
    margin: auto;
    margin-top: auto;
    margin-top: 50px;
  }

  .verify .ofcold__security-code-wrapper .ofcold__security-code-field .form-control {
    border: 2px solid #2090ea !important;
  }
  .verify .spinner{
    display:flex;
    justify-content: center;
    margin-top:20px;
  }
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
