<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 v-translate class="modal-title">Add a new device</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <ul>
            <li v-translate>1. Open Signal Desktop</li>
            <li v-translate>2. Scan the QR code with a QR code scanner</li>
            <li v-translate>3. Paste the QR code here:</li>
          </ul>
          <div class="form-group">
            <label v-translate for="inputQRCode">QR code</label>
            <input id="inputQRCode" v-model="qrCode" type="text" class="form-control" />
          </div>
        </div>
        <div class="modal-footer">
          <button v-translate type="button" class="btn btn-primary" @click="$emit('add', qrCode)">
            Add
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AddDeviceModal',
  emits: ['close', 'add'],
  data() {
    return {
      qrCode: '',
    };
  },
};
</script>
<style scoped>
.modal {
  display: block;
  border: none;
}
.modal-content {
  border-radius: 0px;
}
.modal-header {
  border-bottom: none;
}
.modal-title {
  display: flex;
}
.modal-title > div {
  margin-left: 10px;
}
.modal-footer {
  border-top: 0px;
}
</style>
