<template>
<div>
  <main-header title="Linked devices" :backAllowed="true"></main-header>
  <main class="deviceList">
    <!-- eslint-disable vue/no-use-v-if-with-v-for,vue/no-confusing-v-for-v-if -->
    <div class="row device" v-for="device in devices" v-if="device.id!=1" v-bind:key="device.id">
      <div class="col-10">
        <div class="device-name">{{device.name}}</div>
        <div class="meta">
          <span class="lastSeen"><span v-translate>Last seen:</span> {{humanifyDate(device.lastSeen)}}</span>
        </div>
      </div>
      <div class="col-2 actions">
        <button class="btn" @click="delDevice(device.id)"><font-awesome-icon icon="trash" /></button>
      </div>
    </div>
    <div v-if="devices.length == 0" class="no-entries" v-translate>
      No linked devices
    </div>
    <!-- eslint-enable -->
    <button @click="linkDevice" class="btn start-chat"><font-awesome-icon icon="plus" /></button>
  </main>
</div>
</template>

<script>
import MainHeader from "@/components/Header.vue"
export default {
  name: 'DeviceList',
  components: {
    MainHeader
  },
  props: {
    msg: String
  },
  mounted(){
    this.$store.dispatch("getDevices");
  },
  methods:{
    linkDevice() {
      var result = window.prompt("desktopLink");
      this.showSettingsMenu = false;
      this.$store.dispatch("addDevice", result);
    },
    delDevice(id) {
      this.$store.dispatch("delDevice", id);
    },
    humanifyDate(inputDate){
      var now = new Date();
      var date = new Date(inputDate);
      var diff=(now-date)/1000;
      var seconds = diff;
      if(seconds<60)return "now";
      var minutes = seconds/60;
      if(minutes<60)return Math.floor(minutes)+" minutes ago";
      var hours = minutes/60
      if(hours<24)return Math.floor(hours)+" hours ago";
      return date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate() + " " + date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds()
    },
  },
  computed: {
    devices () {
      return this.$store.state.devices
    }
  }
}
</script>
<style scoped>
.device {
  border-bottom: 1px solid #c2c2c2;
  padding: 10px;
}
.actions {
  display: flex;
  justify-content: flex-end;
}
.lastSeen {
  font-size: 10px;
}
.device-name {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
</style>
