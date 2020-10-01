const { MessagePing } = require('../gen/common_pb.js');

const { ServiceAClient: ServiceAClient } = require('../gen/rest-service_grpc_web_pb.js');
const { ServiceAClient: ServiceAClientBinary } = require('../gen/rest-service_grpc_web_pb.js');


// set xhr2 to nodeJs can call gRPC-Web
global.XMLHttpRequest = require('xhr2');


// const { ServiceExtraClient } = require('../rest-service-extra_grpc_web_pb.js');

const TARGET = {
    ENVOY: "http://localhost:7070"
}

console.log("envoy_host =>", TARGET.ENVOY)

var clientText = new ServiceAClient(TARGET.ENVOY);
var clientBinary = new ServiceAClientBinary(TARGET.ENVOY);

var request = new MessagePing();
request.setTimestamp(11111);

clientText.ping(request, {}, (err, response) => {
    if (err) {
        console.log(err)
    } else {
        console.log(response);
    }
});

clientBinary.ping(request, {}, (err, response) => {
    if (err) {
        console.log(err)
    } else {
        console.log(response);
    }
});