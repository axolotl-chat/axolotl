<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title" ><span v-translate>Add</span> <div v-if="name!=''">{{name}}</div><div v-else v-translate>Contact</div></h5>
          <button type="button" class="close" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <span><strong v-translate>Only add numbers that you are sure are registered Signal accounts!</strong></span>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label for="inputPhone" v-translate>Name</label>
            <input v-model="name" type="text" class="form-control" id="nameInput">
          </div>
          <div class="form-group">
            <label for="inputPhone" v-translate>Phone</label>
            <input v-model="phone" type="text" class="form-control" id="phoneInput" placeholder="+44...">
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-primary" @click="$emit('add',{'name':name, 'phone':phone})" v-translate>Add</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AddContact',
  data() {
    return {
      phone: "",
      name: ""
    }
  },
}
</script>
<style scoped>
.modal {
    display: block;
    border:none;
}
.modal-content {
  border-radius:0px;
}
.modal-header {
  border-bottom:none;
}
.modal-title{
  display:flex;
}
.modal-title > div{
  margin-left:10px;
}
.modal-footer{
  border-top:0px;
}
</style>
