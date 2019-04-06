function Api(wsuri) {
    var sock = null,
        connectInterval = 5 * 1000,
        correlationID=0,
        requests = {},
        subscriptions = {};

    var ampPubslish = 0,
        ampSubscribe = 1,
        ampRequest = 2,
        ampResponse = 3,
        ampPing = 4,
        ampPong = 5,
        ampAlive = 6;

    var wsOpen = 1,
        statusConnectionClosed = -256,
        statusUnknown = -257;

    function subscribe() {
        var msg = {t: 1, b: subscriptions};
        try {
            sock.send(JSON.stringify(msg));
        } catch(e) {
            console.log(e);
        }
    }

    function connect() {
        sock = new WebSocket(wsuri);

        sock.onopen = function() {
            console.log("connected to " + wsuri);
            subscriptions = {"math.v1/i": 0};
            subscribe();
        };

        sock.onclose = function(e) {
            setTimeout(connect, connectInterval);
            console.log("connection closed",  e.code , e);
        };

        sock.onmessage = function(e) {
            try{
                onmessage(e.data);
            }catch(e) {
                console.log(e);
            }
        };
    };

    function onmessage(data) {
        var p = data.split("\n"),
            header = unpackHeader(JSON.parse(p[0])),
            body = {};
        console.log("message received", data);
        if (p.length > 1 && p[1]) {
            body = JSON.parse(p[1]);
        }
        switch (header.type) {
        case ampAlive:
            break;
        case ampResponse:
            var r = requests[header.correlationID];
            if (r) {
                delete requests[header.correlationID];
                if (header.error) {
                    r.fail(body, header);
                } else {
                    r.ok(body);
                }
            }
        }
    }

    function request(uri, payload, ok, fail) {
        if (sock.readyState !== wsOpen) {
            fail(undefined, {error: "connection closed", errorCode: statusConnectionClosed, wsReadyState: sock.readyState});
            return;
        }
        correlationID++;
        var header = {t: ampRequest, u: uri, i: correlationID};
        var msg = JSON.stringify(header) + "\n" + JSON.stringify(payload);
        requests[correlationID] = {ok: ok, fail: fail};
        try {
            sock.send(msg);
        } catch(e) {
            fail(undefined, {error: e, errorCode: statusUnknown, wsReadyState: sock.readyState});
        }
    }

    function unpackHeader(o) {
        var keys = {
            "t": "type",
            "i": "correlationID",
            "e": "error",
            "c": "errorCode",
        };

        for (var short in keys) {
            var long = keys[short];
            if (o[short] !== undefined) {
                o[long] = o[short];
                delete o[short];
            }
        }

        return o;
    }

    connect();

    return {
        request: request,
    };
}
