<template>
  <div id="import-unavailable-modal" class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" v-translate>Import contacts feature unavailable</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <span v-translate>We are sorry, but due to a change in the Signal protocol, automatic import of contacts is currently unavailble.</span>
          <span v-translate>If you are <strong>really sure that the persons you want to text with have Signal</strong>, you can manually add them using the "plus" action on that page.<br/>
          Be careful, if you add too many numbers which aren't Signal accounts, you won't be able to add contacts anymore.</span>
          <div class="modal-footer">
            <button type="button" class="btn btn-primary" @click="$emit('close')" v-translate>Close</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
  export default {
    name: 'ImportUnavailableModal'
  }
</script>

<style>
#import-unavailable-modal{
  display: block;
}
.modal-footer button {
  margin: auto;
}
</style>
