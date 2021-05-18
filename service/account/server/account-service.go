package server

import (
	"context"

	pb "account/api"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"go.uber.org/zap"
)

type accountServiceImpl struct {
	dal    *postgres.DataAccessLayer
	logger log.Factory
}

var _ pb.AccountServiceServer = (*accountServiceImpl)(nil)

func NewAccountService(dal *postgres.DataAccessLayer) pb.AccountServiceServer {
	return &accountServiceImpl{
		dal:    dal,
		logger: log.With(zap.String("srv", "account")),
	}
}

// CreateAccount
func (u *accountServiceImpl) Create(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	return nil, nil
}

func (u *accountServiceImpl) Delete(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	return nil, nil
}

func (u *accountServiceImpl) Update(context.Context, *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	return nil, nil
}

// ListAccounts
func (u *accountServiceImpl) List(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	return nil, nil
}

func (u *accountServiceImpl) ListStream(*pb.ListAccountsRequest, pb.AccountService_ListStreamServer) error {
	return nil
}
