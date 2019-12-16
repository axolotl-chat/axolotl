<template>
  <div id="app">
    <header-comp></header-comp>
    <div class="container">
      <error-modal  v-if="error"/>
      <router-view />
    </div>
  </div>

</template>
<script>
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import HeaderComp from "@/components/Header.vue"
import ErrorModal from "@/components/ErrorModal.vue"
export default {
  name: 'axolotl-web',
  components: {
    HeaderComp,
    ErrorModal
  },
  mounted(){
  },
  computed: {
    error () {
      return this.$store.state.error
    }
  }

}
</script>
<style>
::-webkit-scrollbar {
    display: none;
}
#app{
  font-family:"ubuntu"
}
#app >.container{
  padding-top: 50px;
  position:relative;
}
.btn:focus{
  box-shadow:none;
}
.btn{
  border-radius:0px;
}
.btn-primary {
  background-color: #2090ea;
}
.no-entries {
    height: 90vh;
    display: flex;
    justify-content: center;
    align-items: center;
}
</style>
