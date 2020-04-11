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
    <div class="custom-control darkmode-switch">
      <input type="checkbox" class="custom-control-input" id="darkmode-switch" @change="toggleDarkMode()">
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
    },
    toggleDarkMode(){
      this.$store.dispatch("setDarkMode", !this.darkMode);
    }
  },
  mounted(){
    this.$store.dispatch("getConfig")
  },
  data() {
    return {
      showConfirmationModal:false
    };
  },
  computed: mapState(['config', 'darkMode'])
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
