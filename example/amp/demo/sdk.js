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

/***/ "./src/amp.js":
/*!********************!*\
  !*** ./src/amp.js ***!
  \********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("var messageType = {\n    publish: 0,\n    subscribe: 1,\n    request: 2,\n    response: 3,\n    ping: 4,\n    pong: 5,\n    alive: 6\n};\n\nvar updateType = {\n  diff: 0,\n  full: 1,\n  append: 2,\n  update: 3,\n  close: 4\n};\n\nvar keys = {\n  \"t\": \"type\",\n  \"i\": \"correlationID\",\n  \"e\": \"error\",\n  \"c\": \"errorCode\",\n  \"u\": \"uri\",\n  \"s\": \"ts\",\n  \"p\": \"updateType\",\n  \"b\": \"subscriptions\"\n};\n\nfunction unpackHeader(o) {\n  var header = {\n    // setting defaults\n    type: messageType.publish,\n    updateType: updateType.diff\n  };\n  for (var short in keys) {\n    var long = keys[short];\n    if (o[short] !== undefined) {\n      header[long] = o[short];\n    }\n  }\n\n  return header;\n}\n\nfunction pack(o) {\n  var header = {},\n      body = o.body;\n\n  for (var short in keys) {\n    var long = keys[short];\n    if (o[long] !== undefined) {\n      header[short] = o[long];\n    }\n  }\n\n  var buf = JSON.stringify(header);\n  if (body) {\n    buf = buf + \"\\n\";\n    buf = buf + JSON.stringify(body);\n  }\n\n  return buf;\n}\n\nfunction unpack(data) {\n  var p = data.split(\"\\n\"),\n      msg = null;\n\n  try {\n    msg = unpackHeader(JSON.parse(p[0]));\n  }catch(e){\n    return null;\n  }\n\n  if (p.length > 1 && p[1]) {\n    try{\n      var body = JSON.parse(p[1]);\n      msg[\"body\"] = body;\n    }catch(e) {\n      //console.error(e);\n    }\n  }\n\n  return msg;\n}\n\nmodule.exports = {\n  messageType: messageType,\n  updateType: updateType,\n  unpack: unpack,\n  pack: pack,\n}\n\n\n//# sourceURL=webpack://mnu5/./src/amp.js?");

/***/ }),

/***/ "./src/main.js":
/*!*********************!*\
  !*** ./src/main.js ***!
  \*********************/
/*! exports provided: api */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, \"api\", function() { return api; });\nvar amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\");\nvar sub = __webpack_require__(/*! ./subscriptions.js */ \"./src/subscriptions.js\");\nvar req = __webpack_require__(/*! ./requests.js */ \"./src/requests.js\");\n\nvar sock = null,\n    wsURI = \"\",\n    wsOnChange = undefined,\n    connectInterval = 5 * 1000;\n\nvar wsOpen = 1,\n    statusConnectionClosed = -256,\n    statusUnknown = -257;\n\nfunction defaultFail(body, e) {\n  console.error(body, {error: e, errorCode: statusUnknown, wsReadyState: sock.readyState});\n};\n\n\nfunction send(msg, fail) {\n  if (fail === undefined) {\n    fail = defaultFail;\n  }\n  if (!sock) {\n    fail(\"connection uninitialized\");\n    return;\n  }\n  if (sock.readyState !== wsOpen) {\n    fail(\"connection closed\");\n    return;\n  }\n  var data = amp.pack(msg);\n  try {\n    sock.send(data);\n  } catch(e) {\n    fail(e);\n  }\n}\n\nfunction subscribe(msg) {\n  if (msg === undefined) {\n    msg = sub.message();\n  }\n  send(msg);\n}\n\nfunction connect() {\n  sock = new WebSocket(wsURI);\n\n  sock.onopen = function() {\n    console.log(\"connected to \" + wsURI);\n    subscribe();\n    if (wsOnChange) {\n      wsOnChange(sock.readyState);\n    }\n  };\n\n  sock.onclose = function(e) {\n    setTimeout(connect, connectInterval);\n    console.log(\"connection closed\",  e.code , e);\n    if (wsOnChange) {\n      wsOnChange(sock.readyState);\n    }\n  };\n\n  sock.onmessage = function(e) {\n    try{\n      onmessage(e.data);\n    }catch(e) {\n      console.log(e);\n    }\n  };\n};\n\nfunction onmessage(data) {\n  var m = amp.unpack(data);\n  if (m === null) {\n    return;\n  }\n  switch (m.type) {\n  case amp.messageType.publish:\n    sub.publish(m);\n    break;\n  case amp.messageType.response:\n    req.response(m);\n    break;\n  case amp.messageType.alive:\n    break;\n  case amp.messageType.ping:\n    // TODO return pong message\n    break;\n  case amp.messageType.pong:\n    break;\n  }\n}\n\nfunction request(uri, payload, ok, fail) {\n  if (ok === undefined) {\n    ok = function(){};\n  }\n  if (fail === undefined) {\n    fail = defaultFail;\n  }\n  var msg = req.request(uri, payload, ok, fail);\n  send(msg, fail);\n}\n\nfunction api(uri, onChange) {\n  sub.init(subscribe);\n  wsURI = uri;\n  wsOnChange = onChange;\n  connect();\n  return {\n    request: request,\n    subscribe: sub.add,\n    unSubscribe: sub.remove,\n  };\n}\n\n\n//# sourceURL=webpack://mnu5/./src/main.js?");

