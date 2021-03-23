<template>
  <div class="verify">
    <h3 v-translate>Enter your registration pin</h3>
    <div
      v-if="verificationError == 'RegistrationLockFailure' || requestPin"
      class="verify"
    >
      <p v-translate>or disable it on Android/IOs</p>
      <input v-model="pin" type="text" />
      <button v-translate class="btn btn-primary" @click="sendPin()">
        Send pin
      </button>
    </div>
    <div v-if="!requestPin" class="verify">
      <VueOtpBox v-on:inputValue="updateCode($event)" class="codeInput" />
      <button
        v-translate
        :disabled="inProgress"
        class="btn btn-primary"
        @click="sendCode()"
      >
        Send code
      </button>
    </div>
    <div
      v-if="inProgress && verificationError == null && !requestPin"
      class="spinner"
    >
      <div class="spinner-border" role="status">
        <span v-translate class="sr-only">Loading...</span>
      </div>
    </div>
    <div v-if="verificationError == 404">
      <div v-translate>Wrong code entered. Restart for another try.</div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import VueOtpBox from "@/components/VueOtpBox";
export default {
  name: "Verification",
  components: {
    VueOtpBox,
  },
  props: {
    msg: String,
  },
  data() {
    return {
      code: "",
      pin: "",
      inProgress: false,
    };
  },
  computed: mapState(["verificationError", "requestPin", "registrationStatus"]),
  mounted() {
  },
  methods: {
    updateCode(code){
      this.code = code
    },
    sendCode() {
      if (this.code.length == 6) {
        this.$store.dispatch("sendCode", this.code);
        this.inProgress = true;
      }
    },
    sendPin() {
      if (this.code.length == 6) {
        this.$store.dispatch("sendPin", this.pin);
        this.inProgress = true;
      }
    },
  },
};
</script>
<style>
.verify {
  display: flex;
  flex-direction: column;
  padding-top: 30px;
}
.verify h3 {
  text-align: center;
}
.verify .codeInput {
  margin-top: 30px;
}
.verify .btn {
  max-width: 300px;
  margin: auto;
  margin-top: auto;
  margin-top: 50px;
}

.verify
  .ofcold__security-code-wrapper
  .ofcold__security-code-field
  .form-control {
  border: 2px solid #2090ea !important;
}
.verify .spinner {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
