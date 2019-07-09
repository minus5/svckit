window["minus5"] = window["minus5"] || {}; window["minus5"]["api"] =
/******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId]) {
/******/ 			return installedModules[moduleId].exports;
/******/ 		}
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			i: moduleId,
/******/ 			l: false,
/******/ 			exports: {}
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.l = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// define getter function for harmony exports
/******/ 	__webpack_require__.d = function(exports, name, getter) {
/******/ 		if(!__webpack_require__.o(exports, name)) {
/******/ 			Object.defineProperty(exports, name, { enumerable: true, get: getter });
/******/ 		}
/******/ 	};
/******/
/******/ 	// define __esModule on exports
/******/ 	__webpack_require__.r = function(exports) {
/******/ 		if(typeof Symbol !== 'undefined' && Symbol.toStringTag) {
/******/ 			Object.defineProperty(exports, Symbol.toStringTag, { value: 'Module' });
/******/ 		}
/******/ 		Object.defineProperty(exports, '__esModule', { value: true });
/******/ 	};
/******/
/******/ 	// create a fake namespace object
/******/ 	// mode & 1: value is a module id, require it
/******/ 	// mode & 2: merge all properties of value into the ns
/******/ 	// mode & 4: return value when already ns object
/******/ 	// mode & 8|1: behave like require
/******/ 	__webpack_require__.t = function(value, mode) {
/******/ 		if(mode & 1) value = __webpack_require__(value);
/******/ 		if(mode & 8) return value;
/******/ 		if((mode & 4) && typeof value === 'object' && value && value.__esModule) return value;
/******/ 		var ns = Object.create(null);
/******/ 		__webpack_require__.r(ns);
/******/ 		Object.defineProperty(ns, 'default', { enumerable: true, value: value });
/******/ 		if(mode & 2 && typeof value != 'string') for(var key in value) __webpack_require__.d(ns, key, function(key) { return value[key]; }.bind(null, key));
/******/ 		return ns;
/******/ 	};
/******/
/******/ 	// getDefaultExport function for compatibility with non-harmony modules
/******/ 	__webpack_require__.n = function(module) {
/******/ 		var getter = module && module.__esModule ?
/******/ 			function getDefault() { return module['default']; } :
/******/ 			function getModuleExports() { return module; };
/******/ 		__webpack_require__.d(getter, 'a', getter);
/******/ 		return getter;
/******/ 	};
/******/
/******/ 	// Object.prototype.hasOwnProperty.call
/******/ 	__webpack_require__.o = function(object, property) { return Object.prototype.hasOwnProperty.call(object, property); };
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(__webpack_require__.s = "./src/main.js");
/******/ })
/************************************************************************/
/******/ ({

/***/ "../../../../../../../../usr/local/lib/node_modules/webpack/buildin/global.js":
/*!***********************************!*\
  !*** (webpack)/buildin/global.js ***!
  \***********************************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("var g;\n\n// This works in non-strict mode\ng = (function() {\n\treturn this;\n})();\n\ntry {\n\t// This works if eval is allowed (see CSP)\n\tg = g || new Function(\"return this\")();\n} catch (e) {\n\t// This works if the window reference is available\n\tif (typeof window === \"object\") g = window;\n}\n\n// g can still be undefined, but nothing to do about it...\n// We return undefined, instead of nothing here, so it's\n// easier to handle this case. if(!global) { ...}\n\nmodule.exports = g;\n\n\n//# sourceURL=webpack://minus5.api/(webpack)/buildin/global.js?");

/***/ }),

