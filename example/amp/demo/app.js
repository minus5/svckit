var api = minus5.api();

var mathTopic = "math.v1/i",
    mathHanderl = function(data) {
      if (!data.merged) {
        console.log(mathTopic, data);
        return;
      }
      var prev;
      if (data.merged._xChange) {
        prev = data.merged._xChange.previous;
      }
      console.log(mathTopic, prev, "=>", data.merged.x);
    };

function mathSubscribe() {
  api.subscribe(mathTopic, mathHanderl);
}

function mathUnsubscribe() {
  api.unSubscribe(mathTopic, mathHanderl);
}

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
