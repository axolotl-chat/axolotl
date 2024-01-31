<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="register">
      <div class="page1 info">
        <img class="logo" src="/axolotl.png" alt="Axolotl logo" />
        <h1 class="title">Axolotl Beta</h1>
        <h2 v-translate class="subtitle">A cross-platform Signal client</h2>
        <div v-if="registrationError" class="alert alert-danger" role="alert">
          {{ registrationError }}
        </div>
        <div class="description">
          <div class="bob">
            Hey! Mr. Tambourine Man, play a song for me,
            <br />
            In the jingle jangle morning I'll come following you.
          </div>
          <div v-translate class="bold mt-3">Please be aware:</div>
          <div v-translate>Registering your phone number with Axolotl will</div>
          <div v-translate>de-register your existing Signal account and also</div>
          <div v-translate>de-link your Signal Desktop.</div>
          <a
            href="https://axolotl.chat"
            class="mt-3"
            @click="openExtern($event, 'https://axolotl.chat')"
          >
            https://axolotl.chat
          </a>
          <br />
          <font-awesome-icon id="heart" icon="heart" />
        </div>
        <button
          v-if="globalConfig?.secondaryRegistration"
          v-translate
          class="btn btn-primary"
          @click="registerAsSecondaryDevice()"
        >
          Register as secondary device (like signal desktop)
        </button>
        <button
          v-if="globalConfig.primaryRegistration"
          v-translate
          class="btn btn-primary"
          @click="register()"
        >
          Register with phone number
        </button>
      </div>
    </div>
  </component>
</template>

<script>
import { mapState } from 'vuex'
import config from '@/config.js'
import { ref } from 'vue'

export default {
  name: 'OnBoarding',
  components: {},
  setup() {
    const globalConfig = ref(config)
    return { globalConfig }
  },
  data() {
    return {}
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
    const userLang = navigator.language || navigator.userLanguage
    this.$language.current = userLang
    if (this.captchaToken !== null && !this.captchaTokenSent) {
      this.$store.dispatch('sendCaptchaToken')
    }
  },
  methods: {
    updatePhone(e) {
      if (typeof e === 'string') this.phone = e
    },
    openExtern(e, url) {
      if (this.gui === 'ut') {
        e.preventDefault()
        // eslint-disable-next-line no-alert
        alert(url)
      }
    },
    registerAsSecondaryDevice() {
      this.$store.dispatch('registerSecondaryDevice')
      this.$router.push('/qr')
    },
    register() {
      window.location = 'https://signalcaptchas.org/registration/generate.html'
    },
  },
}
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
  margin: 10px auto;
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

.bob {
  font-style: italic;
}
.bold {
  font-weight: bold;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
