let assert = require("assert");
let errors = require("../src/errors.js");

describe('errors', function() {

  it('should create ws error message', function() {
    let str = "something happened";
    let e = errors.ws(str);
    assert.equal(str, e.message);
    assert(e.isTransport);
    assert(!e.isApplication);
  });

  it('should create server side error message', function() {
    let str = "something happened";
    let msg = {error: {message: str}};
    let e = errors.server(msg);
    assert.equal(msg.error.message, e.message);
    assert(!e.isTransport);
    assert(e.isApplication);

  });

  it('should create server side transport message', function() {
    let msg = {error: {message: "something happened", source: 1}};
    let e = errors.server(msg);
    assert(e.isTransport);
  });

});
