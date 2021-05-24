package server

import (
	"context"
	"fmt"
	"time"

	pb "account/api"
	errorSrv "account/error"
	"account/model"

	"github.com/1412335/grpc-rest-microservice/pkg/dal/postgres"
	"github.com/1412335/grpc-rest-microservice/pkg/errors"
	"github.com/1412335/grpc-rest-microservice/pkg/log"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// get user by id from redis & db
func (u *accountServiceImpl) getAccountByID(ctx context.Context, account *model.Account) error {
	logger := u.logger.With(zap.String("id", account.ID), zap.String("userId", account.UserID)).For(ctx)
	// get from cache
	if e := account.GetCache(); e != nil {
		logger.Error("Get account cache", zap.Error(e))
	} else {
		return nil
	}
	return u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		if e := tx.First(account).Error; e == gorm.ErrRecordNotFound {
			return errorSrv.ErrAccountNotFound
		} else if e != nil {
			logger.Error("Lookup account", zap.Error(e))
			return errorSrv.ErrConnectDB
		}
		// cache
		if e := account.Cache(); e != nil {
			logger.Error("Cache account", zap.Error(e))
		}
		return nil
	})
}

// build query statement & get list users
func (u *accountServiceImpl) getAccounts(ctx context.Context, req *pb.ListAccountsRequest) ([]*pb.Account, error) {
	var accounts []model.Account
	// build sql statement
	psql := u.dal.GetDatabase().WithContext(ctx)
	if req.GetUserId() != nil {
		psql = psql.Where("user_id = ?", req.GetUserId().Value)
	}
	if req.GetId() != nil {
		psql = psql.Where("id = ?", req.GetId().Value)
	}
	if req.GetName() != nil {
		psql = psql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", req.GetName().Value))
	}
	if req.GetBalanceMin() != nil {
		psql = psql.Where("balance >= ?", req.GetBalanceMin().Value)
	}
	if req.GetBalanceMax() != nil {
		psql = psql.Where("balance <= ?", req.GetBalanceMax().Value)
	}
	if req.GetCreatedSince() != nil {
		psql = psql.Where("created_at >= ?", req.GetCreatedSince().AsTime())
	}
	if req.GetOlderThen() != nil {
		psql = psql.Where("created_at >= ?", time.Now().Add(req.GetOlderThen().AsDuration()))
	}
	// exec
	if err := psql.Order("created_at desc").Find(&accounts).Error; err != nil {
		u.logger.For(ctx).Error("Lookup accounts", zap.Error(err))
		return nil, errorSrv.ErrConnectDB
	}
	// check empty from db
	if len(accounts) == 0 {
		return nil, errorSrv.ErrAccountNotFound
	}
	// filter
	rsp := make([]*pb.Account, len(accounts))
	for i, account := range accounts {
		rsp[i] = account.Transform2GRPC()
	}
	return rsp, nil
}

//nolint:unused
func (u *accountServiceImpl) getAccountsByUserID(ctx context.Context, userID string) ([]*pb.Account, error) {
	// validate request
	if userID == "" {
		return nil, errorSrv.ErrMissingUserID
	}
	var accounts []model.Account
	// lookup user by id
	if e := u.dal.GetDatabase().Where(&model.Account{UserID: userID}).Order("created_at desc").Find(&accounts).Error; e != nil {
		u.logger.For(ctx).Error("Lookup accounts", zap.String("userID", userID), zap.Error(e))
		return nil, errorSrv.ErrConnectDB
	}
	if len(accounts) == 0 {
		return nil, errorSrv.ErrAccountNotFound
	}
	// fetch accounts belong to the user
	rsp := make([]*pb.Account, len(accounts))
	for i, acc := range accounts {
		rsp[i] = acc.Transform2GRPC()
	}
	return rsp, nil
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
			u.logger.For(ctx).Error("Validate account", zap.Error(err))
			return err
		}
		if err := tx.Create(acc).Error; err != nil {
			u.logger.For(ctx).Error("Create account", zap.Any("data", acc), zap.Error(err))
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
		u.logger.For(ctx).Error("Set header X-Http-Code", zap.Error(err))
	}
	return rsp, nil
}

func (u *accountServiceImpl) Delete(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	if req.GetId() == "" {
		return nil, errorSrv.ErrMissingAccountID
	}
	err := u.dal.GetDatabase().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.Account{ID: req.GetId()}).Error; err != nil {
			u.logger.For(ctx).Error("Delete account", zap.Error(err))
			return errorSrv.ErrConnectDB
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.DeleteAccountResponse{
		Id: req.GetId(),
	}, nil
}

// nolint:goconst
func (u *accountServiceImpl) Update(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	if req.GetAccount().GetId() == "" {
		return nil, errorSrv.ErrMissingAccountID
	}
	account := &model.Account{
		UserID: req.GetAccount().GetUserId(),
		ID:     req.GetAccount().GetId(),
	}
	rsp := &pb.UpdateAccountResponse{}
	err := u.dal.GetDatabase().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// find user by id
		if e := u.getAccountByID(tx.Statement.Context, account); e != nil {
			return e
		}
		u.logger.For(ctx).Info("mask", zap.Strings("path", req.GetUpdateMask().GetPaths()))
		// If there is no update mask do a regular update
		if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
			account.UpdateFromGRPC(req.GetAccount())
		} else {
			for _, path := range req.GetUpdateMask().GetPaths() {
				switch path {
				case "id":
					return errorSrv.ErrUpdateAccountID
				case "user_id":
					return errorSrv.ErrUpdateAccountUserID
				case "bank":
					return errorSrv.ErrUpdateAccountBank
				case "name":
					account.Name = req.GetAccount().GetName()
				case "balance":
					account.Balance = req.GetAccount().GetBalance()
				default:
					return errors.BadRequest("invalid field specified", map[string]string{
						"update_mask": fmt.Sprintf("account does not have field %q", path),
					})
				}
			}
		}
		if err := account.Validate(); err != nil {
			u.logger.For(ctx).Error("Validate account", zap.Error(err))
			return err
		}
		// response
		rsp.Account = account.Transform2GRPC()
		rsp.Account.UpdatedAt = timestamppb.New(account.UpdatedAt)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rsp, err
}

func (u *accountServiceImpl) List(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	accounts, err := u.getAccounts(ctx, req)
	if err != nil {
		return nil, err
	}
	rsp := &pb.ListAccountsResponse{
		Accounts: accounts,
	}
	return rsp, nil
}

func (u *accountServiceImpl) ListStream(req *pb.ListAccountsRequest, streamSrv pb.AccountService_ListStreamServer) error {
	accounts, err := u.getAccounts(streamSrv.Context(), req)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if err := streamSrv.Send(account); err != nil {
			return err
		}
	}
	return nil
}
