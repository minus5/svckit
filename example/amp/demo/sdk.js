window["mnu5"] =
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

eval("var g;\n\n// This works in non-strict mode\ng = (function() {\n\treturn this;\n})();\n\ntry {\n\t// This works if eval is allowed (see CSP)\n\tg = g || new Function(\"return this\")();\n} catch (e) {\n\t// This works if the window reference is available\n\tif (typeof window === \"object\") g = window;\n}\n\n// g can still be undefined, but nothing to do about it...\n// We return undefined, instead of nothing here, so it's\n// easier to handle this case. if(!global) { ...}\n\nmodule.exports = g;\n\n\n//# sourceURL=webpack://mnu5/(webpack)/buildin/global.js?");

/***/ }),

/***/ "./node_modules/nanoajax/index.js":
/*!****************************************!*\
  !*** ./node_modules/nanoajax/index.js ***!
  \****************************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("/* WEBPACK VAR INJECTION */(function(global) {// Best place to find information on XHR features is:\n// https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest\n\nvar reqfields = [\n  'responseType', 'withCredentials', 'timeout', 'onprogress'\n]\n\n// Simple and small ajax function\n// Takes a parameters object and a callback function\n// Parameters:\n//  - url: string, required\n//  - headers: object of `{header_name: header_value, ...}`\n//  - body:\n//      + string (sets content type to 'application/x-www-form-urlencoded' if not set in headers)\n//      + FormData (doesn't set content type so that browser will set as appropriate)\n//  - method: 'GET', 'POST', etc. Defaults to 'GET' or 'POST' based on body\n//  - cors: If your using cross-origin, you will need this true for IE8-9\n//\n// The following parameters are passed onto the xhr object.\n// IMPORTANT NOTE: The caller is responsible for compatibility checking.\n//  - responseType: string, various compatability, see xhr docs for enum options\n//  - withCredentials: boolean, IE10+, CORS only\n//  - timeout: long, ms timeout, IE8+\n//  - onprogress: callback, IE10+\n//\n// Callback function prototype:\n//  - statusCode from request\n//  - response\n//    + if responseType set and supported by browser, this is an object of some type (see docs)\n//    + otherwise if request completed, this is the string text of the response\n//    + if request is aborted, this is \"Abort\"\n//    + if request times out, this is \"Timeout\"\n//    + if request errors before completing (probably a CORS issue), this is \"Error\"\n//  - request object\n//\n// Returns the request object. So you can call .abort() or other methods\n//\n// DEPRECATIONS:\n//  - Passing a string instead of the params object has been removed!\n//\nexports.ajax = function (params, callback) {\n  // Any variable used more than once is var'd here because\n  // minification will munge the variables whereas it can't munge\n  // the object access.\n  var headers = params.headers || {}\n    , body = params.body\n    , method = params.method || (body ? 'POST' : 'GET')\n    , called = false\n\n  var req = getRequest(params.cors)\n\n  function cb(statusCode, responseText) {\n    return function () {\n      if (!called) {\n        callback(req.status === undefined ? statusCode : req.status,\n                 req.status === 0 ? \"Error\" : (req.response || req.responseText || responseText),\n                 req)\n        called = true\n      }\n    }\n  }\n\n  req.open(method, params.url, true)\n\n  var success = req.onload = cb(200)\n  req.onreadystatechange = function () {\n    if (req.readyState === 4) success()\n  }\n  req.onerror = cb(null, 'Error')\n  req.ontimeout = cb(null, 'Timeout')\n  req.onabort = cb(null, 'Abort')\n\n  if (body) {\n    setDefault(headers, 'X-Requested-With', 'XMLHttpRequest')\n\n    if (!global.FormData || !(body instanceof global.FormData)) {\n      setDefault(headers, 'Content-Type', 'application/x-www-form-urlencoded')\n    }\n  }\n\n  for (var i = 0, len = reqfields.length, field; i < len; i++) {\n    field = reqfields[i]\n    if (params[field] !== undefined)\n      req[field] = params[field]\n  }\n\n  for (var field in headers)\n    req.setRequestHeader(field, headers[field])\n\n  req.send(body)\n\n  return req\n}\n\nfunction getRequest(cors) {\n  // XDomainRequest is only way to do CORS in IE 8 and 9\n  // But XDomainRequest isn't standards-compatible\n  // Notably, it doesn't allow cookies to be sent or set by servers\n  // IE 10+ is standards-compatible in its XMLHttpRequest\n  // but IE 10 can still have an XDomainRequest object, so we don't want to use it\n  if (cors && global.XDomainRequest && !/MSIE 1/.test(navigator.userAgent))\n    return new XDomainRequest\n  if (global.XMLHttpRequest)\n    return new XMLHttpRequest\n}\n\nfunction setDefault(obj, key, value) {\n  obj[key] = obj[key] || value\n}\n\n/* WEBPACK VAR INJECTION */}.call(this, __webpack_require__(/*! ./../../../../../../../../../../usr/local/lib/node_modules/webpack/buildin/global.js */ \"../../../../../../../../usr/local/lib/node_modules/webpack/buildin/global.js\")))\n\n//# sourceURL=webpack://mnu5/./node_modules/nanoajax/index.js?");

/***/ }),

