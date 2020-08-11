var messageType = {
    publish: 0,
    subscribe: 1,
    request: 2,
    response: 3,
    ping: 4,
    pong: 5,
    alive: 6,
    meta: 9,
};

var updateType = {
  diff: 0,
  full: 1,
  append: 2,
  update: 3,
  close: 4,
	burstStart: 5,
  burstEnd: 6,
  event: 8,
};

var keys = {
  "t": "type",
  "i": "correlationID",
  "e": "error",
  "u": "uri",
  "s": "ts",
  "p": "updateType",
  "b": "subscriptions",
  "m": "meta",
};

var keysV1 = {
  "t": "type",
  "s": "stream",
  "n": "no",
  "f": "full",
  "u": "subscriptions",
};

var errorKeys = {
  "s": "source",
  "m": "message",
  "c": "code"
};

function unpackHeader(o, v1) {
  var header = {
    // setting defaults
    type: messageType.publish,
  };

  unpackObject(o, header, v1 ? keysV1 : keys);
  if (header.error) {
    header.error = {};
    unpackObject(o.e, header.error, errorKeys);
  }

  if (header.full) {
    header.updateType = updateType.full;
  }
  if (header.stream) {
    header.uri = header.stream;
  }

  if (header.type == messageType.publish && header.updateType === undefined)  {
    header.updateType = updateType.diff;
  }
  return header;
}

function unpackObject(source, dest, keys) {
  for (var short in keys) {
    var long = keys[short];
    if (source[short] !== undefined) {
      dest[long] = source[short];
    }
  }
}

function pack(o, v1) {
  var header = {},
      body = o.body;

  var k = v1 ? keysV1 : keys;
  for (var short in k) {
    var long = k[short];
    if (o[long] !== undefined) {
      header[short] = o[long];
    }
  }

  var buf = JSON.stringify(header);
  if (body) {
    buf = buf + "\n";
    buf = buf + JSON.stringify(body);
  }

  return buf;
}

function unpackMsg(data, v1) {
  var p = data.split("\n"),
      msg = null;

  try {
    msg = unpackHeader(JSON.parse(p[0]), v1);
  }catch(e){
    return null;
  }

  if (p.length > 1 && p[1]) {
    try{
      var body = JSON.parse(p[1]);
      msg["body"] = body;
    }catch(e) {
      //console.error(e);
    }
  }

  return msg;
}

function unpack(data, v1) {
  if (!data) {
    return null;
  }
  var p = data.split("\n\n");
  var msgs = [];
  for(var i=0; i<p.length; i++) {
    var m = unpackMsg(p[i], v1);
    if (m) {
      msgs.push(m);
    }
  }
  return msgs;
}

function now() {
  return (new Date()).getTime();
}

module.exports = {
  messageType: messageType,
  updateType: updateType,
  unpack: unpack,
  unpackMsg: unpackMsg,
  pack: pack,
  ping: function(ts) {return {type: messageType.ping, ts: (ts || now()) }; },
  pong: function() {return {type: messageType.pong}; }
}
