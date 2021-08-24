import { createApp } from 'vue'
import App from './App.vue'
import VueNativeSock from 'vue-native-websocket-vue3'
import store from './store/store'
import { router } from "./router/router";
import { createGettext } from "@jshmrtn/vue3-gettext";
import translations from '../translations/translations.json'
import { library } from '@fortawesome/fontawesome-svg-core'
import LinkifyHtml from 'linkifyjs/html'
import {
  faArrowLeft,
  faEllipsisV,
  faPencilAlt,
  faTrash,
  faUserFriends,
  faPaperPlane,
  faTimes,
  faCheck,
  faVolumeMute,
  faHeart,
  faSearch,
  faArrowDown,
  faPlus
} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
library.add(faArrowLeft, faEllipsisV, faPencilAlt, faPlus, faTrash, faPaperPlane,
  faUserFriends, faTimes, faCheck, faVolumeMute, faHeart, faSearch, faArrowDown)

const app = createApp(App)
app.component('FontAwesomeIcon', FontAwesomeIcon)
app.mixin({
  methods: {
    linkify: function (content) {
      return LinkifyHtml(content);
    },
  },
})
const gettext = createGettext({
  defaultLanguage: "en",
  translations,
});
app.use(gettext);
app.config.productionTip = false

// set backend adress
var websocketAdress = "ws://";
if (window.location.protocol === "https:") {
  websocketAdress = "wss://";
}
websocketAdress += window.location.host;
websocketAdress += "/ws";

if (process.env.NODE_ENV === "development") {
  if (process.env.VUE_APP_WS_ADDRESS) {
    websocketAdress = 'ws://' + process.env.VUE_APP_WS_ADDRESS + ':9080/ws';
  } else {
    websocketAdress = 'ws://localhost:9080/ws';
  }
}
// initialise connection to the backend
app.use(VueNativeSock, websocketAdress,
  {
    store: store,
    // format: 'json',
    reconnection: true, // (Boolean) whether to reconnect automatically (false)
    reconnectionAttempts: 5, // (Number) number of reconnection attempts before giving up (Infinity),
    reconnectionDelay: 3000, // (Number) how long to initially wait before attempting a new (1000) }
  }
)
app.use(store)
app.use(router)
app.mount('#app')

export default app