<template>
  <div class="settings">
    <div class="profile">
      <div class="avatar">
      </div>
      <div class="name" v-translate> Registered number
      </div>
      <div class="number">
        {{config.RegisteredNumber}}
      </div>
    </div>
    <router-link class="btn btn-primary" :to="'/devices/'" v-translate>
      Linked devices
    </router-link>
    <router-link  class="btn btn-primary" :to="'/setPassword/'" v-translate>
      Set password
    </router-link>

    <button class="btn btn-danger" @click="showConfirmationModal=true" v-translate>
      Unregister
    </button>
    <div class="custom-control custom-switch darkmode-switch">
      <input type="checkbox" v-model="darkMode" class="custom-control-input" id="darkmode-switch" @change="toggleDarkMode()">
      <label class="custom-control-label" for="darkmode-switch" v-translate>Dark mode</label>
    </div>
    <confirmation-modal
    v-if="showConfirmationModal"
    @close="showConfirmationModal=false"
    @confirm="unregister"
    title="Unregister"
    text="Do you really want to unregister? Everything will be deleted!" />
    <div class="about w-100">
      <router-link  class="btn btn-primary" :to="'/about'" v-translate>
        About Axolotl
      </router-link>
    </div>
  </div>
</template>

<script>
import ConfirmationModal from "@/components/ConfirmationModal.vue"
import { mapState } from 'vuex';
export default {
  name: 'settings',
  components: {
    ConfirmationModal
  },
  props: {
    msg: String
  },
  methods:{
    unregister(){
      this.$store.dispatch("unregister");
      localStorage.removeItem("registrationStatus");
    },
    toggleDarkMode(){
      var c = this.getCookie("darkMode")
      if((this.getCookie("darkMode") === 'false'))c = true
      else c = false
      this.$store.dispatch("setDarkMode", c);
    },
    getCookie(cname) {
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
    },
  },
  mounted(){
    this.$store.dispatch("getConfig")
    this.darkMode = (this.getCookie("darkMode") === 'true')
  },
  data() {
    return {
      showConfirmationModal:false,
      darkMode:false,
    };
  },
  computed: mapState(['config'])
}
</script>
<style scoped>
.settings{
  display:flex;
  flex-direction: column;
  justify-content:center;
  align-items: center;
}
.btn{
  margin-bottom: 10px;
}
.profile {
  margin: 40px 0px;
  border-bottom: 1px solid #bbb;
  width: 100%;
  text-align: center;
  padding-bottom: 10px;
}
.name {
  font-weight: bold;
}
.number {
  font-size: 1.8rem;
  color: #2090ea;
}
.about{
  margin-top:20px;
  padding-top:20px;
  border-top: 1px solid #bbb;
  text-align:center;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
