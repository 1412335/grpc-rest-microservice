const { MessagePing, MessagePong, StreamingMessagePing, StreamingMessagePong } = require('../gen/common_pb.js');

const { ServiceAClient: ServiceAClient } = require('../gen/rest-service_grpc_web_pb.js');
const { ServiceAClient: ServiceAClientBinary } = require('../gen/rest-service_grpc_web_pb_binary.js');

const { ServiceExtraClient } = require('../gen/rest-service-extra_grpc_web_pb');
const { ServiceExtraClient: ServiceExtraClientBinary } = require('../gen/rest-service-extra_grpc_web_pb_binary');

// set xhr2 to nodeJs can call gRPC-Web
global.XMLHttpRequest = require('xhr2');


// interceptor implementation
const StreamResponseInterceptor = function () { }

const TARGET = {
    ENVOY: "http://localhost:7070"
}

// request testing
console.log("envoy_host =>", TARGET.ENVOY)

var clientText = new ServiceAClient(TARGET.ENVOY);
var clientBinary = new ServiceAClientBinary(TARGET.ENVOY);
var clientExtra = new ServiceExtraClient(TARGET.ENVOY);
var clientExtraBinary = new ServiceExtraClientBinary(TARGET.ENVOY);

var request = new MessagePing();
request.setTimestamp(11111);

clientText.ping(request, {}, (err, response) => {
    if (err) {
        console.log(err)
    } else {
        console.log("grpcwebtext ping =>", response);
    }
});

clientBinary.ping(request, {}, (err, response) => {
    if (err) {
        console.log(err)
    } else {
        console.log("grpcWeb ping =>", response);
    }
});


// 
const { Service } = require('./service');
console.log('Init services with grpcwebtext mode...')
var service = new Service(
    {
        clientText,
        clientBinary,
        clientExtra,
        clientExtraBinary,
    },
    {
        MessagePing: MessagePing,
        MessagePong: MessagePong,
        StreamingMessagePing: StreamingMessagePing,
    }
)

service.load()