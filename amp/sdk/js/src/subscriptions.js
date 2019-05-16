var amp = require("./amp.js"),
    merge = require("./merge.js");

module.exports = function(onChangeHandler) {

  var subscriptions = {};

  function add(key, handler) {
    var s = subscriptions[key];
    if (s === undefined) {
      s = {
        key: key,
        ts: 0,
        handlers: [],
        full: null,
        diff: null
      };
      subscriptions[key] = s;
      onChange();
    } else {
      handler(s.full, null);
    }
    s.handlers.push(handler);
  }

  function remove(key, handler) {
    var s = subscriptions[key];
    if (s === undefined) {
      return;
    }
    var i = s.handlers.indexOf(handler);
    if (i > -1) {
      s.handlers.splice(i, 1);
    }
    if (s.handlers.length === 0) {
      delete subscriptions[key];
      onChange();
    }
  }

  function message() {
    var s = {};
    for(var key in subscriptions) {
      s[key] = subscriptions[key].ts;
    }
    return {type: amp.messageType.subscribe, subscriptions: s};
  }

  function publish(msg) {
    var key = msg.uri;
    var s = subscriptions[key];

    if (!s) {
      console.log("topic not found", key);
      return;
    }

    if (msg.updateType === amp.updateType.close) {
      delete subscriptions[key];
      s.handlers.forEach(function(handler){
        handler(null, null); // signals close of the topic
      });
      return;
    }

    s.ts = msg.ts;
    if (msg.updateType === amp.updateType.full ||
        msg.updateType === amp.updateType.append ||
        msg.updateType === amp.updateType.update) {
      s.full = msg.body;
      s.diff = null;
    }
    if (msg.updateType === amp.updateType.diff) {
      if (!s.full) {
        s.full = {};
      }
      s.diff = msg.body;
      merge(s.full, msg.body);
    }
    s.handlers.forEach(function(handler){
      handler(s.full, s.diff);
    });
  }

  function onChange() {
    if (onChangeHandler == null) {
      return;
    }
    onChangeHandler(message());
  }

  return {
    add: add,
    remove: remove,
    publish: publish,
    message: message
  };

};