/***/ "./src/amp.js":
/*!********************!*\
  !*** ./src/amp.js ***!
  \********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("var messageType = {\n    publish: 0,\n    subscribe: 1,\n    request: 2,\n    response: 3,\n    ping: 4,\n    pong: 5,\n    alive: 6\n};\n\nvar updateType = {\n  diff: 0,\n  full: 1,\n  append: 2,\n  update: 3,\n  close: 4\n};\n\nvar keys = {\n  \"t\": \"type\",\n  \"i\": \"correlationID\",\n  \"e\": \"error\",\n  \"c\": \"errorCode\",\n  \"u\": \"uri\",\n  \"s\": \"ts\",\n  \"p\": \"updateType\",\n  \"b\": \"subscriptions\"\n};\n\nfunction unpackHeader(o) {\n  var header = {\n    // setting defaults\n    type: messageType.publish,\n    updateType: updateType.diff\n  };\n  for (var short in keys) {\n    var long = keys[short];\n    if (o[short] !== undefined) {\n      header[long] = o[short];\n    }\n  }\n\n  return header;\n}\n\nfunction pack(o) {\n  var header = {},\n      body = o.body;\n\n  for (var short in keys) {\n    var long = keys[short];\n    if (o[long] !== undefined) {\n      header[short] = o[long];\n    }\n  }\n\n  var buf = JSON.stringify(header);\n  if (body) {\n    buf = buf + \"\\n\";\n    buf = buf + JSON.stringify(body);\n  }\n\n  return buf;\n}\n\nfunction unpackMsg(data) {\n  var p = data.split(\"\\n\"),\n      msg = null;\n\n  try {\n    msg = unpackHeader(JSON.parse(p[0]));\n  }catch(e){\n    return null;\n  }\n\n  if (p.length > 1 && p[1]) {\n    try{\n      var body = JSON.parse(p[1]);\n      msg[\"body\"] = body;\n    }catch(e) {\n      //console.error(e);\n    }\n  }\n\n  return msg;\n}\n\nfunction unpack(data) {\n  var p = data.split(\"\\n\\n\");\n  var msgs = [];\n  for(var i=0; i<p.length; i++) {\n    var m = unpackMsg(p[i]);\n    msgs[i] = m;\n  }\n  return msgs;\n}\n\nmodule.exports = {\n  messageType: messageType,\n  updateType: updateType,\n  unpack: unpack,\n  unpackMsg: unpackMsg,\n  pack: pack,\n}\n\n\n//# sourceURL=webpack://mnu5/./src/amp.js?");

/***/ }),

/***/ "./src/main.js":
/*!*********************!*\
  !*** ./src/main.js ***!
  \*********************/
