<template>
  <div id="app">
    <header-comp></header-comp>
    <div class="container">
      <router-view />
    </div>
  </div>

</template>
<script>
import store from './store/store'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import HeaderComp from "@/components/Header.vue"
export default {
  name: 'axolotl-web',
  components: {
    HeaderComp
  },
  mounted(){
    // document.getElementsByClassName("header")[0].innerHTML=
    // "Your screen resolution is: " + screen.width + "x" + screen.height+"<br>"+navigator.userAgent;
    // const viewportmeta = document.querySelector('meta[name=viewport]');
    // viewportmeta.setAttribute('content', "initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0");
  }
}
</script>
<style>
#app{
  padding-top:50px;

}
</style>
