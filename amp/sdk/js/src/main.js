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
      ready() { return !!transport.current; },
      name: function(t) {
        return t === transport.ws ? "ws" : t === transport.pooling ? "pooling" : "none";
      },
      send: function(msg, fail)  {
        // pooling is always available
        // use it while ws-pooling handshake is done
        var tr = transport.current || transport.pooling;
        tr.send(msg, fail);
      }
    };

var failHandlers = {
  default: function(e) { console.error(e);},
  ignore:  function(e) {}
};

function send(msg, fail) {
  fail = fail || failHandlers.default;
  transport.send(msg, fail);
}

function subscribe(msg) {
  if (!transport.ready()) {
    // skip if not ready, will be send later
    // when transport is set
    // console.info("subscribe while transport not ready, queueing...");
    return;
  }
  msg = msg || sub.message();
  send(msg, failHandlers.ignore);
}

function onMessage(data) {
  if (!data) {
    return false;
  }
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
      send(amp.pong(), failHandlers.ignore);
      break;
    case amp.messageType.pong:
      pongReceived = true;
      break;
    }
  }
  return pongReceived;
}

function request(uri, payload, ok, fail) {
  ok = ok ||  function(){};
  fail = fail || failHandlers.default;
  var msg = req.request(uri, payload, ok, fail);
  send(msg, fail);
}

function onWsChange(status) {
  transport.previous = transport.current;
  transport.current = status.success ?  transport.ws : transport.pooling;
  if ((transport.current === transport.ws && status.connected) ||
    (transport.current === transport.pooling)) {
    subscribe();
  }
  if (transport.previous === transport.pooling && transport.current=== transport.ws) {
    transport.pooling.stop();
  }

  if (status.connected)  {
    logger.info(status);
  } else {
    logger.error(status);
  }

  transport.onChange( {
    transport: transport.name(transport.current),
    previousTransport: transport.name(transport.previous),
    status: status});
}

var urls = {
  port: function() {
    return (location.port === '' || location.port === '80') ? '' : (':' + location.port);
  },
  path: function() {
    var pn = location.pathname;
    var path = pn.substring(0, pn.lastIndexOf('/') + 1);
    if (path.length === 0)  {
      path = "/";
    }
    return path;
  },
  ws() {
    var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
    return protocol + location.hostname + urls.port() + urls.path() + 'api';
  },
  log: function() {
    return location.protocol + "//" + location.hostname + urls.port() + urls.path() + 'log';
  },
  pooling: function() {
    return location.protocol + "//" + location.hostname + urls.port() + urls.path() + 'pooling';
  },
  forcePooling: function() {
    return location.search.search("forcePooling") > -1;
  }
};

export function api(onTransportChange) {
  logger = log.init(urls.log());

  transport.onChange = function(status) {
    if (onTransportChange) {
      try {
        onTransportChange(status);
      }catch(e){
        console.error(e);
      }
    }
  };

  transport.pooling = pooling.init(urls.pooling(), onMessage);
  if (urls.forcePooling()) {
    transport.current = transport.pooling;
  } else {
    transport.ws = ws.init(urls.ws(), onMessage, onWsChange);
  }

  sub.init(subscribe);
  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
