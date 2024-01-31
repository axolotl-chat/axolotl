<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="settings">
      <div class="profile">
        <div class="avatar" />
        <div v-if="config.e164">
          <div v-translate class="name">Registered number</div>
          <div class="number">
            {{ config.e164 }}
          </div>
        </div>
      </div>
      <!-- <router-link v-translate class="btn btn-primary" :to="'/devices/'">
        Linked devices
      </router-link>
      <router-link v-translate class="btn btn-primary" :to="'/setPassword/'">
        Set password
      </router-link> -->

      <!-- <div class="custom-control form-check custom-switch darkmode-switch g-2">
        <input
          id="darkmode-switch"
          v-model="darkMode"
          type="checkbox"
          class="form-check-input"
          @change="toggleDarkMode()"
        />
        <label v-translate class="form-check-label" for="darkmode-switch">
          Dark mode
        </label>
      </div> -->
      <!-- <div class="row g-3 mt-1 align-items-center">
        <div class="col-auto">
          <select
            v-model="loglevel"
            class="form-select"
            aria-label="Loglevel select"
            @change="setLogLevel($event)"
          >
            <option v-translate value="info">Info</option>
            <option v-translate value="warn">Warnings</option>
            <option v-translate value="error">Errors</option>
            <option v-translate value="panic">No logs</option>
            <option v-translate value="debug">Debugging</option>
          </select>
        </div>
        <div class="col-auto">
          <label v-translate for="loglevel" class="col-form-label"> Loglevel </label>
        </div>
      </div> -->
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
      <button v-translate class="btn btn-danger" @click="showConfirmationModal = true">
        Unregister
      </button>
      <div class="warning-box">
        <span v-translate>
          Due to technical limitations, Axolotl doesn't support push notifications. Keep the app
          open to be notified in real time. In Ubuntu Touch, use UT Tweak Tool to set Axolotl on
          "Prevent app suspension".
        </span>
      </div>
    </div>
  </component>
</template>

<script>
import ConfirmationModal from "@/components/ConfirmationModal.vue";
import { mapState } from "vuex";
export default {
  name: "SettingsPage",
  components: {
    ConfirmationModal,
  },
  data() {
    return {
      showConfirmationModal: false,
      darkMode: false,
      loglevel: "info",
    };
  },
  computed: mapState(["config"]),
  mounted() {
    this.$store.dispatch("getConfig");
    this.darkMode = this.getCookie("darkMode") === "true";
    this.loglevel = this.config.LogLevel;
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
    setLogLevel(e) {
      this.$store.dispatch("setLogLevel", e.target.value);
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
