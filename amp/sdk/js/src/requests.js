var amp = require("./amp.js");

var correlationID=0,
    requests = {};

// find response handlers and call them
function response(m) {
  var r = requests[m.correlationID];
  if (r) {
    delete requests[m.correlationID];
    if (m.error) {
      r.fail(m.body, m);
    } else {
      r.ok(m.body);
    }
  }
}

// create request message
function request(uri, payload, ok, fail) {
  correlationID++;
  var msg = {
    type: amp.messageType.request,
    uri: uri,
    correlationID: correlationID,
    body: payload
  };
  requests[correlationID] = {ok: ok, fail: fail};
  return msg;
}

module.exports = {
  request: request,
  response: response
};
