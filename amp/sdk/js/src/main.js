const amp       = require("./amp.js");
const Sub       = require("./subscriptions.js");
const Req       = require("./requests.js");
const Ws        = require("./ws.js");

//export function api(config)
//
module.exports = function(config) {
  let urls = {
    port: function() {
      return (location.port === '' || location.port === '80') ? '' : (':' + location.port);
    },
    path: function(relative) {
      if (relative[0] == '/' )  { // if the path has prefix / 
        return relative;          // than it relative from site root
      }
      // othervise we will add relative to the current page location
      let pn = location.pathname;
      let path = pn.substring(0, pn.lastIndexOf('/') + 1);
      if (path.length === 0)  {
        path = "/";
      }
      return path + relative;
    },
    ws() {
      let protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
      return urls.addMeta(protocol + location.hostname + urls.port() + urls.path(urls.paths.api));
    },
    meta: {},
    paths: {
      api: 'api',
    },
    addMeta: function(url) { // add meta key/values as query string to the url
      let queryStr = "";
      for (let key in urls.meta) {
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
    meta: {},
    v1: false,
  };

  if (config.meta) {
    for (let key in config.meta) {
      urls.meta[key] = config.meta[key].toString();
    }
  }
  if (config.paths) {
    for (let key in config.paths) {
      urls.paths[key] = config.paths[key];
    }
  }

  function onMessage(data) {
    if (!data) {
      return false;
    }
    let msgs = amp.unpack(data, config.v1);
    let pongReceived = false;
    for (let i=0; i<msgs.length; i++) {
      let m = msgs[i];
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

  const transport = {
    ready: false,
    ws: undefined
  };

  function onWsChange(status) {
    if (!status) {
      return;
    }
    if (status.connected) {
      transport.ready = true;
      subscribe();
    } else {
      transport.ready = false;
    }
  }

  transport.ws = Ws(urls.ws(), onMessage, onWsChange, config.v1);
  let sub    = Sub(subscribe, config.v1, config.transformBody);
  let req    = Req();

  let failHandlers = {
    default: function(e) { console.error(e);},
    ignore:  function(e) {}
  };

  function send(msg, fail) {
    fail = fail || failHandlers.default;
    transport.ws.send(msg, fail);
  }

  function subscribe(msg) {
    if (!transport.ready) {
      return;
    }
    msg = msg || sub.message();
    send(msg, failHandlers.ignore);
  }

  function request(uri, payload, ok, fail) {
    ok = ok ||  function(){};
    fail = fail || failHandlers.default;
    let msg = req.request(uri, payload, ok, fail);
    send(msg, fail);
  }

  function setMeta(meta) {
    for (let key in meta) {
      urls.meta[key] = meta[key].toString();
    }
    let msg = {
      type: amp.messageType.meta,
      meta,
    };
    send(msg);
  }
  
  function close() {
    transport.ws.close();
  }

  return {
    request: request,
    setMeta: setMeta,
    subscribe: sub.add,
    unSubscribe: sub.remove,
    close: close,
  };
}
