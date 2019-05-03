var amp     = require("./amp.js");
var sub     = require("./subscriptions.js");
var req     = require("./requests.js");
var ws      = require("./ws.js");
var log     = require("./log.js");
var pooling = require("./pooling.js");

var logger = undefined,
    transport = {
      current: undefined,
      previous: undefined,
      ws: undefined,
      pooling: undefined,
      onChange: function(){},
      name: function(t) {
        return t === transport.ws ? "ws" : t === transport.pooling ? "pooling" : "none";
      },
      send: function(msg, fail)  {
        if (fail === undefined) {
          fail = defaultFail;
        }
        if (transport.current === undefined) {  // TODO
          fail("connecting...");
          return;
        }
        transport.current.send(msg, fail);
      }
    };

function defaultFail(body, e) {
  console.error(body, {error: e});
};

function ignoreFail() {}

function send(msg, fail) {
  transport.send(msg, fail);
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

function onWsChange(status) {
  if (status.connected)  {
    logger.info(status);
  } else {
    logger.error(status);
  }

  transport.previous = transport.current;
  transport.current = status.success ?  transport.ws : transport.pooling;
  if ((transport.current === transport.ws && status.connected) ||
    (transport.current === transport.pooling)) {
    subscribe();
  }
  if (transport.previous === transport.pooling && transport.current=== transport.ws) {
    transport.pooling.stop();
  }

  transport.onChange( {
    transport: transport.name(transport.current),
    previousTransport: transport.name(transport.previous),
    status: status});

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

function poolingUrl() {
  return location.protocol + "//" + location.hostname + port() + path() + 'pooling';
}

export function api(onTransportChange) {
  logger = log.init(logUrl());

  transport.onChange = function(status) {
    if (onTransportChange) {
      try {
        onTransportChange(status);
      }catch(e){
        console.error(e);
      }
    }
  };
  transport.ws = ws.init(wsUrl(), onMessage, onWsChange);
  transport.pooling = pooling.init(poolingUrl(), onMessage);


  sub.init(subscribe);
  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
