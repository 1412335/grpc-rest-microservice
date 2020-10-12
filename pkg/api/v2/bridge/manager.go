package bridge

import (
	"context"
	"errors"
	"time"

	api_v2 "grpc-rest-microservice/pkg/api/v2/gen/grpc-gateway/gen"
	"grpc-rest-microservice/pkg/utils"

	"log"

	cmap "github.com/orcaman/concurrent-map"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
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
	interceptor ClientInterceptor
}

type PoolClient struct {
	// interceptor ClientInterceptor
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

// func (poolClient *PoolClient) newFactoryClient() (*grpc.ClientConn, error) {
// 	clientInterceptor := poolClient.interceptor
// 	if clientInterceptor == nil {
// 		clientInterceptor = &SimpleClientInterceptor{}
// 		poolClient.interceptor = clientInterceptor
// 	}
// 	conn, err := grpc.Dial(
// 		poolClient.host,
// 		grpc.WithInsecure(),
// 		grpc.WithUnaryInterceptor(clientInterceptor.Unary()),
// 		grpc.WithStreamInterceptor(clientInterceptor.Stream()),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return conn, nil
// }

// new factory client with interceptor
func (poolClient *PoolClient) newFactoryClientWithInterceptor(clientInterceptor ClientInterceptor) func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		conn, err := grpc.Dial(
			poolClient.host,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(clientInterceptor.Unary()),
			grpc.WithStreamInterceptor(clientInterceptor.Stream()),
		)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

// factory client with credentials
func (poolClient *PoolClient) newFactoryClientWithCredentials(clientInterceptor ClientInterceptor) func() (*grpc.ClientConn, error) {
	return func() (*grpc.ClientConn, error) {
		auth := utils.Authentication{
			Username: "tengido",
			Password: "matkhaune",
		}
		conn, err := grpc.Dial(
			poolClient.host,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(clientInterceptor.Unary()),
			grpc.WithStreamInterceptor(clientInterceptor.Stream()),
			grpc.WithPerRPCCredentials(&auth),
		)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
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
		client:   api_v2.NewServiceExtraClient(conn.ClientConn),
		conn:     conn,
		ctx:      ctx,
		username: "admin",
		password: "admin",
	}, nil
}

func (poolClient *PoolClient) closePool() {
	poolClient.pool.Close()
}

func (managerClient *ManagerClientImpl) loadInterceptor() {
	if managerClient.interceptor == nil {
		interceptor := &SimpleClientInterceptor{}
		managerClient.interceptor = interceptor
	}
}

func (managerClient *ManagerClientImpl) newPoolClient(host string) (*PoolClient, error) {
	poolClient := &PoolClient{
		host: host,
	}

	// load client interceptor
	managerClient.loadInterceptor()

	// // grpc pool with authentication interceptor
	// p, err := grpcpool.New(poolClient.newFactoryClientWithInterceptor(managerClient.interceptor), managerClient.maxPoolSize, managerClient.maxPoolSize, time.Duration(managerClient.timeOut)*time.Second)
	// if err != nil {
	// 	return nil, err
	// }

	// grpc pool with credentials
	p, err := grpcpool.New(poolClient.newFactoryClientWithCredentials(managerClient.interceptor), managerClient.maxPoolSize, managerClient.maxPoolSize, time.Duration(managerClient.timeOut)*time.Second)
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
		poolImpl, err := managerClient.newPoolClient(host)
		if err != nil {
			log.Printf("[Manager Client] Get client with host '%v' error %+v\n", host, err)
			return nil, err
		}
		pool = poolImpl

		client, err := pool.newClient()
		if err != nil {
			managerClient.removePoolClient(host)
			log.Printf("[Manager Client] Remove pool client with host '%v' error %+v\n", host, err)
			return nil, err
		}

		// custom interceptor
		// load interceptor with client implemetation service
		// set client & gen jwt token
		const v2ServicePath = "/v2.ServiceExtra/"
		authMethods := map[string]bool{
			v2ServicePath + "Post": true,
		}
		refreshDuration := 60 * time.Second
		managerClient.interceptor.(*SimpleClientInterceptor).Load(client, authMethods, refreshDuration)

		// reload grpcpool on init
		pool, err = managerClient.addPoolClient(host)
		if err != nil {
			log.Printf("[Manager Client] Get client with host '%v' error %+v\n", host, err)
			return nil, err
		}
		client, err = pool.newClient()
		if err != nil {
			managerClient.removePoolClient(host)
			log.Printf("[Manager Client] Remove pool client with host '%v' error %+v\n", host, err)
			return nil, err
		}
		return client, nil
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
