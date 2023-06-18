<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="set-username">
      <h5 v-translate>Info</h5>
      <p v-translate>Setting a username is required. Please choose one.</p>
      <div class="form-group">
        <label v-translate for="username" class="text-primary"> Username </label>
        <input id="username" v-model="username" required type="text" class="form-control" />
      </div>
      <button v-translate class="btn btn-primary" @click="setUsername()">Set username</button>
    </div>
  </component>
</template>

<script>
export default {
  name: "SetUsername",
  components: {},
  data() {
    return {
      username: "",
    };
  },
  mounted() {
    // To be sure that this page isn't hidden by the loader
    let loader = document.getElementById("initial-loader");
    if (loader !== undefined) {
      loader.remove();
    }
  },
  methods: {
    setUsername() {
      const { username } = this;
      if (username != "") {
        this.$store.dispatch("setUsername", username);
      }
    },
  },
};
</script>
<style scoped>
.set-username {
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

<!-- Add "scoped" attribute to limit CSS to this component only -->
