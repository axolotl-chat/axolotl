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
          <b v-translate>Your key </b>
          <br />
          <div >{{identity.me}}</div>
          <br />
          <br />
          <b>{{currentChat.Name}}<span v-translate>'s key </span></b>
          <br />
          <div >{{identity.their}}</div>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" @click="$emit('confirm')" v-translate>Close</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'IdentityModal',
    methods: {
    },
    computed: {
      identity() {
        return this.$store.state.identity
      },
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
</style>
