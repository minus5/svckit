function merge(full, diff, skipEnrichment) {
  for(let key in diff) {
    let d = diff[key];
    if (d === null || d === undefined) {
      delete full[key];
      delete full["_"+key+"Change"];
      continue;
    }
    if (typeof d === 'object' && !Array.isArray(d)) {
      // no enrichment: new subtrees never had _collection/_parent/_key and consumers iterate their keys
      if (typeof full[key] !== 'object' || full[key] === null || Array.isArray(full[key])) {
        delete full["_"+key+"Change"];
        full[key] = {};
        merge(full[key], d, true);
        continue;
      }
      let parent = full["_collection"];
      if (!!d["_isStruct"]){
        parent = full;
      }
      let current = full[key];
      if (!skipEnrichment) {
        current["_collection"] = full;
        current["_parent"] = parent;
        current["_key"] = key;
        delete current["_list"];
      }
      merge(current, d, skipEnrichment);
      continue;
    }
    let prev = full[key];
    full[key] = d;
    if (!skipEnrichment && prev !== undefined && d !== prev) {
      full["_"+key+"Change"] = {
        previous: prev,
        changedAt: (new Date).getTime(),
      };
    }
  }
}

function sortCollection(parent) {
  if (parent["_list"] !== undefined) {
    return parent["_list"];
  }
  let list = [];
  for(let key in parent) {
    let child = parent[key];
    if (typeof child === 'object' && key.indexOf("_") !== 0) {
      list.push(child);
    }
  }
  list.sort(function(x, y) {
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
  for(let key in parent) {
    let child = parent[key];
    if (child && typeof child === 'object' && key.indexOf("_") !== 0) {
      if (child._isMap === true) {
        let listKey = key+"List",
            col = child;
        parent[listKey] = function() {
          return sortCollection(col);
        };
      }
      addLists(child);
    }
  }
};

module.exports = function (full, diff, skipLists, skipEnrichment) {
  merge(full, diff, skipEnrichment);
  if (!skipLists) { 
    addLists(full);
  }
};
