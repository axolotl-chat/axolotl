<template>
  <div class="about">
    <img class="logo" src="/axolotl.png" />
    <h1 class="title">Axolotl Beta {{config.Version}}</h1>
    <h2 class="subtitle" v-translate>A cross-plattform signal client</h2>
    <div class="description">
      <translate>This is a free and open source signal client written in golang and vuejs.</translate>
      <br />
      <translate class="mr-1">
        You can support the development of axolotl either by filling
      </translate>
      <a href="https://github.com/nanu-c/axolotl/issues" @click="openExtern($event, 'https://github.com/nanu-c/axolotl/issues')" v-translate >issues at the bug tracker</a>.
      <br />
      <translate class="mr-1">or by becoming a</translate>
      <a href="https://www.patreon.com/bePatron?u=11219559" @click="openExtern($event, 'https://www.patreon.com/bePatron?u=11219559')"><span v-translate>sponsor on patreon.</span></a><br />
      <br />
      <font-awesome-icon id="heart" icon="heart" />
      <br />
      <a href="https://axolotl.chat" @click="openExtern($event, 'https://axolotl.chat')">https://axolotl.chat</a>
    </div>
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
  data() {
    return {
      showConfirmationModal:false
    };
  },
  methods:{
    openExtern(e, url){
      if(this.gui=="ut"){
        e.preventDefault();
        alert(url)
      }
    }
  },
  computed: mapState(['config', 'gui'])
}
</script>
<style scoped>
.about{
  display:flex;
  flex-direction: column;
  justify-content:center;
  align-items: center;
  text-align:center;
}
.logo{
  margin: 20px auto;
  border-radius: 10px;
}
#heart{
  font-size: 2rem;
  color:#2090ea;
}
h1{
  font-size:1.5rem;
}
h2{
  font-size:1.3rem;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
