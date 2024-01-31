import { createRouter, createWebHistory } from 'vue-router'
import store from '@/store/store'
import Legacy from '@/layouts/Legacy.vue';
import Default from '@/layouts/Default.vue';


export const router = new createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      alias: "/chatList",
      name: "chatList",
      meta: {
        layout: Default,
        hasMenu: true,
        hasBackButton: false,
      },
      component: () => import("@/pages/ChatList.vue")
    },
    {
      path: "/chat/:id",
      name: "chat",
      meta: {
        layout: Legacy,
      },
      props: route => ({
        chatId: route.params.id,
      }),
      component: () => import("@/pages/MessageList.vue")
    },
    {
      path: "/profile/:id",
      name: "profile",
      meta: {
        layout: Default,
        hasMenu: true,
        hasBackButton: true,
      },
      props: route => ({
        profileId: route.params.id,
      }),
      component: () => import("@/pages/Profile.vue")
    },
    {
      path: "/editContact",
      name: "editContact",
      meta: {
        layout: Default,
        hasBackButton: true,
      },
      component: () => import("@/pages/EditContact.vue")
    },
    {
      // register page is where the registration starts
      path: "/register",
      name: "register",
      meta: {
        layout: Default,
        textCenter: true,
        hasBackButton: true,
      },
      component: () => import("@/pages/Register.vue")
    },
    {
      // verify page is for entering the sms pin
      path: "/verify",
      name: "verify",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Verification.vue")
    },
    {
      // password page is for entering the password for database decryption
      path: "/password",
      name: "password",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Password.vue")
    },
    {
      // pin page is for entering the signal registration pin. this is currently broken
      // and handled by the verification page
      path: "/pin",
      name: "pin",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Verification.vue")
    },
    {
      path: "/setPassword",
      name: "setPassword",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/SetPassword.vue")
    },
    {
      path: "/setUsername",
      name: "setUsername",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/SetUsername.vue")
    },
    {
      path: "/contacts",
      name: "contacts",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Contacts.vue")
    },
    {
      path: "/settings",
      name: "settings",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Settings.vue")
    },
    {
      path: "/about",
      name: "about",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/About.vue")
    },
    {
      path: "/devices",
      name: "devices",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/DeviceList.vue")
    },
    {
      path: "/newGroup",
      name: "newGroup",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/NewGroup.vue")
    },
    {
      path: "/editGroup/:id",
      name: "editGroup",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/EditGroup.vue")
    },
    {
      path: "/qr",
      name: "qr",
      meta: {
        layout: Default,
        hasMenu: false,
        hasBackButton: true,
      },
      component: () => import("@/pages/DeviceLinking.vue")
    },
    {
      path: "/onboarding",
      name: "onboarding",
      meta: {
        layout: Default,
        hasMenu: false,
        hasBackButton: true,
      },
      component: () => import("@/pages/OnBoarding.vue")
    },
    {
      path: "/debug",
      name: "debug",
      meta: {
        layout: Legacy,
      },
      component: () => import("@/pages/Debug.vue")
    },
    {
      path: '/a', redirect: () => {
        return "/ws"
        // the function receives the target route as the argument
        // return redirect path/location here.
      }
    }
  ]
});

router.beforeEach((to, from, next) => {
  if (to.query.token) {
    store.dispatch("setCaptchaToken",
      to.query.token)
  }
  if (to.path === "/debug") {
    return next();
  }

  if (store.state.registrationStatus === null) {
    store.dispatch("getRegistrationStatus");
    store.watch((state) => state.registrationStatus, function () {
      proceed(to, next);
    });
  } else {
    proceed(to, next);
  }
});

function proceed(to, next) {
  const registrationPages = ['/register', '/verify', '/password', '/pin', '/setUsername', '/qr', '/onboarding'];
  const registrationStatus = store.state.registrationStatus;
  //disable routes when registration is not finished yet
  if (registrationStatus === null && !registrationPages.includes(to.path)) {
    return next('/onboarding');
  } else if (registrationStatus === "not_registered") {
    if (to.path === '/qr' || to.path === '/register' || to.path === '/onboarding') {
      let loader = document.getElementById('initial-loader');
      if (typeof loader !== "undefined" && loader !== null) {
        loader.remove();
      }
      return next();
    }
    let loader = document.getElementById('initial-loader');

    if (typeof loader !== "undefined" && loader !== null) {
      loader.remove();
    }
    return next('/onboarding');
  }
  else if ((registrationStatus === null || registrationStatus === "phoneNumber") && to.path !== '/register') {
    return next('/register');
  } else if (registrationStatus === "verificationCode" && to.path !== '/verify') {
    return next('/verify');
  } else if (registrationStatus === "pin" && to.path !== '/pin') {
    return next('/pin');
  } else if (registrationStatus === "password" && to.path !== '/password') {
    return next('/password');
  } else if (registrationStatus === "getUsername" && to.path !== '/setUsername') {
    return next('/setUsername');
  } else if (registrationStatus === "registered" && registrationPages.includes(to.path)) {
    // We are registered. And are trying to access a registration page, redirect to home
    return next('/');
  } else {
    next();
    // The screen can be displayed ;)
    let loader = document.getElementById('initial-loader');
    if (typeof loader !== "undefined" && loader !== null) {
      loader.remove();
    }
  }
}
