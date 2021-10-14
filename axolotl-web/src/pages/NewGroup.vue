<template>
  <div class="new-group">
    <div v-if="!creatingGroup" class="new-group-form">
      <div class="form-group">
        <label for="group-name"><b v-translate>Group name</b></label>
        <input
          id="group-name"
          v-model="newGroupName"
          type="text"
          class="form-control"
          placeholder="Enter group name"
          @change="setGroupName"
        >
      </div>
      <p v-translate>Note, you can't add yourself to a group.</p>
      <button class="btn add-group-members" @click="addMembersModal = true">
        <font-awesome-icon icon="plus" /> <span v-translate>Members</span>
      </button>
      <button class="btn create-group" @click="createGroup">
        <font-awesome-icon icon="check" /> <span v-translate>Create group</span>
      </button>
      <add-group-members-modal
        v-if="addMembersModal"
        :already-added="newGroupMembers"
        @add="addGroupMember"
        @close="addMembersModal = false"
      />
      <div v-for="(m, i) in newGroupMembers" :key="m" class="member row">
        <div class="row col-10">
          <div class="name col-12">
            {{ m.Name }}
          </div>
          <div class="tel col-12">
            {{ m.Tel }}
          </div>
        </div>
        <div class="col-2 rm">
          <button type="button" class="remove btn" @click="removeMember(i)">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
      </div>
    </div>
    <div v-else v-translate class="">Creating group</div>
  </div>
</template>

<script>
import AddGroupMembersModal from "@/components/AddGroupMembersModal.vue";
import { mapState } from "vuex";

export default {
  name: "NewGroup",
  components: {
    AddGroupMembersModal,
  },
  data() {
    return {
      newGroupName: null,
      addMembersModal: false,
      newGroupMembers: [],
      creatingGroup: false,
    };
  },
  computed: mapState(["config"]),
  mounted() {
    this.$store.dispatch("getConfig");
    this.$store.dispatch("getContacts");
  },
  methods: {
    setGroupName() {},
    addGroupMember(groupMember) {
      const found = this.newGroupMembers.find(function (element) {
        return element.Tel === groupMember.Tel;
      });
      if (
        typeof found === "undefined" &&
        groupMember.Tel !== this.config.RegisteredNumber
      )
        this.newGroupMembers.push(groupMember);
    },
    removeMember(i) {
      if (this.newGroupMembers.length > 1)
        this.newGroupMembers = this.newGroupMembers.filter((item, j) => j !== i);
      else this.newGroupMembers = [];
    },
    createGroup() {
      if (this.newGroupName !== null && this.newGroupMembers.length > 0) {
        this.creatingGroup = true;
        const members = [];
        this.newGroupMembers.forEach((m) => {
          if (m.Tel !== this.config.RegisteredNumber) members.push(m.Tel);
        });
        if (members.length > 0)
          this.$store.dispatch("createNewGroup", {
            name: this.newGroupName,
            members: members,
          });
      }
    },
  },
};
</script>
<style scoped>
.new-group {
  margin-top: 30px;
}
.name {
  font-weight: 600;
  font-size: 20px;
}
.member {
  padding: 12px 15px;
  border-bottom: 1px solid #c6c6c6;
}
.rm {
  justify-content: center;
  text-align: center;
}
.remove {
  font-size: 2rem;
}
</style>

<!-- Add "scoped" attribute to limit CSS to this component only -->
