//var wsuri = "ws://127.0.0.1:8080";
//var wsuri = "ws://localhost:8080";
//var wsuri = "ws://10.211.55.2:8080";
var apiWsUri = "ws://" + location.hostname + "/api";
var api = Api(apiWsUri);

function add(x, y) {
    var ok = function(rsp) {
        console.log("ok", x, "+", y, "=", rsp.z);
    };
    var fail = function(rsp, header) {
        console.log("fail", rsp, header);
    };
    api.request("math.req/add", {x: x, y: y}, ok, fail);
}
