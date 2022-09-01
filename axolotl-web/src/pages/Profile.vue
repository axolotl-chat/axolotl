<template>
  <component :is="$route.meta.layout || 'div'">
    <template #header>
      <div class="profile-header" v-if="profile.Recipient">
        <span>{{
          profile.Recipient.ProfileGivenName !== ""
            ? profile.Recipient.ProfileGivenName
            : profile.Recipient.Username
        }}</span>
      </div>
    </template>
    <template #menu>
      <div class="ms-2">
        <div v-translate @click="editContact()">Edit Contact</div>
      </div>
    </template>
    <div class="profile" v-if="profile.Recipient">
      <div class="profile-image">
        <img
          v-if="hasProfileImage"
          class="avatar-img"
          :src="'http://localhost:9080/avatars?recipient=' + profileId"
          alt="Avatar image"
          @error="onImageError($event)"
        />
        <div v-else class="profile-name">
          <span>{{
            profile.Recipient.ProfileGivenName !== ""
              ? profile.Recipient.ProfileGivenName[0]
              : profile.Recipient.Username[0]
          }}</span>
        </div>
      </div>
      <div class="infos">
        <div v-if="profile.Recipient.E164" class="number">
          {{ profile.Recipient.E164 }}
        </div>
        <div v-if="profile.Recipient.Username != ''" class="number">
          {{ profile.Recipient.Username }}
        </div>
      </div>
    </div>
  </component>
</template>

<script>
export default {
  props: {
    profileId: Number,
  },
  data: () => ({
    hasProfileImage: true,
  }),
  computed: {
    profile() {
      return this.$store.state.profile;
    },
  },
  mounted() {
    this.$store.dispatch("getProfile", this.profileId);
  },
  methods: {
    onImageError(event) {
      this.hasProfileImage = false;
    },
    editContact() {
      this.$router.push({
        name: "editContact",
      });
    },
  },
};
</script>

<style scoped>
.profile-image {
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 20px 0;
}
.profile-image img {
  max-width: 80vw;
  width: 150px;
  border-radius: 50px;
}
.profile-name {
  font-size: 1.5rem;
  font-weight: bold;
  text-align: center;
  max-width: 80vw;
  width: 150px;
  height: 150px;
  border-radius: 50px;
  background-color: rgb(240, 240, 240);

  display: flex;
  justify-content: center;
  align-items: center;
  color: blue;
}
.profile-header {
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
.infos {
  text-align: center;
}
</style>