/*
Defines errors raised during communication.
There are two main types of errors:
  * transport failures
  * errors response from server side

In the first case message has never reached server side, or server has responded within 200 range.
In the second case we successfully received response but there is server side error during procesing request.

If application needs to distinguish these types of errors it should call isTransport or isServer helpers.
For example to decide weather the error is transient or permanent. In the case of transport failures retry can help.
*/
var sources = {
  application: 0,
  transport: 1
};

function create(source, message) {
  return {
    message: message,
    isTransport: source == sources.transport,
    isApplication: source == sources.application
  };
}

function pooling(code, responseText) {
  var desc = "code: " + code + ", " + responseText;
  return create(sources.transport, desc);
}

function ws(desc) {
  return create(sources.transport, desc);
}

function server(msg) {
  var e = msg ? msg.error : {};
  return create(e.source || sources.application, e.message);
}

module.exports = {
  pooling,
  ws,
  server
};