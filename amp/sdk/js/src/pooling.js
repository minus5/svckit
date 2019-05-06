var sub  = require("./subscriptions.js");
var amp  = require("./amp.js");
var errors  = require("./errors.js");
var nanoajax = require('nanoajax');

var _processes = {},
    _uri,
    _onMessages = undefined,
    _stopped = false;

function ajax(msg, success, fail) {
  var data = amp.pack(msg);

  nanoajax.ajax({
    url: _uri,
    method: 'POST',
    body: data,
  }, function(code, responseText){
    if (code>=200&&code<300) {
      success(responseText);
    }else {
      fail(errors.pooling(code, responseText));
    }
  });
}

function subscribe(msg) {
  for (var key in  msg.subscriptions) {
    if (_processes[key] !== undefined) {
      continue;
    }

    var ts = msg.subscriptions[key];
    _processes[key] = ts;
    var m = {type: amp.messageType.subscribe, subscriptions: {}};
    m.subscriptions[key] = ts;
    ajax(m, function(data) {
      delete _processes[key];
      if (data) {
        _onMessages(data);
      }
      if (_stopped) {
        return;
      }
      subscribe(sub.message());
    },function(code, rsp) {
      delete _processes[key];
      if (code >= 400 && code < 500){
        return; // bad request and friends
      }
      console.error(code, rsp);
      if (_stopped) {
        return;
      }
      setTimeout(function() {
        subscribe(sub.message());
      }, 4 * 1000);
    });
  }
}

function send(msg, fail) {
  if (msg.type == amp.messageType.subscribe) {
    subscribe(msg);
    return;
  }
  ajax(msg, _onMessages, fail);
}

export function init(uri, onMessages) {
  _uri = uri;
  _onMessages = onMessages;

  return {
    send: send,
    stop: function() { _stopped = true; }
  };
}
