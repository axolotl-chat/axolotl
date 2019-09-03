<template>
  <div>
    <div id="noise"></div>
  </div>
</template>

<script>
export default {
  name: 'socketTest',
  props: {
    msg: String
  },
  mounted(){
    var noise = document.getElementById("noise");

      let socket = new WebSocket("ws://172.17.0.2:8080/ws");
      noise.innerHTML = '<b>Attempting Connection...</b>';
      console.log("Attempting Connection...");

      socket.onopen = () => {
          console.log("Successfully Connected");
          socket.send("Hi From the Client!")
          noise.innerHTML = '<b>Successfully Connected</b>';

      };

      socket.onclose = event => {
          console.log("Socket Closed Connection: ", event);
          socket.send("Client Closed!")
          noise.innerHTML = '<b>Socket Closed Connection: '+event+'</b>';

      };

      socket.onerror = error => {
          console.log("Socket Error: ", error);
          noise.innerHTML = '<b>Socket Error: '+error+'</b>';

      };
      socket.onmessage = function (e) {
        console.log('Server: ' + e.data);
        noise.innerHTML = '<b>message: '+e.data+'</b>';

      };
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