/***/ "./node_modules/nanoajax/index.js":
/*!****************************************!*\
  !*** ./node_modules/nanoajax/index.js ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("/* WEBPACK VAR INJECTION */(function(global) {// Best place to find information on XHR features is:\n// https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest\n\nvar reqfields = [\n  'responseType', 'withCredentials', 'timeout', 'onprogress'\n]\n\n// Simple and small ajax function\n// Takes a parameters object and a callback function\n// Parameters:\n//  - url: string, required\n//  - headers: object of `{header_name: header_value, ...}`\n//  - body:\n//      + string (sets content type to 'application/x-www-form-urlencoded' if not set in headers)\n//      + FormData (doesn't set content type so that browser will set as appropriate)\n//  - method: 'GET', 'POST', etc. Defaults to 'GET' or 'POST' based on body\n//  - cors: If your using cross-origin, you will need this true for IE8-9\n//\n// The following parameters are passed onto the xhr object.\n// IMPORTANT NOTE: The caller is responsible for compatibility checking.\n//  - responseType: string, various compatability, see xhr docs for enum options\n//  - withCredentials: boolean, IE10+, CORS only\n//  - timeout: long, ms timeout, IE8+\n//  - onprogress: callback, IE10+\n//\n// Callback function prototype:\n//  - statusCode from request\n//  - response\n//    + if responseType set and supported by browser, this is an object of some type (see docs)\n//    + otherwise if request completed, this is the string text of the response\n//    + if request is aborted, this is \"Abort\"\n//    + if request times out, this is \"Timeout\"\n//    + if request errors before completing (probably a CORS issue), this is \"Error\"\n//  - request object\n//\n// Returns the request object. So you can call .abort() or other methods\n//\n// DEPRECATIONS:\n//  - Passing a string instead of the params object has been removed!\n//\nexports.ajax = function (params, callback) {\n  // Any variable used more than once is var'd here because\n  // minification will munge the variables whereas it can't munge\n  // the object access.\n  var headers = params.headers || {}\n    , body = params.body\n    , method = params.method || (body ? 'POST' : 'GET')\n    , called = false\n\n  var req = getRequest(params.cors)\n\n  function cb(statusCode, responseText) {\n    return function () {\n      if (!called) {\n        callback(req.status === undefined ? statusCode : req.status,\n                 req.status === 0 ? \"Error\" : (req.response || req.responseText || responseText),\n                 req)\n        called = true\n      }\n    }\n  }\n\n  req.open(method, params.url, true)\n\n  var success = req.onload = cb(200)\n  req.onreadystatechange = function () {\n    if (req.readyState === 4) success()\n  }\n  req.onerror = cb(null, 'Error')\n  req.ontimeout = cb(null, 'Timeout')\n  req.onabort = cb(null, 'Abort')\n\n  if (body) {\n    setDefault(headers, 'X-Requested-With', 'XMLHttpRequest')\n\n    if (!global.FormData || !(body instanceof global.FormData)) {\n      setDefault(headers, 'Content-Type', 'application/x-www-form-urlencoded')\n    }\n  }\n\n  for (var i = 0, len = reqfields.length, field; i < len; i++) {\n    field = reqfields[i]\n    if (params[field] !== undefined)\n      req[field] = params[field]\n  }\n\n  for (var field in headers)\n    req.setRequestHeader(field, headers[field])\n\n  req.send(body)\n\n  return req\n}\n\nfunction getRequest(cors) {\n  // XDomainRequest is only way to do CORS in IE 8 and 9\n  // But XDomainRequest isn't standards-compatible\n  // Notably, it doesn't allow cookies to be sent or set by servers\n  // IE 10+ is standards-compatible in its XMLHttpRequest\n  // but IE 10 can still have an XDomainRequest object, so we don't want to use it\n  if (cors && global.XDomainRequest && !/MSIE 1/.test(navigator.userAgent))\n    return new XDomainRequest\n  if (global.XMLHttpRequest)\n    return new XMLHttpRequest\n}\n\nfunction setDefault(obj, key, value) {\n  obj[key] = obj[key] || value\n}\n\n/* WEBPACK VAR INJECTION */}.call(this, __webpack_require__(/*! ./../../../../../../../../../../usr/local/lib/node_modules/webpack/buildin/global.js */ \"../../../../../../../../usr/local/lib/node_modules/webpack/buildin/global.js\")))\n\n//# sourceURL=webpack://minus5.api/./node_modules/nanoajax/index.js?");

/***/ }),

