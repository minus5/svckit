var amp = require("./amp.js");

var errors = require("./errors.js");

var nanoajax = require('nanoajax');

module.exports = function (uri, onMessages, subMessage) {
  var processes = {},
      stopped = false;

  function ajax(msg, success, fail) {
    var data = amp.pack(msg);
    nanoajax.ajax({
      url: uri,
      method: 'POST',
      body: data
    }, function (code, responseText) {
      if (code >= 200 && code < 300) {
        success(responseText);
      } else {
        fail(errors.pooling(code, responseText));
      }
    });
  }

  function subscribe(msg) {
    for (const key in msg.subscriptions) {
      if (processes[key] !== undefined) {
        continue;
      }

      var ts = msg.subscriptions[key];
      processes[key] = ts;
      var m = {
        type: amp.messageType.subscribe,
        subscriptions: {}
      };
      m.subscriptions[key] = ts;
      ajax(m, function (data) {
        delete processes[key];

        if (data) {
          onMessages(data);
        }

        if (stopped) {
          return;
        }

        subscribe(subMessage());
      }, function (code, rsp) {
        delete processes[key];

        if (code >= 400 && code < 500) {
          return; // bad request and friends
        }

        console.error(code, rsp);

        if (stopped) {
          return;
        }

        setTimeout(function () {
          subscribe(subMessage());
        }, 4 * 1000);
      });
    }
  }

  function send(msg, fail) {
    if (msg.type == amp.messageType.subscribe) {
      subscribe(msg);
      return;
    }

    ajax(msg, onMessages, fail);
  }

  return {
    send: send,
    stop: function () {
      stopped = true;
    }
  };
};