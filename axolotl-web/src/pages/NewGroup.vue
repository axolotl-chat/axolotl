<template>
  <div class="new-group">
    <div class="form-group">
      <label for="group-name">Group name</label>
      <input type="text" v-model="newGroupName" @change="setGroupName" class="form-control" id="group-name" placeholder="Enter group name">

    </div>
    <button class="btn add-group-members" @click="addMembersModal=true"><font-awesome-icon icon="plus" /> Members</button>
    <add-group-members-modal  v-if="addMembersModal"
    :allreadyAdded="newGroupMembers"
    @add="addGroupMemeber"
    @close="addMembersModal=false"/>
    <div class="member row" v-for="(m, i) in newGroupMembers">
      <div class="row col-10">
        <div class="name col-12">
          {{m.Name}}
        </div>
        <div class="tel col-12">
          {{m.Tel}}
        </div>
      </div>
      <div class="col-2 rm">
        <button type="button" class="remove btn" @click="removeMember(i)">
          <span aria-hidden="true">&times;</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import AddGroupMembersModal from "@/components/AddGroupMembersModal.vue"

export default {
  name: 'newGroup',
  components: {
    AddGroupMembersModal
  },
  props: {
    msg: String
  },
  methods:{
    setGroupName(){
      this.$store.dispatch("setNewGroupName", this.newGroupName)
    },
    addGroupMemeber(groupMember){
      var found = this.newGroupMembers.find(function(element) {
          return element.Tel == groupMember.Tel;
      });
      if(typeof found =="undefined")
      this.newGroupMembers.push(groupMember)
    },
    removeMember(i){
      if(this.newGroupMembers.length>1)
      this.newGroupMembers= this.newGroupMembers.filter((item,j)=>j!=i)
      else this.newGroupMembers = [];
    }
  },
  data() {
    return {
      newGroupName:null,
      addMembersModal:false,
      newGroupMembers:[]
    };
  },
}
</script>
<style scoped>
  .new-group{
    margin-top:30px;
  }
  .name{
    font-weight:600;
    font-size:20px;
  }
  .member{
    padding: 12px 15px;
    border-bottom: 1px solid #c6c6c6;
  }
  .rm{
    justify-content:center;
    text-align:center;
  }
  .remove{
    font-size: 2rem;
  }
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
