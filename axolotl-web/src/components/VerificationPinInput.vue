<template>
  <div id="wrapper" :style="{ justifyContent: direction }">
    <div v-for="(char, index) in arraySize" :key="'otp_' + index">
      <input
        :key="index"
        :ref="'otp' + index"
        v-model="arraySize[index]"
        class="inputBox"
        type="tel"
        maxlength="1"
        :autofocus="index === 0"
        @keydown="
          handleEnterKey($event);
          handleKeyDown($event, index);
        "
        @input="handleInput($event, index)"
        @paste="onPaste"
        @focus="newColor(index)"
        @blur="defaultColor(index)"
      >
    </div>
  </div>
</template>
<script>
export default {
  name: "VerificationPinInputComponent",
  props: {
    numberOfBoxes: {
      type: Number,
      default: 6
    },
    position: {
      type: String,
      default: "center"
    },
    color: {
      type: String,
      default: ""
    },
    clearInput: {
      type: Boolean,
      default: false,
    },
  },
  emits:['input-value','enter','clearValue'],

  data() {
    return {
      arraySize: null,
      boxLength: null,
      direction: null,
      boxColor: "#2090ea",
      clearFlag: false,
    };
  },
  computed: {},
  mounted() {
    this.handleBoxes();

    if (typeof this.color !== "undefined" && this.color) {
      this.boxColor = this.color;
    }
    this.clearFlag = this.clearInput;
    if (this.clearFlag) {
      this.$emit("clearValue", "");
    }
    this.handleAlignment();
  },
  methods: {
    handleBoxes() {
      if (typeof this.numberOfBoxes === "undefined" && !this.numberOfBoxes) {
        this.boxLength = 6;
        this.arraySize = Array(this.boxLength).fill("");
      } else if (typeof this.numberOfBoxes === "number") {
        this.boxLength = this.numberOfBoxes;
        this.arraySize = Array(this.boxLength).fill("");
      } else {
        this.boxLength = 6;
        this.arraySize = Array(this.boxLength).fill("");
      }
    },
    handleAlignment() {
      switch (this.position) {
        case "left":
          this.direction = "flex-start";
          break;
        case "right":
          this.direction = "flex-end";
          break;
        case "center":
          this.direction = "center";
          break;
        default:
          this.direction = "center";
      }
    },
    newColor(index) {
      const i = "otp" + index;
      this.$refs[i][0].style.boxShadow = ` 0 0 5px  ${this.boxColor} inset`;
      this.$refs[i][0].style.border = `1px solid ${this.boxColor}`;
    },
    defaultColor(index) {
      const i = "otp" + index;
      this.$refs[i][0].style.boxShadow = " 0 0 5px #ccc inset";
      this.$refs[i][0].style.border = "solid 1px #ccc";
    },
    focusElement(index) {
      const i = "otp" + index;
      this.$refs[i][0].focus();
    },
    handleEnterKey(event) {
      if (event.key === "Enter") {
        this.$emit("enter");
        event.stopPropagation();
      }
    },
    sanitizeKeyData(key) {
      return key === "Unidentified" ? undefined : key;
    },
    emitInput() {
      const result = this.arraySize.join("").slice(0, this.numberOfBoxes);
      this.$emit("input-value", result);
    },
    handleKeyDown(event, index) {
      const key = this.sanitizeKeyData(event.key);
      if (!key) {
        return;
      }
      if (key === "Backspace") {
        if (this.arraySize[index]) {
          return (this.arraySize[index] = "");
        }

        if (index > 0) {
          this.focusElement(index - 1);
        }
      } else if (!event.shiftKey && (key === "ArrowRight" || key === "Right")) {
        if (index < this.arraySize.length - 1) {
          this.focusElement(index + 1);
        }
      } else if (!event.shiftKey && (key === "ArrowLeft" || key === "Left")) {
        if (index > 0) {
          this.focusElement(index - 1);
        }
      } else if (key.length === 1 && this.arraySize[index]) {
        this.arraySize[index] = key;
        this.$forceUpdate();
        if (index < this.arraySize.length - 1) {
          this.focusElement(index + 1);
        }
        this.emitInput();
      }
    },
    handleInput(event, index) {
      const value = this.arraySize[index];

      if (value) {
        if (value.length > 1) {
          this.arraySize[index] = value[value.length - 1];
        }
        if (index < this.arraySize.length - 1) {
          this.focusElement(index + 1);
        }
      }

      this.emitInput();
    },
    onPaste(event) {
      const clipboardData = event.clipboardData || window.clipboardData;
      if (!clipboardData) {
        return;
      }
      event.preventDefault();
      const code =
        clipboardData.getData("Text") || clipboardData.getData("text/plain");
      this.fillCode(code);
    },
    fillCode(code) {
      code = code.trim();
      code = code.slice(0, this.boxLength);
      const parts = code.split("");
      parts.length = this.boxLength;
      this.arraySize = parts;
      const last = code.length - 1;
      setTimeout(() => {
        this.arraySize[last] =
          this.arraySize[last] && this.arraySize[last].slice(0, 1);
        this.$forceUpdate();
        this.$refs["otp" + (this.arraySize.length - 1)][0].focus();
      }, 0);
    },
  },
};
</script>
<style>
#wrapper {
  width: 100%;
  margin: 8px auto 2px;
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

#wrapper input {
  margin: 0 5px !important;
  text-align: center;
  line-height: 30px !important;
  font-size: 35px !important;
  border: solid 1px #ccc;
  box-shadow: 0 0 5px #ccc inset;
  outline: none;
  width: 38px !important;
  -webkit-transition: all 0.2s ease-in-out;
  transition: all 0.2s ease-in-out;
  border-radius: 3px;
}

#wrapper input::-moz-selection {
  background: transparent;
}

#wrapper input::selection {
  background: transparent;
}

input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

input[type="tel"] {
  -moz-appearance: textfield;
}
</style>