/***/ }),

/***/ "./src/merge.js":
/*!**********************!*\
  !*** ./src/merge.js ***!
  \**********************/
/*! no static exports found */
/***/ (function(module, exports) {

eval("function merge(full, diff) {\n  for(var key in diff) {\n    var d = diff[key];\n    if (d === null || d === undefined) {\n      delete full[key];\n      delete full[\"_\"+key+\"Change\"];\n      continue;\n    }\n    if (typeof d === 'object') {\n      if (full[key] === undefined) {\n        full[key] = {};\n      }\n      var parent = full[\"_collection\"];\n      if (!!d[\"_isStruct\"]){\n        parent = full;\n      }\n      var current = full[key];\n      current[\"_collection\"] = full;\n      current[\"_parent\"] = parent;\n      current[\"_key\"] = key;\n      delete current[\"_list\"];\n      merge(current, d);\n      continue;\n    }\n    var prev = full[key];\n    full[key] = d;\n    if (prev !== undefined && d !== prev) {\n      full[\"_\"+key+\"Change\"] = {\n        previous: prev,\n        changedAt: (new Date).getTime(),\n      };\n    }\n  }\n}\n\nfunction sortCollection(parent) {\n  if (parent[\"_list\"] !== undefined) {\n    return parent[\"_list\"];\n  }\n  var list = [];\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      //console.log(key);\n      list.push(child);\n    }\n  }\n  list.sort(function(x, y) {\n    if (x.order === undefined) {\n      x[\"order\"] = 0;\n    }\n    if (y.order === undefined) {\n      y[\"order\"] = 0;\n    }\n    if (x.order !== y.order) {\n      return x.order - y.order;\n    }\n    if (x.name > y.name) {\n      return 1;\n    }\n    if (x.name < y.name) {\n      return -1;\n    }\n    return 0;\n  });\n  parent[\"_list\"] = list;\n  return list;\n}\n\nfunction addLists(parent) {\n  for(var key in parent) {\n    var child = parent[key];\n    if (typeof child === 'object' && key.indexOf(\"_\") !== 0) {\n      if (child._isMap === true) {\n        var listKey = key+\"List\",\n            col = child;\n        parent[listKey] = function() {\n          return sortCollection(col);\n        };\n      }\n      addLists(child);\n    }\n  }\n};\n\nmodule.exports = function(full, diff) {\n  merge(full, diff);\n  addLists(full);\n};\n\n\n//# sourceURL=webpack://mnu5/./src/merge.js?");

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

eval("var amp = __webpack_require__(/*! ./amp.js */ \"./src/amp.js\"),\n    merge = __webpack_require__(/*! ./merge.js */ \"./src/merge.js\");\n\nvar subscriptions = {},\n    onChangeHandler = null;\n\nfunction add(key, handler) {\n  var s = subscriptions[key];\n  if (s === undefined) {\n    s = {\n      key: key,\n      ts: 0,\n      handlers: [],\n      full: null,\n      diff: null\n    };\n    subscriptions[key] = s;\n    onChange();\n  } else {\n    handler(s.full, null);\n  }\n  s.handlers.push(handler);\n}\n\nfunction remove(key, handler) {\n  var s = subscriptions[key];\n  if (s === undefined) {\n    return;\n  }\n  var i = s.handlers.indexOf(handler);\n  if (i > -1) {\n    s.handlers.splice(i, 1);\n  }\n  if (s.handlers.length === 0) {\n    delete subscriptions[key];\n    onChange();\n  }\n}\n\nfunction message() {\n  var s = {};\n  for(var key in subscriptions) {\n    s[key] = subscriptions[key].ts;\n  }\n  return {type: amp.messageType.subscribe, subscriptions: s};\n}\n\nfunction publish(msg) {\n  var key = msg.uri;\n  var s = subscriptions[key];\n\n  if (!s) {\n    console.log(\"topic not found\", key);\n    return;\n  }\n\n  if (msg.updateType === amp.updateType.close) {\n    delete subscriptions[key];\n    s.handlers.forEach(function(handler){\n      handler(null, null); // signals close of the topic\n    });\n    return;\n  }\n\n  s.ts = msg.ts;\n  if (msg.updateType === amp.updateType.full) {\n    s.full = msg.body;\n    s.diff = null;\n  } else {\n    if (!s.full) {\n      s.full = {};\n    }\n    s.diff = msg.body;\n    merge(s.full, msg.body);\n  }\n\n  s.handlers.forEach(function(handler){\n    handler(s.full, s.diff);\n  });\n}\n\nfunction onChange() {\n  if (onChangeHandler == null) {\n    return;\n  }\n  onChangeHandler(message());\n}\n\nmodule.exports = {\n  add: add,\n  remove: remove,\n  publish: publish,\n  message: message,\n  init: function(handler) {\n    onChangeHandler = handler;\n  }\n};\n\n\n//# sourceURL=webpack://mnu5/./src/subscriptions.js?");

/***/ })

/******/ });