var api = minus5.api(function(status){
  console.log("ws status", status);
});

api.subscribe("math.v1/non-existing", function(data) {});
