<template>
  <component :is="$route.meta.layout || 'div'">
    <div class="device-linking-page">
      <div v-translate>Please scan the QR code with your primary signal device.</div>
      <div class="d-flex justify-content-center align-items-center">
        <canvas id="qrcode" />
      </div>
    </div>
  </component>
</template>

<script>
import QRCode from 'qrcode';
import { mapState } from 'vuex';

export default {
  name: 'DeviceLinking',
  computed: mapState(['deviceLinkCode']),
  watch: {
    deviceLinkCode() {
      this.updateQrCode();
    },
  },
  mounted() {
    this.updateQrCode();
  },
  methods: {
    updateQrCode() {
      if (this.deviceLinkCode) {
        QRCode.toCanvas(
          document.getElementById('qrcode'),
          [
            {
              data: this.deviceLinkCode,
              mode: 'url',
            },
          ],
          { errorCorrectionLevel: 'L' },
        );
      }
    },
  },
};
</script>
<style scoped>
.device-linking-page {
  padding: 1rem;
  text-align: center;
}
</style>
