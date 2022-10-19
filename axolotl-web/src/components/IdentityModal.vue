<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 v-translate class="modal-title">Verify identity</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <div class="qr-code-container">
            <canvas id="qrcode" />
          </div>
          <b><span v-translate>Safety numbers of you and</span>
            {{ SessionNames[currentChat.ID].Name }}:</b>
          <div class="row fingerprint">
            <div
              v-for="(part, i) in fingerprint.numbers"
              :key="'fingerprint_' + i"
              class="col-3"
            >
              {{ part }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              v-translate
              type="button"
              class="btn btn-primary"
              @click="$emit('confirm')"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import QRCode from "qrcode";
import { mapState } from "vuex";
export default {
  name: "IdentityModal",
  emits: ["close", "confirm"],
  data() {
    return {
      errorMessage: null,
    };
  },
  computed: {
    ...mapState(["fingerprint", "SessionNames"]),
    currentChat() {
      return this.$store.state.currentChat;
    },
  },
  watch: {
    fingerprint() {
      QRCode.toCanvas(
        document.getElementById("qrcode"),
        [
          {
            data: this.fingerprint.qrCode,
            mode: "byte",
          },
        ],
        { errorCorrectionLevel: "L" }
      );
    },
  },
  methods: {},
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
.modal-body {
  text-align: left;
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
.qr-code-container {
  width: 100%;
  justify-content: center;
  display: flex;
}
</style>
