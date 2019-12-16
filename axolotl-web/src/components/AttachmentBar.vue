<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Send Attachments</h5>
          <button type="button" class="close" @click="close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <button class="btn btn-primary" @click="$emit('send','photo')">Photo</button>
          <button class="btn btn-primary" @click="$emit('send','video')">Video</button>
          <button class="btn btn-primary" @click="$emit('send','audio')">Audio</button>
          <button class="btn btn-primary" @click="$emit('send','document')">Document</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'AddContact',
    methods: {

      close() {
        this.$emit('close');
      }
    },
    computed: {
      gui() {
        return this.$store.state.gui
      },
    }
  }
</script>
<style scoped>
  .modal {
    display: block;
    border: none;
  }

  .modal-content {
    border-radius: 0px;
    /* border: 4px solid red; */

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
  .modal-body{
    color: red;
  }

  .modal-footer {
    border-top: 0px;
  }
</style>
