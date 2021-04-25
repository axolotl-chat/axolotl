<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            <span v-translate>Edit</span>
            <div v-if="contact.Name !== ''">{{ contact.Name }}</div>
            <div v-else v-translate>Contact</div>
          </h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <input
              id="nameInput"
              v-model="contact.Name"
              type="text"
              class="form-control"
              placeholder="Enter name"
            >
          </div>
          <div class="form-group">
            <label v-translate for="nameInput">Phone</label>
            <input
              id="phoneInput"
              v-model="contact.Tel"
              type="text"
              class="form-control"
              placeholder="+44..."
            >
          </div>
        </div>
        <div class="modal-footer">
          <button
            v-translate
            type="button"
            class="btn btn-primary"
            @click="$emit('save', { id: id, contact: contact })"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "AddContact",
  props: {
    contact: Object,
    id: String,
  },
  data() {
    return {
      phone: "",
      name: "",
    };
  },
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
</style>
