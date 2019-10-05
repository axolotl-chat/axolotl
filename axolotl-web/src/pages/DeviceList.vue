
<template>
  <div class="deviceList">
    <div v-if="devices.length>1" class="row">
      <div v-for="device in devices" v-if="device.id!=1" v-bind:key="device.id"
        class="col-12 device row">
        <div class="col-10">
          {{device.name}} <br/>
          <div class="meta">
            <span class="lastSeen">Last seen: {{humanifyDate(device.lastSeen)}}</span>
          </div>
        </div>
        <div class="col-2 actions">
          <button class="btn" @click="delDevice(device.id)"><font-awesome-icon icon="trash" /></button>
        </div>
      </div>
    </div>
    <div v-else class="no-entries" >
      No devices available
    </div>
    <button @click="linkDevice" class="btn start-chat"><font-awesome-icon icon="plus" /></button>

  </div>
</template>

<script>
export default {
  name: 'DeviceList',
  props: {
    msg: String
  },
  created(){
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
<style>
.avatar {
    justify-content: center;
    display: flex;
    align-items: center;
}
.badge-name{
  background-color: #2090ea;
  /* padding: 14px; */
  width:50px;
  height:50px;
  border-radius: 50%;
  color: #FFF;
  font-weight: bold;
  text-transform: uppercase;
  font-size: 16px;
  display:flex;
  justify-content: center;
  align-items:center;
}
.meta{
  text-align:left;
}
.meta p{
  margin:0px;
}
.meta .name{
  font-weight:bold;
  font-size:20px;
}
.meta .preview{
  font-size:15px;
}
.row.device{
  border-bottom:1px solid grey;
  border-bottom: 1px solid #c2c2c2;
  padding: 10px;
}
a.chat{
  color:#000;
}
a:hover.chat{
  text-decoration:none;
}
.btn.start-chat {
  position: fixed;
  bottom: 16px;
  right: 10px;
  background-color: #2090ea;
  color: #FFF;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.actions{
  display: flex;
  justify-content: flex-end;
}
.lastSeen{
  font-size:10px;
}
</style>
