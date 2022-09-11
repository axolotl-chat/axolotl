<template>
  <component :is="$route.meta.layout || 'div'">
    <template #header>
      <div v-if="profile.Recipient" class="profile-header">
        <span>{{ name }}</span>
      </div>
    </template>
    <template #menu>
      <div class="ms-2">
        <div v-translate @click="editContact()">Edit Contact</div>
      </div>
    </template>
    <div v-if="profile.Recipient" class="profile">
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
            name[0]
          }}</span>
        </div>
      </div>
      <div class="infos">
        <div v-if="profile.Recipient.E164" class="number">
          {{ profile.Recipient.E164 }}
        </div>
        <div v-if="profile.Recipient.Username != ''" class="username">
          {{ profile.Recipient.Username }}
        </div>
        <div v-if="profile.Recipient.UUID != ''" class="uuid">
          {{ profile.Recipient.UUID }}
        </div>
        <div v-if="profile.Recipient.About != ''" class="about">
          {{ profile.Recipient.About }}
        </div>
        <div v-if="profile.Recipient.AboutEmoji != ''" class="about-emoji">
          {{ profile.Recipient.AboutEmoji }}
        </div>
        <div class="btn btn-primary create-chat mt-4" @click="createChat()">
          <div v-translate>Create private chat with {{ name }}</div>
        </div>
      </div>
    </div>
  </component>
</template>

<script>
export default {
    name: "ProfilePage",
  props: {
    profileId:{
      type: Number,
      required: true,
      default: -1
    },
  },
  data: () => ({
    hasProfileImage: true,
  }),
  computed: {
    profile() {
      return this.$store.state.profile;
    },
    name(){
      return      this.profile.Recipient.ProfileGivenName !== ""
            ? this.profile.Recipient.ProfileGivenName
            : this.profile.Recipient.Username?this.profile.Recipient.Username:"Unknow user"
    }
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
    createChat() {
      this.$store.dispatch("createChatForRecipient", {
        id: this.profile.Recipient.Id,
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
}
.profile-name span{
  font-size: 3.5rem;
  color: #5fb5ea
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
.create-chat {
  color: #fff;
}
</style>