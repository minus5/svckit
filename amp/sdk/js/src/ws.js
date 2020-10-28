const amp  = require("./amp.js");
const errors  = require("./errors.js");

function now() {
  return (new Date()).getTime();
}

module.exports = function(uri, onMessage_, onChange_, v1) { // TODO get rid of this suffix_

  function onChange(status) {
    try{
      onChange_(status);
    } catch(e) {
      console.log(e);
    }
  };

  function onMessage(data) { // expecting that onMessage returns true if it is a pong message
    try{
      return onMessage_(data);
    } catch(e) {
      console.log(e);
    }
    return false;
  };

  let ws = null,
  pong = {
    timer: undefined,
    schedule: function(handler) {
      let interval = 16 * 1000;
      pong.timer = setTimeout(handler, interval);
    },
    clear: function() {
      clearTimeout(pong.timer);
    },
    onMessage: function(isPong) {
      if (isPong) {
        pong.clear();
      }
    },
    start: function() {
      pong.schedule(function() {
        if (ws.readyState != WebSocket.OPEN) { // connection is closed
          return;
        }
        status.event("pongTimeout");
        ws.close();
      });
    }
  },
  ping = {
    timer: undefined,
    no: 0,
    lastMessage: 0,
    afterPongInterval: 16 * 1000,
    beforePongInterval: 4 * 1000,
    interval: 4 * 1000,
    clear: function() {
      clearTimeout(ping.timer);
    },
    start: function() {
      ping.interval = ping.beforePongInterval;
      ping.lastMessage = 0;
      ping.loop();
    },
    loop: function() {
      if (now() - ping.lastMessage > ping.interval / 2) {
        ping.no++;
        send(amp.ping(ping.no), function(e) {
          status.event("pingError", e);
        });
      }
      ping.timer = setTimeout(ping.loop, ping.interval);
    },
    onMessage: function(isPong) {
      ping.lastMessage = now();
      if (isPong) {
        ping.interval = ping.afterPongInterval;
      }
    }
  },
  status = {
    success: false,
    opened: false,
    connected: false,
    start: now(),
    startConnect: now(),
    messages: 0,
    connects: 0,
    retries: 0,
    events: [],
    onMessage: function(isPong) {
      status.messages++;
      if (isPong && !status.connected) {
        // handle first pong message
        // pong messages are only send as reply to ping
        // indicates that connection works in both directions
        status.event("pong");
        status.success = true;
        status.connected = true;
        status.retries = status.connects;
        status.change(); // signal success
      }
    },
    event: function(name, e) {
      let o = {name: name, sinceStart: now() - status.start, sinceConnect: now() - status.startConnect};
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
        if (e.name) {
          o["name"] = e.name;
        }
        o["error"] = e.toString();
      }
      status.events.push(o);
    },
    change: function() {
      onChange(status);
    },
    // calculates exponential increasing interval based on number of connects
    connectInterval: function() {
      let p = status.connects || 1;
      if (p > 12) {
        p = 12; // 4096 max
      }
      return  Math.pow(2, p) * 1000;
    }
  };

  function createOpenGuard() {
    let resolve;
    return {
      guard: new Promise(r => resolve = r),
      resolve
    };
  }

  let openGuard = createOpenGuard();

  async function send(msg, fail) {
    function err(no, msg, e) {
      fail(errors.ws(msg));
      status.event("sendError"+no, e);
    }
    await openGuard.guard;
    if (ws.readyState >= WebSocket.CLOSING) {
      err(2, "connection closed readyState: " + ws.readyState);
      return;
    }
    let data = amp.pack(msg, v1);
    try {
      ws.send(data);
    } catch(e) {
      err(3, e.toString(), e);
    }
  }

  function connect() {
    function reconnect() {
      pong.clear();
      ping.clear();
      setTimeout(connect, status.connectInterval());
      status.connected = false;
    }

    status.connects += 1;
    status.startConnect = now();

    try {
      ws = new WebSocket(uri);
    } catch (e) {
      reconnect();
      status.event("wsError", e);
      return;
    }

    ws.onopen = function() {
      status.opened = true;
      pong.start();
      ping.start();
      status.event("open");
      openGuard.resolve();
    };

    ws.onclose = function(e) {
      openGuard = createOpenGuard();
      reconnect();
      status.event("close", e);
    };

    ws.onmessage = function(e) {
      let isPong = onMessage(e.data);
      status.onMessage(isPong);
      ping.onMessage(isPong);
      pong.onMessage(isPong);
    };

  };

  function close() {
    ws.onopen = null;
    ws.onclose = function(e) {
      status.event("close", e);
      status.opened = false;
      openGuard = createOpenGuard();
    };
    ws.onmessage = null;
    ws.close();
  }

  connect();

  return {
    send: send,
    close: close
  };
}
