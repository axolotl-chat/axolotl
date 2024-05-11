<template>
  <component :is="$route.meta.layout || 'div'">
    <template #header>
      <div v-if="profile?.name" class="profile-header">
        <span>{{ profile?.name }}</span>
      </div>
    </template>
    <template #menu>
      <div class="ms-2">
        <div v-translate @click="editContact()">Edit Contact</div>
      </div>
    </template>
    <div v-if="profile?.name" class="profile">
      <div class="profile-image">
        <img
          v-if="hasProfileImage"
          class="avatar-img"
          :src="'http://localhost:9080/avatars?recipient=' + profileId"
          alt="Avatar image"
          @error="onImageError($event)"
        />
        <div v-else class="profile-name">
          <span>{{ profile ? profile.name[0] : 'Unknown' }}</span>
        </div>
      </div>
      <div class="infos">
        <div v-if="profile?.phone_number" class="number">
          <a
            href="tel:{{ `+${profile?.phone_number.code.value}${profile?.phone_number.national.value}` }}"
          >
            {{ `+${profile?.phone_number.code.value}${profile?.phone_number.national.value}` }}
          </a>
        </div>
        <div v-if="profile?.username !== ''" class="username">
          {{ profile?.username }}
        </div>
        <div v-if="profile?.uuid !== ''" class="uuid">
          {{ profile?.uuid }}
        </div>
        <div v-if="profile?.about !== ''" class="about">
          {{ profile?.about }}
        </div>
        <div v-if="profile?.aboutEmoji !== ''" class="about-emoji">
          {{ profile?.aboutEmoji }}
        </div>
        <div class="btn btn-primary create-chat mt-4" @click="createChat()">
          <div v-translate>Create private chat with {{ profile?.name }}</div>
        </div>
      </div>
    </div>
  </component>
</template>

<script>
export default {
  name: 'ProfilePage',
  props: {
    profileId: {
      type: String,
      required: true,
      default: '',
    },
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
    this.$store.dispatch('getProfile', this.profileId);
  },
  methods: {
    onImageError(event) {
      this.hasProfileImage = false;
    },
    editContact() {
      this.$router.push({
        name: 'editContact',
      });
    },
    createChat() {
      this.$router.push(`/chat/${JSON.stringify({ Contact: this.profileId })}`);
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
.profile-name span {
  font-size: 3.5rem;
  color: #5fb5ea;
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