/*! exports provided: api */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"api\", function() { return api; });\nvar amp  = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar sub  = __webpack_require__(/*! ./subscriptions.js */ \"./src/subscriptions.js\");\nvar req  = __webpack_require__(/*! ./requests.js */ \"./src/requests.js\");\nvar ws   = __webpack_require__(/*! ./ws.js */ \"./src/ws.js\");\nvar pool = __webpack_require__(/*! ./pooling.js */ \"./src/pooling.js\");\n\nvar _transport = undefined,\n    _onTransportChange = undefined,\n    //_pool,\n    statusUnknown = -257;          // TODO define/rethink this error codes\n\nfunction defaultFail(body, e) {\n  console.error(body, {error: e, errorCode: statusUnknown, transportState: _transport.state()});\n};\n\nfunction ignoreFail() {}\n\nfunction send(msg, fail) {\n  if (fail === undefined) {\n    fail = defaultFail;\n  }\n  _transport.send(msg, fail);\n}\n\nfunction subscribe(msg) {\n  if (msg === undefined) {\n    msg = sub.message();\n  }\n  send(msg, ignoreFail);\n\n  //_pool.send(msg, ignoreFail);\n}\n\nfunction onMessage(data) {\n  var msgs = amp.unpack(data);\n  for (var i=0; i<msgs.length; i++) {\n    var m = msgs[i];\n    if (m === null) {\n      return;\n    }\n    switch (m.type) {\n    case amp.messageType.publish:\n      sub.publish(m);\n      break;\n    case amp.messageType.response:\n      req.response(m);\n      break;\n    case amp.messageType.alive:\n      break;\n    case amp.messageType.ping:\n      // TODO return pong message\n      break;\n    case amp.messageType.pong:\n      break;\n    }\n  }\n}\n\nfunction request(uri, payload, ok, fail) {\n  if (ok === undefined) {\n    ok = function(){};\n  }\n  if (fail === undefined) {\n    fail = defaultFail;\n  }\n  var msg = req.request(uri, payload, ok, fail);\n  send(msg, fail);\n}\n\nfunction onChange(status) {\n  if (status == 1) {\n    subscribe();\n  }\n  if (_onTransportChange) {\n    _onTransportChange(status);\n  }\n}\n\n\n\nfunction api(uri, onTransportChange) {\n  _onTransportChange = onTransportChange;\n  //_transport = ws.init(uri, onMessage, onChange);\n  _transport = pool.init(\"http://localhost/pooling\", onMessage);\n\n  sub.init(subscribe);\n  return {\n    request: request,\n    subscribe: sub.add,\n    unSubscribe: sub.remove,\n  };\n}\n\n\n//# sourceURL=webpack://mnu5/./src/main.js?");

/***/ }),

/***/ "./src/merge.js":
/*!**********************!*\
  !*** ./src/merge.js ***!
  \**********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("function merge(full, diff) {\n  for(var key in diff) {\n    var d = diff[key];\n    if (d === null || d === undefined) {\n      delete full[key];\n      delete full[\"_\"+key+\"Change\"];\n      continue;\n    }\n    if (typeof d === 'object') {\n      if (full[key] === undefined) {\n        full[key] = {};\n      }\n      var parent = full[\"_collection\"];\n      if (!!d[\"_isStruct\"]){\n        parent = full;\n      }\n      var current = full[key];\n      current[\"_collection\"] = full;\n      current[\"_parent\"] = parent;\n      current[\"_key\"] = key;\n      delete current[\"_list\"];\n      merge(current, d);\n      continue;\n    }\n    var prev = full[key];\n    full[key] = d;\n    if (prev !== undefined && d !== prev) {\n      full[\"_\"+key+\"Change\"] = {\n        previous: prev,\n        changedAt: (new Date).getTime(),\n      };\n    }\n  }\n}\n\nfunction sortCollection(parent) {\n  if (parent[\"_list\"] !== undefined) {\n    return parent[\"_list\"];\n  }\n  var list = [];\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      //console.log(key);\n      list.push(child);\n    }\n  }\n  list.sort(function(x, y) {\n    if (x.order === undefined) {\n      x[\"order\"] = 0;\n    }\n    if (y.order === undefined) {\n      y[\"order\"] = 0;\n    }\n    if (x.order !== y.order) {\n      return x.order - y.order;\n    }\n    if (x.name > y.name) {\n      return 1;\n    }\n    if (x.name < y.name) {\n      return -1;\n    }\n    return 0;\n  });\n  parent[\"_list\"] = list;\n  return list;\n}\n\nfunction addLists(parent) {\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      if (child._isMap === true) {\n        var listKey = key+\"List\",\n            col = child;\n        parent[listKey] = function() {\n          return sortCollection(col);\n        };\n      }\n      addLists(child);\n    }\n  }\n};\n\nmodule.exports = function(full, diff) {\n  merge(full, diff);\n  addLists(full);\n};\n\n\n//# sourceURL=webpack://mnu5/./src/merge.js?");

/***/ }),

