"use strict";

var nanoajax = require('nanoajax');

function ajax(url, msg, success, fail) {
  nanoajax.ajax({
    url: url,
    method: 'POST',
    body: JSON.stringify(msg)
  }, function (code, responseText) {
    if (code >= 200 && code < 300) {// success(responseText);
    } else {
      // fail(code, responseText);
      console.error(code, responseText);
    }
  });
}

module.exports = function (uri) {
  return {
    info: function info(msg) {
      ajax(uri + "/info", msg);
    },
    error: function error(msg) {
      ajax(uri + "/error", msg);
    }
  };
};