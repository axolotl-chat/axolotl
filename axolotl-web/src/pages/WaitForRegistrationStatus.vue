<template>
  <div><!-- At this point, the loader hasn't been removed by App.vue --></div>
</template>

<script>
import { mapState } from 'vuex';
import checkRegistrationStatus from '@/helpers/registrationStatus'
export default {
  name: 'WaitForRegistrationStatus',
  computed: mapState(['registrationStatus']),
  watch:{
    registrationStatus(oldValue, newValue) {
      checkRegistrationStatus(newValue)
    }
  },
  mounted(){
    this.$store.dispatch("getRegistrationStatus");
  }
}
</script>