/***/ "./src/amp.js":
/*!********************!*\
  !*** ./src/amp.js ***!
  \********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("var messageType = {\n    publish: 0,\n    subscribe: 1,\n    request: 2,\n    response: 3,\n    ping: 4,\n    pong: 5,\n    alive: 6\n};\n\nvar updateType = {\n  diff: 0,\n  full: 1,\n  append: 2,\n  update: 3,\n  close: 4\n};\n\nvar keys = {\n  \"t\": \"type\",\n  \"i\": \"correlationID\",\n  \"e\": \"error\",\n  \"u\": \"uri\",\n  \"s\": \"ts\",\n  \"p\": \"updateType\",\n  \"b\": \"subscriptions\"\n};\n\nvar errorKeys = {\n  \"s\": \"source\",\n  \"m\": \"message\",\n  \"c\": \"code\"\n};\n\nfunction unpackHeader(o) {\n  var header = {\n    // setting defaults\n    type: messageType.publish,\n  };\n\n  unpackObject(o, header, keys);\n  if (header.error) {\n    header.error = {};\n    unpackObject(o.e, header.error, errorKeys);\n  }\n\n  if (header.type == messageType.publish && header.updateType === undefined)  {\n    header[\"updateType\"] = updateType.diff;\n  }\n  return header;\n}\n\nfunction unpackObject(source, dest, keys) {\n  for (var short in keys) {\n    var long = keys[short];\n    if (source[short] !== undefined) {\n      dest[long] = source[short];\n    }\n  }\n}\n\nfunction pack(o) {\n  var header = {},\n      body = o.body;\n\n  for (var short in keys) {\n    var long = keys[short];\n    if (o[long] !== undefined) {\n      header[short] = o[long];\n    }\n  }\n\n  var buf = JSON.stringify(header);\n  if (body) {\n    buf = buf + \"\\n\";\n    buf = buf + JSON.stringify(body);\n  }\n\n  return buf;\n}\n\nfunction unpackMsg(data) {\n  var p = data.split(\"\\n\"),\n      msg = null;\n\n  try {\n    msg = unpackHeader(JSON.parse(p[0]));\n  }catch(e){\n    return null;\n  }\n\n  if (p.length > 1 && p[1]) {\n    try{\n      var body = JSON.parse(p[1]);\n      msg[\"body\"] = body;\n    }catch(e) {\n      //console.error(e);\n    }\n  }\n\n  return msg;\n}\n\nfunction unpack(data) {\n  if (!data) {\n    return null;\n  }\n  var p = data.split(\"\\n\\n\");\n  var msgs = [];\n  for(var i=0; i<p.length; i++) {\n    var m = unpackMsg(p[i]);\n    msgs[i] = m;\n  }\n  return msgs;\n}\n\nfunction now() {\n  return (new Date()).getTime();\n}\n\nmodule.exports = {\n  messageType: messageType,\n  updateType: updateType,\n  unpack: unpack,\n  unpackMsg: unpackMsg,\n  pack: pack,\n  ping: function(ts) {return {type: messageType.ping, ts: (ts || now()) }; },\n  pong: function() {return {type: messageType.pong}; }\n}\n\n\n//# sourceURL=webpack://minus5.api/./src/amp.js?");

/***/ }),

/***/ "./src/errors.js":
/*!***********************!*\
  !*** ./src/errors.js ***!
  \***********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("/*\nDefines errors raised during communication.\nThere are two main types of errors:\n  * transport failures\n  * errors response from server side\n\nIn the first case message has never reached server side, or server has responded within 200 range.\nIn the second case we successfully received response but there is server side error during procesing request.\n\nIf application needs to distinguish these types of errors it should call isTransport or isServer helpers.\nFor example to decide weather the error is transient or permanent. In the case of transport failures retry can help.\n*/\n\nvar sources = {\n  application: 0,\n  transport: 1\n};\n\nfunction create(source, message) {\n  return {\n    message: message,\n    isTransport: source == sources.transport,\n    isApplication: source == sources.application,\n  };\n}\n\nfunction pooling(code, responseText) {\n  var desc = \"code: \" + code + \", \" + responseText;\n  return create(sources.transport, desc);\n}\n\nfunction ws(desc) {\n  return create(sources.transport, desc);\n}\n\nfunction server(msg) {\n  var e = msg ? msg.error : {};\n  return create((e.source || sources.application), e.message);\n}\n\nmodule.exports = {\n  pooling,\n  ws,\n  server,\n};\n\n\n//# sourceURL=webpack://minus5.api/./src/errors.js?");

