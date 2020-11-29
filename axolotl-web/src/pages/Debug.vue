<template>
  <div>
    <h1>For developers only, use with care!</h1>
    <section>
      <h2>Registration data</h2>
      <p>Current registration status in local storage: {{localRegistrationStatus}}</p>
      <p>Current registration status in $store: {{registrationStatus}}</p>
      <button @click="clearRegistrationFromLocalStorage">Clear registration from localStorage</button>
    </section>
    <button @click="clearLocalStorage">Clear all localStorage</button>
  </div>
</template>

<script>
import { mapState } from 'vuex';
export default {
  name: 'debug',
  methods:{
    clearRegistrationFromLocalStorage(){
      localStorage.removeItem('registrationStatus');
    },
    clearLocalStorage(){
      localStorage.clear();
    }
  },
  mounted(){
    // To be sure that this page isn't hidden by the loader
    let loader = document.getElementById('initial-loader');
    if (loader != undefined) {
      loader.remove();
    }
  },
  computed: {
    localRegistrationStatus() {
      return localStorage.getItem('registrationStatus');
    },
    ...mapState(['registrationStatus'])
  }
}
</script>
