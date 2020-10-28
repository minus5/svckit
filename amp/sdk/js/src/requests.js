const amp = require("./amp.js");
const errors  = require("./errors.js");

module.exports = function() {

  let correlationID=0;
  let requests = {};

  // find response handlers and call them
  function response(m) {
    let r = requests[m.correlationID];
    if (!r) {
    }

    delete requests[m.correlationID];
    if (m.error) {
      r.fail(errors.server(m));
    } else {
      r.ok(m.body);
    }
  }

  // create request message and store handlers (ok, fail) into requests
  function request(uri, payload, ok, fail) {
    correlationID++;
    let msg = {
      type: amp.messageType.request,
      uri: uri,
      correlationID: correlationID,
      body: payload
    };
    requests[correlationID] = {ok: ok, fail: fail};
    return msg;
  }

  return {
    request: request,
    response: response
  };
  
};