/***/ }),

/***/ "./src/log.js":
/*!********************!*\
  !*** ./src/log.js ***!
  \********************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var nanoajax = __webpack_require__(/*! nanoajax */ \"./node_modules/nanoajax/index.js\");\n\nfunction ajax(url, msg, success, fail) {\n  nanoajax.ajax({\n    url: url,\n    method: 'POST',\n    body: JSON.stringify(msg),\n  }, function(code, responseText){\n    if (code>=200&&code<300) {\n      // success(responseText);\n    }else {\n      // fail(code, responseText);\n      console.error(code, responseText);\n    }\n  });\n}\n\nmodule.exports = function(uri) {\n  return {\n    info:  function(msg) { ajax(uri+\"/info\", msg); },\n    error: function(msg) { ajax(uri+\"/error\", msg); },\n  };\n}\n\n\n//# sourceURL=webpack://minus5.api/./src/log.js?");

/***/ }),

/***/ "./src/main.js":
/*!*********************!*\
  !*** ./src/main.js ***!
  \*********************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp       = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar Sub       = __webpack_require__(/*! ./subscriptions.js */ \"./src/subscriptions.js\");\nvar Req       = __webpack_require__(/*! ./requests.js */ \"./src/requests.js\");\nvar Log       = __webpack_require__(/*! ./log.js */ \"./src/log.js\");\nvar Ws        = __webpack_require__(/*! ./ws.js */ \"./src/ws.js\");\nvar Pooling   = __webpack_require__(/*! ./pooling.js */ \"./src/pooling.js\");\n\nvar urls = {\n  port: function() {\n    return (location.port === '' || location.port === '80') ? '' : (':' + location.port);\n  },\n  path: function() {\n    var pn = location.pathname;\n    var path = pn.substring(0, pn.lastIndexOf('/') + 1);\n    if (path.length === 0)  {\n      path = \"/\";\n    }\n    return path;\n  },\n  ws() {\n    var protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';\n    return protocol + location.hostname + urls.port() + urls.path() + 'api';\n  },\n  log: function() {\n    return location.protocol + \"//\" + location.hostname + urls.port() + urls.path() + 'log';\n  },\n  pooling: function() {\n    return location.protocol + \"//\" + location.hostname + urls.port() + urls.path() + 'pooling';\n  },\n  forcePooling: function() {\n    return location.search.search(\"forcePooling\") > -1;\n  }\n};\n\n\n//export function api(config)\n//\nmodule.exports = function(config) {\n  config = config || {\n    onTransportChange: function() {},\n    logTransportChanges: false\n  };\n\n  var sub    = Sub(subscribe);\n  var logger = null;\n  var req    = Req();\n\n  if (config.logTransportChanges) {\n    logger = Log(urls.log());\n  }\n\n  var transport = {\n        current: undefined,\n        previous: undefined,\n        ws: undefined,\n        pooling: undefined,\n        onChange: function(){},\n        ready() { return !!transport.current; },\n        name: function(t) {\n          return t === transport.ws ? \"ws\" : t === transport.pooling ? \"pooling\" : \"none\";\n        },\n        send: function(msg, fail)  {\n          // pooling is always available\n          // use it while ws-pooling handshake is done\n          var tr = transport.current || transport.pooling;\n          tr.send(msg, fail);\n        }\n      };\n  var failHandlers = {\n    default: function(e) { console.error(e);},\n    ignore:  function(e) {}\n  };\n\n  function send(msg, fail) {\n    fail = fail || failHandlers.default;\n    transport.send(msg, fail);\n  }\n\n  function subscribe(msg) {\n    if (!transport.ready()) {\n      // skip if not ready, will be send later\n      // when transport is set\n      // console.info(\"subscribe while transport not ready, queueing...\");\n      return;\n    }\n    msg = msg || sub.message();\n    send(msg, failHandlers.ignore);\n  }\n\n  function onMessage(data) {\n    if (!data) {\n      return false;\n    }\n    var msgs = amp.unpack(data),\n        pongReceived = false;\n    for (var i=0; i<msgs.length; i++) {\n      var m = msgs[i];\n      if (m === null) {\n        return pongReceived;\n      }\n      switch (m.type) {\n      case amp.messageType.publish:\n        sub.publish(m);\n        break;\n      case amp.messageType.response:\n        req.response(m);\n        break;\n      case amp.messageType.alive:\n        break;\n      case amp.messageType.ping:\n        send(amp.pong(), failHandlers.ignore);\n        break;\n      case amp.messageType.pong:\n        pongReceived = true;\n        break;\n      }\n    }\n    return pongReceived;\n  }\n\n  function request(uri, payload, ok, fail) {\n    ok = ok ||  function(){};\n    fail = fail || failHandlers.default;\n    var msg = req.request(uri, payload, ok, fail);\n    send(msg, fail);\n  }\n\n  function onWsChange(status) {\n    transport.previous = transport.current;\n    transport.current = status.success ?  transport.ws : transport.pooling;\n    if ((transport.current === transport.ws && status.connected) ||\n        (transport.current === transport.pooling)) {\n      subscribe();\n    }\n    if (transport.previous === transport.pooling && transport.current=== transport.ws) {\n      transport.pooling.stop();\n    }\n\n    if (logger) {\n      if (status.connected)  {\n        logger.info(status);\n      } else {\n        logger.error(status);\n      }\n    }\n\n    transport.onChange( {\n      transport: transport.name(transport.current),\n      previousTransport: transport.name(transport.previous),\n      status: status});\n  }\n\n  transport.onChange = function(status) {\n    if (config.onTransportChange) {\n      try {\n        config.onTransportChange(status);\n      }catch(e){\n        console.error(e);\n      }\n    }\n  };\n\n  transport.pooling = Pooling(urls.pooling(), onMessage, sub.message);\n  if (urls.forcePooling()) {\n    transport.current = transport.pooling;\n  } else {\n    transport.ws = Ws(urls.ws(), onMessage, onWsChange);\n  }\n\n  return {\n    request: request,\n    subscribe: sub.add,\n    unSubscribe: sub.remove,\n  };\n}\n\n\n//# sourceURL=webpack://minus5.api/./src/main.js?");

