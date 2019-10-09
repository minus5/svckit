function merge(full, diff) {
  for (var key in diff) {
    var d = diff[key];

    if (d === null || d === undefined) {
      delete full[key];
      delete full["_" + key + "Change"];
      continue;
    }

    if (typeof d === 'object') {
      if (full[key] === undefined) {
        full[key] = {};
      }

      var parent = full["_collection"];

      if (!!d["_isStruct"]) {
        parent = full;
      }

      var current = full[key];
      current["_collection"] = full;
      current["_parent"] = parent;
      current["_key"] = key;
      delete current["_list"];
      merge(current, d);
      continue;
    }

    var prev = full[key];
    full[key] = d;

    if (prev !== undefined && d !== prev) {
      full["_" + key + "Change"] = {
        previous: prev,
        changedAt: new Date().getTime()
      };
    }
  }
}

function sortCollection(parent) {
  if (parent["_list"] !== undefined) {
    return parent["_list"];
  }

  var list = [];

  for (var key in parent) {
    var child = parent[key];

    if (typeof child === 'object' && key.indexOf("_") !== 0) {
      //console.log(key);
      list.push(child);
    }
  }

  list.sort(function (x, y) {
    if (x.order === undefined) {
      x["order"] = 0;
    }

    if (y.order === undefined) {
      y["order"] = 0;
    }

    if (x.order !== y.order) {
      return x.order - y.order;
    }

    if (x.name > y.name) {
      return 1;
    }

    if (x.name < y.name) {
      return -1;
    }

    return 0;
  });
  parent["_list"] = list;
  return list;
}

function addLists(parent) {
  for (var key in parent) {
    var child = parent[key];

    if (typeof child === 'object' && key.indexOf("_") !== 0) {
      if (child._isMap === true) {
        var listKey = key + "List",
            col = child;

        parent[listKey] = function () {
          return sortCollection(col);
        };
      }

      addLists(child);
    }
  }
}

;

module.exports = function (full, diff) {
  merge(full, diff);
  addLists(full);
};