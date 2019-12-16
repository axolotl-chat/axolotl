<template>
  <div class="verify">
    <Sms v-model="cc" class="codeInput"></Sms>
    <button class="btn btn-primary" @click="sendCode()"> send Code </button>
  </div>
</template>

<script>
import Sms from 'ofcold-security-code';
export default {
  name: 'verification',
  components: {
    Sms
  },
  props: {
    msg: String
  },
  methods:{
    sendCode(){
      // console.log(this.cc)
      // if(typeof this.cc!="undefined")
      this.$store.dispatch("sendCode",this.cc);
    },
    updatePhone(e){
      this.cc=e.formattedNumber;
    }
  },
  data() {
    return {
      cc:""
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
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
