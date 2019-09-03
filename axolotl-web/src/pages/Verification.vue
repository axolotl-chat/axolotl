<template>
  <div>
    <Sms v-model="cc" class="codeInput"></Sms>
    <button class="btn" @click="sendCode()"> send Code </button>
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
  .codeInput{
    margin-top:30px;
  }
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
