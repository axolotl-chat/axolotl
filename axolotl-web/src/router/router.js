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
    component:  () => import("@/pages/Chat.vue")
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
  ]
});
