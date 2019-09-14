<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Error</h5>
          <button type="button" class="close" @click="close">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          {{error}}
          <div class="modal-footer">
            <button type="button" class="btn btn-danger" @click="unregister">Unregister</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
</template>

<script>
  export default {
    name: 'AddContact',
    methods: {
      unregister() {
        this.$store.dispatch("unregister");
      },
      close() {
        this.$store.state.error = null
      }
    },
    computed: {
      error() {
        return this.$store.state.error
      }
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
    border: 4px solid red;
    
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