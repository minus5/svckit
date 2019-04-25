var amp  = require("./amp.js");

function now() {
  return (new Date()).getTime();
}

var sock = null,
    _uri = "",
    _onChange = function(){},
    _onMessage = function(){},
    _connectInterval = 4 * 1000,
    _pingInterval = 16 * 1000,
    _pingTimer = undefined,
    _pingNo = 0,
    _status = {
      start: now(),
      firstMessage: 0,
      lastMessage: 0,
      readyState: 0,
      messages: 0,
      connects: 0,
      changes: [],
      errors: [],
      connected: function() {
        return _status.messages > 0 && _status.readyState === 1;
      }
    };

function send(msg, fail) {
  if (!sock) {
    fail("connection uninitialized");
    return;
  }
  if (sock.readyState !== WebSocket.OPEN) {
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
  _status.connects++;

  var logStatus = function() {
    if (sock) {
      _status.readyState = sock.readyState;
      _status.changes.push(sock.readyState);
      if (sock.readyState === WebSocket.OPEN && _status.messages === 0) {
        return; // wait for ping message
      }
    }
    _onChange(_status);
  };

  if (_status.connects % 8 === 0 && _status.messages === 0) {
    logStatus();
  }

  function pingLoop() {
    if (now() - _status.lastMessage > 16 * 1000) {
      _pingNo++;
      send(amp.ping(_pingNo), function(e) {
        _status.errors.push({method: "ping", error: e});
        logStatus();
      });
    }
    _pingTimer = setTimeout(pingLoop, _pingInterval);
  }


  try {
    sock = new WebSocket(_uri);
  } catch (e) {
    _status.errors.push({method: "connect", error: e, uri: _uri});
    logStatus();
    return;
  }

  sock.onopen = function() {
    // console.log("connected to " + _uri);
    pingLoop();
  };

  sock.onclose = function(e) {
    clearTimeout(_pingTimer);
    setTimeout(connect, _connectInterval);
    _status.errors.push({method: "close", error: e, uri: _uri});
    logStatus();
    // console.log("connection closed",  e.code , e);
  };

  sock.onmessage = function(e) {
    try{
      _status.messages++;
      _status.lastMessage = now();
      if (_status.messages === 1) {
        _status.firstMessage = now();
        logStatus();
      }
      _onMessage(e.data);
    } catch(e) {
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
