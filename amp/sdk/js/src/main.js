var amp       = require("./amp.js");
var Sub       = require("./subscriptions.js");
var Req       = require("./requests.js");
var Log       = require("./log.js");
var Ws        = require("./ws.js");
var Pooling   = require("./pooling.js");

//export function api(config)
//
module.exports = function(config) {
  var urls = {
    port: function() {
      return (location.port === '' || location.port === '80') ? '' : (':' + location.port);
    },
    path: function(relative) {
      if (relative[0] == '/' )  { // if the path has prefix / 
        return relative;          // than it relative from site root
      }
      // othervise we will add relative to the current page location
      var pn = location.pathname;
      var path = pn.substring(0, pn.lastIndexOf('/') + 1);
      if (path.length === 0)  {
        path = "/";
      }
      return path + relative;
    },
    ws() {
      var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
      return urls.addMeta(protocol + location.hostname + urls.port() + urls.path(urls.paths.api));
    },
    log: function() {
      return location.protocol + "//" + location.hostname + urls.port() + urls.path(urls.paths.log);
    },
    pooling: function() {
      return urls.addMeta(location.protocol + "//" + location.hostname + urls.port() + urls.path(urls.paths.pooling));
    },
    forcePooling: function() {
      return location.search.search("forcePooling") > -1;
    },
    meta: {},
    paths: {
      api: 'api',
      pooling: 'pooling',
      log: 'log',
    },
    addMeta: function(url) { // add meta key/values as query string to the url
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
  };

  config = config || {
    onTransportChange: function() {},
    logTransportChanges: false,
    meta: {},
    v1: false,
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

  var sub    = Sub(subscribe, config.v1, config.transformBody);
  var logger = null;
  var req    = Req();

  if (config.logTransportChanges) {
    logger = Log(urls.log());
  }

  var transport = {
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
    var msgs = amp.unpack(data, config.v1),
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
    // send updated metadata in case it changed before ws was initialized
    if (transport.current === transport.ws && transport.previous !== transport.ws && transport.metaChanged) {
      setMeta(urls.meta);
    }

    if (logger) {
      if (status.connected)  {
        logger.info(status);
      } else {
        logger.error(status);
      }
    }

    transport.onChange( {
      transport: transport.name(transport.current),
      previousTransport: transport.name(transport.previous),
      status: status});
  }

  transport.onChange = function(status) {
    if (config.onTransportChange) {
      try {
        config.onTransportChange(status);
      }catch(e){
        console.error(e);
      }
    }
  };

  transport.pooling = Pooling(urls.pooling(), onMessage, sub.message, config.v1);
  if (urls.forcePooling()) {
    transport.current = transport.pooling;
  } else {
    transport.ws = Ws(urls.ws(), onMessage, onWsChange, config.v1);
  }

  function setMeta(meta) {
    for (var key in meta) {
      urls.meta[key] = meta[key].toString();
    }
    transport.metaChanged = true;

    transport.pooling.stop();
    transport.pooling = Pooling(urls.pooling(), onMessage, sub.message, config.v1);

    if (transport.current === transport.ws) {
      let msg = {
        type: amp.messageType.meta,
        meta,
      };
      send(msg);
    }
  }
  
  function close() {
    if (transport.ws) {
      transport.ws.close();
    }
  }

  return {
    request: request,
    setMeta: setMeta,
    subscribe: sub.add,
    unSubscribe: sub.remove,
    close: close,
  };
}
