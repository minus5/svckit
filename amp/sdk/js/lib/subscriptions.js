"use strict";

var amp = require("./amp.js"),
    merge = require("./merge.js");

module.exports = function (onChangeHandler) {
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

    for (var key in subscriptions) {
      s[key] = subscriptions[key].ts;
    }

    return {
      type: amp.messageType.subscribe,
      subscriptions: s
    };
  }

  function publish(msg) {
    var key = msg.uri;
    var s = subscriptions[key];

    if (!s) {
      console.error("topic not found", key);
      return;
    }

    var data = null;

    switch (msg.updateType) {
      case amp.updateType.close:
        delete subscriptions[key];
        data = {
          close: true
        };
        break;

      case amp.updateType.burstStart:
        data = {
          burstStart: true
        };
        break;

      case amp.updateType.burstEnd:
        data = {
          burstEnd: true
        };
        break;

      case amp.updateType.full:
        s.full = msg.body;
        data = {
          full: msg.body,
          diff: null,
          merged: msg.body
        };
        break;

      case amp.updateType.diff:
        if (!s.full) {
          s.full = {};
        }

        merge(s.full, msg.body);
        data = {
          full: null,
          diff: msg.body,
          merged: s.full
        };
        break;

      case amp.updateType.append:
        data = {
          append: msg.body
        };
        break;

      case amp.updateType.update:
        data = {
          update: msg.body
        };
        break;

      default:
        console.error("unknown update type", msg.updateType);
        return;
    }

    s.ts = msg.ts;
    s.handlers.forEach(function (handler) {
      handler(data);
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