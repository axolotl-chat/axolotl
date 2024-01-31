<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="set-password">
      <h5 v-translate>Info</h5>
      <p v-translate>Setting a password is not advised on devices short in memory.</p>
      <p v-translate>Restart is required!</p>
      <div v-if="passwordError" v-translate class="alert alert-danger" role="alert">
        Passwords don't match
      </div>
      <div v-if="passwordUnsafe" v-translate class="alert alert-danger" role="alert">
        Unsafe password
      </div>
      <div v-if="currentPasswordWrong" v-translate class="alert alert-danger" role="alert">
        Current password is wrong
      </div>
      <div class="form-group">
        <label v-translate for="passwordRepeat" class="text-primary"> New password </label>
        <input
          id="setPassword"
          v-model="password"
          type="password"
          name="password"
          class="form-control"
          :secure-length="7"
        />
      </div>
      <div class="form-group">
        <label v-translate for="passwordRepeat" class="text-primary"> Repeat password </label>
        <input
          id="passwordRepeat"
          v-model="passwordRepeat"
          required
          type="password"
          name="passwordRepeat"
          class="form-control"
        />
      </div>
      <div class="form-group">
        <label v-translate for="passwordCurrent" class="text-primary"> Current password </label>
        <input
          id="passwordCurrent"
          v-model="passwordCurrent"
          required
          type="password"
          name="passwordCurrent"
          class="form-control"
        />
      </div>
      <button v-translate class="btn btn-primary" @click="setPassword()">Set password</button>
    </div>
  </component>
</template>

<script>
export default {
  name: "SetPassword",
  components: {},
  data() {
    return {
      password: "",
      passwordCurrent: "",
      currentRepeat: "",
      passwordRepeat: null,
      passwordError: false,
      passwordUnsafe: false,
      currentPasswordWrong: false,
    };
  },
  methods: {
    setPassword() {
      const { password, passwordRepeat } = this;
      if (password.localeCompare(passwordRepeat) !== 0) {
        this.passwordError = true;
      } else if (password.length < 7 && password.length > 0) {
        this.passwordUnsafe = true;
      } else {
        this.$store.dispatch("setPassword", {
          pw: this.password,
          cPw: this.passwordCurrent,
        });
      }
    },
  },
};
</script>
<style scoped>
.set-password {
  display: flex;
  flex-direction: column;
  padding-top: 30px;
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
</style>
<style>
.Password {
  width: 100% !important;
  max-width: 100% !important;
  margin: 0 auto;
}
.Password__group #setPassword {
  display: block;
  width: 100%;
  height: calc(1.5em + 0.75rem + 2px);
  padding: 0.375rem 0.75rem;
  font-size: 1rem;
  font-weight: 400;
  line-height: 1.5;
  color: #495057;
  background-color: #fff;
  background-clip: padding-box;
  border: 1px solid #ced4da;
  border-radius: 0.25rem;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
