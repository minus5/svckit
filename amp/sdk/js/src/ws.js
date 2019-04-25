var amp  = require("./amp.js");

function now() {
  return (new Date()).getTime();
}

var _sock = null,
    _uri = "",
    _connectInterval = 4 * 1000,

    _onChange = function(){},
    _onMessage = function(){},

    _firstMessage = {
      interval: 16 * 1000,
      timer: undefined,
    },

    _ping = {
      interval: 16 * 1000,
      timer: undefined,
      no: 0,
    },
    _status = {
      success: false,
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
      },
      onMessage: function() {
        _status.messages++;
        _status.lastMessage = now();
        if (_status.messages === 1) {
          _status.firstMessage = now();
          _status.success = true;
          logStatus();
        }
      }
    };

function send(msg, fail) {
  if (!_sock) {
    fail("connection uninitialized");
    return;
  }
  if (_sock.readyState !== WebSocket.OPEN) {
    fail("connection closed readyState:" + _sock.readyState);
    return;
  }
  var data = amp.pack(msg);
  try {
    _sock.send(data);
  } catch(e) {
    fail(e);
  }
}

function logStatus() {
  if (_sock) {
    _status.readyState = _sock.readyState;
    _status.changes.push(_sock.readyState);
    if (_sock.readyState === WebSocket.OPEN && _status.messages === 0) {
      return; // wait for pong message
    }
  }
  _onChange(_status);
}

function pingLoop() {
  if (now() - _status.lastMessage > _ping.interval) {
    _ping.no++;
    send(amp.ping(_ping.no), function(e) {
      _status.errors.push({method: "ping", error: e});
      logStatus();
    });
  }
  _ping.timer = setTimeout(pingLoop, _ping.interval);
}

function connect() {
  _status.connects++;
  if (_status.connects % 8 === 0 && _status.messages === 0) {
    logStatus();
  }

  try {
    _sock = new WebSocket(_uri);
  } catch (e) {
    _status.errors.push({method: "connect", error: e, uri: _uri});
    logStatus();
    // TODO retry or give up, currently giving up
    return;
  }

  _sock.onopen = function() {
    console.log("connected to " + _uri);
    _status.lastMessage = 0;
    _firstMessage.timer = setTimeout(function() {
      console.log("closing connection");
      clearTimeout(_ping.timer);
      _status.errors.push({method: "firstMessageTimeout", interval: _firstMessage.interval});
      logStatus();
      _sock.close();
    }, _firstMessage.interval);
    pingLoop();
  };

  _sock.onclose = function(e) {
    clearTimeout(_ping.timer);
    setTimeout(connect, _connectInterval);
    //_status.errors.push({method: "close", error: e, uri: _uri});
    //logStatus();
    console.log("connection closed",  e.code , e);
  };

  _sock.onmessage = function(e) {
    _status.onMessage();
    clearTimeout(_firstMessage.timer);
    _onMessage(e.data);
  };

  _sock.onerror = function(event) {
    console.error("error observed:", event);
  };
};

export function init(uri, onMessage, onChange) {
  _uri = uri;
  _onChange = function(status) {
    try{
      onChange(status);
    } catch(e) {
      console.log(e);
    }
  };
  _onMessage = function(data) {
    try{
      onMessage(data);
    } catch(e) {
      console.log(e);
    }
  };
  connect();

  return {
    send: send
  };
}
