<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 v-translate class="modal-title">Send attachments</h5>
          <button type="button" class="close btn" @click="close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <button v-translate class="btn btn-primary" @click="$emit('send', 'photo')">Photo</button>
          <button v-translate class="btn btn-primary" @click="$emit('send', 'video')">Video</button>
          <button v-translate class="btn btn-primary" @click="$emit('send', 'audio')">Audio</button>
          <button v-translate class="btn btn-primary" @click="$emit('send', 'document')">
            Document
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AttachmentBarComponent',
  emits: ['send', 'close'],
  methods: {
    close() {
      this.$emit('close');
    },
  },
};
</script>
<style scoped lang="scss">
.modal {
  display: block;
  border: none;

  .modal-content {
    border-radius: 0px;
    min-height: 50vh;
  }

  .modal-header {
    border-bottom: none;
  }

  .modal-title {
    display: flex;

    & > div {
      margin-left: 10px;
    }
  }

  .modal-body {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-direction: column;

    .btn {
      color: #fff;
    }
  }

  .modal-footer {
    border-top: 0px;
  }
}
</style>
