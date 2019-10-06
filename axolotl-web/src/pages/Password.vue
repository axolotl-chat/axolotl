<template>
  <div class="password">
    <input v-model="pw"  type="password" class="codeInput form-control" />
    <button class="btn btn-primary" @click="sendPassword()"> Decrypt</button>
  </div>
</template>

<script>
export default {
  name: 'password',
  methods:{
    sendPassword(){
      this.$store.dispatch("sendPassword",this.pw);
    },
  },
  data() {
    return {
      pw:""
    };
  },
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
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
