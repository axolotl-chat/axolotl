<template>
  <div class="settings">
    <div class="profile">
      <div class="avatar">
      </div>
      <div class="name"> Registered number
      </div>
      <div class="number">
        {{config.RegisteredNumber}}
      </div>
    </div>
    <router-link class="btn" :to="'/devices/'" @click="showSettingsMenu=false">
      Linked devices
    </router-link>
    <router-link  class="btn" :to="'/setPassword/'" @click="showSettingsMenu=false">
      Set password
    </router-link>
    <button class="btn btn-danger" @click="unregister">Unregister</button>

  </div>
</template>

<script>
import { mapState } from 'vuex';
export default {
  name: 'settings',
  components: {
  },
  props: {
    msg: String
  },
  methods:{
  },
  mounted(){
    this.$store.dispatch("getConfig")
  },
  data() {
    return {
      cc:""
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
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
