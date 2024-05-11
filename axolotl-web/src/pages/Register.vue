<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="register">
      <div v-if="rateLimitError !== null" class="rateLimit-error">
        <div class="error">
          {{ rateLimitError }}
        </div>
      </div>
      <div v-if="registrationError !== null" class="registration-error">
        <div class="error">
          {{ registrationError }}
        </div>
      </div>
      <div v-else class="registration">
        <VueTelInput id="phoneInput" mode="international" class="phoneInput" @input="updatePhone" />
        <button v-translate class="btn btn-primary mt-3" @click="requestCode()">
          Request code
        </button>
      </div>
    </div>
  </component>
</template>

<script>
import { VueTelInput } from 'vue3-tel-input';
import 'vue3-tel-input/dist/vue3-tel-input.css';
import { mapState } from 'vuex';

export default {
  name: 'RegisterPage',
  components: {
    VueTelInput,
  },
  data() {
    return {
      phone: '',
    };
  },
  computed: mapState([
    'gui',
    'rateLimitError',
    'registrationStatus',
    'captchaToken',
    'captchaTokenSent',
    'registrationError',
  ]),
  mounted() {
    const userLang = navigator.language || navigator.userLanguage;
    this.$language.current = userLang;
    if (this.captchaToken !== null && !this.captchaTokenSent) {
      this.$store.dispatch('sendCaptchaToken');
    }
  },
  methods: {
    requestCode() {
      this.$store.dispatch('requestCode', this.phone.replace(/\s/g, ''));
    },
    updatePhone(e) {
      if (typeof e === 'string') this.phone = e;
    },
    openExtern(e, url) {
      if (this.gui === 'ut') {
        e.preventDefault();
        alert(url);
      }
    },
  },
};
</script>
<style scoped>
.info,
.register {
  display: flex;
  flex-direction: column;
  text-align: center;
}
.info {
  position: fixed;
  width: 100vw;
  height: 100vh;
  top: 0px;
  left: 0px;
  z-index: 12;
  text-align: center;
}
h1 {
  font-size: 1.5rem;
}
h2 {
  font-size: 1.3rem;
}
.phoneInput {
  margin-top: 30px;
}
.btn {
  max-width: 300px;
  margin: auto;
  margin-top: auto;
  color: #fff;
}
.logo {
  margin: 20px auto;
  border-radius: 10px;
}
#heart {
  font-size: 2rem;
  color: #2090ea;
}
.rateLimit-error {
  width: 90%;
  height: 90vh;
  color: red;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
