<template>
  <div id="app" @click="checkDebug">
    <header-comp></header-comp>
    <main class="container">
      <error-modal  v-if="error"/>
      <router-view />
    </main>
  </div>
</template>

<script>
window.getCookie = function(cname) {
      var name = cname + "=";
      var ca = document.cookie.split(';');
      for(var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
          c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
          return c.substring(name.length, c.length);
        }
      }
      return false;
}
if (window.getCookie("darkMode") === 'true') {
  import ('./assets/dark.scss');
} else {
  import ('./assets/light.scss');
}
import { router } from './router/router';
import HeaderComp from "@/components/Header.vue"
import ErrorModal from "@/components/ErrorModal.vue"
export default {
  name: 'axolotl-web',
  data() {
    return {
      lastTappedForDebug: new Date(),
      nbTappedForDebug: 0
    }
  },
  components: {
    HeaderComp,
    ErrorModal
  },
  mounted(){
    var userLang = navigator.language || navigator.userLanguage;
    this.$language.current = userLang;

    // If we have a registration status, remove the loader
    if (localStorage.getItem('registrationStatus') != null) {
      let loader = document.getElementById('initial-loader');
      if (loader != undefined) {
        loader.remove();
      }
    }
  },
  computed: {
    error () {
      return this.$store.state.error
    }
  },
  methods: {
    checkDebug() {
      if (this.lastTappedForDebug.getTime() + 1000 > Date.now()) {
        this.nbTappedForDebug++;
      } else {
        this.nbTappedForDebug = 1;
      }
      this.lastTappedForDebug = new Date();
    }
  },
  watch: {
    nbTappedForDebug() {
      if (this.nbTappedForDebug >= 9) {
        router.push('/debug');
      }
    }
  }
}
</script>

<style>
::-webkit-scrollbar {
    display: none;
}
html,
body,
#app {
  height: 100%;
  font-family:"ubuntu";
  display: flex;
  flex-direction: column;
}
main {
  height: calc(100% - 50px); /* This is needed for Ubuntu Touch because the Blink engine is too old (chrome 61). */
  overflow: auto;
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
