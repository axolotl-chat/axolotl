<template>
  <div><!-- At this point, the loader hasn't been removed by App.vue --></div>
</template>

<script>
import { router } from '../router/router';
import { mapState } from 'vuex';

export default {
  name: 'WaitForRegistrationStatus',
  computed: mapState(['registrationStatus']),
  watch:{
    registrationStatus() {
      const registrationStatus = this.registrationStatus;
      localStorage.setItem("registrationStatus", registrationStatus);
      let loader = document.getElementById('initial-loader');
      if (loader != undefined) {
        loader.remove();
      }

      let newRoute;
      if (registrationStatus == "registered") {
        newRoute = "chatList";
      } else if (registrationStatus == "phoneNumber") {
        newRoute = "register";
      } else if (registrationStatus == "verificationCode" || registrationStatus == "pin") {
        newRoute = "verify";
      } else if (registrationStatus == "password") {
        newRoute = "verify";
      }
      router.push('/' + newRoute);
    }
  },
  mounted(){
    this.$store.dispatch("getRegistrationStatus");
  }
}
</script>
