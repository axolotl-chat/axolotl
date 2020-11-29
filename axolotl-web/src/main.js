import Vue from 'vue'
import App from './App.vue'
import VueNativeSock from 'vue-native-websocket'
import store from './store/store'
import {router} from "./router/router";
import BootstrapVue from 'bootstrap-vue'
import VueChatScroll from 'vue-chat-scroll'
import linkify from 'vue-linkify'
import GetTextPlugin from 'vue-gettext'
import translations from '../translations/translations.json'
import { library } from '@fortawesome/fontawesome-svg-core'
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

import { longClickDirective } from 'vue-long-click'

const longClickInstance = longClickDirective({ delay: 800, interval: 0 })
Vue.directive('longclick', longClickInstance)
Vue.component('font-awesome-icon', FontAwesomeIcon)
Vue.use(VueChatScroll)
Vue.use(BootstrapVue)
Vue.use(GetTextPlugin, { translations: translations, defaultLanguage: 'en', })
Vue.directive('linkified', linkify)
Vue.config.productionTip = false
var websocketAdress = "ws://";
if (window.location.protocol === "https:") {
  websocketAdress = "wss://";
}
websocketAdress += window.location.host;
websocketAdress += "/ws";

if (process.env.NODE_ENV == "development") {
  if (process.env.VUE_APP_WS_ADDRESS) {
    websocketAdress = 'ws://' + process.env.VUE_APP_WS_ADDRESS + ':9080/ws';
  } else {
    websocketAdress = 'ws://localhost:9080/ws';
  }
}
Vue.use(VueNativeSock, websocketAdress,
  {
    store: store,
    // format: 'json',
    reconnection: true, // (Boolean) whether to reconnect automatically (false)
    reconnectionAttempts: 5, // (Number) number of reconnection attempts before giving up (Infinity),
    reconnectionDelay: 3000, // (Number) how long to initially wait before attempting a new (1000) }
  }
)

export default new Vue({
  store,
  router,
  render: h => h(App),
}).$mount('#app')

// Vue.use(VueSocketio, `//${window.location.host}`, store);
