<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" v-translate>
            Starting new chats is currently broken in Axolotl.
          </h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <p v-translate>
            New conversations are temporarily disabled. You can create a new
            chat from the linked Desktop Client. Alternatively, your
            counterparty could start the discussion.
          </p>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-primary"
              @click="$emit('close')"
              v-translate
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
  name: "StartChat",
  methods: {},
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
