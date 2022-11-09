<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 v-translate class="modal-title">
            Error communicating with Signal servers
          </h5>
          <button type="button" class="close btn" @click="close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <p v-translate>
            Axolotl encountered an error communicating with Signal servers.
            Please try again.
          </p>
          <p v-translate>
            If you're seeing that error multiple times, check the
            <a href="https://status.signal.org/" target="_blank">status of Signal servers</a>. If it's fine, then please tell us there is a problem by
            <a href="https://github.com/nanu-c/axolotl/issues" target="_blank">opening an issue</a>.
          </p>
          <p v-translate>
            If you think that something is wrong on your side, you can
            <a @click="unregister">unregister</a> and register again.
            Be careful,
            <strong>your encryption key will change and you will lose all your
              messages</strong>
            if you choose to do that.
          </p>
          <div class="modal-footer">
            <button
              v-translate
              type="button"
              class="btn btn-primary"
              @click="close"
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
export default {
  name: "ErrorModal",
  computed: {
    error() {
      return this.$store.state.error;
    },
  },
  methods: {
    unregister() {
      if (
        confirm(
          "You are about to lose all your messages, are you sure you want to unregister?"
        )
      ) {
        this.$store.dispatch("unregister");
        location.reload();
      }
    },
    close() {
      this.$store.state.error = null;
    },
  },
};
</script>
<style scoped>
.modal {
  display: block;
}
.modal-content {
  border-radius: 0px;
  border: 4px solid red;
}
.modal-header {
  border-bottom: none;
}
.modal-body strong {
  color: red;
}
.modal-footer {
  border-top: 0px;
}
</style>