/***/ "./src/pooling.js":
/*!************************!*\
  !*** ./src/pooling.js ***!
  \************************/
/*! exports provided: init */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"init\", function() { return init; });\nvar sub  = __webpack_require__(/*! ./subscriptions.js */ \"./src/subscriptions.js\");\nvar amp  = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar nanoajax = __webpack_require__(/*! nanoajax */ \"./node_modules/nanoajax/index.js\");\nvar _processes = {},\n    _uri,\n    _onMessages = undefined;\n    ;\n\n\nfunction ajax(msg, success, fail) {\n  var data = amp.pack(msg);\n\n  nanoajax.ajax({\n    url: _uri,\n    method: 'POST',\n    body: data,\n  }, function(code, responseText){\n    if (code>=200&&code<300) {\n      success(responseText);\n    }else {\n      fail(code, responseText);\n      console.error(code, responseText);\n    }\n  });\n}\n\n\nfunction subscribe(msg) {\n  for (var key in  msg.subscriptions) {\n    if (_processes[key] !== undefined) {\n      continue;\n    }\n\n    var ts = msg.subscriptions[key];\n    _processes[key] = ts;\n    var m = {type: amp.messageType.subscribe, subscriptions: {}};\n    m.subscriptions[key] = ts;\n    ajax(m, function(data) {\n      delete _processes[key];\n      _onMessages(data);\n      subscribe(sub.message());\n    },function(code, rsp) {\n      delete _processes[key];\n      console.error(code, rsp);\n      //subscribe(sub.message());\n    });\n  }\n}\n\n\nfunction send(msg, fail) {\n  if (msg.type == amp.messageType.subscribe) {\n    subscribe(msg);\n    return;\n  }\n  ajax(msg, _onMessages, fail);\n}\n\nfunction init(uri, onMessages) {\n  _uri = uri;\n  _onMessages = onMessages;\n\n  return {\n    send: send,\n    state: function() { return 0; } // TODO napravi ping prvi put\n  };\n}\n\n\n//# sourceURL=webpack://mnu5/./src/pooling.js?");

/***/ }),

/***/ "./src/requests.js":
/*!*************************!*\
  !*** ./src/requests.js ***!
  \*************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\n\nvar correlationID=0,\n    requests = {};\n\n// find response handlers and call them\nfunction response(m) {\n  var r = requests[m.correlationID];\n  if (r) {\n    delete requests[m.correlationID];\n    if (m.error) {\n      r.fail(m.body, m);\n    } else {\n      r.ok(m.body);\n    }\n  }\n}\n\n// create request message\nfunction request(uri, payload, ok, fail) {\n  correlationID++;\n  var msg = {\n    type: amp.messageType.request,\n    uri: uri,\n    correlationID: correlationID,\n    body: payload\n  };\n  requests[correlationID] = {ok: ok, fail: fail};\n  return msg;\n}\n\nmodule.exports = {\n  request: request,\n  response: response\n};\n\n\n//# sourceURL=webpack://mnu5/./src/requests.js?");

/***/ }),

