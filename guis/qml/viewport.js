function addViewPortShim() {

    var x = document.querySelector('meta[name="viewport"]');
    if (false) {
      console.log("VIEW PORT EALREADY EXISTS");
    } else {
        var meta = document.createElement('meta');
        meta.name = "viewport";
        meta.content = "width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=0";
        document.getElementsByTagName('head')[0].appendChild(meta);
    }
}

addViewPortShim();
