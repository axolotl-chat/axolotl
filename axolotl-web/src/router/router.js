import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

const base = "/";

export default new Router({
  mode: "history",
  base,
  routes: [
  {
    path: "/chatList",
    name: "chatList",
    component:  () => import("@/pages/ChatList.vue")
  },
  {
    path: "/chat/:id",
    name: "chat",
    props: route => ({
      chatId: route.params.id,
    }),
    component:  () => import("@/pages/MessageList.vue")
  },
  {
    path: "/",
    name: "register",
    component:  () => import("@/pages/Register.vue")
  },
  {
    path: "/verify",
    name: "verify",
    component:  () => import("@/pages/Verification.vue")
  },
  {
    path: "/contacts",
    name: "contacts",
    component:  () => import("@/pages/Contacts.vue")
  },
  {
    path: "/devices",
    name: "contacts",
    component:  () => import("@/pages/DeviceList.vue")
  },
      { path: '/a', redirect: to => {
        const { hash, params, query } = to
         return "/ws"
      // the function receives the target route as the argument
      // return redirect path/location here.
    }}
  ]
});
