import Vue from "vue";
import Router from "vue-router";

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
        chatId: route.params.id,
      }),
      component: () => import("@/pages/MessageList.vue")
    },
    {
      path: "/waitForRegistrationStatus",
      name: "waitForRegistrationStatus",
      component: () => import("@/pages/WaitForRegistrationStatus.vue")
    },
    {
      path: "/register",
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
  if (to.path === "/debug") {
    return next();
  }

  const registrationStatus = localStorage.getItem('registrationStatus');

  // If we're going to the waiting status page
  if (to.path === "/waitForRegistrationStatus") {
    // And we still don't have the registration status of the user
    if (registrationStatus == null) {
      // Then we go to the registration page
      return next();
    } else {
      // Else we go to the home
      return next("/");
    }
  }

  // We're not requesting the registration page.
  // But we should do so if we don't have registration status yet, let's redirect the user
  if (registrationStatus == null) {
    // The information about logging isn't there yet, go to the waiting page
    return next('/waitForRegistrationStatus');
  }

  // We have a status (regisered or not). Let's check access right
  const publicPages = ['/register', '/verify', '/password'];
  const authRequired = !publicPages.includes(to.path);

  // redirect to registration page if not registered and trying to access a restricted page
  if (authRequired && registrationStatus !== "registered") {
    return next('/register');
  } else if (!authRequired && registrationStatus === "registered") {
    // If we request the registration pages but are already registered, we're redirected
    return next("/");
  }

  next();
});
