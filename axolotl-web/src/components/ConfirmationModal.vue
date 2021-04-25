<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">{{ title }}</h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          {{ text }}
          <div class="modal-footer">
            <button
              v-translate
              type="button"
              class="btn btn-primary"
              @click="$emit('confirm')"
            >
              Confirm
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "ConfirmModal",
  props: {
    title: String,
    text: String,
  },
  methods: {},
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
.modal-body {
  text-align: left;
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
