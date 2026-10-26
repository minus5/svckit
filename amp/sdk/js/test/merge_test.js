let assert = require("assert");
let merge = require("../src/merge.js");

describe('merge', function() {

  it("should drop null keys of a new subtree", function() {
    let full = {};
    merge(full, {m: {a: null, b: 1}}, true, true);
    assert.deepStrictEqual(full, {m: {b: 1}});
  });

  it("should add a key again after a new subtree deleted it", function() {
    let full = {};
    merge(full, {m: {a: null}}, true);
    merge(full, {m: {a: {x: 1}}}, true);
    assert.strictEqual(full.m.a.x, 1);
  });

  it("should add a key again after a new subtree deleted it without enrichment", function() {
    let full = {};
    merge(full, {m: {a: null}}, true, true);
    merge(full, {m: {a: {x: 1}}}, true, true);
    assert.deepStrictEqual(full, {m: {a: {x: 1}}});
  });

  it("should not keep diff objects in the merged state", function() {
    let full = {};
    let diff = {m: {a: 1}};
    merge(full, diff, true, true);
    diff.m.a = 2;
    assert.strictEqual(full.m.a, 1);
  });

  it("should not add enrichment keys to a new subtree", function() {
    let full = {};
    merge(full, {m: {o: {odds: 1}}}, true);
    assert.deepStrictEqual(full, {m: {o: {odds: 1}}});
    assert.doesNotThrow(() => JSON.stringify(full));
  });

  it("should replace a primitive with an object", function() {
    let full = {};
    merge(full, {m: 5}, true);
    merge(full, {m: 6}, true);
    merge(full, {m: {x: 1}}, true);
    assert.deepStrictEqual(full, {m: {x: 1}});
  });

  it("should replace an array with an object", function() {
    let full = {};
    merge(full, {m: [1, 2]}, true, true);
    merge(full, {m: {x: 1}}, true, true);
    assert.deepStrictEqual(full, {m: {x: 1}});
  });

});
