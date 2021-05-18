package server

import (
	"context"

	pb "account/api"
	errorSrv "account/error"
	"account/model"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
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
	// validate request
	if req.GetUserId() == "" {
		return nil, errorSrv.ErrMissingUserID
	}
	if req.GetBalance() < 0 {
		return nil, errorSrv.ErrInvalidAccountBalance
	}

	// response
	rsp := &pb.CreateAccountResponse{}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		// create account
		acc := &model.Account{
			ID:      uuid.New().String(),
			UserID:  req.GetUserId(),
			Name:    req.GetName(),
			Bank:    req.GetBank().String(),
			Balance: req.GetBalance(),
		}
		if err := acc.Validate(); err != nil {
			u.logger.For(ctx).Error("Error validate account", zap.Error(err))
			return err
		}
		if err := tx.Create(acc).Error; err != nil {
			u.logger.For(ctx).Error("Error create account", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		//
		rsp.Account = acc.Transform2GRPC()
		return nil
	})
	if err != nil {
		return nil, err
	}
	// set header in your handler
	md := metadata.Pairs("X-Http-Code", "201")
	if err := grpc.SetHeader(ctx, md); err != nil {
		u.logger.For(ctx).Error("Error set header X-Http-Code", zap.Error(err))
	}
	return rsp, nil
}

func (u *accountServiceImpl) Delete(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	return nil, nil
}

func (u *accountServiceImpl) Update(context.Context, *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	return nil, nil
}

// ListAccounts
func (u *accountServiceImpl) List(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	// validate request
	if req.GetUserId() == nil {
		return nil, errorSrv.ErrMissingUserID
	}

	var account model.Account[]
	// lookup user by id
	if e := u.dal.GetDatabase().Where(&model.Account{UserID: req.GetUserId().Value}).Find(&accounts).Error; e == gorm.ErrRecordNotFound {
		return nil, errorSrv.ErrUserNotFound
	} else if e != nil {
		u.logger.For(ctx).Error("Error find user by id", zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}
	rsp := &pb.ListAccountsResponse{}
	// fetch accounts belong to the user
	rsp.Accounts = make([]*pb.Account, len(accounts))
	for i, acc := range accounts {
		rsp.Accounts[i] = acc.Transform2GRPC()
	}
	return rsp, nil
}

func (u *accountServiceImpl) ListStream(*pb.ListAccountsRequest, pb.AccountService_ListStreamServer) error {
	return nil
}
