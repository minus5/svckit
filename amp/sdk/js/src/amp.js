var messageType = {
    publish: 0,
    subscribe: 1,
    request: 2,
    response: 3,
    ping: 4,
    pong: 5,
    alive: 6
};

var updateType = {
  diff: 0,
  full: 1,
  append: 2,
  update: 3,
  close: 4,
	burstStart: 5,
	burstEnd: 6
};

var keys = {
  "t": "type",
  "i": "correlationID",
  "e": "error",
  "u": "uri",
  "s": "ts",
  "p": "updateType",
  "b": "subscriptions"
};

var errorKeys = {
  "s": "source",
  "m": "message",
  "c": "code"
};

function unpackHeader(o) {
  var header = {
    // setting defaults
    type: messageType.publish,
  };

  unpackObject(o, header, keys);
  if (header.error) {
    header.error = {};
    unpackObject(o.e, header.error, errorKeys);
  }

  if (header.type == messageType.publish && header.updateType === undefined)  {
    header["updateType"] = updateType.diff;
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

function pack(o) {
  var header = {},
      body = o.body;

  for (var short in keys) {
    var long = keys[short];
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

function unpackMsg(data) {
  var p = data.split("\n"),
      msg = null;

  try {
    msg = unpackHeader(JSON.parse(p[0]));
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

function unpack(data) {
  if (!data) {
    return null;
  }
  var p = data.split("\n\n");
  var msgs = [];
  for(var i=0; i<p.length; i++) {
    var m = unpackMsg(p[i]);
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
