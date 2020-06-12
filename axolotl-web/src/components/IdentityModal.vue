<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" v-translate>Verify identity</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <div class="qr-code-container">
            <canvas id="qrcode"></canvas>
          </div>
          <b><span v-translate>Safety numbers of you and</span> {{currentChat.Name}}:</b>
          <div class="row fingerprint">
            <div class="col-3" v-for="(part,i) in fingerprint" v-bind:key="'fingerprint_'+i">
                {{ part }}
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" @click="$emit('confirm')" v-translate>Close</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  import QRCode from 'qrcode'
  import { mapState } from 'vuex';
  export default {
    name: 'IdentityModal',
    methods: {
    },
    data() {
      return {
        errorMessage:null
      }
    },
    watch:{
      fingerprint(){
        QRCode.toCanvas(document.getElementById('qrcode'), this.fingerprint, function (error) {
            if (error) this.errorMesssage = error;//console.error(error)
            // console.log('success!');
          })
      }
    },
    computed: {
      ...mapState(['fingerprint']),
      currentChat() {
        return this.$store.state.currentChat
      },
    },
  }
</script>
<style scoped>
  .modal {
    display: block;
    border: none;
  }

  .modal-content {
    border-radius: 0px;
  }
  .modal-body{
    text-align: left;
  }
  .modal-header {
    border-bottom: none;
  }

  .modal-title {
    display: flex;
  }

  .modal-title>div {
    margin-left: 10px;
  }

  .modal-footer {
    border-top: 0px;
  }
  .qr-code-container{
    width:100%;
    justify-content: center;
    display: flex;
  }
</style>
