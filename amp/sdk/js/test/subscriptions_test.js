var assert = require("assert");
var amp = require("../src/amp.js");
var Sub = require("../src/subscriptions.js");

//import Sub from '../src/subscriptions.js';
var sub = new Sub();

describe('subscriptions', function() {

  it("should add topics to the subscribe message", function(){
    sub.add("one", function(){});
    sub.add("one", function(){});
    sub.add("two", function(){});

    var m = sub.message();
    assert.equal(m.type, amp.messageType.subscribe);
    assert.equal(Object.keys(m.subscriptions).length, 2);
    assert.equal(m.subscriptions.one, 0);
    assert.equal(m.subscriptions.two, 0);
  });

  it("should update ts after publish", function(){
    sub.publish({uri: "one", ts: 123, body: {foo: 1}, updateType: amp.updateType.full });
    sub.publish({uri: "two", ts: 234, body: {foo: 2}, updateType: amp.updateType.full });
    var m = sub.message();
    assert.equal(m.subscriptions.one, 123);
    assert.equal(m.subscriptions.two, 234);
  });

  it("should call handler on new subscribe, publish and close", function(){
    var called = 0;
    var handler = function(data){
      called++;
      if (called == 1)  {
        assert.equal(data.full.foo, 1);
        assert.equal(data.diff, null);
        assert.equal(data.merged.foo, 1);
      }
      if (called == 2)  {
        assert.equal(data.full.foo, 4);
        assert.equal(data.diff, null);
        assert.equal(data.merged.foo, 4);
      }
      if (called == 3)  {
        assert.equal(data.full, null);
        assert.equal(data.diff, null);
        assert.equal(data.merged, null);
      }
    };
    sub.add("one", handler);
    assert.equal(called, 1);

    sub.publish({uri: "one", ts: 123, body: {foo: 4}, updateType: amp.updateType.full });
    assert.equal(called, 2);

    sub.remove("one", handler);
    sub.publish({uri: "one", ts: 123, body: {foo: 4}, updateType: amp.updateType.full });
    assert.equal(called, 2);

    called = 1;
    sub.add("one", handler);  // on sub handler will be called with full
    sub.publish({uri: "one", ts: 456, updateType: amp.updateType.close}); // on close it will be called with null, null
    assert.equal(called, 3);
  });

  describe("onChange handler", function() {
    var called = 0;
    var sub2 = new Sub(function(){ called++; });
    var threeHandler = function(){};

    it("should be called on new topics", function(){

      //sub.init(function(){ called++; });
      assert.equal(called, 0);

      // sub.add("two", function(){}); // this is not a new topic
      // assert.equal(called, 0);

      sub2.add("three",threeHandler);
      assert.equal(called, 1);
    });


    it("and on remove of the handler", function(){
      sub2.remove("three", threeHandler);
      assert.equal(called, 2);
    });
  });



});
