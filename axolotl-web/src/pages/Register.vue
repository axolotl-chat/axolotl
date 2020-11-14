<template>
  <div class="register">
    <div v-if="infoPage" class="page1 info">
      <img class="logo" src="/axolotl.png" />
      <h1 class="title">Axolotl Beta</h1>
      <h2 class="subtitle" v-translate>A cross-plattform signal client</h2>
      <div class="description">Hey! Mr. Tambourine Man, play a song for me,
        <br />
        In the jingle jangle morning I'll come following you.
        <br />
        It's beta, expect lot's of things not working.
        <br />
        <a href="https://axolotl.chat" @click="openExtern($event, 'https://axolotl.chat')">https://axolotl.chat</a>
        <br />
        <font-awesome-icon id="heart" icon="heart" />
      </div>
      <button class="btn btn-primary" @click="infoPage=false" v-translate>Next</button>
    </div>
    <div class="rateLimit-error" v-if = "ratelimitError!=null">
      <div class="error">
        {{ratelimitError}}
      </div>
    </div>
    <div v-else class="registration">
      <VuePhoneNumberInput v-model="phone" @update="updatePhone" :callingCode="cc" class="phoneInput" />
      <button class="btn btn-primary" @click="requestCode()" v-translate> Request code </button>
    </div>
  </div>
</template>

<script>
import VuePhoneNumberInput from 'vue-phone-number-input';
import 'vue-phone-number-input/dist/vue-phone-number-input.css';
import { mapState } from 'vuex';
import checkRegistrationStatus from '@/helpers/registrationStatus'


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
      if(typeof this.cc!="undefined")
      this.$store.dispatch("requestCode",this.cc);
    },
    updatePhone(e){
      this.cc=e.formattedNumber;
    },
    openExtern(e,url){
      if(this.gui=="ut"){
        e.preventDefault();
        alert(url)
      }
    }
  },
  mounted(){
    var userLang = navigator.language || navigator.userLanguage;
    this.$language.current = userLang;
    document.getElementById("VuePhoneNumberInput_phone_number").focus();
  },
  data() {
    return {
      phone: '',
      cc:"",
      infoPage:true,
    };
  },
  computed: mapState(['gui', 'ratelimitError', 'registrationStatus']),
  watch:{
    cc(){
      document.getElementById("VuePhoneNumberInput_phone_number").focus()
    },
    registrationStatus() {
      checkRegistrationStatus(this.registrationStatus)
    }
  }

}
</script>
<style scoped>
  .info,
  .register{
    display:flex;
    flex-direction: column;
    text-align:center;
  }
  .info{
    position:fixed;
    width:100vw;
    height:100vh;
    top:0px;
    left:0px;
    z-index:12;
    text-align:center;
  }
  h1{
    font-size:1.5rem;
  }
  h2{
    font-size:1.3rem;
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
  .logo{
    margin: 20px auto;
    border-radius: 10px;
  }
  #heart{
    font-size: 2rem;
    color:#2090ea;
  }
  .rateLimit-error{
    width:90%;
    height: 90vh;
    color:red;
    display:flex;
    justify-content:center;
    align-items:center;
  }
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
