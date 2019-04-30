var amp  = require("./amp.js");

function now() {
  return (new Date()).getTime();
}
var wsOPEN = 1;
var _sock = null,
    _uri = "",

    _onChange = function(){},
    _onMessage = function(){},

    _pongMessage = {
      timer: undefined,
      schedule: function(handler) {
        var interval = 16 * 1000;
        _pongMessage.timer = setTimeout(handler, interval);
      },
      clear: function() {
        clearTimeout(_pongMessage.timer);
      }
    },
    _ping = {
      timer: undefined,
      no: 0,
      schedule: function(handler) {
        var interval = 16 * 1000;
        _ping.timer = setTimeout(handler, interval);
      },
      clear: function() {
        clearTimeout(_ping.timer);
      }
    },
    _status = {
      success: false,
      opened: false,
      start: now(),
      startConnect: now(),
      firstMessage: 0,
      lastMessage: 0,
      readyState: 0,
      messages: 0,
      connects: 0,
      pings: 0,
      retries: 0,
      changes: [],
      errors: [],
      closes: [],
      opens: [],
      connected: function() {
        return _status.messages > 0 && _status.readyState === wsOPEN;
      },
      onMessage: function(isPong) {
        _status.messages++;
        _status.lastMessage = now();
        if (isPong && !_status.success) {
          // handle first pong message
          // pong messages are only send as reply to ping
          // indicates that connection works in both directions
          _pongMessage.clear();
          _status.firstMessage = now() - _status.start;
          _status.success = true;
          _status.retries = _status.connects;
          _status.log();
        }
      },
      error: function(method, e) {
        var o = {method: method, ts: now()-_status.start};
        if (e) {
          if (e.code) {
            o["code"] = e.code;
          }
          if (e.type) {
            o["type"] = e.type;
          }
          if (e.reason) {
            o["reason"] = e.reason;
          }
          if (e.message) {
            o["message"] = e.message;
          }
        }
        _status.errors.push(o);
      },
      log: function() {
        if (_sock) {
          _status.readyState = _sock.readyState;
          _status.changes.push(_sock.readyState);
          if (_sock.readyState === wsOPEN && _status.messages === 0) {
            return; // wait for pong message
          }
        }
        _onChange(_status);
      },
      giveUp: function() {
        if (_status.success) { // ako je ikada uspio
          return false;
        }
        return _status.connects > 16;
      },
      quit: function() {
        _status.connects++;
        if (_status.giveUp()) {
          _status.log();
          return true;
        }
        _status.startConnect = now();
        return false;
      },
      // calculates exponential increasing interval
      connectInterval: function() {
        var p = _status.connects || 1;
        if (p > 12) {
          p = 12; // 4096 max
        }
        return  Math.pow(2, p);
      }
    };

function send(msg, fail) {
  if (!_sock) {
    fail("connection uninitialized");
    return;
  }
  if (_sock.readyState !== wsOPEN) {
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

function pingLoop() {
  if (now() - _status.lastMessage > _ping.interval / 2) {
    _ping.no++;
    _status.pings++;
    send(amp.ping(_ping.no), function(e) {
      _status.error("pingError", e);
      _status.log();
    });
  }
  _ping.schedule(pingLoop);
}

function connect() {
  if (_status.quit()) {
    return;
  }

  try {
    _sock = new WebSocket(_uri);
  } catch (e) {
    _status.error("wsConnectError", e);
    _status.log();
    return;
  }

  _sock.onopen = function() {
    _status.opened = true;
    if (!_status.success) {
      _status.lastMessage = 0;
      _status.opens.push(now() - _status.start);
      _pongMessage.schedule(function() {
        if (_sock.readyState != wsOPEN) { // connection is closed
          return;
        }
        _status.error("firstMessageTimeout");
        _status.log();
        _sock.close();
      });
    }
    pingLoop();
  }; 

  _sock.onclose = function(e) {
    _ping.clear();
    _pongMessage.clear();
    setTimeout(connect, _status.connectInterval());
    _status.error("onclose", e);
    _status.closes.push(now() - _status.startConnect);
  };

  _sock.onmessage = function(e) {
    var isPong = _onMessage(e.data);
    _status.onMessage(isPong);
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
      return onMessage(data);
    } catch(e) {
      console.log(e);
    }
    return false;
  };
  connect();

  return {
    send: send
  };
}
