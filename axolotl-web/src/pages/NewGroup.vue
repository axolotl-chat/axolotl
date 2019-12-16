<template>
  <div class="new-group">
    <div v-if="!creatingGroup" class="new-group-form">
      <div class="form-group">
        <label for="group-name"><b>Group name</b></label>
        <input type="text" v-model="newGroupName" @change="setGroupName" class="form-control" id="group-name" placeholder="Enter group name">

      </div>
      <p>Note, you can't add yourself to a group.</p>
      <button class="btn add-group-members" @click="addMembersModal=true">
        <font-awesome-icon icon="plus" /> Members</button>
      <button class="btn create-group" @click="createGroup">
        <font-awesome-icon icon="check" /> Create group
      </button>
      <add-group-members-modal  v-if="addMembersModal"
      :allreadyAdded="newGroupMembers"
      @add="addGroupMemeber"
      @close="addMembersModal=false"/>
      <div class="member row" v-for="(m, i) in newGroupMembers"
      v-bind:key="m">
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
    <div v-else class="">
      Creating Group
    </div>
  </div>
</template>

<script>
import AddGroupMembersModal from "@/components/AddGroupMembersModal.vue"
import { mapState } from 'vuex';

export default {
  name: 'newGroup',
  components: {
    AddGroupMembersModal
  },
  props: {
    msg: String
  },
  mounted(){
    this.$store.dispatch("getConfig")
    this.$store.dispatch("getContacts")
  },
  methods:{
    setGroupName(){
    },
    addGroupMemeber(groupMember){
      var found = this.newGroupMembers.find(function(element) {
          return element.Tel == groupMember.Tel;
      });
      if(typeof found =="undefined"&&groupMember.Tel != this.config.RegisteredNumber)
      this.newGroupMembers.push(groupMember)
    },
    removeMember(i){
      if(this.newGroupMembers.length>1)
      this.newGroupMembers= this.newGroupMembers.filter((item,j)=>j!=i)
      else this.newGroupMembers = [];
    },
    createGroup(){
      if(this.newGroupName!=null&&this.newGroupMembers.length>0){
        this.creatingGroup = true;
        var members = [];
        this.newGroupMembers.forEach(m=>{
          if(m.Tel!=this.config.RegisteredNumber)
          members.push(m.Tel)
        })
        if(members.length>0)
        this.$store.dispatch("createNewGroup",{
          name: this.newGroupName,
          members: members
        })
      }
    }
  },
  data() {
    return {
      newGroupName:null,
      addMembersModal:false,
      newGroupMembers:[],
      creatingGroup:false
    };
  },
  computed: mapState(['config'])
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
