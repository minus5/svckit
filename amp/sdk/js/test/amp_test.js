var assert = require("assert");
var amp = require("../src/amp.js");

describe('amp', function() {
  it('should export messageType and updateType', function() {
    assert(typeof amp.messageType === 'object');
    assert(typeof amp.updateType === 'object');
  });

  describe('message unpack', function() {
    it('should unpack header attirbutes', function() {
      var buf='{"u":"topic/method", "i":123, "s":234, "p": 3, "t": 3}\n{"no":1}';
      var m = amp.unpackMsg(buf);

      assert.equal(m.type, amp.messageType.response);
      assert.equal(m.updateType, amp.updateType.update);
      assert.equal(m.correlationID, 123);
      assert.equal(m.uri, "topic/method");
      assert.equal(m.ts, 234);

      var keys = Object.keys(m).sort();
      //assert.equal(keys.length, 7);
      assert.deepEqual(keys, [ 'body',
                               'correlationID',
                               'ts',
                               'type',
                               'updateType',
                               'uri']);
    });

    it('should set messageType and updateType defaults', function() {
      var buf='{"o":"one"}\n{"no":1}';
      var m = amp.unpackMsg(buf);

      assert.equal(m.type, amp.messageType.publish);
      assert.equal(m.updateType, amp.updateType.diff);
    });

    it('should unmarshal json body', function() {
      var buf='{"u":"uri"}\n{"no":1}';
      var m = amp.unpackMsg(buf);
      assert.equal(m.body.no, 1);
    });

    it('should handle non JSON body message', function() {
      var buf='{"u":"uri"}\nwrong';
      var m = amp.unpackMsg(buf);
      assert.equal(m.body, undefined);
    });

    it('should handle message without separator', function() {
      var buf='{"u":"uri"}';
      var m = amp.unpackMsg(buf);
      assert.equal(m.uri, "uri");
    });

    it('should handle message without header', function() {
      var buf='without header';
      var m = amp.unpackMsg(buf);
      assert.equal(m, null);
    });

    it('should unpack pooling message', function() {
      var buf = '{"u":"math.v1/i","s":1555517084318,"p":1}\n{"x":2911,"y":2911}\n\n{"u":"math.v1/i","s":1555517085318}\n{"x":2912}\n\n{"u":"math.v1/i","s":1555517086323}\n{"x":2913}\n\n{"u":"math.v1/i","s":1555517087325}\n{"x":2914}\n\n{"u":"math.v1/i","s":1555517088317}\n{"x":2915}';

      var msgs = amp.unpack(buf);
      assert.equal(5, msgs.length);

      for(var i=0; i<msgs.length; i++) {
        var m = msgs[i];
        assert.equal(m.type, amp.messageType.publish);
        assert.equal(m.uri, "math.v1/i");
        if (i==0) {
          assert.equal(m.updateType, amp.updateType.full);
        } else {
          assert.equal(m.updateType, amp.updateType.diff);
        }
        //console.log(m);
      }
    });

    it('should unpack one pooling message', function() {
      var buf='{"u":"one"}\n{"no":1}';
      var msgs = amp.unpack(buf);

      assert.equal(1, msgs.length);
      var m = msgs[0];
      assert.equal(m.type, amp.messageType.publish);
      assert.equal(m.updateType, amp.updateType.diff);
      assert.equal(m.uri, "one");
    });

    it('should unpack error message', function() {
      var buf='{"t":3,"e":{"m":"message","c":123}}';
      var msgs = amp.unpack(buf);

      var m = msgs[0];
      //console.log(m);
      assert.equal(m.error.message, "message");
      assert.equal(m.error.code, 123);
    });
  });

  describe('message pack', function() {
    it('should pack header and body', function() {
      var m = {type: amp.messageType.request, uri: "topic/method", correlationID: 123, body: {one: 1, two: 2}};
      var buf = amp.pack(m);

      var p = buf.split("\n");
      assert.equal(p[0], '{"t":2,"i":123,"u":"topic/method"}');
      assert.equal(p[1], '{"one":1,"two":2}');
    });
  });
});
