var apiWsUri = "ws://" + location.hostname + "/api";

var api = mnu5.api(apiWsUri, function(){
  start();
});

function start() {
  var topic = "math.v1/i";
  api.subscribe(topic, function(data) {
    console.log("sub", topic, data);
  });
}

function add(x, y) {
  var ok = function(rsp) {
    console.log("ok", x, "+", y, "=", rsp.z);
  };
  var fail = function(rsp, header) {
    console.log("fail", rsp, header);
  };
  api.request("math.req/add", {x: x, y: y}, ok, fail);
}
