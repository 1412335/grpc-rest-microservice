/**
 * @fileoverview gRPC-Web generated client stub for v2
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var common_pb = require('./common_pb.js')
const proto = {};
proto.v2 = require('./rest-service_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.v2.ServiceAClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.v2.ServiceAPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'binary';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.v2.MessagePing,
 *   !proto.v2.MessagePong>}
 */
const methodDescriptor_ServiceA_Ping = new grpc.web.MethodDescriptor(
  '/v2.ServiceA/Ping',
  grpc.web.MethodType.UNARY,
  common_pb.MessagePing,
  common_pb.MessagePong,
  /**
   * @param {!proto.v2.MessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.MessagePong.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.v2.MessagePing,
 *   !proto.v2.MessagePong>}
 */
const methodInfo_ServiceA_Ping = new grpc.web.AbstractClientBase.MethodInfo(
  common_pb.MessagePong,
  /**
   * @param {!proto.v2.MessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.MessagePong.deserializeBinary
);


/**
 * @param {!proto.v2.MessagePing} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.v2.MessagePong)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.v2.MessagePong>|undefined}
 *     The XHR Node Readable Stream
 */
proto.v2.ServiceAClient.prototype.ping =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/v2.ServiceA/Ping',
      request,
      metadata || {},
      methodDescriptor_ServiceA_Ping,
      callback);
};


/**
 * @param {!proto.v2.MessagePing} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.v2.MessagePong>}
 *     Promise that resolves to the response
 */
proto.v2.ServiceAPromiseClient.prototype.ping =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/v2.ServiceA/Ping',
      request,
      metadata || {},
      methodDescriptor_ServiceA_Ping);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.v2.MessagePing,
 *   !proto.v2.MessagePong>}
 */
const methodDescriptor_ServiceA_Post = new grpc.web.MethodDescriptor(
  '/v2.ServiceA/Post',
  grpc.web.MethodType.UNARY,
  common_pb.MessagePing,
  common_pb.MessagePong,
  /**
   * @param {!proto.v2.MessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.MessagePong.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.v2.MessagePing,
 *   !proto.v2.MessagePong>}
 */
const methodInfo_ServiceA_Post = new grpc.web.AbstractClientBase.MethodInfo(
  common_pb.MessagePong,
  /**
   * @param {!proto.v2.MessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.MessagePong.deserializeBinary
);


/**
 * @param {!proto.v2.MessagePing} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.v2.MessagePong)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.v2.MessagePong>|undefined}
 *     The XHR Node Readable Stream
 */
proto.v2.ServiceAClient.prototype.post =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/v2.ServiceA/Post',
      request,
      metadata || {},
      methodDescriptor_ServiceA_Post,
      callback);
};


/**
 * @param {!proto.v2.MessagePing} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.v2.MessagePong>}
 *     Promise that resolves to the response
 */
proto.v2.ServiceAPromiseClient.prototype.post =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/v2.ServiceA/Post',
      request,
      metadata || {},
      methodDescriptor_ServiceA_Post);
};


module.exports = proto.v2;

