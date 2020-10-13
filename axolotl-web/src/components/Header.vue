<template>
  <header>
    <div class="actions-container">
      <div class="action" v-if="backAllowed" @click="back()">
        <font-awesome-icon icon="arrow-left" />
      </div>
      <h1>{{title}}</h1>
    </div>
    <div class="actions-container">
      <slot />
      <div class="dropdown" v-if="globalMenuItems != undefined">
        <div class="action" @click="showMenu = !showMenu">
          <font-awesome-icon icon="ellipsis-v" />
        </div>
        <nav v-if="showMenu" class="dropdown-menu">
          <router-link v-for="item in globalMenuItems" :key="item.label" class="dropdown-item" :to="item.url">
            {{item.label}}
          </router-link>
        </nav>
      </div>
    </div>
  </header>
</template>

<script>
export default {
  name: "Header",
  data() {
    return {
      showMenu: false
    }
  },
  props: {
    title: String,
    backAllowed: Boolean,
    globalMenuItems: Array
  },
  methods: {
    back() {
      this.$router.go(-1);
    }
  }
};
</script>

<style lang="scss" scoped>
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  color: #fff;
  background-color: #2090ea;
  z-index: 2;
  box-shadow: 0px -11px 14px 7px rgba(0, 0, 0, 0.75);
}
h1 {
  font-size: 1.6rem;
  margin: 0;
  padding: 0 8px;
}
.action {
  cursor: pointer;
  padding: 0 8px;
}
.actions-container {
  display: flex;
  align-items: center;
}
// Temp
.dropdown-menu {
  display: block;
  right: 0;
  left: unset;
}
</style>
