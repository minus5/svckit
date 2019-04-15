var apiWsUri = "ws://" + location.hostname + "/api";

var api = mnu5.api(apiWsUri, function(status){
  console.log("ws status changed to", status);
});

var topic = "math.v1/i";
api.subscribe(topic, function(data) {
  var prev;
  if (data._xChange) {
    prev = data._xChange.previous;
  }
  //console.log("sub", topic, data);
  console.log(topic, prev, "=>", data.x);
});


function startPeriodicRequests() {
  var x = 1;
  setInterval(function() {
    x++;
    add(x, parseInt(Math.random() * 1000));
  } , 1000);
}

function add(x, y) {
  var ok = function(rsp) {
    console.log("add", x, "+", y, "=", rsp.z);
  };
  var fail = function(rsp, header) {
    console.log("add fail", rsp, header);
  };
  api.request("math.req/add", {x: x, y: y}, ok, fail);
}
