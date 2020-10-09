package bridge

import (
	"context"
	"errors"
	"fmt"
	"time"

	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"

	"log"

	cmap "github.com/orcaman/concurrent-map"
	grpcpool "github.com/processout/grpc-go-pool"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	IDLE              = "IDLE"
	CONNECTING        = "CONNECTING"
	READY             = "READY"
	TRANSIENT_FAILURE = "TRANSIENT_FAILURE"
	SHUTDOWN          = "SHUTDOWN"
	INVALID           = "Invalid-State"
)

type ManagerClient interface {
	GetClient(host string) (Client, error)
	Close()
}

type ManagerClientImpl struct {
	maxPoolSize int
	timeOut     int
	poolClients cmap.ConcurrentMap
}

type PoolClient struct {
	pool *grpcpool.Pool
	host string
}

func NewManagerClient(maxPoolSize, timeOut int) ManagerClient {
	return &ManagerClientImpl{
		maxPoolSize: maxPoolSize,
		timeOut:     timeOut,
		poolClients: cmap.New(),
	}
}

// client interceptor for unary request
func unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	start := time.Now()

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	// fetch response header
	var header metadata.MD
	opts = append(opts, grpc.Header(&header))

	// invoke request
	err = invoker(ctx, method, req, reply, cc, opts...)

	// get x-response-id header
	xrespid := header.Get("x-response-id")

	log.Printf("[gRPC client] Invoked RPC method=%s, xrid=%s, xrespid=%v, duration=%v, resp='%+v', error='%v'", method, xrid, xrespid, time.Since(start), reply, err)

	return err
}

// client streaming interceptor
func streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	start := time.Now()

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	clientStream, err = streamer(ctx, desc, cc, method, opts...)

	// get x-response-id header
	// NOT WORK: not using stream context
	// md, ok := metadata.FromIncomingContext(clientStream.Context())
	// if !ok {
	// 	return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	// }
	// xrespid := md.Get("x-response-id")

	var xrespid []string
	var customHeader []string
	header, ok := clientStream.Header()
	if ok == nil {
		xrespid = header.Get("x-response-id")
		customHeader = header.Get("custom-resp-header")
	}

	log.Printf("[gRPC client] Stream RPC method=%s, xrid=%s, xrespid=%v, customHeader=%v, duration=%v, error='%v'", method, xrid, xrespid, customHeader, time.Since(start), err)

	return clientStream, err
}

func (poolClient *PoolClient) newFactoryClient() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		poolClient.host,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(unaryClientInterceptor),
		grpc.WithStreamInterceptor(streamClientInterceptor),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (poolClient *PoolClient) newClient() (Client, error) {
	ctx := context.Background()

	conn, err := poolClient.pool.Get(ctx)
	if err != nil {
		return nil, err
	}

	// check status conn
	state := conn.GetState().String()
	log.Println("[PoolClient] State connection: " + state)
	if state == TRANSIENT_FAILURE || state == SHUTDOWN || state == INVALID {
		return nil, errors.New("Pool connection failed")
	}

	return &ClientImpl{
		client: api_v2.NewServiceExtraClient(conn.ClientConn),
		conn:   conn,
		ctx:    ctx,
	}, nil
}

func (poolClient *PoolClient) closePool() {
	poolClient.pool.Close()
}

func (managerClient *ManagerClientImpl) newPoolClient(host string) (*PoolClient, error) {
	poolClient := &PoolClient{
		host: host,
	}

	p, err := grpcpool.New(poolClient.newFactoryClient, managerClient.maxPoolSize, managerClient.maxPoolSize, time.Duration(managerClient.timeOut)*time.Second)
	if err != nil {
		return nil, err
	}

	poolClient.pool = p

	return poolClient, nil
}

func (managerClient *ManagerClientImpl) addPoolClient(host string) (*PoolClient, error) {
	poolClient, err := managerClient.newPoolClient(host)
	if err != nil {
		return nil, err
	}

	managerClient.poolClients.Set(host, poolClient)

	return poolClient, nil
}

func (managerClient *ManagerClientImpl) getPoolClient(host string) *PoolClient {
	pool, ok := managerClient.poolClients.Get(host)
	if !ok {
		return nil
	}

	return pool.(*PoolClient)
}

func (managerClient *ManagerClientImpl) removePoolClient(host string) {
	managerClient.poolClients.Remove(host)
}

func (managerClient *ManagerClientImpl) GetClient(host string) (Client, error) {
	pool := managerClient.getPoolClient(host)
	if pool == nil {
		poolImpl, err := managerClient.addPoolClient(host)
		if err != nil {
			log.Printf("[Manager Client] Get client with host '%v' error %+v\n", host, err)
			return nil, err
		}
		pool = poolImpl
	}

	client, err := pool.newClient()
	if err != nil {
		managerClient.removePoolClient(host)
		log.Printf("[Manager Client] Remove pool client with host '%v' error %+v\n", host, err)
		return nil, err
	}
	return client, nil
}

func (managerClient *ManagerClientImpl) Close() {
	for item := range managerClient.poolClients.Iter() {
		poolClient := item.Val.(*PoolClient)
		poolClient.closePool()
	}
}
