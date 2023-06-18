<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="password">
      <div v-if="error" v-translate class="alert alert-danger">Password is wrong</div>
      <input
        id="passwordInput"
        v-model="pw"
        type="password"
        class="codeInput form-control"
        @keydown="checkEnter($event)"
      />
      <button v-translate class="btn btn-primary" @click="sendPassword">Decrypt</button>
      <button v-if="error" v-translate class="btn btn-danger" @click="unregister">
        Unregister
      </button>
    </div>
  </component>
</template>

<script>
import { mapState } from "vuex";

export default {
  name: "PasswordPage",
  data() {
    return {
      pw: "",
    };
  },
  computed: {
    ...mapState(["registrationStatus"]),
    error() {
      return this.$store.state.loginError;
    },
  },
  mounted() {
    document.getElementById("passwordInput").focus();
  },
  methods: {
    sendPassword() {
      this.$store.dispatch("sendPassword", this.pw);
      this.pw = null;
    },
    checkEnter(e) {
      if (e.keyCode === 13) this.sendPassword();
    },
    unregister() {
      this.$store.dispatch("unregister");
    },
  },
};
</script>
<style scoped>
.password {
  display: flex;
  flex-direction: column;
}
.codeInput {
  margin-top: 30px;
}
.btn {
  max-width: 300px;
  margin: auto;
  margin-top: auto;
  margin-top: 50px;
}
.alert {
  border: none;
  border-radius: 0px;
  margin-top: 20px;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
