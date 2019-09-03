import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

const base = "/";

let router = new Router({
  mode: "history",
  base,
  routes: [
  {
    path: "/",
    name: "chatList",
    component:  () => import("@/components/ChatList.vue")
  },
  {
    path: "/chat/:id",
    name: "chat",
    props: route => ({
      chatId: route.params.id,
      sidebar: true,
    }),
    component:  () => import("@/components/Chat.vue")
  },
  ]
});
export default router
