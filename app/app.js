var express = require("express")
var router = require("./router")

var app = express();
app.use(router)

var server = app.listen(3333, function () {
  var host = server.address().address;
  var port = server.address().port;

  console.log('Example app listening at http://%s:%s', host, port);
});
