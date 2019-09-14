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
<style scoped>
  .verify{
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
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
