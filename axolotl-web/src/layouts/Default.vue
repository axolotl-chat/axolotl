<template>
  <div class="base-layout">
    <header :class="currentRoute + ' header fixed-top'">
      <div class="container">
        <div class="header-row row">
          <div class="col-2 header-left">
            <button
              v-if="$route.meta.hasBackButton"
              class="btn"
              @click="back()"
            >
              <font-awesome-icon icon="arrow-left" />
            </button>
          </div>
          <div class="col-8 header-center">
            <slot name="header" />
          </div>
          <div class="col-2 header-right justify-content-end d-flex">
            <div class="dropdown">
              <button
                id="dropdownMenuButton"
                class="btn btn-primary"
                type="button"
                data-bs-toggle="dropdown"
                aria-expanded="false"
              >
                <font-awesome-icon icon="ellipsis-v" />
              </button>
              <ul class="dropdown-menu" aria-labelledby="dropdownMenuButton">
                <slot name="menu" />
              </ul>
            </div>
          </div>
        </div>
      </div>
    </header>
    <main class="container">
      <slot />
    </main>
    <footer>
      <slot name="footer" />
    </footer>
  </div>
</template>
<script>
export default {
  name: "DefaultLayout",
  data: () => ({
    showMenu: false,
  }),
  computed: {
    currentRoute() {
      return this.$route.name;
    },
  },
  methods: {
    back() {
      this.$router.back();
    },
    toggleMenu() {
      this.showMenu = !this.showMenu;
    },
  },
};
</script>
<style scoped>
.header {
  padding: 5px 0;
  background-color: #2090ea;
  z-index: 2;
  box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
  min-height: 49px;
}
#dropdownMenuButton {
  color: #fff;
}
.dropdown-menu {
  border: 1px solid #2090ea;
}
main {
  padding-top: 50px;
}
</style>