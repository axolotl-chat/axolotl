<template>
  <div id="import-unavailable-modal" class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" v-translate>Import contacts feature unavailable</h5>
        </div>
        <div class="modal-body">
          <span v-translate>
            We are sorry, but due to a change in the Signal protocol, automatic import of contacts is currently unavailble.
          </span>
        </div>
        <div class="modal-body">
          <span v-translate>
            If you are really sure that the persons you want to text with have Signal, you can manually add them using the "plus" action on that page.
          </span>
        </div>
        <div class="modal-body">
          <span v-translate>
            Be careful, if you add too many numbers which aren't Signal accounts, you won't be able to add contacts anymore.
          </span>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-primary" @click="$emit('close')" v-translate>Close</button>
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