/***/ }),

/***/ "./src/merge.js":
/*!**********************!*\
  !*** ./src/merge.js ***!
  \**********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("function merge(full, diff) {\n  for(var key in diff) {\n    var d = diff[key];\n    if (d === null || d === undefined) {\n      delete full[key];\n      delete full[\"_\"+key+\"Change\"];\n      continue;\n    }\n    if (typeof d === 'object') {\n      if (full[key] === undefined) {\n        full[key] = {};\n      }\n      var parent = full[\"_collection\"];\n      if (!!d[\"_isStruct\"]){\n        parent = full;\n      }\n      var current = full[key];\n      current[\"_collection\"] = full;\n      current[\"_parent\"] = parent;\n      current[\"_key\"] = key;\n      delete current[\"_list\"];\n      merge(current, d);\n      continue;\n    }\n    var prev = full[key];\n    full[key] = d;\n    if (prev !== undefined && d !== prev) {\n      full[\"_\"+key+\"Change\"] = {\n        previous: prev,\n        changedAt: (new Date).getTime(),\n      };\n    }\n  }\n}\n\nfunction sortCollection(parent) {\n  if (parent[\"_list\"] !== undefined) {\n    return parent[\"_list\"];\n  }\n  var list = [];\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      //console.log(key);\n      list.push(child);\n    }\n  }\n  list.sort(function(x, y) {\n    if (x.order === undefined) {\n      x[\"order\"] = 0;\n    }\n    if (y.order === undefined) {\n      y[\"order\"] = 0;\n    }\n    if (x.order !== y.order) {\n      return x.order - y.order;\n    }\n    if (x.name > y.name) {\n      return 1;\n    }\n    if (x.name < y.name) {\n      return -1;\n    }\n    return 0;\n  });\n  parent[\"_list\"] = list;\n  return list;\n}\n\nfunction addLists(parent) {\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      if (child._isMap === true) {\n        var listKey = key+\"List\",\n            col = child;\n        parent[listKey] = function() {\n          return sortCollection(col);\n        };\n      }\n      addLists(child);\n    }\n  }\n};\n\nmodule.exports = function(full, diff) {\n  merge(full, diff);\n  addLists(full);\n};\n\n\n//# sourceURL=webpack://minus5.api/./src/merge.js?");

/***/ }),

