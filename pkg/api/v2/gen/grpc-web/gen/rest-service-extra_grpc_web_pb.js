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


// var google_api_annotations_pb = require('./google/api/annotations_pb.js')

var common_pb = require('./common_pb.js')
const proto = {};
proto.v2 = require('./rest-service-extra_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.v2.ServiceExtraClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

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
proto.v2.ServiceExtraPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

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
const methodDescriptor_ServiceExtra_Ping = new grpc.web.MethodDescriptor(
  '/v2.ServiceExtra/Ping',
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
const methodInfo_ServiceExtra_Ping = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.v2.ServiceExtraClient.prototype.ping =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/v2.ServiceExtra/Ping',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_Ping,
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
proto.v2.ServiceExtraPromiseClient.prototype.ping =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/v2.ServiceExtra/Ping',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_Ping);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.v2.MessagePing,
 *   !proto.v2.MessagePong>}
 */
const methodDescriptor_ServiceExtra_Post = new grpc.web.MethodDescriptor(
  '/v2.ServiceExtra/Post',
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
const methodInfo_ServiceExtra_Post = new grpc.web.AbstractClientBase.MethodInfo(
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
proto.v2.ServiceExtraClient.prototype.post =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/v2.ServiceExtra/Post',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_Post,
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
proto.v2.ServiceExtraPromiseClient.prototype.post =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/v2.ServiceExtra/Post',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_Post);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.v2.StreamingMessagePing,
 *   !proto.v2.StreamingMessagePong>}
 */
const methodDescriptor_ServiceExtra_StreamingPing = new grpc.web.MethodDescriptor(
  '/v2.ServiceExtra/StreamingPing',
  grpc.web.MethodType.SERVER_STREAMING,
  common_pb.StreamingMessagePing,
  common_pb.StreamingMessagePong,
  /**
   * @param {!proto.v2.StreamingMessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.StreamingMessagePong.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.v2.StreamingMessagePing,
 *   !proto.v2.StreamingMessagePong>}
 */
const methodInfo_ServiceExtra_StreamingPing = new grpc.web.AbstractClientBase.MethodInfo(
  common_pb.StreamingMessagePong,
  /**
   * @param {!proto.v2.StreamingMessagePing} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  common_pb.StreamingMessagePong.deserializeBinary
);


/**
 * @param {!proto.v2.StreamingMessagePing} request The request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.v2.StreamingMessagePong>}
 *     The XHR Node Readable Stream
 */
proto.v2.ServiceExtraClient.prototype.streamingPing =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/v2.ServiceExtra/StreamingPing',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_StreamingPing);
};


/**
 * @param {!proto.v2.StreamingMessagePing} request The request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!grpc.web.ClientReadableStream<!proto.v2.StreamingMessagePong>}
 *     The XHR Node Readable Stream
 */
proto.v2.ServiceExtraPromiseClient.prototype.streamingPing =
    function(request, metadata) {
  return this.client_.serverStreaming(this.hostname_ +
      '/v2.ServiceExtra/StreamingPing',
      request,
      metadata || {},
      methodDescriptor_ServiceExtra_StreamingPing);
};


module.exports = proto.v2;

