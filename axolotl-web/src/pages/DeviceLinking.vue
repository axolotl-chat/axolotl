<template>
  <component :is="$route.meta.layout || 'div'">
    Device linking
    {{ deviceLinkCode }}
    <div>
      <canvas id="qrcode" />
    </div>
  </component>
</template>

<script>
import QRCode from "qrcode";
import { mapState } from "vuex";

export default {
  name: "DeviceLinking",
  computed: mapState(["deviceLinkCode"]),
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
      QRCode.toCanvas(
        document.getElementById("qrcode"),
        [
          {
            data: this.deviceLinkCode,
            mode: "url",
          },
        ],
        { errorCorrectionLevel: "L" }
      );
    },
  },
};
</script>
<style>

</style>

