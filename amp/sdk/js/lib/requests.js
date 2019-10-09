var amp = require("./amp.js");

var errors = require("./errors.js");

module.exports = function () {
  var correlationID = 0,
      requests = {}; // find response handlers and call them

  function response(m) {
    var r = requests[m.correlationID];

    if (!r) {}

    delete requests[m.correlationID];

    if (m.error) {
      r.fail(errors.server(m));
    } else {
      r.ok(m.body);
    }
  } // create request message and store handlers (ok, fail) into requests


  function request(uri, payload, ok, fail) {
    correlationID++;
    var msg = {
      type: amp.messageType.request,
      uri: uri,
      correlationID: correlationID,
      body: payload
    };
    requests[correlationID] = {
      ok: ok,
      fail: fail
    };
    return msg;
  }

  return {
    request: request,
    response: response
  };
};