/***/ "./src/pooling.js":
/*!************************!*\
  !*** ./src/pooling.js ***!
  \************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp  = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar errors  = __webpack_require__(/*! ./errors.js */ \"./src/errors.js\");\nvar nanoajax = __webpack_require__(/*! nanoajax */ \"./node_modules/nanoajax/index.js\");\n\nmodule.exports = function(uri, onMessages, subMessage) {\n\n  var processes = {},\n      stopped = false;\n\n  function ajax(msg, success, fail) {\n    var data = amp.pack(msg);\n\n    nanoajax.ajax({\n      url: uri,\n      method: 'POST',\n      body: data,\n    }, function(code, responseText){\n      if (code>=200&&code<300) {\n        success(responseText);\n      }else {\n        fail(errors.pooling(code, responseText));\n      }\n    });\n  }\n\n  function subscribe(msg) {\n    for (var key in  msg.subscriptions) {\n      if (processes[key] !== undefined) {\n        continue;\n      }\n\n      var ts = msg.subscriptions[key];\n      processes[key] = ts;\n      var m = {type: amp.messageType.subscribe, subscriptions: {}};\n      m.subscriptions[key] = ts;\n      ajax(m, function(data) {\n        delete processes[key];\n        if (data) {\n          onMessages(data);\n        }\n        if (stopped) {\n          return;\n        }\n        subscribe(subMessage());\n      },function(code, rsp) {\n        delete processes[key];\n        if (code >= 400 && code < 500){\n          return; // bad request and friends\n        }\n        console.error(code, rsp);\n        if (stopped) {\n          return;\n        }\n        setTimeout(function() {\n          subscribe(subMessage());\n        }, 4 * 1000);\n      });\n    }\n  }\n\n  function send(msg, fail) {\n    if (msg.type == amp.messageType.subscribe) {\n      subscribe(msg);\n      return;\n    }\n    ajax(msg, onMessages, fail);\n  }\n\n  return {\n    send: send,\n    stop: function() { stopped = true; }\n  };\n}\n\n\n//# sourceURL=webpack://minus5.api/./src/pooling.js?");

/***/ }),

/***/ "./src/requests.js":
/*!*************************!*\
  !*** ./src/requests.js ***!
  \*************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar errors  = __webpack_require__(/*! ./errors.js */ \"./src/errors.js\");\n\nmodule.exports = function() {\n\n  var correlationID=0,\n      requests = {};\n\n  // find response handlers and call them\n  function response(m) {\n    var r = requests[m.correlationID];\n    if (!r) {\n    }\n\n    delete requests[m.correlationID];\n    if (m.error) {\n      r.fail(errors.server(m));\n    } else {\n      r.ok(m.body);\n    }\n  }\n\n  // create request message and store handlers (ok, fail) into requests\n  function request(uri, payload, ok, fail) {\n    correlationID++;\n    var msg = {\n      type: amp.messageType.request,\n      uri: uri,\n      correlationID: correlationID,\n      body: payload\n    };\n    requests[correlationID] = {ok: ok, fail: fail};\n    return msg;\n  }\n\n  return {\n    request: request,\n    response: response\n  };\n  \n};\n\n\n//# sourceURL=webpack://minus5.api/./src/requests.js?");

/***/ }),

