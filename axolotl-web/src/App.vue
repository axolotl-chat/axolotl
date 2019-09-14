<template>
  <div id="app">
    <header-comp></header-comp>
    <div class="container">
      <router-view />
    </div>
  </div>

</template>
<script>
import store from './store/store'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import HeaderComp from "@/components/Header.vue"
import qwebchannel from 'qwebchannel'
export default {
  name: 'axolotl-web',
  components: {
    HeaderComp
  },
  mounted(){
    console.log(process.env.NODE_ENV)
    // new qwebchannel.QWebChannel(qt.webChannelTransport, function(channel) {
    // // all published objects are available in channel.objects under
    // // the identifier set in their attached WebChannel.id property
    // var foo = channel.objects.foo;
    //
    // // access a property
    // alert(foo.hello);
    //
    // // connect to a signal
    // foo.someSignal.connect(function(message) {
    //         alert("Got signal: " + message);
    //     });
    //
    //     // invoke a method, and receive the return value asynchronously
    //        foo.someMethod("bar", function(ret) {
    //        alert("Got return value: " + ret);
    //     });
    // });
    // document.getElementsByClassName("header")[0].innerHTML=
    // "Your screen resolution is: " + screen.width + "x" + screen.height+"<br>"+navigator.userAgent;
    // const viewportmeta = document.querySelector('meta[name=viewport]');
    // viewportmeta.setAttribute('content', "initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0");
    console.log('websocket: start injection');
    var that = this;
    function initWebSocket() {
        window.blobsaverSocket = new WebSocket('ws://localhost:12345');

        window.blobsaverSocket.onclose = function() {
            console.error('BlobSaver: websocket closed');
        };

        window.blobsaverSocket.onerror = function(error) {
            console.error('BlobSaver: websocket', error);
        };

        window.blobsaverSocket.onopen = function() {
            console.log('BlobSaver: websocket opened');
        };
        window.blobsaverSocket.onmessage = function (event) {
            console.log(event.data);
            that.$store.dispatch("addDesktopSync",event.data);
          }
    }

    window.addEventListener('load', initWebSocket, false);

    window.URL.createObjectURL = function(obj) {
        // TODO check if the obj is a blob

        console.log('BlobSaver: createObjectURL interceptor');

        var reader = new FileReader();
        reader.readAsDataURL(obj);
        reader.onloadend = function() {
            console.log('BlobSaver: createObjectURL sending message');
            window.blobsaverSocket.send({type:file,
                                         data:reader.result});
        };

        throw 'stop'; // Throw an error here to stop execution (continuing execution would likely result in an error from the download manager)
    };

    console.log('websocket: end injection');
  }
}
</script>
<style>
#app{
  padding-top:50px;
}
#app >.container{
  position:relative;
}
.btn:focus{
  box-shadow:none;
}
.btn{
  border-radius:0px;
}
.btn-primary {
  background-color: #2090ea;
}
.no-entries {
    height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
}
</style>
