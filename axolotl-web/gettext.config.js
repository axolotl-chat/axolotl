// This is the gettext config file.
// See here for documentation: https://jshmrtn.github.io/vue3-gettext/extraction.html

module.exports = {
  input: {
    path: "./src",
    include: ["**/*.vue"],
  },
  output: {
    path: "../po",
    potPath: "./textsecure.nanuc.pot",
    jsonPath: "../axolotl-web/translations/translations.json",
    locales: [
      "ar", "be", "bg", "cs", "da", "de", "el", "es", "fa", "fi", "fr", "hr", "hu", "in",
      "it", "iw", "ja", "kn-rIN", "ko", "mk", "nb", "nl", "no", "pl", "pt-BR", "pt", "ro", "ru",
      "sk", "sl", "sr", "sv", "ta", "tr", "vi", "zh-rCN"
    ],
    flat: true,
    linguas: true,
    silent: true
  },
};