/***/ "./src/subscriptions.js":
/*!******************************!*\
  !*** ./src/subscriptions.js ***!
  \******************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\"),\n    merge = __webpack_require__(/*! ./merge.js */ \"./src/merge.js\");\n\nmodule.exports = function(onChangeHandler) {\n\n  var subscriptions = {};\n\n  function add(key, handler) {\n    var s = subscriptions[key];\n    if (s === undefined) {\n      s = {\n        key: key,\n        ts: 0,\n        handlers: [],\n        full: null,\n        diff: null\n      };\n      subscriptions[key] = s;\n      onChange();\n    } else {\n      handler(s.full, null);\n    }\n    s.handlers.push(handler);\n  }\n\n  function remove(key, handler) {\n    var s = subscriptions[key];\n    if (s === undefined) {\n      return;\n    }\n    var i = s.handlers.indexOf(handler);\n    if (i > -1) {\n      s.handlers.splice(i, 1);\n    }\n    if (s.handlers.length === 0) {\n      delete subscriptions[key];\n      onChange();\n    }\n  }\n\n  function message() {\n    var s = {};\n    for(var key in subscriptions) {\n      s[key] = subscriptions[key].ts;\n    }\n    return {type: amp.messageType.subscribe, subscriptions: s};\n  }\n\n  function publish(msg) {\n    var key = msg.uri;\n    var s = subscriptions[key];\n\n    if (!s) {\n      console.log(\"topic not found\", key);\n      return;\n    }\n\n    if (msg.updateType === amp.updateType.close) {\n      delete subscriptions[key];\n      s.handlers.forEach(function(handler){\n        handler(null, null); // signals close of the topic\n      });\n      return;\n    }\n\n    s.ts = msg.ts;\n    if (msg.updateType === amp.updateType.full ||\n        msg.updateType === amp.updateType.append ||\n        msg.updateType === amp.updateType.update) {\n      s.full = msg.body;\n      s.diff = null;\n    }\n    if (msg.updateType === amp.updateType.diff) {\n      if (!s.full) {\n        s.full = {};\n      }\n      s.diff = msg.body;\n      merge(s.full, msg.body);\n    }\n    s.handlers.forEach(function(handler){\n      handler(s.full, s.diff);\n    });\n  }\n\n  function onChange() {\n    if (onChangeHandler == null) {\n      return;\n    }\n    onChangeHandler(message());\n  }\n\n  return {\n    add: add,\n    remove: remove,\n    publish: publish,\n    message: message\n  };\n\n};\n\n\n//# sourceURL=webpack://minus5.api/./src/subscriptions.js?");

/***/ }),

