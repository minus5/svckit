var assert = require("assert");
var errors = require("../src/errors.js");

describe('errors', function() {

  it('should create ws error message', function() {
    var str = "something happened";
    var e = errors.ws(str);
    assert.equal(str, e.message);
    assert(e.isTransport);
    assert(!e.isApplication);
  });

  it('should create pooling error message', function() {
    var str = "code: 503, service unavaiable";
    var e = errors.pooling(503, "service unavaiable");
    assert.equal(str, e.message);
    assert(e.isTransport);
    assert(!e.isApplication);
  });

  it('should create server side error message', function() {
    var str = "something happened";
    var msg = {error: {message: str}};
    var e = errors.server(msg);
    assert.equal(msg.error.message, e.message);
    assert(!e.isTransport);
    assert(e.isApplication);

  });

  it('should create server side transport message', function() {
    var msg = {error: {message: "something happened", source: 1}};
    var e = errors.server(msg);
    assert(e.isTransport);
  });

});
