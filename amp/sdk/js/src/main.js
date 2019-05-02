var amp  = require("./amp.js");
var sub  = require("./subscriptions.js");
var req  = require("./requests.js");
var ws   = require("./ws.js");
var log   = require("./log.js");
//var pool = require("./pooling.js");

var _transport = undefined,
    _onTransportChange = undefined,
    _log = undefined,
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
}

function onMessage(data) {
  var msgs = amp.unpack(data),
      pongReceived = false;
  for (var i=0; i<msgs.length; i++) {
    var m = msgs[i];
    if (m === null) {
      return pongReceived;
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
      send(amp.pong());
      break;
    case amp.messageType.pong:
      pongReceived = true;
      break;
    }
  }
  return pongReceived;
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
  if (status.connected)  {
    _log.info(status);
  } else {
    _log.error(status);
  }
  if (status.connected) {
    subscribe();
  }
  if (_onTransportChange) {
    _onTransportChange(status);
  }
}

function port() {
  return (location.port === '' || location.port === '80') ? '' : (':' + location.port);
}

function path() {
  var pn = location.pathname;
  var path = pn.substring(0, pn.lastIndexOf('/') + 1);
  if (path.length === 0)  {
    path = "/";
  }
  return path;
}

function wsUrl() {
  var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
  return protocol + location.hostname + port() + path() + 'api';
}

function logUrl() {
  return location.protocol + "//" + location.hostname + port() + path() + 'log';
}

export function api(onTransportChange) {
  _onTransportChange = onTransportChange;

  _log = log.init(logUrl());
  _transport = ws.init(wsUrl(), onMessage, onChange);

  sub.init(subscribe);
  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
