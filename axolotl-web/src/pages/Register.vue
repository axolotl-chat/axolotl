<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="register">
      <div v-if="infoPage" class="page1 info">
        <img class="logo" src="/axolotl.png" alt="Axolotl logo">
        <h1 class="title">Axolotl Beta</h1>
        <h2 v-translate class="subtitle">A cross-platform Signal client</h2>
        <div class="description">
          Hey! Mr. Tambourine Man, play a song for me,
          <br>
          In the jingle jangle morning I'll come following you.
          <br>
          It's beta, expect lot's of things not working.
          <br>
          Please be aware:
          <br>
          Registering your phone number with Axolotl will
          <br>
          de-register your existing Signal account and also
          <br>
          de-link your Signal Desktop.
          <br>
          <a
            href="https://axolotl.chat"
            @click="openExtern($event, 'https://axolotl.chat')"
          >
            https://axolotl.chat
          </a>
          <br>
          <font-awesome-icon id="heart" icon="heart" />
        </div>
        <button v-translate class="btn btn-primary" @click="infoPage = false">
          Next
        </button>
      </div>
      <div v-if="rateLimitError !== null" class="rateLimit-error">
        <div class="error">
          {{ rateLimitError }}
        </div>
      </div>
      <div v-else class="registration">
        <VueTelInput
          id="phoneInput"
          mode="international"
          class="phoneInput"
          @input="updatePhone"
        />
        <button v-translate class="btn btn-primary" @click="requestCode()">
          Request code
        </button>
      </div>
    </div>
  </component>
</template>

<script>
import { VueTelInput } from "vue3-tel-input";
import "vue3-tel-input/dist/vue3-tel-input.css";
import { mapState } from "vuex";

export default {
  name: "RegisterPage",
  components: {
    VueTelInput,
  },
  data() {
    return {
      phone: "",
      infoPage: true,
    };
  },
  computed: mapState([
    "gui",
    "rateLimitError",
    "registrationStatus",
    "captchaToken",
    "captchaTokenSent",
  ]),
  mounted() {
    const userLang = navigator.language || navigator.userLanguage;
    this.$language.current = userLang;
    if (this.captchaToken !== null && !this.captchaTokenSent) {
      this.$store.dispatch("sendCaptchaToken");
    }
  },
  methods: {
    requestCode() {
      this.$store.dispatch("requestCode", this.phone.replace(/\s/g, ""));
    },
    updatePhone(e) {
      if(typeof e === "string")
      this.phone = e;
    },
    openExtern(e, url) {
      if (this.gui === "ut") {
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
  margin-top: 50px;
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
