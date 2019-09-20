import Vue from 'vue'
import App from './App.vue'
import VueNativeSock from 'vue-native-websocket'
import store from './store/store'
import router from "./router/router";
import BootstrapVue from 'bootstrap-vue'
import VueChatScroll from 'vue-chat-scroll'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faArrowLeft,
         faEllipsisV,
         faPencilAlt,
         faTrash,
         faUserFriends,
         faPaperPlane,
        faPlus} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
library.add(faArrowLeft, faEllipsisV, faPencilAlt, faPlus, faTrash, faPaperPlane,faUserFriends)

Vue.component('font-awesome-icon', FontAwesomeIcon)
Vue.use(VueChatScroll)
Vue.use(BootstrapVue)
Vue.config.productionTip = false
var websocketAdress = "ws://[::1]:9080/ws";
if(process.env.NODE_ENV=="development")
  websocketAdress =  'ws://10.0.10.196:9080/ws';
Vue.use(VueNativeSock, websocketAdress,
  { store: store,
    // format: 'json',
    reconnection: true, // (Boolean) whether to reconnect automatically (false)
    reconnectionAttempts: 5, // (Number) number of reconnection attempts before giving up (Infinity),
    reconnectionDelay: 3000, // (Number) how long to initially wait before attempting a new (1000) }
  }
)
Vue.directive('longpress', {
    bind: function (el, binding, vNode) {
        // Make sure expression provided is a function
        if (typeof binding.value !== 'function') {
            // Fetch name of component
            const compName = vNode.context.name
            // pass warning to console
            let warn = `[longpress:] provided expression '${binding.expression}' is not a function, but has to be`
            if (compName) { warn += `Found in component '${compName}' ` }

            console.warn(warn)
        }

        // Define variable
        let pressTimer = null

        // Define funtion handlers
        // Create timeout ( run function after 1s )
        let start = (e) => {

            if (e.type === 'click' && e.button !== 0) {
                return;
            }

            if (pressTimer === null) {
                pressTimer = setTimeout(() => {
                    // Run function
                    handler()
                }, 1000)
            }
        }

        // Cancel Timeout
        let cancel = (e) => {
            // Check if timer has a value or not
            if (pressTimer !== null) {
                clearTimeout(pressTimer)
                pressTimer = null
            }
        }
        // Run Function
        const handler = (e) => {
            binding.value(e)
        }

        // Add Event listeners
        el.addEventListener("mousedown", start);
        el.addEventListener("touchstart", start);
        // Cancel timeouts if this events happen
        el.addEventListener("click", cancel);
        el.addEventListener("mouseout", cancel);
        el.addEventListener("touchend", cancel);
        el.addEventListener("touchcancel", cancel);
    }
})
export default new Vue({
  store,
  router,
  render: h => h(App),
}).$mount('#app')

// Vue.use(VueSocketio, `//${window.location.host}`, store);
