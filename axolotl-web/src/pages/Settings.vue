<template>
  <div class="settings">
    <div class="profile">
      <div class="avatar" />
      <div v-translate class="name">Registered number</div>
      <div class="number">
        {{ config.RegisteredNumber }}
      </div>
    </div>
    <router-link v-translate class="btn btn-primary" :to="'/devices/'">
      Linked devices
    </router-link>
    <router-link v-translate class="btn btn-primary" :to="'/setPassword/'">
      Set password
    </router-link>

    <button
      v-translate
      class="btn btn-danger"
      @click="showConfirmationModal = true"
    >
      Unregister
    </button>
    <div class="custom-control custom-switch darkmode-switch">
      <input
        id="darkmode-switch"
        v-model="darkMode"
        type="checkbox"
        class="custom-control-input"
        @change="toggleDarkMode()"
      >
      <label v-translate class="custom-control-label" for="darkmode-switch">Dark mode</label>
    </div>
    <confirmation-modal
      v-if="showConfirmationModal"
      title="Unregister"
      text="Do you really want to unregister? Everything will be deleted!"
      @close="showConfirmationModal = false"
      @confirm="unregister"
    />
    <div class="about w-100">
      <router-link v-translate class="btn btn-primary" :to="'/about'">
        About Axolotl
      </router-link>
    </div>
    <div class="warning-box">
      <span v-translate>
        Due to technical limitations, Axolotl doesn't support push
        notifications. Keep the app open to be notified in real time. In Ubuntu
        Touch, use UT Tweak Tool to set Axolotl on "Prevent app suspension".
      </span>
    </div>
  </div>
</template>

<script>
import ConfirmationModal from "@/components/ConfirmationModal.vue";
import { mapState } from "vuex";
export default {
  name: "Settings",
  components: {
    ConfirmationModal,
  },
  data() {
    return {
      showConfirmationModal: false,
      darkMode: false,
    };
  },
  computed: mapState(["config"]),
  mounted() {
    this.$store.dispatch("getConfig");
    this.darkMode = this.getCookie("darkMode") === "true";
  },
  methods: {
    unregister() {
      this.$store.dispatch("unregister");
      localStorage.removeItem("registrationStatus");
    },
    toggleDarkMode() {
      let c = this.getCookie("darkMode");
      if (this.getCookie("darkMode") === "false") c = true;
      else c = false;
      this.$store.dispatch("setDarkMode", c);
    },
    getCookie(cname) {
      const name = cname + "=";
      const ca = document.cookie.split(";");
      for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === " ") {
          c = c.substring(1);
        }
        if (c.indexOf(name) === 0) {
          return c.substring(name.length, c.length);
        }
      }
      return false;
    },
  },
};
</script>
<style scoped>
.settings {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
.btn {
  margin-bottom: 10px;
}
.profile {
  margin: 40px 0px;
  border-bottom: 1px solid #bbb;
  width: 100%;
  text-align: center;
  padding-bottom: 10px;
}
.name {
  font-weight: bold;
}
.number {
  font-size: 1.8rem;
  color: #2090ea;
}
.about {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #bbb;
  text-align: center;
}
.warning-box {
  margin-top: 0.5rem;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
