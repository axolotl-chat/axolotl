<template>
  <div class="verify">
    <div v-if="verificationError=='RegistrationLockFailure'||requestPin"
        class="verify">
      <h2>Enter your registration pin</h2>
      <p>or disable it on Android</p>
      <input v-model="pin" type="text"/>
      <button  class="btn btn-primary" @click="sendPin()">Send pin</button>

    </div>
    <div  class="verify">
      <Sms v-model="code" class="codeInput"></Sms>
      <button :disabled="inProgress" class="btn btn-primary" @click="sendCode()"> send Code </button>
    </div>
    <div v-if="inProgress&&verificationError==null" class="spinner">
      <div class="spinner-border" role="status">
          <span class="sr-only">Loading...</span>
      </div>
    </div>
    <div v-if="verificationError==404">
      Wrong code entered. Restart for another try.
      {{verificationError}}
    </div>

  </div>
</template>

<script>
import Sms from 'ofcold-security-code';
import { mapState } from 'vuex';
export default {
  name: 'verification',
  components: {
    Sms
  },
  props: {
    msg: String
  },
  mounted(){
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
  computed: mapState(['verificationError', 'requestPin']),
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
