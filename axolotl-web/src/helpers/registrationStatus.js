import { router } from '../router/router';

function checkRegistrationStatus(registrationStatus) {
  if (registrationStatus != undefined) {
    localStorage.setItem("registrationStatus", registrationStatus);
    let loader = document.getElementById('initial-loader');
    if (loader != undefined) {
      loader.remove();
    }

    let newRoute;
    if (registrationStatus == "registered") {
      newRoute = "chatList";
    } else if (registrationStatus == "phoneNumber") {
      newRoute = "register";
    } else if (registrationStatus == "verificationCode" || registrationStatus == "pin") {
      newRoute = "verify";
    } else if (registrationStatus == "password") {
      newRoute = "verify";
    }
    if (router.currentRoute.path != "/" + newRoute)
      router.push('/' + newRoute);
  }

}
export default checkRegistrationStatus
