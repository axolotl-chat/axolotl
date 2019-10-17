<template>
  <div class="register">
    <div v-if="infoPage" class="page1 info">
      <img class="logo" src="/axolotl.png" />
      <h1 class="title">Axolotl Alpha</h1>
      <h2 class="subtitle">A cross-plattform signal client</h2>
      <div class="description">Hey! Mr. Tambourine Man, play a song for me,
        <br />
        In the jingle jangle morning I'll come following you.
        <br />
        <a href="https://axolotl.chat"@click="openExtern($event, 'https://axolotl.chat')">https://axolotl.chat</a>
        <br />
        <font-awesome-icon id="heart" icon="heart" />
      </div>
      <button class="btn btn-primary" @click="infoPage=false">Next</button>
    </div>
    <div class="registration">
      <VuePhoneNumberInput v-model="phone" @update="updatePhone" :callingCode="cc" class="phoneInput" />
      <button class="btn btn-primary" @click="requestCode()"> request code </button>
    </div>
  </div>
</template>

<script>
import VuePhoneNumberInput from 'vue-phone-number-input';
import 'vue-phone-number-input/dist/vue-phone-number-input.css';
import { mapState } from 'vuex';

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
    },
    openExtern(e,url){
      if(this.gui=="ut"){
        e.preventDefault();
        alert(url)
      }
    }
  },
  data() {
    return {
      phone: '',
      cc:"",
      infoPage:true,
    };
  },
  computed: mapState(['gui'])

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
    background-color:#FFF;
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
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
