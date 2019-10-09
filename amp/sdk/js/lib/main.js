"use strict";

var amp = require("./amp.js");

var Sub = require("./subscriptions.js");

var Req = require("./requests.js");

var Log = require("./log.js");

var Ws = require("./ws.js");

var Pooling = require("./pooling.js");

var urls = {
  port: function port() {
    return location.port === '' || location.port === '80' ? '' : ':' + location.port;
  },
  path: function path(relative) {
    if (relative[0] == '/') {
      // if the path has prefix / 
      return relative; // than it relative from site root
    } // othervise we will add relative to the current page location


    var pn = location.pathname;
    var path = pn.substring(0, pn.lastIndexOf('/') + 1);

    if (path.length === 0) {
      path = "/";
    }

    return path + relative;
  },
  ws: function ws() {
    var protocol = location.protocol === 'https:' ? 'wss://' : 'ws://';
    return urls.addMeta(protocol + location.hostname + urls.port() + urls.path(urls.paths.api));
  },
  log: function log() {
    return location.protocol + "//" + location.hostname + urls.port() + urls.path(urls.paths.log);
  },
  pooling: function pooling() {
    return urls.addMeta(location.protocol + "//" + location.hostname + urls.port() + urls.path(urls.paths.pooling));
  },
  forcePooling: function forcePooling() {
    return location.search.search("forcePooling") > -1;
  },
  meta: {},
  paths: {
    api: 'api',
    pooling: 'pooling',
    log: 'log'
  },
  addMeta: function addMeta(url) {
    // add meta key/values as query string to the url
    var queryStr = "";

    for (var key in urls.meta) {
      if (queryStr) {
        queryStr += "&";
      }

      queryStr += key + "=" + urls.meta[key];
    }

    if (queryStr) {
      return url + "?" + encodeURI(queryStr);
    }

    return url;
  }
}; //export function api(config)
//

module.exports = function (config) {
  config = config || {
    onTransportChange: function onTransportChange() {},
    logTransportChanges: false,
    meta: {}
  };

  if (config.meta) {
    for (var key in config.meta) {
      urls.meta[key] = config.meta[key].toString();
    }
  }

  if (config.paths) {
    for (key in config.paths) {
      urls.paths[key] = config.paths[key];
    }
  }

  var sub = Sub(subscribe);
  var logger = null;
  var req = Req();

  if (config.logTransportChanges) {
    logger = Log(urls.log());
  }

  var transport = {
    current: undefined,
    previous: undefined,
    ws: undefined,
    pooling: undefined,
    onChange: function onChange() {},
    ready: function ready() {
      return !!transport.current;
    },
    name: function name(t) {
      return t === transport.ws ? "ws" : t === transport.pooling ? "pooling" : "none";
    },
    send: function send(msg, fail) {
      // pooling is always available
      // use it while ws-pooling handshake is done
      var tr = transport.current || transport.pooling;
      tr.send(msg, fail);
    }
  };
  var failHandlers = {
    "default": function _default(e) {
      console.error(e);
    },
    ignore: function ignore(e) {}
  };

  function send(msg, fail) {
    fail = fail || failHandlers["default"];
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

    for (var i = 0; i < msgs.length; i++) {
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
    ok = ok || function () {};

    fail = fail || failHandlers["default"];
    var msg = req.request(uri, payload, ok, fail);
    send(msg, fail);
  }

  function onWsChange(status) {
    transport.previous = transport.current;
    transport.current = status.success ? transport.ws : transport.pooling;

    if (transport.current === transport.ws && status.connected || transport.current === transport.pooling) {
      subscribe();
    }

    if (transport.previous === transport.pooling && transport.current === transport.ws) {
      transport.pooling.stop();
    }

    if (logger) {
      if (status.connected) {
        logger.info(status);
      } else {
        logger.error(status);
      }
    }

    transport.onChange({
      transport: transport.name(transport.current),
      previousTransport: transport.name(transport.previous),
      status: status
    });
  }

  transport.onChange = function (status) {
    if (config.onTransportChange) {
      try {
        config.onTransportChange(status);
      } catch (e) {
        console.error(e);
      }
    }
  };

  transport.pooling = Pooling(urls.pooling(), onMessage, sub.message);

  if (urls.forcePooling()) {
    transport.current = transport.pooling;
  } else {
    transport.ws = Ws(urls.ws(), onMessage, onWsChange);
  }

  return {
    request: request,
    subscribe: sub.add,
    unSubscribe: sub.remove
  };
};