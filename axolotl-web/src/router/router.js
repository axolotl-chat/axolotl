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
      component: () => import("@/pages/ChatList.vue")
    },
    {
      path: "/chat/:id",
      name: "chat",
      props: route => ({
        chatId: route.params.id,
      }),
      component: () => import("@/pages/MessageList.vue")
    },
    {
      path: "/",
      name: "register",
      component: () => import("@/pages/Register.vue")
    },
    {
      path: "/verify",
      name: "verify",
      component: () => import("@/pages/Verification.vue")
    },
    {
      path: "/password",
      name: "password",
      component: () => import("@/pages/Password.vue")
    },
    {
      path: "/setPassword",
      name: "setPassword",
      component: () => import("@/pages/SetPassword.vue")
    },
    {
      path: "/contacts",
      name: "contacts",
      component: () => import("@/pages/Contacts.vue")
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import("@/pages/Settings.vue")
    },
    {
      path: "/about",
      name: "about",
      component: () => import("@/pages/About.vue")
    },
    {
      path: "/devices",
      name: "devices",
      component: () => import("@/pages/DeviceList.vue")
    },
    {
      path: "/newGroup",
      name: "newGroup",
      component: () => import("@/pages/NewGroup.vue")
    },
    {
      path: "/editGroup/:id",
      name: "editGroup",
      component: () => import("@/pages/EditGroup.vue")
    },
    {
      path: '/a', redirect: () => {
        // { path: '/a', redirect: to => {
        // const { hash, params, query } = to
        return "/ws"
        // the function receives the target route as the argument
        // return redirect path/location here.
      }
    }
  ]
});