/***/ "./src/subscriptions.js":
/*!******************************!*\
  !*** ./src/subscriptions.js ***!
  \******************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

eval("var amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\"),\n    merge = __webpack_require__(/*! ./merge.js */ \"./src/merge.js\");\n\nvar subscriptions = {},\n    onChangeHandler = null;\n\nfunction add(key, handler) {\n  var s = subscriptions[key];\n  if (s === undefined) {\n    s = {\n      key: key,\n      ts: 0,\n      handlers: [],\n      full: null,\n      diff: null\n    };\n    subscriptions[key] = s;\n    onChange();\n  } else {\n    handler(s.full, null);\n  }\n  s.handlers.push(handler);\n}\n\nfunction remove(key, handler) {\n  var s = subscriptions[key];\n  if (s === undefined) {\n    return;\n  }\n  var i = s.handlers.indexOf(handler);\n  if (i > -1) {\n    s.handlers.splice(i, 1);\n  }\n  if (s.handlers.length === 0) {\n    delete subscriptions[key];\n    onChange();\n  }\n}\n\nfunction message() {\n  var s = {};\n  for(var key in subscriptions) {\n    s[key] = subscriptions[key].ts;\n  }\n  return {type: amp.messageType.subscribe, subscriptions: s};\n}\n\nfunction publish(msg) {\n  var key = msg.uri;\n  var s = subscriptions[key];\n\n  if (!s) {\n    console.log(\"topic not found\", key);\n    return;\n  }\n\n  if (msg.updateType === amp.updateType.close) {\n    delete subscriptions[key];\n    s.handlers.forEach(function(handler){\n      handler(null, null); // signals close of the topic\n    });\n    return;\n  }\n\n  s.ts = msg.ts;\n  if (msg.updateType === amp.updateType.full ||\n      msg.updateType === amp.updateType.append ||\n      msg.updateType === amp.updateType.update) {\n    s.full = msg.body;\n    s.diff = null;\n  }\n  if (msg.updateType === amp.updateType.diff) {\n    if (!s.full) {\n      s.full = {};\n    }\n    s.diff = msg.body;\n    merge(s.full, msg.body);\n  }\n  s.handlers.forEach(function(handler){\n    handler(s.full, s.diff);\n  });\n}\n\nfunction onChange() {\n  if (onChangeHandler == null) {\n    return;\n  }\n  onChangeHandler(message());\n}\n\nmodule.exports = {\n  add: add,\n  remove: remove,\n  publish: publish,\n  message: message,\n  init: function(handler) {\n    onChangeHandler = handler;\n  }\n};\n\n\n//# sourceURL=webpack://mnu5/./src/subscriptions.js?");

/***/ }),

/***/ "./src/ws.js":
/*!*******************!*\
  !*** ./src/ws.js ***!
  \*******************/
/*! exports provided: init */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"init\", function() { return init; });\nvar amp  = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\n\nvar sock = null,\n    _uri = \"\",\n    _onChange = undefined,\n    _onMessage = undefined,\n    _connectInterval = 5 * 1000;\n\nvar wsOpen = 1;\n\nfunction send(msg, fail) {\n  if (!sock) {\n    fail(\"connection uninitialized\");\n    return;\n  }\n  if (sock.readyState !== wsOpen) {\n    fail(\"connection closed\");\n    return;\n  }\n  var data = amp.pack(msg);\n  try {\n    sock.send(data);\n  } catch(e) {\n    fail(e);\n  }\n}\n\nfunction connect() {\n  sock = new WebSocket(_uri);\n\n  sock.onopen = function() {\n    console.log(\"connected to \" + _uri);\n    if (_onChange) {\n      _onChange(sock.readyState);\n    }\n  };\n\n  sock.onclose = function(e) {\n    setTimeout(connect, _connectInterval);\n    console.log(\"connection closed\",  e.code , e);\n    if (_onChange) {\n      _onChange(sock.readyState);\n    }\n  };\n\n  sock.onmessage = function(e) {\n    try{\n      _onMessage(e.data);\n    }catch(e) {\n      console.log(e);\n    }\n  };\n};\n\nfunction init(uri, onMessage, onChange) {\n  _uri = uri;\n  _onChange = onChange;\n  _onMessage = onMessage;\n  connect();\n\n  return {\n    send: send,\n    state: function() { return sock.readyState; }\n  };\n}\n\n\n//# sourceURL=webpack://mnu5/./src/ws.js?");

/***/ })

/******/ });