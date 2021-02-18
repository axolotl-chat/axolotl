<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" v-translate>Add a new device</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <ul>
            <li v-translate>1. Open Signal Desktop</li>
            <li v-translate>2. Scan the qr-Code with a qr-Code scanner</li>
            <li v-translate>3. Paste the Qr-Code here:</li>
          </ul>
          <div class="form-group">
            <label for="inputQRCode" v-translate>QR-Code</label>
            <input v-model="qrCode" type="text" class="form-control" id="inputQRCode" >
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-primary" @click="$emit('add',qrCode)" v-translate>Add</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AddDeviceModal',
  data() {
    return {
      qrCode: "",
    }
  },
}
</script>
<style scoped>
.modal {
    display: block;
    border:none;
}
.modal-content {
  border-radius:0px;
}
.modal-header {
  border-bottom:none;
}
.modal-title{
  display:flex;
}
.modal-title > div{
  margin-left:10px;
}
.modal-footer{
  border-top:0px;
}
</style>
