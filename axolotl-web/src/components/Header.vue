<template>
  <div :class="route()+' header '">
    <div class="container">
      <div class="row">
        <div v-if="route()!='chatList' && route()!='register' ">
          <button class="back btn" @click="back()"><font-awesome-icon icon="arrow-left" /></button>
        </div>
        <div v-else class="">
          <button class="btn"><font-awesome-icon icon="ellipsis-v" /></button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Header',
  props: {
    msg: String
  },
  methods:{
    route(){
      return this.$route.name
    },
    back(){
      this.$router.go(-1)
    }
  },
  mounted(){
    window.router = this.$router;
    console.info('App this router:', this.$router)
  }
}
</script>

<style scoped>
  .header{
    position:fixed;
    width:100%;
    height:50px;
    background-color:#2090ea;
    top: 0px;
    z-index:2;
  }
  .btn{
    color:#FFF;

  }
  .back{
    font-weight:bold;
    font-size: 30px;
  }
</style>
