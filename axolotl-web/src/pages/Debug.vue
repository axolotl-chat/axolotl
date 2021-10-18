<template>
  <div>
    <h1 v-translate>Debug screen</h1>
    <p v-translate class="warning-box">For developers only, use with care!</p>
    <p style="text-align: center">
      <a v-translate href="/" class="btn btn-primary">Exit</a>
    </p>
    <section>
      <h2 v-translate>Registration data</h2>
      <p v-translate>
        Current registration status in local storage:
        {{ localRegistrationStatus }}
      </p>
      <p v-translate>Current registration status in $store: {{ registrationStatus }}</p>
      <button v-translate @click="clearRegistrationFromLocalStorage">
        Clear registration from localStorage
      </button>
    </section>
    <button v-translate @click="clearLocalStorage">Clear all localStorage</button>
  </div>
</template>

<script>
import { mapState } from "vuex";
export default {
  name: "Debug",
  computed: {
    localRegistrationStatus() {
      return localStorage.getItem("registrationStatus");
    },
    ...mapState(["registrationStatus"]),
  },
  mounted() {
    // To be sure that this page isn't hidden by the loader
    let loader = document.getElementById("initial-loader");
    if (loader !== undefined) {
      loader.remove();
    }
  },
  methods: {
    clearRegistrationFromLocalStorage() {
      localStorage.removeItem("registrationStatus");
    },
    clearLocalStorage() {
      localStorage.clear();
    },
  },
};
</script>
