var amp  = require("./amp.js");

var sock = null,
    _uri = "",
    _onChange = undefined,
    _onMessage = undefined,
    _connectInterval = 5 * 1000;

var wsOpen = 1;

function send(msg, fail) {
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

function connect() {
  sock = new WebSocket(_uri);

  sock.onopen = function() {
    console.log("connected to " + _uri);
    if (_onChange) {
      _onChange(sock.readyState);
    }
  };

  sock.onclose = function(e) {
    setTimeout(connect, _connectInterval);
    console.log("connection closed",  e.code , e);
    if (_onChange) {
      _onChange(sock.readyState);
    }
  };

  sock.onmessage = function(e) {
    try{
      _onMessage(e.data);
    }catch(e) {
      console.log(e);
    }
  };
};

export function init(uri, onMessage, onChange) {
  _uri = uri;
  _onChange = onChange;
  _onMessage = onMessage;
  connect();

  return {
    send: send,
    state: function() { return sock.readyState; }
  };
}
