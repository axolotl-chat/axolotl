import { includeIgnoreFile } from "@eslint/compat";
import pluginVue from 'eslint-plugin-vue';
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const gitignorePath = path.resolve(__dirname, ".gitignore");

export default [
    ...pluginVue.configs['flat/recommended'],
    includeIgnoreFile(gitignorePath),
    {
        languageOptions: {
            globals: {},
        },

        files: ["**/*.js", "**/*.ts", "**/*.vue"],

        rules: {
            "no-console": process.env.NODE_ENV === "production" ? "error" : "off",
            "no-debugger": process.env.NODE_ENV === "production" ? "error" : "off",
            "comma-dangle": "off",
            "class-methods-use-this": "off",
            "import/no-unresolved": "off",
            "import/extensions": "off",
            "implicit-arrow-linebreak": "off",
            "import/prefer-default-export": "off",
            "vue/no-mutating-props": "off",
            "vue/singleline-html-element-content-newline": "off",
            "vue/max-attributes-per-line": "off",
            "vue/component-name-in-template-casing": "off",

            "vue/html-self-closing": ["error", {
                html: {
                    void: "any",
                    normal: "always",
                    component: "always",
                },

                svg: "always",
                math: "always",
            }],
        },
    }];
