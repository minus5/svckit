var amp = require("./amp.js");
var sub = require("./subscriptions.js");
var req = require("./requests.js");

var sock = null,
    connectInterval = 5 * 1000,
    lang = "en";

var wsOpen = 1,
    statusConnectionClosed = -256,
    statusUnknown = -257;

function defaultFail(body, e) {
  console.error(body, {error: e, errorCode: statusUnknown, wsReadyState: sock.readyState});
};


function send(msg, fail) {
  if (fail === undefined) {
    fail = defaultFail;
  }
  if (!sock) {
    fail("connection uninitialized");
    return;
  }
  if (sock.readyState !== wsOpen) {
    fail("connection closed");
    return;
  }
  var data = amp.pack(msg);
  try {
    sock.send(data);
  } catch(e) {
    fail(e);
  }
}

function subscribe(msg) {
  if (msg === undefined) {
    msg = sub.message();
  }
  send(msg);
}

function connect(uri, onChange) {
  sock = new WebSocket(uri);

  sock.onopen = function() {
    console.log("connected to " + uri);
    subscribe();
    if (onChange) {
      onChange(sock.readyState);
    }
  };

  sock.onclose = function(e) {
    setTimeout(connect, connectInterval);
    console.log("connection closed",  e.code , e);
    if (onChange) {
      onChange(sock.readyState);
    }
  };

  sock.onmessage = function(e) {
    try{
      onmessage(e.data);
    }catch(e) {
      console.log(e);
    }
  };
};

function onmessage(data) {
  var m = amp.unpack(data);
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
    break;
  case amp.messageType.pong:
    break;
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

function setLang(l) {
  if (l !== undefined) {
    lang = l;
  }
  return lang;
}

export function api(uri, onChange) {
  sub.init(subscribe);
  connect(uri, onChange);
  return {
    request: request,
    language: setLang,
    subscribe: sub.add,
    unSubscribe: sub.remove,
  };
}
