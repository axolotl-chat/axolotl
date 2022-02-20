<template>
  <div class="modal" tabindex="-1" role="dialog">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            <span v-translate>Adding contacts</span>
          </h5>
          <button type="button" class="close btn" @click="$emit('close')">
            <span aria-hidden="true">&times;</span>
          </button>
        </div>
        <div class="modal-body">
          <span><strong v-translate>Contacts can be added through a vcf contacts file. If an contact has multiple numbers, all of them will be added as separate contacts.</strong></span>
        </div>
        <div class="modal-body">
          <input
            id="addVcf"
            type="file"
            style="position: fixed; top: -100em"
            accept=".vcf"
            @change="readVcf"
          />
        </div>
        <div class="modal-footer">
          <button
            v-translate
            type="button"
            class="btn btn-primary"
            @click="refreshContacts()"
          >
            Import vcf
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "ImportVcfModal",
  props: {
    number: {
      type: String,
      required: false,
      default: null
    },
    uuid: {
      type: String,
      required: false,
      default: null
    },
  },
  emits: ["close", "add"],
  data() {
    return {
      phone: "",
      name: "",
    };
  },
  mounted() {
    if (this.number) {
      this.phone = this.number;
      this.name = "";
    } 
  },
  methods: {
    readVcf(evt) {
      const f = evt.target.files[0];
      if (f) {
        const r = new FileReader();
        const that = this;
        r.onload = function (e) {
          const vcf = e.target.result;
          that.$store.dispatch("uploadVcf", vcf);
        };
        r.readAsText(f);
        this.$emit("close");
      } else {
        alert("Failed to load file");
      }
    },
    refreshContacts() {
      this.$store.state.importingContacts = true;
      // console.log("Import contacts for gui " + this.gui)
      this.showSettingsMenu = false;
      if (this.gui === "ut") {
        const result = window.prompt("refreshContacts");
        if (result !== "canceled")
          this.$store.dispatch("refreshContacts", result);
      } else {
        this.showSettingsMenu = false;
        document.getElementById("addVcf").click();
      }
    },
  },
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
  justify-content: center;
}
</style>
