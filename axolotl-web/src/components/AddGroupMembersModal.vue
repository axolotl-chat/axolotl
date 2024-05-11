<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 v-translate class="modal-title">Add members</h5>
          <div v-if="!searchActive" class="actions">
            <button type="button" class="btn search" @click="searchActive = true">
              <font-awesome-icon icon="search" />
            </button>
            <button type="button" class="btn" @click="$emit('close')">
              <font-awesome-icon icon="times" />
            </button>
          </div>
          <div v-if="searchActive" class="actions">
            <div class="input-container">
              <input
                v-model="contactsFilter"
                type="text"
                class="form-control"
                @change="filterContacts()"
                @keyup="filterContacts()"
              />
            </div>
            <button type="button" class="btn" @click="searchActive = false">
              <font-awesome-icon icon="times" />
            </button>
          </div>
        </div>
        <div class="modal-body">
          <div class="contact-list">
            <div v-if="contacts.length > 0 && contactsFilter === ''">
              <div v-for="contact in contacts" :key="contact.UUID" class="btn col-12 chat">
                <div class="row chat-entry">
                  <div class="avatar col-3" @click="contactClick(contact)">
                    <div v-if="contact.Name" class="badge-name">
                      {{ contact.Name[0] + contact.Name[1] }}
                    </div>
                  </div>
                  <div class="meta col-7" @click="$emit('add', contact)">
                    <p class="name">{{ contact.Name }}</p>
                    <p class="number">{{ contact.Tel }}</p>
                  </div>
                </div>
              </div>
            </div>
            <div v-else-if="contactsFilter !== ''">
              <div
                v-for="contact in contactsFiltered"
                :key="'filter_' + contact.UUID"
                class="btn col-12 chat"
              >
                <div class="row chat-entry">
                  <div class="avatar col-3" @click="contactClick(contact)">
                    <div class="badge-name">
                      {{ contact.Name[0] + contact.Name[1] }}
                    </div>
                  </div>
                  <div class="meta col-7" @click="$emit('add', contact)">
                    <p class="name">{{ contact.Name }}</p>
                    <p class="number">{{ contact.Tel }}</p>
                  </div>
                </div>
              </div>
            </div>
            <div v-else>
              <span v-translate>Add contacts first</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { validateUUID } from '@/helpers/uuidCheck';

export default {
  name: 'AddGroupMembersModal',
  props: {
    alreadyAdded: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['add', 'close'],
  data() {
    return {
      contacts: [],
      searchActive: false,
      contactsFilter: '',
    };
  },
  computed: {
    contactsFiltered() {
      return this.filterForOnlyContactsWithUUID(this.$store.state.contactsFiltered);
    },
  },
  watch: {
    alreadyAdded() {
      this.contacts = this.filterForOnlyContactsWithUUID(this.$store.state.contacts);
    },
  },
  mounted() {
    this.contacts = this.filterForOnlyContactsWithUUID(this.$store.state.contacts);
  },
  methods: {
    validateUUID,
    contactClick(contact) {
      this.$store.dispatch('addNewGroupMember', contact);
    },
    filterContacts() {
      if (this.contactsFilter !== '') {
        this.$store.dispatch('filterContactsForGroup', this.contactsFilter);
      } else {
        this.$store.dispatch('clearFilterContacts');
      }
    },
    filterForOnlyContactsWithUUID(contacts) {
      return contacts.filter((c) => {
        const isValid = this.validateUUID(c.UUID);

        if (isValid) {
          return false;
        }
        const found = this.alreadyAdded.find((element) => {
          return element.Tel === c.Tel;
        });
        return typeof found === 'undefined';
      });
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn.add-contact {
  position: fixed;
  bottom: 16px;
  right: 10px;
  background-color: #2090ea;
  color: #fff;
  border-radius: 50%;
  width: 45px;
  height: 45px;
  font-size: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
}
.chat {
  padding: 0;
}
.number {
  font-size: 14px;
}
.actions-header {
  position: fixed;
  background-color: #cacaca;
  width: 100%;
  left: 0;
  display: flex;
  justify-content: flex-end;
  z-index: 2;
  top: 0;
  height: 51px;
}
.hide-actions {
  padding-right: 40px;
}
.col-2.actions {
  position: absolute;
  display: flex;
  right: 0;
  justify-content: center;
  align-items: center;
}
.col-2.actions .btn {
  font-size: 15px;
  padding: 5px;
}
.modal {
  display: block;
  border: none;
}
.modal-content {
  border-radius: 0;
}
.modal-body {
  max-height: 80vh;
  overflow: auto;
}
.modal-header {
  border-bottom: none;
  background-color: #2090ea;
  border-radius: 0;
  color: #fff;
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
.actions .btn {
  color: #fff;
  opacity: 1;
}
.actions {
  display: flex;
}
</style>