/***/ "./src/ws.js":
/*!*******************!*\
  !*** ./src/ws.js ***!
  \*******************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp  = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar errors  = __webpack_require__(/*! ./errors.js */ \"./src/errors.js\");\n\nfunction now() {\n  return (new Date()).getTime();\n}\n\nmodule.exports = function(uri, onMessage_, onChange_) { // TODO get rid of this suffix_\n\n  function onChange(status) {\n    try{\n      onChange_(status);\n    } catch(e) {\n      console.log(e);\n    }\n  };\n\n  function onMessage(data) { // expecting that onMessage returns true if it is a pong message\n    try{\n      return onMessage_(data);\n    } catch(e) {\n      console.log(e);\n    }\n    return false;\n  };\n\n  var ws = null,\n      pong = {\n        timer: undefined,\n        schedule: function(handler) {\n          var interval = 16 * 1000;\n          pong.timer = setTimeout(handler, interval);\n        },\n        clear: function() {\n          clearTimeout(pong.timer);\n        },\n        onMessage: function(isPong) {\n          if (isPong) {\n            pong.clear();\n          }\n        },\n        start: function() {\n          pong.schedule(function() {\n            if (ws.readyState != WebSocket.OPEN) { // connection is closed\n              return;\n            }\n            status.event(\"pongTimeout\");\n            ws.close();\n          });\n        }\n      },\n      ping = {\n        timer: undefined,\n        no: 0,\n        lastMessage: 0,\n        afterPongInterval: 16 * 1000,\n        beforePongInterval: 4 * 1000,\n        interval: 4 * 1000,\n        clear: function() {\n          clearTimeout(ping.timer);\n        },\n        start: function() {\n          ping.interval = ping.beforePongInterval;\n          ping.lastMessage = 0;\n          ping.loop();\n        },\n        loop: function() {\n          if (now() - ping.lastMessage > ping.interval / 2) {\n            ping.no++;\n            send(amp.ping(ping.no), function(e) {\n              status.event(\"pingError\", e);\n            });\n          }\n          ping.timer = setTimeout(ping.loop, ping.interval);\n        },\n        onMessage: function(isPong) {\n          ping.lastMessage = now();\n          if (isPong) {\n            ping.interval = ping.afterPongInterval;\n          }\n        }\n      },\n      status = {\n        success: false,\n        opened: false,\n        fallback: false,\n        giveup: false,\n        connected: false,\n        supported: false,\n        start: now(),\n        startConnect: now(),\n        messages: 0,\n        connects: 0,\n        retries: 0,\n        events: [],\n        onMessage: function(isPong) {\n          status.messages++;\n          if (isPong && !status.connected) {\n            // handle first pong message\n            // pong messages are only send as reply to ping\n            // indicates that connection works in both directions\n            status.event(\"pong\");\n            status.success = true;\n            status.connected = true;\n            status.retries = status.connects;\n            status.change(); // signal success\n          }\n        },\n        event: function(name, e) {\n          var o = {name: name, sinceStart: now() - status.start, sinceConnect: now() - status.startConnect};\n          if (e) {\n            if (e.code) {\n              o[\"code\"] = e.code;\n            }\n            if (e.type) {\n              o[\"type\"] = e.type;\n            }\n            if (e.reason) {\n              o[\"reason\"] = e.reason;\n            }\n            if (e.message) {\n              o[\"message\"] = e.message;\n            }\n            if (e.name) {\n              o[\"name\"] = e.name;\n            }\n            o[\"error\"] = e.toString();\n          }\n          status.events.push(o);\n        },\n        change: function() {\n          onChange(status);\n        },\n        shouldQuit: function() {\n          status.connects++;\n          status.startConnect = now();\n          if (status.success) {\n            return false;\n          }\n          if (status.connects > 32) {\n            status.giveup = true;\n            status.change(); // signal give up\n            return true;\n          }\n          if (status.connects === 5) {\n            status.fallback = true;\n            status.change(); // signal fallback\n          }\n          return false;\n        },\n        // calculates exponential increasing interval based on number of connects\n        connectInterval: function() {\n          var p = status.connects || 1;\n          if (p > 12) {\n            p = 12; // 4096 max\n          }\n          return  Math.pow(2, p);\n        }\n      };\n\n  function send(msg, fail) {\n    function err(no, msg, e) {\n      fail(errors.ws(msg));\n      status.event(\"sendError\"+no, e);\n    }\n\n    if (!ws) {\n      err(1, \"connection uninitialized\");\n      return;\n    }\n    if (ws.readyState !== WebSocket.OPEN) {\n      err(2, \"connection closed readyState: \" + ws.readyState);\n      return;\n    }\n    var data = amp.pack(msg);\n    try {\n      ws.send(data);\n    } catch(e) {\n      err(3, e.toString(), e);\n    }\n  }\n\n  function connect() {\n    if (status.shouldQuit()) {\n      return;\n    }\n\n    function reconnect() {\n      pong.clear();\n      ping.clear();\n      setTimeout(connect, status.connectInterval());\n      status.connected = false;\n    }\n\n    try {\n      ws = new WebSocket(uri);\n    } catch (e) {\n      reconnect();\n      status.event(\"wsError\", e);\n      return;\n    }\n\n    ws.onopen = function() {\n      status.opened = true;\n      pong.start();\n      ping.start();\n      status.event(\"open\");\n    };\n\n    ws.onclose = function(e) {\n      reconnect();\n      status.event(\"close\", e);\n    };\n\n    ws.onmessage = function(e) {\n      var isPong = onMessage(e.data);\n      status.onMessage(isPong);\n      ping.onMessage(isPong);\n      pong.onMessage(isPong);\n    };\n\n  };\n\n  status.supported = (\"WebSocket\" in window && window.WebSocket != undefined);\n  if (!status.supported) {\n    onChange();\n    return undefined;\n  }\n\n  connect();\n\n  return {\n    send: send\n  };\n}\n\n\n//# sourceURL=webpack://minus5.api/./src/ws.js?");

/***/ })

/******/ });