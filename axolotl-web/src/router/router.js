import Vue from "vue";
import Router from "vue-router";
import store from '../store/store'

Vue.use(Router);

const base = "/";

export const router = new Router({
  mode: "history",
  base,
  routes: [
    {
      path: "/",
      alias: "/chatList",
      name: "chatList",
      component: () => import("@/pages/ChatList.vue")
    },
    {
      path: "/chat/:id",
      name: "chat",
      props: route => ({
        chatId: Number(route.params.id),
      }),
      component: () => import("@/pages/MessageList.vue")
    },
    {
      // register page is where the registration starts
      path: "/register",
      name: "register",
      component: () => import("@/pages/Register.vue")
    },
    {
      // verify page is for entering the sms pin
      path: "/verify",
      name: "verify",
      component: () => import("@/pages/Verification.vue")
    },
    {
      // password page is for entering the password for database decryption
      path: "/password",
      name: "password",
      component: () => import("@/pages/Password.vue")
    },
    {
      // pin page is for entering the signal registration pin. this is currently broken
      // and handled by the verification page
      path: "/pin",
      name: "pin",
      component: () => import("@/pages/Verification.vue")
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
      path: "/debug",
      name: "debug",
      component: () => import("@/pages/Debug.vue")
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

router.beforeEach((to, from, next) => {
  if(to.query.token){
    store.dispatch("setCaptchaToken",
    to.query.token)
  }
  if (to.path === "/debug" || to.path ==="/verify") {
    return next();
  }

  if (store.state.registrationStatus == null) {
    store.dispatch("getRegistrationStatus");
    store.watch((state) => state.registrationStatus, function() {
      proceed(to, next);
    });
  } else {
    proceed(to, next);
  }
});

function proceed(to, next) {
  const registrationPages = ['/register', '/verify', '/password', '/pin', '/captcha'];
  const registrationStatus = store.state.registrationStatus;

  //disable routes when registration is not finished yet
  if ((registrationStatus == null || registrationStatus == "phoneNumber") && to.path != '/register') {
    return next('/register');
  } else if (registrationStatus == "verificationCode" && to.path != '/verify') {
    return next('/verify');
  } else if (registrationStatus == "pin" && to.path != '/pin'){
    return next('/pin');
  } else if (registrationStatus == "password" && to.path != '/password'){
    return next('/password');
  } else if (registrationStatus == "registered" && registrationPages.includes(to.path)){
    // We are registered. And are trying to access a registration page, redirect to home
      return next('/');
  } else {
    next();
    // The screen can be displayed ;)
    let loader = document.getElementById('initial-loader');
    if (loader != undefined) {
      loader.remove();
    }
  }
}
