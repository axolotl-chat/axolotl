<template>
  <div class="register">
    <VuePhoneNumberInput v-model="phone" @update="updatePhone" :callingCode="cc" class="phoneInput" />
    <button class="btn btn-primary" @click="requestCode()"> request code </button>
  </div>
</template>

<script>
import VuePhoneNumberInput from 'vue-phone-number-input';
import 'vue-phone-number-input/dist/vue-phone-number-input.css';
export default {
  name: 'register',
  components: {
    VuePhoneNumberInput
  },
  props: {
    msg: String
  },
  methods:{
    requestCode(){
      // console.log(this.cc)
      if(typeof this.cc!="undefined")
      this.$store.dispatch("requestCode",this.cc);
    },
    updatePhone(e){
      this.cc=e.formattedNumber;
    }
  },
  data() {
    return {
      phone: '',
      cc:""
    };
  },
}
</script>
<style scoped>
  .register{
    display:flex;
    flex-direction: column;
  }
  .phoneInput{
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
