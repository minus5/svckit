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
      send(amp.pong());
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
  if (status.connected()) {
    subscribe();
  }
  if (status.messages > 0)  {
    _log.info(status);
  } else {
    _log.error(status);
  }
  if (_onTransportChange) {
    _onTransportChange(status);
  }
}

function port() {
  return (location.port === '' || location.port === '80') ? '' : (':' + location.port);
}

function wsUrl() {
  var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
  return protocol + location.hostname + port() + '/api';
}

function logUrl() {
  return location.protocol + "//" + location.hostname + port() + '/log';
}

export function api(onTransportChange) {
  _onTransportChange = onTransportChange;
  _transport = ws.init(wsUrl(), onMessage, onChange);
  _log = log.init(logUrl());

  sub.init(subscribe);
  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
