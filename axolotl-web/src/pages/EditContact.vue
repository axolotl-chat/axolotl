<template>
  <component :is="$route.meta.layout || 'div'">
    <template #header>
      <div v-if="profile.Recipient" class="profile-edit-header">
        <span>
          {{
            profile.Recipient.ProfileGivenName !== ''
              ? profile.Recipient.ProfileGivenName
              : profile.Recipient.Username
          }}
        </span>
      </div>
    </template>
    <div v-translate class="mt-4">Update name for contact</div>
    <input
      id="nameInput"
      v-model="name"
      type="text"
      class="form-control"
      placeholder="Enter name"
    />
    <div class="btn btn-primary mt-4" @click="save()">Save</div>
  </component>
</template>

<script>
export default {
  data() {
    return {
      name: '',
    };
  },
  computed: {
    profile() {
      this.updateName();

      return this.$store.state.profile;
    },
  },
  mounted() {
    this.updateName();
  },
  methods: {
    updateName() {
      if (this.profile && this.profile.Recipient) {
        this.name = this.profile.Recipient.ProfileGivenName
          ? this.profile.Recipient.ProfileGivenName
          : this.profile.Recipient.Username;
      }
    },
    save() {
      this.$store.dispatch('updateProfileName', {
        id: this.profile.Recipient.Id,
        name: this.name,
      });
      this.$router.back();
    },
  },
};
</script>

<style scoped>
.profile-edit-header {
  height: 100%;
  display: flex;
  align-items: baseline;
  justify-content: center;
  flex-direction: column;
}
span {
  color: #fff;
  font-size: 18px;
  padding: 0;
}
</style>
