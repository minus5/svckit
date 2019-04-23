var amp  = require("./amp.js");
var sub  = require("./subscriptions.js");
var req  = require("./requests.js");
var ws   = require("./ws.js");
var pool = require("./pooling.js");

var _transport = undefined,
    _onTransportChange = undefined,
    //_pool,
    statusUnknown = -257;          // TODO define/rethink this error codes

function defaultFail(body, e) {
  console.error(body, {error: e, errorCode: statusUnknown, transportState: _transport.state()});
};

function ignoreFail() {}

function send(msg, fail) {
  if (fail === undefined) {
    fail = defaultFail;
  }
  _transport.send(msg, fail);
}

function subscribe(msg) {
  if (msg === undefined) {
    msg = sub.message();
  }
  send(msg, ignoreFail);

  //_pool.send(msg, ignoreFail);
}

function onMessage(data) {
  var msgs = amp.unpack(data);
  for (var i=0; i<msgs.length; i++) {
    var m = msgs[i];
    if (m === null) {
      return;
    }
    switch (m.type) {
    case amp.messageType.publish:
      sub.publish(m);
      break;
    case amp.messageType.response:
      req.response(m);
      break;
    case amp.messageType.alive:
      break;
    case amp.messageType.ping:
      // TODO return pong message
      break;
    case amp.messageType.pong:
      break;
    }
  }
}

function request(uri, payload, ok, fail) {
  if (ok === undefined) {
    ok = function(){};
  }
  if (fail === undefined) {
    fail = defaultFail;
  }
  var msg = req.request(uri, payload, ok, fail);
  send(msg, fail);
}

function onChange(status) {
  if (status == 1) {
    subscribe();
  }
  if (_onTransportChange) {
    _onTransportChange(status);
  }
}



export function api(uri, onTransportChange) {
  _onTransportChange = onTransportChange;
  //_transport = ws.init(uri, onMessage, onChange);
  _transport = pool.init("http://localhost/pooling", onMessage);

  sub.init(subscribe);
  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
